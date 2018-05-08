//
// Copyright (c) 2018 Red Hat, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//    http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//

package broker

import (
	"crypto/md5"
	"encoding/hex"
	"errors"
	"fmt"
	"reflect"
	"regexp"
	"strings"

	"github.com/automationbroker/bundle-lib/bundle"
	"github.com/automationbroker/bundle-lib/registries"
	"github.com/automationbroker/config"
	"github.com/openshift/ansible-service-broker/pkg/dao"
	"github.com/openshift/ansible-service-broker/pkg/metrics"
	"github.com/pborman/uuid"
	log "github.com/sirupsen/logrus"
)

var (
	// ErrorAlreadyProvisioned - Error for when an service instance has already been provisioned
	ErrorAlreadyProvisioned = errors.New("already provisioned")
	// ErrorDuplicate - Error for when a duplicate service instance already exists
	ErrorDuplicate = errors.New("duplicate instance")
	// ErrorNotFound  - Error for when a service instance is not found. (either etcd or kubernetes)
	ErrorNotFound = errors.New("not found")
	// ErrorBindingExists - Error for when deprovision is called on a service
	// instance with active bindings, or bind requested for already-existing
	// binding
	ErrorBindingExists = errors.New("binding exists")
	// ErrorProvisionInProgress - Error for when provision is called on a service instance that has a provision job in progress
	ErrorProvisionInProgress = errors.New("provision in progress")
	// ErrorDeprovisionInProgress - Error for when deprovision is called on a service instance that has a deprovision job in progress
	ErrorDeprovisionInProgress = errors.New("deprovision in progress")
	// ErrorUpdateInProgress - Error for when update is called on a service instance that has an update job in progress
	ErrorUpdateInProgress = errors.New("update in progress")
	// ErrorPlanNotFound - Error for when plan for update not found
	ErrorPlanNotFound = errors.New("plan not found")
	// ErrorParameterNotUpdatable - Error for when parameter in update request is not updatable
	ErrorParameterNotUpdatable = errors.New("parameter not updatable")
	// ErrorParameterNotFound - Error for when a parameter for update is not found
	ErrorParameterNotFound = errors.New("parameter not found")
	// ErrorParameterUnknownEnum - Error for when an unknown enum param has been requested
	ErrorParameterUnknownEnum = errors.New("unknown enum parameter value requested")
	// ErrorPlanUpdateNotPossible - Error when a Plan Update request cannot be satisfied
	ErrorPlanUpdateNotPossible = errors.New("plan update not possible")
	// ErrorNoUpdateRequested - Error for when no valid updates are requested
	ErrorNoUpdateRequested = errors.New("no valid updates requested")
	// ErrorUnbindingInProgress - Error when unbind is called that has an unbinding job in progress
	ErrorUnbindingInProgress = errors.New("unbinding in progress")
)

const (
	// fqNameRegex - regular expression used when forming FQName.
	fqNameRegex = "[/.:-]"
)

// Broker - A broker is used to to complete all the tasks that a broker must be able to do.
type Broker interface {
	Bootstrap() (*BootstrapResponse, error)
	Catalog() (*CatalogResponse, error)
	Provision(uuid.UUID, *ProvisionRequest, bool, UserInfo) (*ProvisionResponse, error)
	Update(uuid.UUID, *UpdateRequest, bool, UserInfo) (*UpdateResponse, error)
	Deprovision(bundle.ServiceInstance, string, bool, bool, UserInfo) (*DeprovisionResponse, error)
	Bind(bundle.ServiceInstance, uuid.UUID, *BindRequest, bool, UserInfo) (*BindResponse, bool, error)
	Unbind(bundle.ServiceInstance, bundle.BindInstance, string, bool, bool, UserInfo) (*UnbindResponse, bool, error)
	LastOperation(uuid.UUID, *LastOperationRequest) (*LastOperationResponse, error)
	Recover() (string, error)
	GetServiceInstance(uuid.UUID) (bundle.ServiceInstance, error)
	GetBindInstance(uuid.UUID) (bundle.BindInstance, error)
	GetBind(bundle.ServiceInstance, uuid.UUID) (*BindResponse, error)
}

// Config - Configuration for the broker.
type Config struct {
	DevBroker           bool   `yaml:"dev_broker"`
	LaunchApbOnBind     bool   `yaml:"launch_apb_on_bind"`
	BootstrapOnStartup  bool   `yaml:"bootstrap_on_startup"`
	Recovery            bool   `yaml:"recovery"`
	OutputRequest       bool   `yaml:"output_request"`
	SSLCertKey          string `yaml:"ssl_cert_key"`
	SSLCert             string `yaml:"ssl_cert"`
	RefreshInterval     string `yaml:"refresh_interval"`
	AutoEscalate        bool   `yaml:"auto_escalate"`
	ClusterURL          string `yaml:"cluster_url"`
	DashboardRedirector string `yaml:"dashboard_redirector"`
}

// DevBroker - Interface for the development broker.
type DevBroker interface {
	AddSpec(spec bundle.Spec) (*CatalogResponse, error)
	RemoveSpec(specID string) error
	RemoveSpecs() error
}

// AnsibleBroker - Broker using ansible and images to interact with oc/kubernetes/etcd
type AnsibleBroker struct {
	dao          dao.Dao
	registry     []registries.Registry
	engine       *WorkEngine
	brokerConfig Config
	namespace    string
}

// NewAnsibleBroker - Creates a new ansible broker
func NewAnsibleBroker(dao dao.Dao,
	registry []registries.Registry,
	engine WorkEngine,
	brokerConfig *config.Config,
	namespace string) (*AnsibleBroker, error) {

	broker := &AnsibleBroker{
		dao:      dao,
		registry: registry,
		engine:   &engine,
		brokerConfig: Config{
			DevBroker:           brokerConfig.GetBool("dev_broker"),
			LaunchApbOnBind:     brokerConfig.GetBool("launch_apb_on_bind"),
			BootstrapOnStartup:  brokerConfig.GetBool("bootstrap_on_startup"),
			Recovery:            brokerConfig.GetBool("recovery"),
			OutputRequest:       brokerConfig.GetBool("output_request"),
			SSLCertKey:          brokerConfig.GetString("ssl_cert_key"),
			SSLCert:             brokerConfig.GetString("ssl_cert"),
			RefreshInterval:     brokerConfig.GetString("refresh_interval"),
			AutoEscalate:        brokerConfig.GetBool("auto_escalate"),
			ClusterURL:          brokerConfig.GetString("cluster_url"),
			DashboardRedirector: brokerConfig.GetString("dashboard_redirector"),
		},
		namespace: namespace,
	}
	return broker, nil
}

// GetServiceInstance - retrieve the service instance for a instanceID.
func (a AnsibleBroker) GetServiceInstance(instanceUUID uuid.UUID) (bundle.ServiceInstance, error) {
	instance, err := a.dao.GetServiceInstance(instanceUUID.String())

	if err != nil {
		if a.dao.IsNotFoundError(err) {
			log.Infof("Could not find a service instance in dao - %v", err)
			return bundle.ServiceInstance{}, ErrorNotFound
		}
		log.Info("Couldn't find a service instance: ", err)
		return bundle.ServiceInstance{}, err
	}

	dashboardURL := a.getDashboardURL(instance)
	if dashboardURL != "" {
		instance.DashboardURL = dashboardURL
	}

	return *instance, nil

}

// GetBindInstance - retrieve the bind instance for a bindUUID
func (a AnsibleBroker) GetBindInstance(bindUUID uuid.UUID) (bundle.BindInstance, error) {
	instance, err := a.dao.GetBindInstance(bindUUID.String())
	if err != nil {
		if a.dao.IsNotFoundError(err) {
			return bundle.BindInstance{}, ErrorNotFound
		}
		return bundle.BindInstance{}, err
	}
	return *instance, nil
}

// Bootstrap - Loads all known specs from a registry into local storage for reference
// Potentially a large download; on the order of 10s of thousands
// TODO: How do we handle a large amount of data on this side as well? Pagination?
func (a AnsibleBroker) Bootstrap() (*BootstrapResponse, error) {
	log.Info("AnsibleBroker::Bootstrap")
	var err error
	var specs []*bundle.Spec
	var imageCount int

	// Remove all non apb-push sourced specs that have been saved.
	pushedSpecs := []*bundle.Spec{}
	dir := "/spec"
	specs, err = a.dao.BatchGetSpecs(dir)
	if err != nil {
		log.Errorf("Something went real bad trying to retrieve batch specs for deletion... - %v", err)
		return nil, err
	}
	// Save all apb-push sourced specs
	for _, spec := range specs {
		if strings.HasPrefix(spec.FQName, "apb-push") {
			log.Infof("Saving apb-push sourced spec to prevent deletion: %v", spec.FQName)
			pushedSpecs = append(pushedSpecs, spec)
		}
	}

	err = a.dao.BatchDeleteSpecs(specs)
	if err != nil {
		log.Errorf("Something went real bad trying to delete batch specs... - %v", err)
		return nil, err
	}
	specs = []*bundle.Spec{}
	//Metrics calls.
	metrics.SpecsLoadedReset()
	metrics.SpecsReset()
	//re-add the apb-push metrics.
	metrics.SpecsLoaded(apbPushRegName, len(pushedSpecs))

	// Load Specs for each registry
	registryErrors := []error{}
	for _, r := range a.registry {
		s, count, err := r.LoadSpecs()
		if err != nil && r.Fail(err) {
			log.Errorf("registry caused bootstrap failure - %v", err)
			return nil, err
		}
		if err != nil {
			log.Warningf("registry: %v was unable to complete bootstrap - %v",
				r.RegistryName, err)
			registryErrors = append(registryErrors, err)
		}
		imageCount += count
		// this will also update the plan id
		addNameAndIDForSpec(s, r.RegistryName())
		metrics.SpecsLoaded(r.RegistryName(), len(s))
		specs = append(specs, s...)
	}
	// Add apb-push sourced specs back to the list
	for _, spec := range pushedSpecs {
		specs = append(specs, spec)
	}
	if len(registryErrors) == len(a.registry) {
		return nil, errors.New("all registries failed on bootstrap")
	}
	specManifest := map[string]*bundle.Spec{}
	planNameManifest := map[string]string{}

	for _, s := range specs {
		specManifest[s.ID] = s

		// each of the plans from all of the specs gets its own uuid. even
		// though the names may be the same we want them to be globally unique.
		for _, p := range s.Plans {
			if p.ID == "" {
				log.Errorf("We have a plan that did not get its id generated: %v", p.Name)
				continue
			}
			planNameManifest[p.ID] = p.Name
		}
	}
	if err := a.dao.BatchSetSpecs(specManifest); err != nil {
		return nil, err
	}

	bundle.AddSecrets(specs)

	return &BootstrapResponse{SpecCount: len(specs), ImageCount: imageCount}, nil
}

// addNameAndIDForSpec - will create the unique spec name and id
// and set it for each spec
func addNameAndIDForSpec(specs []*bundle.Spec, registryName string) {
	for _, spec := range specs {
		// need to make / a hyphen to allow for global uniqueness
		// but still match spec.

		re := regexp.MustCompile(fqNameRegex)
		spec.FQName = re.ReplaceAllLiteralString(
			fmt.Sprintf("%v-%v", registryName, spec.FQName),
			"-")
		spec.FQName = fmt.Sprintf("%.51v", spec.FQName)
		if strings.HasSuffix(spec.FQName, "-") {
			spec.FQName = spec.FQName[:len(spec.FQName)-1]
		}

		// ID Will be a md5 hash of the fully qualified spec name.
		hasher := md5.New()
		hasher.Write([]byte(spec.FQName))
		spec.ID = hex.EncodeToString(hasher.Sum(nil))

		// update the id on the plans, doing it here avoids looping through the
		// specs array again
		addIDForPlan(spec.Plans, spec.FQName)
	}
}

// addIDForPlan - for each of the plans create a new ID
func addIDForPlan(plans []bundle.Plan, FQSpecName string) {

	// need to use the index into the array to actually update the struct.
	for i, plan := range plans {
		//plans[i].ID = uuid.New()
		FQPlanName := fmt.Sprintf("%s-%s", FQSpecName, plan.Name)
		hasher := md5.New()
		hasher.Write([]byte(FQPlanName))
		plans[i].ID = hex.EncodeToString(hasher.Sum(nil))
	}
}

// Recover - Will recover the broker.
func (a AnsibleBroker) Recover() (string, error) {
	var emptyToken string
	// do we have any jobs that are still running?
	// get all in progress jobs
	// For each job, check the status of each of their containers to update
	// their status in case any of them finished.

	recoverStatuses, err := a.dao.FindJobStateByState(bundle.StateInProgress)
	if err != nil {
		// no jobs or states to recover, this is OK.
		if a.dao.IsNotFoundError(err) {
			log.Info("No jobs to recover")
			return "", nil
		}
		return emptyToken, err
	}

	/*
		if a job was in progress, we know the instanceuuid & token.
		do we have a podname?

		if no, the job never started
			we should restart the job
		if yes,
			did the job finish?
				yes
					* update status to finished
					* extractCreds if available
				no
					* create a monitoring job to update status
	*/

	// let's see if we need to recover any of these
	for _, rs := range recoverStatuses {

		// We have an in progress job
		instanceID := rs.InstanceID.String()
		instance, err := a.dao.GetServiceInstance(instanceID)
		if err != nil {
			return emptyToken, err
		}

		// Do we have a podname?
		if rs.State.Podname == "" {
			// NO, we do not have a podname

			log.Infof("No podname. Attempting to restart job: %s", instanceID)
			log.Debugf("%v", instance)

			// Handle bad write of service instance
			if instance.Spec == nil || instance.Parameters == nil {
				a.dao.SetState(instanceID, bundle.JobState{
					Token:  rs.State.Token,
					State:  bundle.StateFailed,
					Method: rs.State.Method,
				})
				a.dao.DeleteServiceInstance(instance.ID.String())
				log.Warningf("incomplete ServiceInstance [%s] record, marking job as failed",
					instance.ID)

				// skip to the next item
				continue
			}

			var job Work
			var topic WorkTopic
			if rs.State.Method == bundle.JobMethodProvision {
				job = &ProvisionJob{instance}
				topic = ProvisionTopic
			} else if rs.State.Method == bundle.JobMethodUpdate {
				job = &UpdateJob{instance}
				topic = UpdateTopic
			} else if rs.State.Method == bundle.JobMethodDeprovision {
				job = &DeprovisionJob{instance, false}
				topic = DeprovisionTopic
			} else {
				log.Warningf(
					"Attempted to recover job %s, but found an unrecognized "+
						"MethodType: %s, skipping...",
					rs.State.Token, rs.State.Method,
				)
			}

			// Need to use the same token as before, since that's what the
			// catalog will try to ping.
			_, err := a.engine.StartNewAsyncJob(rs.State.Token, job, topic)
			if err != nil {
				return emptyToken, err
			}

		} else {
			// YES, we have a podname
			log.Infof("We have a pod to recover: %s", rs.State.Podname)

			// did the pod finish?
			extCreds, extErr := bundle.ExtractCredentials(
				rs.State.Podname,
				instance.Context.Namespace,
				instance.Spec.Runtime,
			)

			// NO, pod failed.
			if extErr != nil {
				log.Errorf("broker::Recover error occurred. %s", extErr.Error())
				return emptyToken, extErr
			}

			// YES, pod finished we have creds
			if extCreds != nil {
				log.Debug("broker::Recover, got ExtractedCredentials!")
				a.dao.SetState(instanceID, bundle.JobState{
					Token:   rs.State.Token,
					State:   bundle.StateSucceeded,
					Podname: rs.State.Podname,
					Method:  rs.State.Method,
				})
				err = bundle.SetExtractedCredentials(instanceID, extCreds)
				if err != nil {
					log.Errorf("Could not persist extracted credentials - %s", err.Error())
					return emptyToken, err
				}
			}
		}
	}

	log.Info("Recovery complete")
	return "recover called", nil
}

// Catalog - returns the catalog of services defined
func (a AnsibleBroker) Catalog() (*CatalogResponse, error) {
	log.Info("AnsibleBroker::Catalog")

	var specs []*bundle.Spec
	var err error
	var services []Service
	dir := "/spec"

	if specs, err = a.dao.BatchGetSpecs(dir); err != nil {
		log.Error("Something went real bad trying to retrieve batch specs...")
		return nil, err
	}

	log.Debugf("Filtering secret parameters out of specs...")
	specs, err = bundle.FilterSecrets(specs)
	if err != nil {
		// Should we blow up or warn and continue?
		log.Errorf("Something went real bad trying to load secrets %v", err)
		return nil, err
	}

	services = []Service{}
	for _, spec := range specs {
		ser, err := SpecToService(spec)
		if err != nil {
			log.Errorf("not adding spec %v to list of services due to error transforming to service - %v", spec.FQName, err)
		} else {
			// Bug 1539542 - in order for async bind to work,
			// bindings_retrievable needs to be set to true. We only want to
			// set BindingsRetrievable to true if the service is bindable
			// AND we the broker is configured to launch apbs on bind
			if ser.Bindable && a.brokerConfig.LaunchApbOnBind {
				ser.BindingsRetrievable = true
			}

			services = append(services, ser)
		}
	}

	return &CatalogResponse{services}, nil
}

// Provision  - will provision a service
func (a AnsibleBroker) Provision(instanceUUID uuid.UUID, req *ProvisionRequest, async bool, userInfo UserInfo,
) (*ProvisionResponse, error) {
	////////////////////////////////////////////////////////////
	//type ProvisionRequest struct {

	//-> OrganizationID    uuid.UUID
	//-> SpaceID           uuid.UUID
	// Used for determining where this service should be provisioned. Analogous to
	// OCP's namespaces and projects. Re: OrganizationID, spec mentions
	// "Most brokers will not use this field, it could be helpful in determining
	// the data placement or applying custom business rules"

	//-> PlanID            uuid.UUID
	//-> ServiceID         uuid.UUID
	// ServiceID maps directly to a Spec.Id found in etcd. Can pull Spec via
	// Dao::GetSpec(id string)

	//-> Parameters        map[string]string
	// User provided configuration answers for the AnsibleApp

	// -> AcceptsIncomplete bool
	// true indicates both the SC and the requesting client (sc client). If param
	// is not included in the req, and the broker can only provision an instance of
	// the request plan asynchronously, broker should reject with a 422
	// NOTE: Spec.Async should indicate what level of async support is available for
	// a given ansible app

	//}

	// Summary:
	// For our purposes right now, the ServiceID and the Params should be enough to
	// Provision an ansible app.
	////////////////////////////////////////////////////////////
	// Provision Flow
	// -> Retrieve Spec from etcd (if missing, 400, this returns err missing)
	// -> Make entry in /instance, ID'd by instance. Value should be Instance type
	//    Purpose is to make sure everything need to deprovision is available
	//    in persistence.
	// -> Provision!
	////////////////////////////////////////////////////////////

	/*
		dao GET returns error strings like CODE: message (entity) [#]
		dao SetServiceInstance returns what error?
		dao.SetState returns what error?
		Provision returns what error?
		SetExtractedCredentials returns what error?

		broker
		* normal synchronous return ProvisionResponse
		* normal async return ProvisionResponse
		* if instance already exists with the same params, return ProvisionResponse, AND InstanceExists
		* if instance already exists DIFFERENT param, return nil AND InstanceExists

		handler returns the following
		* synchronous provision return 201 created
		* instance already exists with IDENTICAL parameters to existing instance, 200 OK
		* async provision 202 Accepted
		* instance already exists with DIFFERENT parameters, 409 Conflict {}
		* if only support async and no accepts_incomplete=true passed in, 422 Unprocessable entity

	*/
	var spec *bundle.Spec
	var err error

	// Retrieve requested spec
	specID := req.ServiceID
	if spec, err = a.dao.GetSpec(specID); err != nil {
		// etcd return not found i.e. code 100
		if a.dao.IsNotFoundError(err) {
			return nil, ErrorNotFound
		}
		// otherwise unknown error bubble it up
		return nil, err
	}

	context := &req.Context
	parameters := req.Parameters
	if parameters == nil {
		parameters = make(bundle.Parameters)
	}

	if req.PlanID == "" {
		errMsg :=
			"PlanID from provision request is blank. " +
				"Provision requests must specify PlanIDs"
		return nil, errors.New(errMsg)
	}

	plan, ok := spec.GetPlanFromID(req.PlanID)
	if !ok {
		return nil, ErrorNotFound
	}

	log.Debugf(
		"Injecting PlanID as parameter: { %s: %s }",
		planParameterKey, plan.Name)
	parameters[planParameterKey] = plan.Name
	log.Debugf("Injecting ServiceClassID as parameter: { %s: %s }",
		serviceClassIDKey, req.ServiceID)
	parameters[serviceClassIDKey] = req.ServiceID
	log.Debugf("Injecting ServiceInstanceID as parameter: { %s: %s }",
		serviceInstIDKey, instanceUUID.String())
	parameters[serviceInstIDKey] = instanceUUID.String()
	log.Debugf("Injecting lastRequestingUserKey as parameter: { %s: %s }",
		lastRequestingUserKey, getLastRequestingUser(userInfo))
	parameters[lastRequestingUserKey] = getLastRequestingUser(userInfo)

	// Build and persist record of service instance
	serviceInstance := &bundle.ServiceInstance{
		ID:         instanceUUID,
		Spec:       spec,
		Context:    context,
		Parameters: &parameters,
	}

	// Verify we're not re-provisioning the same instance
	// if err is nil, there is an instance. Let's compare it to the instance
	// we're being asked to provision.
	//
	// if err is not nil, we will just bubble that up

	si, err := a.dao.GetServiceInstance(instanceUUID.String())
	if err != nil && !a.dao.IsNotFoundError(err) {
		return nil, err
	}
	// This will use the package to make sure that if the type is changed
	// away from []byte it can still be evaluated.
	if si != nil && uuid.Equal(si.ID, serviceInstance.ID) {
		if reflect.DeepEqual(si.Parameters, serviceInstance.Parameters) {
			alreadyInProgress, jobToken, err := a.isJobInProgress(serviceInstance.ID.String(), bundle.JobMethodProvision)
			if err != nil {
				return nil, fmt.Errorf("An error occurred while trying to determine if a provision job is already in progress for instance: %s", serviceInstance.ID)
			}
			if alreadyInProgress {
				log.Infof("Provision requested for instance %s, but job is already in progress", serviceInstance.ID)
				return &ProvisionResponse{Operation: jobToken}, ErrorProvisionInProgress
			}
			log.Debug("already have this instance returning 200")
			return &ProvisionResponse{}, ErrorAlreadyProvisioned
		}
		log.Info("we have a duplicate instance with parameters that differ, returning 409 conflict")
		return nil, ErrorDuplicate
	}

	//
	// Looks like this is a new provision, let's get started.
	//
	if err = a.dao.SetServiceInstance(instanceUUID.String(), serviceInstance); err != nil {
		return nil, err
	}

	var token = a.engine.Token()
	pjob := &ProvisionJob{serviceInstance}
	metrics.ActionStarted("provision")

	if async {
		log.Info("ASYNC provisioning in progress")
		// asynchronously provision and return the token for the lastoperation
		token, err = a.engine.StartNewAsyncJob(token, pjob, ProvisionTopic)
		if err != nil {
			log.Errorf("Failed to start new job for async provision\n%s", err.Error())
			return nil, err
		}
	} else {
		log.Info("reverting to synchronous provisioning in progress")
		if err := a.engine.StartNewSyncJob(token, pjob, ProvisionTopic); err != nil {
			log.Errorf("Failed to start new job for sync provision\n%s", err.Error())
			return nil, err
		}
	}

	var response *ProvisionResponse
	dashboardURL := a.getDashboardURL(serviceInstance)
	if dashboardURL != "" {
		response = &ProvisionResponse{Operation: token, DashboardURL: dashboardURL}
	} else {
		response = &ProvisionResponse{Operation: token}
	}

	return response, nil
}

// getDashboardURL - will conditionally return a dashboard redirector url or
// an empty string if the redirector feature is not specified by the APB.
func (a *AnsibleBroker) getDashboardURL(si *bundle.ServiceInstance) string {
	var val interface{}
	var drEnabled, ok bool
	spec := si.Spec

	if len(spec.Alpha) == 0 {
		return ""
	}

	val, ok = spec.Alpha["dashboard_redirect"]
	if !ok {
		return ""
	}

	drEnabled, ok = val.(bool)
	if !ok {
		return ""
	}

	if !drEnabled {
		return ""
	}

	if a.brokerConfig.DashboardRedirector == "" {
		log.Warningf("Attempting to provision %v, which has dashboard redirect enabled, "+
			"but no dashboard_redirector route was found in the broker's configmap. "+
			"Deploying without a dashboard_url.", spec.FQName)
		return ""
	}

	drURL := a.brokerConfig.DashboardRedirector
	if !strings.HasPrefix(drURL, "http") {
		drURL = fmt.Sprintf("http://%s", drURL)
	}

	return fmt.Sprintf("%s/?id=%s", drURL, si.ID)
}

// Deprovision - will deprovision a service.
func (a AnsibleBroker) Deprovision(
	instance bundle.ServiceInstance, planID string, skipApbExecution bool, async bool, userInfo UserInfo,
) (*DeprovisionResponse, error) {
	////////////////////////////////////////////////////////////
	// Deprovision flow
	// -> Lookup bindings by instance ID; 400 if any are active, related issue:
	//    https://github.com/openservicebrokerapi/servicebroker/issues/127
	// -> Atomic deprovision and removal of service entry in etcd?
	//    * broker::Deprovision
	//    Arguments for this? What data do apbs require to deprovision?
	//    * namespace
	//    Maybe just hand off a serialized ServiceInstance and let the apb
	//    decide what's important?
	//    * delete credentials from etcd
	//    * if noerror: delete serviceInstance entry with Dao
	if planID == "" {
		errMsg := "Deprovision request contains an empty plan_id"
		return nil, errors.New(errMsg)
	}
	err := a.validateDeprovision(&instance)
	if err != nil {
		return nil, err
	}

	alreadyInProgress, jobToken, err := a.isJobInProgress(instance.ID.String(), bundle.JobMethodDeprovision)
	if err != nil {
		return nil, fmt.Errorf("An error occurred while trying to determine if a deprovision job is already in progress for instance: %s", instance.ID)
	}

	if alreadyInProgress {
		log.Infof("Deprovision requested for instance %s, but job is already in progress", instance.ID)
		return &DeprovisionResponse{Operation: jobToken}, ErrorDeprovisionInProgress
	}

	provExtCreds, err := bundle.GetExtractedCredentials(instance.ID.String())
	if err != nil && err != bundle.ErrExtractedCredentialsNotFound {
		log.Warningf("unable to retrieve provision time credentials - %v", err)
		return nil, err
	}

	// Add the DB Credentials to the parameters. This will allow the apb to use these credentials
	// if it so chooses.
	if provExtCreds != nil && instance.Parameters != nil {
		params := *instance.Parameters
		params[bundle.ProvisionCredentialsKey] = provExtCreds.Credentials
		instance.Parameters = &params
	}

	// Override the lastRequestingUserKey value in the instance.Parameters
	if instance.Parameters != nil {
		(*instance.Parameters)[lastRequestingUserKey] = getLastRequestingUser(userInfo)
		instance.Parameters.EnsureDefaults()
	}

	var token = a.engine.Token()
	dpjob := &DeprovisionJob{&instance, skipApbExecution}
	metrics.ActionStarted("deprovision")
	if async {
		log.Info("ASYNC deprovision in progress")

		token, err = a.engine.StartNewAsyncJob(token, dpjob, DeprovisionTopic)
		if err != nil {
			log.Errorf("Failed to start new job for async deprovision\n%s", err.Error())
			return nil, err
		}
		return &DeprovisionResponse{Operation: token}, nil
	}

	if !skipApbExecution {
		log.Info("Synchronous deprovision in progress")
		if err := a.engine.StartNewSyncJob(token, dpjob, DeprovisionTopic); err != nil {
			return nil, err
		}
	}
	return &DeprovisionResponse{}, nil
}

func (a AnsibleBroker) validateDeprovision(instance *bundle.ServiceInstance) error {
	// -> Lookup bindings by instance ID; 400 if any are active, related issue:
	//    https://github.com/openservicebrokerapi/servicebroker/issues/127
	if len(instance.BindingIDs) > 0 {
		log.Debugf("Found bindings with ids: %v", instance.BindingIDs)
		return ErrorBindingExists
	}

	return nil
}

func (a AnsibleBroker) isJobInProgress(ID string,
	method bundle.JobMethod) (bool, string, error) {

	allJobs, err := a.dao.GetSvcInstJobsByState(ID, bundle.StateInProgress)
	log.Infof("All Jobs for instance: %v in state:  %v - \n%#v", ID, bundle.StateInProgress, allJobs)
	if err != nil {
		return false, "", err
	}

	var token string
	methodJobs := dao.MapJobStatesWithMethod(allJobs, method)
	if len(methodJobs) > 0 {
		token = methodJobs[0].Token
	}
	return len(methodJobs) > 0, token, nil
}

// GetBind - will return the binding between a service created via an async
// binding event.
func (a AnsibleBroker) GetBind(instance bundle.ServiceInstance, bindingUUID uuid.UUID) (*BindResponse, error) {

	log.Debug("broker.GetBind: entered GetBind")

	provExtCreds, err := bundle.GetExtractedCredentials(instance.ID.String())
	if err != nil && err != bundle.ErrExtractedCredentialsNotFound {
		log.Warningf("unable to retrieve provision time credentials - %v", err)
		return nil, err
	}

	bi, err := a.dao.GetBindInstance(bindingUUID.String())
	if err != nil {
		if a.dao.IsNotFoundError(err) {
			log.Warningf("id: %v - could not find bind instance - %v", bindingUUID, err)
			return nil, ErrorNotFound
		}
		log.Warningf("id: %v - unable to retrieve bind instance - %v", bindingUUID, err)
		return nil, err
	}

	bindExtCreds, err := bundle.GetExtractedCredentials(bi.ID.String())
	if err != nil {
		if err == bundle.ErrExtractedCredentialsNotFound {
			return nil, ErrorNotFound
		}

		return nil, err
	}

	log.Debug("broker.GetBind: we got the bind credentials")
	return NewBindResponse(provExtCreds, bindExtCreds)
}

// Bind - will create a binding between a service. Parameter "async" declares
// whether the caller is willing to have the operation run asynchronously. The
// returned bool will be true if the operation actually ran asynchronously.
func (a AnsibleBroker) Bind(instance bundle.ServiceInstance, bindingUUID uuid.UUID, req *BindRequest, async bool, userInfo UserInfo,
) (*BindResponse, bool, error) {
	// binding_id is the id of the binding.
	// the instanceUUID is the previously provisioned service id.
	//
	// See if the service instance still exists, if not send back a badrequest.

	// GET SERVICE get provision parameters
	params := req.Parameters
	if params == nil {
		params = make(bundle.Parameters)
	}

	// Inject PlanID into parameters passed to APBs
	if req.PlanID == "" {
		errMsg :=
			"PlanID from bind request is blank. " +
				"Bind requests must specify PlanIDs"
		return nil, false, errors.New(errMsg)
	}
	plan, ok := instance.Spec.GetPlanFromID(req.PlanID)
	if !ok {
		log.Debug("Plan not found")
		return nil, false, ErrorNotFound
	}

	log.Debugf(
		"Injecting PlanID as parameter: { %s: %s }",
		planParameterKey, plan.Name)

	params[planParameterKey] = plan.Name

	log.Debugf("Injecting ServiceClassID as parameter: { %s: %s }",
		serviceClassIDKey, req.ServiceID)
	params[serviceClassIDKey] = req.ServiceID

	log.Debugf("Injecting ServiceInstanceID as parameter: { %s: %s }",
		serviceInstIDKey, instance.ID.String())
	params[serviceInstIDKey] = instance.ID.String()

	log.Debugf("Injecting lastRequestingUserKey as parameter: { %s: %s }",
		lastRequestingUserKey, getLastRequestingUser(userInfo))
	params[lastRequestingUserKey] = getLastRequestingUser(userInfo)

	log.Debugf("Injecting ServiceBindingID as parameter: { %s: %s }",
		serviceBindingIDKey, bindingUUID.String())
	params[serviceBindingIDKey] = bindingUUID.String()

	// Create a BindingInstance with a reference to the serviceinstance.
	bindingInstance := &bundle.BindInstance{
		ID:         bindingUUID,
		ServiceID:  instance.ID,
		Parameters: &params,
	}

	// Verify we're not rebinding the same instance. if err is nil, there is an
	// instance. Let's compare it to the instance we're being asked to bind.
	//
	// if err is not nil, we will just bubble that up
	//
	// if binding instance exists, and the parameters are the same return: 200.
	// if binding instance exists, and the parameters are different return: 409.
	//
	// return 201 when we're done.
	provExtCreds, err := bundle.GetExtractedCredentials(instance.ID.String())
	if err != nil && err != bundle.ErrExtractedCredentialsNotFound {
		log.Warningf("unable to retrieve provision time credentials - %v", err)
		return nil, false, err
	}

	if existingBI, err := a.dao.GetBindInstance(bindingUUID.String()); err == nil {
		if existingBI.IsEqual(bindingInstance) {
			bindExtCreds, err := bundle.GetExtractedCredentials(existingBI.ID.String())
			// It's ok if there aren't any bind credentials yet.
			if err != nil && err != bundle.ErrExtractedCredentialsNotFound {
				return nil, false, err
			}
			var createJob bundle.JobState
			if existingBI.CreateJobKey != "" {
				createJob, err = a.dao.GetStateByKey(existingBI.CreateJobKey)
			}

			switch {
			// unknown error
			case err != nil && !a.dao.IsNotFoundError(err):
				return nil, false, err
			// If there is a job in "succeeded" state, or no job at all, or
			// the referenced job no longer exists (we assume it got
			// cleaned up eventually), assume everything is complete.
			case createJob.State == bundle.StateSucceeded, existingBI.CreateJobKey == "", a.dao.IsNotFoundError(err):
				log.Debug("already have this binding instance, returning 200")
				resp, err := NewBindResponse(provExtCreds, bindExtCreds)
				if err != nil {
					return nil, false, err
				}
				return resp, false, ErrorBindingExists
			// If there is a job in any other state, send client through async flow.
			case len(createJob.State) > 0:
				return &BindResponse{Operation: createJob.Token}, true, nil
			// This should not happen unless there is bad data in the data store.
			default:
				err = errors.New("found a JobState with no value for field State")
				log.Error(err.Error())
				return nil, false, err
			}
		}

		// parameters are different
		log.Info("duplicate binding instance diff params, returning 409 conflict")
		return nil, false, ErrorDuplicate
	} else if !a.dao.IsNotFoundError(err) {
		return nil, false, err
	}

	// No existing BindInstance was found above, so proceed with saving this one
	if err := a.dao.SetBindInstance(bindingUUID.String(), bindingInstance); err != nil {
		return nil, false, err
	}

	// Add the DB Credentials. This will allow the apb to use these credentials
	// if it so chooses.
	if provExtCreds != nil {
		params[bundle.ProvisionCredentialsKey] = provExtCreds.Credentials
	}

	// NOTE: We are currently disabling running an APB on bind via
	// 'LaunchApbOnBind' of the broker config, due to lack of async support of
	// bind in Open Service Broker API Currently, the 'launchapbonbind' is set
	// to false in the 'config' ConfigMap

	metrics.ActionStarted("bind")
	var (
		bindExtCreds *bundle.ExtractedCredentials
		token        = a.engine.Token()
		bindingJob   = &BindJob{&instance, bindingUUID.String(), &params}
	)

	if async && a.brokerConfig.LaunchApbOnBind {
		// asynchronous mode, requires that the launch apb config
		// entry is on, and that async comes in from the catalog
		log.Info("ASYNC binding in progress")
		token, err = a.engine.StartNewAsyncJob("", bindingJob, BindingTopic)
		if err != nil {
			log.Errorf("Failed to start new job for async binding\n%s", err.Error())
			return nil, false, err
		}

		bindingInstance.CreateJobKey = fmt.Sprintf("/state/%s/job/%s", bindingUUID.String(), token)
		if err := a.dao.SetBindInstance(bindingUUID.String(), bindingInstance); err != nil {
			return nil, false, err
		}
		return &BindResponse{Operation: token}, true, nil
	} else if a.brokerConfig.LaunchApbOnBind {
		// we are synchronous mode
		log.Info("Broker configured to run APB bind")
		if err := a.engine.StartNewSyncJob(token, bindingJob, BindingTopic); err != nil {
			return nil, false, err
		}
		//TODO are we only setting the bindingUUID if sync?
		instance.AddBinding(bindingUUID)
		if err := a.dao.SetServiceInstance(instance.ID.String(), &instance); err != nil {
			return nil, false, err
		}
	} else {
		log.Warning("Broker configured to *NOT* launch and run APB bind")
		// Create a credentials for the binding using the provision credentials
		bindExtCreds = provExtCreds
		err := bundle.SetExtractedCredentials(bindingUUID.String(), bindExtCreds)
		if err != nil {
			log.Errorf("Unable to create new binding extracted creds from provision creds - %v", err)
			return nil, false, err
		}
		instance.AddBinding(bindingUUID)
		if err := a.dao.SetServiceInstance(instance.ID.String(), &instance); err != nil {
			return nil, false, err
		}
	}

	resp, err := NewBindResponse(provExtCreds, bindExtCreds)
	// If we made it this far, the operation completed synchronously.
	return resp, false, err
}

// Unbind - unbind a service's previous binding. Parameter "async" declares
// whether the caller is willing to have the operation run asynchronously. The
// returned bool will be true if the operation actually ran asynchronously.
func (a AnsibleBroker) Unbind(
	instance bundle.ServiceInstance, bindInstance bundle.BindInstance, planID string, skipApbExecution bool, async bool, userInfo UserInfo,
) (*UnbindResponse, bool, error) {
	if planID == "" {
		errMsg :=
			"PlanID from unbind request is blank. " +
				"Unbind requests must specify PlanIDs"
		return nil, false, errors.New(errMsg)
	}

	jobInProgress, jobToken, err := a.isJobInProgress(bindInstance.ID.String(), bundle.JobMethodUnbind)
	if err != nil {
		log.Errorf("An error occurred while trying to determine if a unbind job is already in progress for instance: %s", instance.ID)
		return nil, false, err
	}
	if jobInProgress {
		log.Infof("Unbind requested for instance %s, but job is already in progress", instance.ID)
		return &UnbindResponse{Operation: jobToken}, false, ErrorUnbindingInProgress
	}

	// Override the lastRequestingUserKey value in the instance.Parameters
	if instance.Parameters != nil {
		(*instance.Parameters)[lastRequestingUserKey] = getLastRequestingUser(userInfo)
	}

	provExtCreds, err := bundle.GetExtractedCredentials(instance.ID.String())
	if err != nil && err != bundle.ErrExtractedCredentialsNotFound {
		return nil, false, err
	}
	bindExtCreds, err := bundle.GetExtractedCredentials(bindInstance.ID.String())
	if err != nil && err != bundle.ErrExtractedCredentialsNotFound {
		return nil, false, err
	}
	// Add the credentials to the parameters so that an APB can choose what
	// it would like to do.
	if provExtCreds == nil && bindExtCreds == nil {
		log.Warningf("Unable to find credentials for instance id: %v and binding id: %v"+
			" something may have gone wrong. Proceeding with unbind.",
			instance.ID, bindInstance.ID)
	}
	serviceInstance, err := a.GetServiceInstance(instance.ID)
	if err != nil {
		log.Debugf("Service instance with id %s does not exist", instance.ID.String())
		return nil, false, err
	}

	// build up unbind parameters
	params := make(bundle.Parameters)
	// Fixes BZ1578319 - put last requesting user at the top level
	// they should be at the top level. We are still keeping the lower level
	// values as well since others might already be using them.
	params[lastRequestingUserKey] = getLastRequestingUser(userInfo)
	params[planParameterKey] = planID
	params[serviceInstIDKey] = serviceInstance.ID.String()
	params[serviceBindingIDKey] = bindInstance.ID.String()

	// TODO: feels like we should be passing in service instance id
	// and binding uuid as well

	if provExtCreds != nil {
		params[bundle.ProvisionCredentialsKey] = provExtCreds.Credentials
	}
	if bindExtCreds != nil {
		params[bundle.BindCredentialsKey] = bindExtCreds.Credentials
	}
	if serviceInstance.Parameters != nil {
		params["provision_params"] = *serviceInstance.Parameters
	}
	metrics.ActionStarted("unbind")

	var (
		token     = a.engine.Token()
		jerr      error
		unbindJob = &UnbindJob{
			&serviceInstance, bindInstance.ID.String(), &params, skipApbExecution}
	)
	if async && a.brokerConfig.LaunchApbOnBind {
		// asynchronous mode, required that the launch apb config
		// entry is on, and that async comes in from the catalog
		log.Info("ASYNC unbinding in progress")

		token, jerr = a.engine.StartNewAsyncJob("", unbindJob, UnbindingTopic)
		if jerr != nil {
			log.Errorf("Failed to start new job for async unbind\n%s", jerr.Error())
			return nil, false, jerr
		}

		return &UnbindResponse{Operation: token}, true, nil

	} else if a.brokerConfig.LaunchApbOnBind {
		// only launch apb if we are always launching the APB.
		if skipApbExecution {
			log.Debug("Skipping unbind apb execution")
			err = nil
		} else {
			log.Debug("Launching apb for unbind in blocking mode")
			if err := a.engine.StartNewSyncJob(token, unbindJob, UnbindingTopic); err != nil {
				return nil, false, err
			}
		}
	} else {
		log.Warning("Broker configured to *NOT* launch and run APB unbind")
		if err := a.dao.DeleteBinding(bindInstance, serviceInstance); err != nil {
			log.Errorf("Failed to delete binding when launch_apb_on_bind is false: %v", err)
			return nil, false, err
		}
		if err := bundle.DeleteExtractedCredentials(bindInstance.ID.String()); err != nil {
			log.Errorf("Failed to delete extracted credentials secret when launch_apb_on_bind is false: %v", err)
			return nil, false, err
		}
	}
	return &UnbindResponse{}, false, nil
}

// Update  - will update a service
func (a AnsibleBroker) Update(instanceUUID uuid.UUID, req *UpdateRequest, async bool, userInfo UserInfo,
) (*UpdateResponse, error) {
	////////////////////////////////////////////////////////////
	//type UpdateRequest struct {

	//-> PreviousValues
	//  -> OrganizationID    uuid.UUID
	//  -> SpaceID           uuid.UUID
	//   Used for determining where this service should be provisioned. Analogous to
	//   OCP's namespaces and projects. Re: OrganizationID, spec mentions
	//   "Most brokers will not use this field, it could be helpful in determining
	//   the data placement or applying custom business rules"
	//   -> PlanID            uuid.UUID
	//   -> ServiceID         uuid.UUID
	// ServiceID maps directly to a Spec.Id found in etcd. Can pull Spec via
	// Dao::GetSpec(id string)

	//-> Parameters        map[string]string
	// User provided configuration answers for the AnsibleApp

	// -> AcceptsIncomplete bool
	// true indicates both the SC and the requesting client (sc client). If param
	// is not included in the req, and the broker can only provision an instance of
	// the request plan asynchronously, broker should reject with a 422
	// NOTE: Spec.Async should indicate what level of async support is available for
	// a given ansible app

	//}

	// Summary:
	// For our purposes right now, the ServiceID and the Params should be enough to
	// Update an ansible app.
	////////////////////////////////////////////////////////////
	// Update Flow
	// -> Retrieve Spec from etcd (if missing, 400, this returns err missing)
	// -> Retrieve Instance from etcd (if missing, 400, this returns err missing)
	// -> Update entry in /instance, ID'd by instance. Value should be Instance type
	//    Purpose is to make sure everything need to deprovision is available
	//    in persistence.
	// -> Update!
	////////////////////////////////////////////////////////////

	/*
	   dao GET returns error strings like CODE: message (entity) [#]
	   dao SetServiceInstance returns what error?
	   dao.SetState returns what error?
	   Provision returns what error?
	   SetExtractedCredentials returns what error?

	   broker
	   * normal synchronous return UpdateResponse
	   * normal async return UpdateResponse

	   handler returns the following
	   * synchronous update return 201 created
	   * instance already exists with IDENTICAL parameters to existing instance, 200 OK
	   * async provision 202 Accepted
	   * if only support async and no accepts_incomplete=true passed in, 422 Unprocessable entity

	*/

	var err error
	var fromPlanName string
	var fromPlan, toPlan bundle.Plan

	si, err := a.dao.GetServiceInstance(instanceUUID.String())
	if err != nil {
		log.Debug("Error retrieving instance")
		return nil, ErrorNotFound
	}

	// update the lastRequestingUserKey value in the si.Parameters
	if *si.Parameters != nil {
		(*si.Parameters)[lastRequestingUserKey] = getLastRequestingUser(userInfo)
	}

	// copy previous params, since the loaded si is mutated during update
	prevParams := make(bundle.Parameters)
	for k, v := range *si.Parameters {
		prevParams[k] = v
	}

	////////////////////////////////////////////////////////////
	// TODO -- HACK!: Update will report a 202 if it finds any jobs
	// in_progress for a particular instance, *even if the requests are different*.
	// This means an update must be completed before a user is able to further
	// request additional, possibly different updates. This should be considered
	// a known issue with our update implementation.
	//
	// The right way to do this is probably to setup an update request queue.
	// When a request comes in, hash it, check to see if there are any jobs in
	// the queue or currently in progress that match the hash. If so, $DO_SENSIBLE_THING,
	// else, add onto the back of the queue. Ensures update operations are not
	// trying to execute concurrently.
	////////////////////////////////////////////////////////////
	alreadyInProgress, jobToken, err := a.isJobInProgress(si.ID.String(), bundle.JobMethodUpdate)
	if err != nil {
		return nil, fmt.Errorf(
			"An error occurred while trying to determine if an update job is already in progress for instance: %s", si.ID)
	}
	if alreadyInProgress {
		log.Infof("Update requested for instance %s, but job is already in progress", si.ID)
		return &UpdateResponse{Operation: jobToken}, ErrorUpdateInProgress
	}
	////////////////////////////////////////////////////////////

	// Retrieve requested spec
	spec, err := a.dao.GetSpec(si.Spec.ID)
	if err != nil {
		// etcd return not found i.e. code 100
		if a.dao.IsNotFoundError(err) {
			return nil, ErrorNotFound
		}
		// otherwise unknown error bubble it up
		return nil, err
	}

	// NOTE: It might be better to actually pull this value from the *request*
	// sent from the catalog for the update, not the ServiceInstance parameters?
	fromPlanName, ok := (*si.Parameters)[planParameterKey].(string)
	if !ok {
		emsg := "Could not retrieve current plan name from parameters for update"
		log.Error(emsg)
		return nil, errors.New(emsg)
	}

	fromPlan, ok = spec.GetPlan(fromPlanName)
	if !ok {
		log.Errorf("The plan %s, specified for updating from on instance %s, does not exist.", fromPlanName, si.ID)
		return nil, ErrorPlanNotFound
	}

	log.Debugf("Update received the following Request.PlanID: [%s]", req.PlanID)

	if req.PlanID == "" {
		// Lock to currentPlan if no plan passed in request
		// No need to decode from FQPlanID -> ServiceClass scoped plan name, since
		// `fromPlanName` in this case is already decoded. Ex: "prod" instead of the md5 hash
		toPlan = fromPlan
	} else {
		// The catalog only identifies plans via their md5(FQPlanID), and will request
		// and update using that hash. If a PlanID is submitted, we'll need to look up
		// the ServiceClass scoped plan name via the passed in hash so the APB
		// will understand what to do with it, since APBs do not understand plan hashes.

		toPlan, ok = spec.GetPlanFromID(req.PlanID)
		if !ok {
			log.Errorf("Could not find requested PlanID %s in plan name lookup table", req.PlanID)
			return nil, ErrorPlanNotFound
		}
	}

	// If a plan transition has been requested, validate it is possible and then
	// update the service instance with the desired next plan
	if fromPlan.Name != toPlan.Name {
		log.Debugf("Validating plan transition from: %s, to: %s", fromPlan.Name, toPlan.Name)
		if ok := a.isValidPlanTransition(fromPlan, toPlan.Name); !ok {
			log.Errorf("The current plan, %s, cannot be updated to the requested plan, %s.", fromPlan.Name, toPlan.Name)
			return nil, ErrorPlanUpdateNotPossible
		}

		log.Debug("Plan transition valid!")
		(*si.Parameters)[planParameterKey] = toPlan.Name
	} else {
		log.Debug("Plan transition NOT requested as part of update")
	}

	req.Parameters, err = a.validateRequestedUpdateParams(req.Parameters, toPlan, prevParams, si)
	if err != nil {
		return nil, err
	}

	if fromPlan.Name == toPlan.Name && len(req.Parameters) == 0 {
		log.Warningf("Returning without running the APB. No changes were actually requested")

		return &UpdateResponse{}, ErrorNoUpdateRequested
	}

	// Parameters look good, update the ServiceInstance values
	for newParamKey, newParamVal := range req.Parameters {
		(*si.Parameters)[newParamKey] = newParamVal
	}

	// We're ready to provision so save
	if err = a.dao.SetServiceInstance(instanceUUID.String(), si); err != nil {
		return nil, err
	}

	var token = a.engine.Token()

	log.Debug("Initiating update with the inputs:")
	log.Debugf("fromPlanName: [%s]", fromPlanName)
	log.Debugf("toPlanName: [%s]", toPlan.Name)
	log.Debugf("PreviousValues: [ %+v ]", req.PreviousValues)
	log.Debugf("ServiceInstance Parameters: [%v]", *si.Parameters)
	ujob := &UpdateJob{si}
	metrics.ActionStarted("update")
	if async {
		log.Info("ASYNC update in progress")
		// asynchronously provision and return the token for the lastoperation
		token, err = a.engine.StartNewAsyncJob(token, ujob, UpdateTopic)
		if err != nil {
			log.Errorf("Failed to start new job for async update\n%s", err.Error())
			return nil, err
		}
	} else {
		log.Info("reverting to synchronous update in progress")
		if err := a.engine.StartNewSyncJob(token, ujob, UpdateTopic); err != nil {
			log.Errorf("Failed to start new job for sync update\n%s", err.Error())
			return nil, err
		}
	}

	return &UpdateResponse{Operation: token}, nil
}

func (a AnsibleBroker) isValidPlanTransition(fromPlan bundle.Plan, toPlanName string) bool {
	// Make sure that we can find the plan we're updating from.
	// This should probably never fail, but cover our tail.
	for _, validToPlanName := range fromPlan.UpdatesTo {
		if validToPlanName == toPlanName {
			return true
		}
	}
	return false
}

func (a AnsibleBroker) validateRequestedUpdateParams(
	reqParams map[string]string,
	toPlan bundle.Plan,
	prevParams map[string]interface{},
	si *bundle.ServiceInstance,
) (map[string]string, error) {
	log.Debugf("Validating update parameters...")
	log.Debugf("Request Params: %v", reqParams)
	log.Debugf("Previous Params: %v", prevParams)

	// The catalog will always pass all parameters for update, so let's filter
	// out parameters that the user has not changed first.
	changedParams := make(map[string]string)
	for reqParam, reqVal := range reqParams {
		if prevVal, ok := prevParams[reqParam]; ok {
			if reqVal == prevVal {
				continue
			}
		}
		changedParams[reqParam] = reqVal
	}
	log.Debugf("Changed Params: %v", changedParams)

	for reqParam := range changedParams {
		pd := toPlan.GetParameter(reqParam)
		if pd == nil {
			// Confirm the parameter actually exists on the plan
			log.Warningf("Removing non-parameter %s, requested for update on instance %s, from request.", reqParam, si.ID)
			return nil, ErrorParameterNotFound
		} else if !pd.Updatable {
			log.Errorf("Request attempted to update non-updatable parameter: %v on ServiceInstance: %v", reqParam, si.ID)
			return nil, ErrorParameterNotUpdatable
		} else if pd.Type == "enum" {
			enums := make(map[string]bool)
			for _, v := range pd.Enum {
				enums[v] = true
			}
			if !enums[changedParams[reqParam]] {
				log.Warningf("Removing invalid enum parameter %s, requested for update on instance %s, from request.", reqParam, si.ID)
				return nil, ErrorParameterUnknownEnum
			}
		}
	}

	log.Debugf("Validated Params: %v", changedParams)
	return changedParams, nil
}

// LastOperation - gets the last operation and status
func (a AnsibleBroker) LastOperation(instanceUUID uuid.UUID, req *LastOperationRequest,
) (*LastOperationResponse, error) {
	/*
		look up the resource in etcd the operation should match what was returned by provision
		take the status and return that.

		process:

		if async, provision: it should create a Job that calls bundle.Provision. And write the output to etcd.
	*/
	log.Debugf("service_id: %s", req.ServiceID)
	log.Debugf("plan_id: %s", req.PlanID)
	log.Debugf("operation:  %s", req.Operation) // Operation is the job token id from the work_engine

	jobstate, err := a.dao.GetState(instanceUUID.String(), req.Operation)
	if err != nil {
		// not sure what we do with the error if we can't find the state
		log.Warningf("unable to find job state: [%s]. error: [%v]", instanceUUID, err.Error())
		if a.dao.IsNotFoundError(err) {
			return &LastOperationResponse{}, ErrorNotFound
		}
	}

	state := StateToLastOperation(jobstate.State)
	log.Debugf("state: %s", state)
	log.Debugf("description: %s", jobstate.Description)
	if jobstate.Error != "" {
		log.Debugf("job state has an error. Assuming that any error here is human readable. err - %v", jobstate.Error)
	}
	return &LastOperationResponse{State: state, Description: jobstate.Description}, err
}

// AddSpec - adding the spec to the catalog for local development
func (a AnsibleBroker) AddSpec(spec bundle.Spec) (*CatalogResponse, error) {
	log.Debug("broker::AddSpec")
	spec.Image = spec.FQName
	addNameAndIDForSpec([]*bundle.Spec{&spec}, apbPushRegName)
	log.Debugf("Generated name for pushed APB: [%s], ID: [%s]", spec.FQName, spec.ID)
	if err := a.dao.SetSpec(spec.ID, &spec); err != nil {
		return nil, err
	}
	bundle.AddSecretsFor(&spec)
	service, err := SpecToService(&spec)
	if err != nil {
		log.Debugf("spec was not added due to issue with transformation to service - %v", err)
		return nil, err
	}
	metrics.SpecsLoaded(apbPushRegName, 1)
	return &CatalogResponse{Services: []Service{service}}, nil
}

// RemoveSpec - remove the spec specified from the catalog/etcd
func (a AnsibleBroker) RemoveSpec(specID string) error {
	spec, err := a.dao.GetSpec(specID)
	if a.dao.IsNotFoundError(err) {
		return ErrorNotFound
	}
	if err != nil {
		log.Errorf("Something went real bad trying to retrieve spec for deletion... - %v", err)
		return err
	}
	err = a.dao.DeleteSpec(spec.ID)
	if err != nil {
		log.Errorf("Something went real bad trying to delete spec... - %v", err)
		return err
	}
	metrics.SpecsUnloaded(apbPushRegName, 1)
	return nil
}

// RemoveSpecs - remove all the specs from the catalog/etcd
func (a AnsibleBroker) RemoveSpecs() error {
	dir := "/spec"
	specs, err := a.dao.BatchGetSpecs(dir)
	if err != nil {
		log.Errorf("Something went real bad trying to retrieve batch specs for deletion... - %v", err)
		return err
	}
	err = a.dao.BatchDeleteSpecs(specs)
	if err != nil {
		log.Errorf("Something went real bad trying to delete batch specs... - %v", err)
		return err
	}
	metrics.SpecsLoadedReset()
	return nil
}
