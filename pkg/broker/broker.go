//
// Copyright (c) 2017 Red Hat, Inc.
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

	"github.com/coreos/etcd/client"
	"github.com/openshift/ansible-service-broker/pkg/apb"
	"github.com/openshift/ansible-service-broker/pkg/config"
	"github.com/openshift/ansible-service-broker/pkg/dao"
	"github.com/openshift/ansible-service-broker/pkg/metrics"
	"github.com/openshift/ansible-service-broker/pkg/registries"
	logutil "github.com/openshift/ansible-service-broker/pkg/util/logging"
	"github.com/pborman/uuid"
)

var (
	// ErrorAlreadyProvisioned - Error for when an service instance has already been provisioned
	ErrorAlreadyProvisioned = errors.New("already provisioned")
	// ErrorDuplicate - Error for when a duplicate service instance already exists
	ErrorDuplicate = errors.New("duplicate instance")
	// ErrorNotFound  - Error for when a service instance is not found. (either etcd or kubernetes)
	ErrorNotFound = errors.New("not found")
	// ErrorBindingExists - Error for when deprovision is called on a service instance with active bindings
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
	// ErrorPlanUpdateNotPossible - Error when a Plan Update request cannot be satisfied
	ErrorPlanUpdateNotPossible = errors.New("plan update not possible")
	log                        = logutil.NewLog()
)

const (
	// provisionCredentialsKey - Key used to pass credentials to apb.
	provisionCredentialsKey = "_apb_provision_creds"
	// bindCredentialsKey - Key used to pas bind credentials to apb.
	bindCredentialsKey = "_apb_bind_creds"
	// fqNameRegex - regular expression used when forming FQName.
	fqNameRegex = "[/.:-]"
)

// Broker - A broker is used to to compelete all the tasks that a broker must be able to do.
type Broker interface {
	Bootstrap() (*BootstrapResponse, error)
	Catalog() (*CatalogResponse, error)
	Provision(uuid.UUID, *ProvisionRequest, bool) (*ProvisionResponse, error)
	Update(uuid.UUID, *UpdateRequest, bool) (*UpdateResponse, error)
	Deprovision(apb.ServiceInstance, string, bool, bool) (*DeprovisionResponse, error)
	Bind(apb.ServiceInstance, uuid.UUID, *BindRequest, bool) (*BindResponse, error)
	Unbind(apb.ServiceInstance, uuid.UUID, string, bool, bool) (*UnbindResponse, error)
	LastOperation(uuid.UUID, *LastOperationRequest) (*LastOperationResponse, error)
	// TODO: consider returning a struct + error
	Recover() (string, error)
	GetServiceInstance(uuid.UUID) (apb.ServiceInstance, error)
	GetBind(apb.ServiceInstance, uuid.UUID) (*BindResponse, error)
}

// Config - Configuration for the broker.
type Config struct {
	DevBroker          bool   `yaml:"dev_broker"`
	LaunchApbOnBind    bool   `yaml:"launch_apb_on_bind"`
	BootstrapOnStartup bool   `yaml:"bootstrap_on_startup"`
	Recovery           bool   `yaml:"recovery"`
	OutputRequest      bool   `yaml:"output_request"`
	SSLCertKey         string `yaml:"ssl_cert_key"`
	SSLCert            string `yaml:"ssl_cert"`
	RefreshInterval    string `yaml:"refresh_interval"`
	AutoEscalate       bool   `yaml:"auto_escalate"`
	ClusterURL         string `yaml:"cluster_url"`
}

// DevBroker - Interface for the development broker.
type DevBroker interface {
	AddSpec(spec apb.Spec) (*CatalogResponse, error)
	RemoveSpec(specID string) error
	RemoveSpecs() error
}

// AnsibleBroker - Broker using ansible and images to interact with oc/kubernetes/etcd
type AnsibleBroker struct {
	dao          *dao.Dao
	registry     []registries.Registry
	engine       *WorkEngine
	brokerConfig Config
}

// NewAnsibleBroker - Creates a new ansible broker
func NewAnsibleBroker(dao *dao.Dao,
	registry []registries.Registry,
	engine WorkEngine,
	brokerConfig *config.Config) (*AnsibleBroker, error) {

	broker := &AnsibleBroker{
		dao:      dao,
		registry: registry,
		engine:   &engine,
		brokerConfig: Config{
			DevBroker:          brokerConfig.GetBool("dev_broker"),
			LaunchApbOnBind:    brokerConfig.GetBool("launch_apb_on_bind"),
			BootstrapOnStartup: brokerConfig.GetBool("bootstrap_on_startup"),
			Recovery:           brokerConfig.GetBool("recovery"),
			OutputRequest:      brokerConfig.GetBool("output_request"),
			SSLCertKey:         brokerConfig.GetString("ssl_cert_key"),
			SSLCert:            brokerConfig.GetString("ssl_cert"),
			RefreshInterval:    brokerConfig.GetString("refresh_interval"),
			AutoEscalate:       brokerConfig.GetBool("auto_escalate"),
			ClusterURL:         brokerConfig.GetString("cluster_url"),
		},
	}
	return broker, nil
}

// GetServiceInstance - retrieve the service instance for a instanceID.
func (a AnsibleBroker) GetServiceInstance(instanceUUID uuid.UUID) (apb.ServiceInstance, error) {
	instance, err := a.dao.GetServiceInstance(instanceUUID.String())
	if err != nil {
		if client.IsKeyNotFound(err) {
			log.Errorf("Could not find a service instance in dao - %v", err)
			return apb.ServiceInstance{}, ErrorNotFound
		}
		log.Error("Couldn't find a service instance: ", err)
		return apb.ServiceInstance{}, err
	}
	return *instance, nil

}

// Bootstrap - Loads all known specs from a registry into local storage for reference
// Potentially a large download; on the order of 10s of thousands
// TODO: Response here? Async?
// TODO: How do we handle a large amount of data on this side as well? Pagination?
func (a AnsibleBroker) Bootstrap() (*BootstrapResponse, error) {
	log.Info("AnsibleBroker::Bootstrap")
	var err error
	var specs []*apb.Spec
	var imageCount int

	// Remove all non apb-push sourced specs that have been saved.
	pushedSpecs := []*apb.Spec{}
	dir := "/spec"
	specs, err = a.dao.BatchGetSpecs(dir)
	if err != nil {
		log.Error("Something went real bad trying to retrieve batch specs for deletion... - %v", err)
		return nil, err
	}
	// Save all apb-push sourced specs
	for _, spec := range specs {
		if strings.HasPrefix(spec.FQName, "apb-push") {
			log.Info("Saving apb-push sourced spec to prevent deletion: %v", spec.FQName)
			pushedSpecs = append(pushedSpecs, spec)
		}
	}

	err = a.dao.BatchDeleteSpecs(specs)
	if err != nil {
		log.Error("Something went real bad trying to delete batch specs... - %v", err)
		return nil, err
	}
	specs = []*apb.Spec{}
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
		specs = append(specs, s...)
	}
	// Add apb-push sourced specs back to the list
	for _, spec := range pushedSpecs {
		specs = append(specs, spec)
	}
	if len(registryErrors) == len(a.registry) {
		return nil, errors.New("all registries failed on bootstrap")
	}
	specManifest := map[string]*apb.Spec{}
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

	// save off the plan names as well
	if err = a.dao.BatchSetPlanNames(planNameManifest); err != nil {
		return nil, err
	}

	apb.AddSecrets(specs)

	return &BootstrapResponse{SpecCount: len(specs), ImageCount: imageCount}, nil
}

// addNameAndIDForSpec - will create the unique spec name and id
// and set it for each spec
func addNameAndIDForSpec(specs []*apb.Spec, registryName string) {
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
func addIDForPlan(plans []apb.Plan, FQSpecName string) {

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
	// At startup we should write a key to etcd.
	// Then in recovery see if that key exists, which means we are restarting
	// and need to try to recover.

	// do we have any jobs that wre still running?
	// get all /state/*/jobs/* == in progress
	// For each job, check the status of each of their containers to update
	// their status in case any of them finished.

	recoverStatuses, err := a.dao.FindJobStateByState(apb.StateInProgress)
	if err != nil {
		// no jobs or states to recover, this is OK.
		if client.IsKeyNotFound(err) {
			log.Info("No jobs to recover")
			return "", nil
		}
		return "", err
	}

	/*
		if job was in progress we know instanceuuid & token. do we have a podname?
		if no, job never started
			restart
		if yes,
			did it finish?
				yes
					* update status
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
			return "", err
		}

		// Do we have a podname?
		if rs.State.Podname == "" {
			// NO, we do not have a podname

			log.Info(fmt.Sprintf("No podname. Attempting to restart job: %s", instanceID))

			log.Debug(fmt.Sprintf("%v", instance))

			// Handle bad write of service instance
			if instance.Spec == nil || instance.Parameters == nil {
				a.dao.SetState(instanceID, apb.JobState{
					Token:  rs.State.Token,
					State:  apb.StateFailed,
					Method: rs.State.Method,
				})
				a.dao.DeleteServiceInstance(instance.ID.String())
				log.Warning(fmt.Sprintf("incomplete ServiceInstance [%s] record, marking job as failed", instance.ID))
				// skip to the next item
				continue
			}

			var job Work
			var topic WorkTopic
			if rs.State.Method == apb.JobMethodProvision {
				job = NewProvisionJob(instance)
				topic = ProvisionTopic
			} else if rs.State.Method == apb.JobMethodUpdate {
				job = NewUpdateJob(instance)
				topic = UpdateTopic
			} else {
				log.Warningf(
					"Attempted to recover job %s, but found an unrecognized "+
						"MethodType: %s, skipping...",
					rs.State.Token, rs.State.Method,
				)
			}

			// Need to use the same token as before, since that's what the
			// catalog will try to ping.
			_, err := a.engine.StartNewJob(rs.State.Token, job, topic)
			if err != nil {
				return "", err
			}

			// HACK: there might be a delay between the first time the state in etcd
			// is set and the job was already started. But I need the token.
			a.dao.SetState(instanceID, apb.JobState{
				Token:  rs.State.Token,
				State:  apb.StateInProgress,
				Method: rs.State.Method,
			})
		} else {
			// YES, we have a podname
			log.Info(fmt.Sprintf("We have a pod to recover: %s", rs.State.Podname))

			// TODO: ExtractCredentials is doing more than it should
			// be and it needs to be broken up.

			// did the pod finish?
			extCreds, extErr := apb.ExtractCredentials(
				rs.State.Podname,
				instance.Context.Namespace,
				instance.Spec.Runtime,
			)

			// NO, pod failed.
			// TODO: do we restart the job or mark it as failed?
			if extErr != nil {
				log.Error("broker::Recover error occurred.")
				log.Error("%s", extErr.Error())
				return "", extErr
			}

			// YES, pod finished we have creds
			if extCreds != nil {
				log.Debug("broker::Recover, got ExtractedCredentials!")
				a.dao.SetState(instanceID, apb.JobState{
					Token:   rs.State.Token,
					State:   apb.StateSucceeded,
					Podname: rs.State.Podname,
					Method:  rs.State.Method,
				})
				err = a.dao.SetExtractedCredentials(instanceID, extCreds)
				if err != nil {
					log.Error("Could not persist extracted credentials")
					log.Error("%s", err.Error())
					return "", err
				}
			}
		}
	}

	// if no pods, do we restart? or just return failed?

	// binding

	log.Info("Recovery complete")
	return "recover called", nil
}

// Catalog - returns the catalog of services defined
func (a AnsibleBroker) Catalog() (*CatalogResponse, error) {
	log.Info("AnsibleBroker::Catalog")

	var specs []*apb.Spec
	var err error
	var services []Service
	dir := "/spec"

	if specs, err = a.dao.BatchGetSpecs(dir); err != nil {
		log.Error("Something went real bad trying to retrieve batch specs...")
		return nil, err
	}

	log.Debugf("Filtering secret parameters out of specs...")
	specs, err = apb.FilterSecrets(specs)
	if err != nil {
		// TODO: Should we blow up or warn and continue?
		log.Errorf("Something went real bad trying to load secrets %v", err)
		return nil, err
	}

	services = []Service{}
	for _, spec := range specs {
		ser, err := SpecToService(spec)
		if err != nil {
			log.Errorf("not adding spec %v to list of services due to error transforming to service - %v", spec.FQName, err)
		} else {
			services = append(services, ser)
		}
	}

	return &CatalogResponse{services}, nil
}

// Provision  - will provision a service
func (a AnsibleBroker) Provision(instanceUUID uuid.UUID, req *ProvisionRequest, async bool,
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
	// the request plan asyncronously, broker should reject with a 422
	// NOTE: Spec.Async should indicate what level of async support is available for
	// a given ansible app

	//}

	// Summary:
	// For our purposes right now, the ServiceID and the Params should be enough to
	// Provision an ansible app.
	////////////////////////////////////////////////////////////
	// Provision Flow
	// -> Retrieve Spec from etcd (if missing, 400, this returns err missing)
	// -> TODO: Check to see if the spec supports or requires async, and reconcile
	//    need a typed error condition so the REST server knows correct response
	//    depending on the scenario
	//    (async requested, unsupported, 422)
	//    (async not requested, required, ?)
	// -> Make entry in /instance, ID'd by instance. Value should be Instance type
	//    Purpose is to make sure everything neeed to deprovision is available
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
	var spec *apb.Spec
	var err error
	var planName string

	// Retrieve requested spec
	specID := req.ServiceID
	if spec, err = a.dao.GetSpec(specID); err != nil {
		// etcd return not found i.e. code 100
		if client.IsKeyNotFound(err) {
			return nil, ErrorNotFound
		}
		// otherwise unknown error bubble it up
		return nil, err
	}

	context := &req.Context
	parameters := req.Parameters
	if parameters == nil {
		parameters = make(apb.Parameters)
	}

	if req.PlanID == "" {
		errMsg :=
			"PlanID from provision request is blank. " +
				"Provision requests must specify PlanIDs"
		return nil, errors.New(errMsg)
	}

	planName, err = a.dao.GetPlanName(req.PlanID)
	if err != nil {
		// etcd return not found i.e. code 100
		if client.IsKeyNotFound(err) {
			return nil, ErrorNotFound
		}
		// otherwise unknown error bubble it up
		return nil, err
	}

	log.Debugf(
		"Injecting PlanID as parameter: { %s: %s }",
		planParameterKey, planName)
	parameters[planParameterKey] = planName
	log.Debugf("Injecting ServiceClassID as parameter: { %s: %s }",
		serviceClassIDKey, req.ServiceID)
	parameters[serviceClassIDKey] = req.ServiceID
	log.Debugf("Injecting ServiceInstanceID as parameter: { %s: %s }",
		serviceInstIDKey, instanceUUID.String())
	parameters[serviceInstIDKey] = instanceUUID.String()

	// Build and persist record of service instance
	serviceInstance := &apb.ServiceInstance{
		ID:         instanceUUID,
		Spec:       spec,
		Context:    context,
		Parameters: &parameters,
	}

	// Verify we're not reprovisioning the same instance
	// if err is nil, there is an instance. Let's compare it to the instance
	// we're being asked to provision.
	//
	// if err is not nil, we will just bubble that up

	if si, err := a.dao.GetServiceInstance(instanceUUID.String()); err == nil {
		// This will use the package to make sure that if the type is changed
		// away from []byte it can still be evaluated.
		if uuid.Equal(si.ID, serviceInstance.ID) {
			if reflect.DeepEqual(si.Parameters, serviceInstance.Parameters) {
				alreadyInProgress, jobToken, err := a.isJobInProgress(serviceInstance, apb.JobMethodProvision)
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
	}

	//
	// Looks like this is a new provision, let's get started.
	//
	if err = a.dao.SetServiceInstance(instanceUUID.String(), serviceInstance); err != nil {
		return nil, err
	}

	var token string

	if async {
		log.Info("ASYNC provisioning in progress")
		// asyncronously provision and return the token for the lastoperation
		pjob := NewProvisionJob(serviceInstance)

		token, err = a.engine.StartNewJob("", pjob, ProvisionTopic)
		if err != nil {
			log.Error("Failed to start new job for async provision\n%s", err.Error())
			return nil, err
		}

		// HACK: there might be a delay between the first time the state in etcd
		// is set and the job was already started. But I need the token.
		a.dao.SetState(instanceUUID.String(), apb.JobState{
			Token:  token,
			State:  apb.StateInProgress,
			Method: apb.JobMethodProvision,
		})
	} else {
		// TODO: do we want to do synchronous provisioning?
		log.Info("reverting to synchronous provisioning in progress")
		_, extCreds, err := apb.Provision(serviceInstance)
		if extCreds != nil {
			log.Debug("broker::Provision, got ExtractedCredentials!")
			err = a.dao.SetExtractedCredentials(instanceUUID.String(), extCreds)
			if err != nil {
				log.Error("Could not persist extracted credentials")
				log.Error("%s", err.Error())
				return nil, err
			}
		}
	}

	// TODO: What data needs to be sent back on a response?
	// Not clear what dashboardURL means in an AnsibleApp context
	// operation should be the task id from the work_engine
	return &ProvisionResponse{Operation: token}, nil
}

// Deprovision - will deprovision a service.
func (a AnsibleBroker) Deprovision(
	instance apb.ServiceInstance, planID string, skipApbExecution bool, async bool,
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

	alreadyInProgress, jobToken, err := a.isJobInProgress(&instance, apb.JobMethodDeprovision)
	if err != nil {
		return nil, fmt.Errorf("An error occurred while trying to determine if a deprovision job is already in progress for instance: %s", instance.ID)
	}

	if alreadyInProgress {
		log.Infof("Deprovision requested for instance %s, but job is already in progress", instance.ID)
		return &DeprovisionResponse{Operation: jobToken}, ErrorDeprovisionInProgress
	}

	var token string

	if async {
		log.Info("ASYNC deprovision in progress")
		// asynchronously provision and return the token for the lastoperation
		dpjob := NewDeprovisionJob(&instance, skipApbExecution, a.dao)

		token, err = a.engine.StartNewJob("", dpjob, DeprovisionTopic)
		if err != nil {
			log.Error("Failed to start new job for async deprovision\n%s", err.Error())
			return nil, err
		}

		// HACK: there might be a delay between the first time the state in etcd
		// is set and the job was already started. But I need the token.
		a.dao.SetState(instance.ID.String(), apb.JobState{
			Token:  token,
			State:  apb.StateInProgress,
			Method: apb.JobMethodDeprovision,
		})
		return &DeprovisionResponse{Operation: token}, nil
	}

	// TODO: do we want to do synchronous deprovisioning?
	if !skipApbExecution {
		log.Info("Synchronous deprovision in progress")
		_, err = apb.Deprovision(&instance)
		if err != nil {
			return nil, err
		}
	}

	err = cleanupDeprovision(&instance, a.dao)
	if err != nil {
		return nil, err
	}
	return &DeprovisionResponse{}, nil
}

func (a AnsibleBroker) validateDeprovision(instance *apb.ServiceInstance) error {
	// -> Lookup bindings by instance ID; 400 if any are active, related issue:
	//    https://github.com/openservicebrokerapi/servicebroker/issues/127
	if len(instance.BindingIDs) > 0 {
		log.Debugf("Found bindings with ids: %v", instance.BindingIDs)
		return ErrorBindingExists
	}

	return nil
}

func (a AnsibleBroker) isJobInProgress(instance *apb.ServiceInstance,
	method apb.JobMethod) (bool, string, error) {

	allJobs, err := a.dao.GetSvcInstJobsByState(instance.ID.String(), apb.StateInProgress)
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
func (a AnsibleBroker) GetBind(instance apb.ServiceInstance, bindingUUID uuid.UUID) (*BindResponse, error) {

	log.Debug("broker.GetBind: entered GetBind")

	provExtCreds, err := a.dao.GetExtractedCredentials(instance.ID.String())
	if err != nil && !client.IsKeyNotFound(err) {
		log.Warningf("unable to retrieve provision time credentials - %v", err)
		return nil, err
	}

	bi, err := a.dao.GetBindInstance(bindingUUID.String())
	if err != nil {
		if client.IsKeyNotFound(err) {
			log.Warningf("id: %v - could not find bind instance - %v", bindingUUID, err)
			return nil, ErrorNotFound
		}
		log.Warningf("id: %v - unable to retrieve bind instance - %v", bindingUUID, err)
		return nil, err
	}

	bindExtCreds, err := a.dao.GetExtractedCredentials(bi.ID.String())
	if err != nil {
		if client.IsKeyNotFound(err) {
			return nil, ErrorNotFound
		}

		return nil, err
	}

	log.Debug("broker.GetBind: we got the bind credentials")
	return a.buildBindResponse(provExtCreds, bindExtCreds, false, "")
}

// Bind - will create a binding between a service.
func (a AnsibleBroker) Bind(instance apb.ServiceInstance, bindingUUID uuid.UUID, req *BindRequest, async bool,
) (*BindResponse, error) {
	// binding_id is the id of the binding.
	// the instanceUUID is the previously provisioned service id.
	//
	// See if the service instance still exists, if not send back a badrequest.

	// GET SERVICE get provision parameters
	params := req.Parameters
	if params == nil {
		params = make(apb.Parameters)
	}

	// Inject PlanID into parameters passed to APBs
	if req.PlanID == "" {
		errMsg :=
			"PlanID from bind request is blank. " +
				"Bind requests must specify PlanIDs"
		return nil, errors.New(errMsg)
	}

	planName, err := a.dao.GetPlanName(req.PlanID)
	if err != nil {
		// etcd return not found i.e. code 100
		if client.IsKeyNotFound(err) {
			return nil, ErrorNotFound
		}
		// otherwise unknown error bubble it up
		return nil, err
	}

	log.Debugf(
		"Injecting PlanID as parameter: { %s: %s }",
		planParameterKey, planName)

	params[planParameterKey] = planName
	log.Debugf("Injecting ServiceClassID as parameter: { %s: %s }",
		serviceClassIDKey, req.ServiceID)

	params[serviceClassIDKey] = req.ServiceID
	log.Debugf("Injecting ServiceInstanceID as parameter: { %s: %s }",
		serviceInstIDKey, instance.ID.String())

	params[serviceInstIDKey] = instance.ID.String()

	// Create a BindingInstance with a reference to the serviceinstance.
	bindingInstance := &apb.BindInstance{
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
	provExtCreds, err := a.dao.GetExtractedCredentials(instance.ID.String())
	if err != nil && !client.IsKeyNotFound(err) {
		log.Warningf("unable to retrieve provision time credentials - %v", err)
		return nil, err
	}
	if bi, err := a.dao.GetBindInstance(bindingUUID.String()); err == nil {
		if uuid.Equal(bi.ID, bindingInstance.ID) {
			if reflect.DeepEqual(bi.Parameters, bindingInstance.Parameters) {
				bindExtCreds, err := a.dao.GetExtractedCredentials(bi.ID.String())
				if err != nil && !client.IsKeyNotFound(err) {
					return nil, err
				}
				log.Debug("already have this binding instance, returning 200")
				// since we have this already, we can set async to false
				return a.buildBindResponse(provExtCreds, bindExtCreds, false, "")
			}

			// parameters are different
			log.Info("duplicate binding instance diff params, returning 409 conflict")
			return nil, ErrorDuplicate
		}
	}

	if err := a.dao.SetBindInstance(bindingUUID.String(), bindingInstance); err != nil {
		return nil, err
	}

	// Add the DB Credentials this will allow the apb to use these credentials
	// if it so chooses.
	if provExtCreds != nil {
		params[provisionCredentialsKey] = provExtCreds.Credentials
	}

	// NOTE: We are currently disabling running an APB on bind via
	// 'LaunchApbOnBind' of the broker config, due to lack of async support of
	// bind in Open Service Broker API Currently, the 'launchapbonbind' is set
	// to false in the 'config' ConfigMap
	var bindExtCreds *apb.ExtractedCredentials
	metrics.ActionStarted("bind")
	var token string

	if async && a.brokerConfig.LaunchApbOnBind {
		// asynchronous mode, requires that the launch apb config
		// entry is on, and that async comes in from the catalog
		log.Info("ASYNC binding in progress")
		bindjob := NewBindingJob(&instance, bindingUUID, &params)

		token, err = a.engine.StartNewJob("", bindjob, BindingTopic)
		if err != nil {
			log.Error("Failed to start new job for async binding\n%s", err.Error())
			return nil, err
		}

		if err := a.dao.SetState(instance.ID.String(), apb.JobState{
			Token:  token,
			State:  apb.StateInProgress,
			Method: apb.JobMethodBind,
		}); err != nil {
			log.Errorf("failed to set initial jobstate for %v, %v", token, err.Error())
		}
	} else if a.brokerConfig.LaunchApbOnBind {
		// we are synchronous mode
		// TODO: decide if we need to keep this at all
		log.Info("Broker configured to run APB bind")
		_, bindExtCreds, err = apb.Bind(&instance, &params)

		if err != nil {
			return nil, err
		}
		instance.AddBinding(bindingUUID)
		if err := a.dao.SetServiceInstance(instance.ID.String(), &instance); err != nil {
			return nil, err
		}
		if bindExtCreds != nil {
			err = a.dao.SetExtractedCredentials(bindingUUID.String(), bindExtCreds)
			if err != nil {
				log.Errorf("Could not persist extracted credentials - %v", err)
				return nil, err
			}
		}
	} else {
		log.Warning("Broker configured to *NOT* launch and run APB bind")
	}

	// TODO: if we're sync we need to return the Credentials, otherwise the
	// Operation with the token
	// maybe we just have 2 returns for each if/else segment
	return a.buildBindResponse(provExtCreds, bindExtCreds, async, token)
}

func (a AnsibleBroker) buildBindResponse(pCreds, bCreds *apb.ExtractedCredentials,
	async bool, token string) (*BindResponse, error) {

	if async {
		if token == "" {
			log.Error("async used but no job token was created")
			return nil, errors.New("no job token created during async")
		}

		return &BindResponse{Operation: token}, nil
	}

	// Can't bind to anything if we have nothing to return to the catalog
	if pCreds == nil && bCreds == nil {
		log.Errorf("No extracted credentials found from provision or bind instance ID")
		return nil, errors.New("No credentials available")
	}

	if bCreds != nil {
		log.Debugf("bind creds: %v", bCreds.Credentials)
		return &BindResponse{Credentials: bCreds.Credentials}, nil
	}

	log.Debugf("provision bind creds: %v", pCreds.Credentials)
	return &BindResponse{Credentials: pCreds.Credentials}, nil
}

// Unbind - unbind a services previous binding
func (a AnsibleBroker) Unbind(
	instance apb.ServiceInstance, bindingUUID uuid.UUID, planID string, skipApbExecution bool, async bool,
) (*UnbindResponse, error) {
	if planID == "" {
		errMsg :=
			"PlanID from unbind request is blank. " +
				"Unbind requests must specify PlanIDs"
		return nil, errors.New(errMsg)
	}

	params := make(apb.Parameters)
	provExtCreds, err := a.dao.GetExtractedCredentials(instance.ID.String())
	if err != nil && !client.IsKeyNotFound(err) {
		return nil, err
	}
	bindExtCreds, err := a.dao.GetExtractedCredentials(bindingUUID.String())
	if err != nil && !client.IsKeyNotFound(err) {
		return nil, err
	}
	// Add the credentials to the parameters so that an APB can choose what
	// it would like to do.
	if provExtCreds == nil && bindExtCreds == nil {
		log.Warningf("Unable to find credentials for instance id: %v and binding id: %v"+
			" something may have gone wrong. Proceeding with unbind.",
			instance.ID, bindingUUID)
	}
	if provExtCreds != nil {
		params[provisionCredentialsKey] = provExtCreds.Credentials
	}
	if bindExtCreds != nil {
		params[bindCredentialsKey] = bindExtCreds.Credentials
	}
	serviceInstance, err := a.GetServiceInstance(instance.ID)
	if err != nil {
		log.Debugf("Service instance with id %s does not exist", instance.ID.String())
		return nil, err
	}
	if serviceInstance.Parameters != nil {
		params["provision_params"] = *serviceInstance.Parameters
	}
	metrics.ActionStarted("unbind")

	var token string
	if async && a.brokerConfig.LaunchApbOnBind {
		// asynchronous mode, required that the launch apb config
		// entry is on, and that async comes in from the catalog
		log.Info("ASYNC unbinding in progress")
		unbindjob := NewUnbindingJob(&serviceInstance, bindingUUID, &params, skipApbExecution)
		token, err := a.engine.StartNewJob("", unbindjob, UnbindingTopic)
		if err != nil {
			log.Error("Failed to start new job for async unbind\n%s", err.Error())
			return nil, err
		}

		if err := a.dao.SetState(serviceInstance.ID.String(), apb.JobState{
			Token:  token,
			State:  apb.StateInProgress,
			Method: apb.JobMethodUnbind,
		}); err != nil {
			log.Errorf("failed to set initial jobstate for %v, %v", token, err.Error())
		}

	} else if a.brokerConfig.LaunchApbOnBind {
		// only launch apb if we are always launching the APB.
		if skipApbExecution {
			log.Debug("Skipping unbind apb execution")
			err = nil
		} else {
			err = apb.Unbind(&serviceInstance, &params)
		}
		if err != nil {
			return nil, err
		}
	} else {
		log.Warning("Broker configured to *NOT* launch and run APB unbind")
	}

	if bindExtCreds != nil {
		err = a.dao.DeleteExtractedCredentials(bindingUUID.String())
		if err != nil {
			return nil, err
		}
	}

	err = a.dao.DeleteBindInstance(bindingUUID.String())
	if err != nil {
		return nil, err
	}

	serviceInstance.RemoveBinding(bindingUUID)
	err = a.dao.SetServiceInstance(instance.ID.String(), &serviceInstance)
	if err != nil {
		return nil, err
	}

	if token != "" {
		return &UnbindResponse{Operation: token}, nil
	}
	return &UnbindResponse{}, nil
}

// Update  - will update a service
func (a AnsibleBroker) Update(instanceUUID uuid.UUID, req *UpdateRequest, async bool,
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
	// the request plan asyncronously, broker should reject with a 422
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
	// -> TODO: Check to see if the spec supports or requires async, and reconcile
	//    need a typed error condition so the REST server knows correct response
	//    depending on the scenario
	//    (async requested, unsupported, 422)
	//    (async not requested, required, ?)
	// -> Update entry in /instance, ID'd by instance. Value should be Instance type
	//    Purpose is to make sure everything neeed to deprovision is available
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
	var fromPlanName, toPlanName string
	var fromPlan, toPlan *apb.Plan

	si, err := a.dao.GetServiceInstance(instanceUUID.String())
	if err != nil {
		log.Debug("Error retrieving instance")
		return nil, ErrorNotFound
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
	alreadyInProgress, jobToken, err := a.isJobInProgress(si, apb.JobMethodUpdate)
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
		if client.IsKeyNotFound(err) {
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

	log.Debugf("Update received the following Request.PlanID: [%s]", req.PlanID)

	if req.PlanID == "" {
		// Lock to currentPlan if no plan passed in request
		// No need to decode from FQPlanID -> ServiceClass scoped plan name, since
		// `fromPlanName` in this case is already decoded. Ex: "prod" instead of the md5 hash
		toPlanName = fromPlanName
	} else {
		// The catalog only identifies plans via their md5(FQPlanID), and will request
		// and update using that hash. If a PlanID is submitted, we'll need to look up
		// the ServiceClass scoped plan name via the passed in hash so the APB
		// will understand what to do with it, since APBs do not understand plan hashes.
		toPlanName, err = a.dao.GetPlanName(req.PlanID)
		if err != nil {
			log.Error("Could not find requested PlanID %s in plan name lookup table", req.PlanID)
			return nil, ErrorPlanNotFound
		}
	}

	// Retrieve from/to plans by name, else respond with appropriate error
	if fromPlan = spec.GetPlan(fromPlanName); fromPlan == nil {
		log.Error("The plan %s, specified for updating from on instance %s, does not exist.", fromPlanName, si.ID)
		return nil, ErrorPlanNotFound
	}
	if toPlan = spec.GetPlan(toPlanName); toPlan == nil {
		log.Error("The plan %s, specified for updating to on instance %s, does not exist.", toPlanName, si.ID)
		return nil, ErrorPlanNotFound
	}

	// If a plan transition has been requested, validate it is possible and then
	// update the service instance with the desired next plan
	if fromPlanName != toPlanName {
		log.Debug("Validating plan transition from: %s, to: %s", fromPlanName, toPlanName)
		if ok := a.isValidPlanTransition(fromPlan, toPlanName); !ok {
			log.Error("The current plan, %s, cannot be updated to the requested plan, %s.", fromPlanName, toPlanName)
			return nil, ErrorPlanUpdateNotPossible
		}

		log.Debug("Plan transition valid!")
		(*si.Parameters)[planParameterKey] = toPlanName
	} else {
		log.Debug("Plan transition NOT requested as part of update")
	}

	req.Parameters = a.validateRequestedUpdateParams(req.Parameters, toPlan, si)

	// Parameters look good, update the ServiceInstance values
	for newParamKey, newParamVal := range req.Parameters {
		(*si.Parameters)[newParamKey] = newParamVal
	}

	// We're ready to provision so save
	if err = a.dao.SetServiceInstance(instanceUUID.String(), si); err != nil {
		return nil, err
	}

	var token string

	log.Debug("Initiating update with the inputs:")
	log.Debugf("fromPlanName: [%s]", fromPlanName)
	log.Debugf("toPlanName: [%s]", toPlanName)
	log.Debugf("PreviousValues: [ %+v ]", req.PreviousValues)
	log.Debugf("ServiceInstance Parameters: [%v]", *si.Parameters)

	if async {
		log.Info("ASYNC update in progress")
		// asyncronously provision and return the token for the lastoperation
		ujob := NewUpdateJob(si)

		token, err = a.engine.StartNewJob("", ujob, UpdateTopic)
		if err != nil {
			log.Error("Failed to start new job for async update\n%s", err.Error())
			return nil, err
		}

		// HACK: there might be a delay between the first time the state in etcd
		// is set and the job was already started. But I need the token.
		a.dao.SetState(instanceUUID.String(), apb.JobState{Token: token, State: apb.StateInProgress, Method: apb.JobMethodUpdate})
	} else {
		// TODO: do we want to do synchronous updating?
		log.Info("reverting to synchronous update in progress")
		_, extCreds, err := apb.Update(si)
		if extCreds != nil {
			log.Debug("broker::Update, got ExtractedCredentials!")
			err = a.dao.SetExtractedCredentials(instanceUUID.String(), extCreds)
			if err != nil {
				log.Error("Could not persist extracted credentials")
				log.Error("%s", err.Error())
				return nil, err
			}
		}
	}

	return &UpdateResponse{Operation: token}, nil
}

func (a AnsibleBroker) isValidPlanTransition(fromPlan *apb.Plan, toPlanName string) bool {
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
	toPlan *apb.Plan,
	si *apb.ServiceInstance,
) map[string]string {
	for requestedParamKey := range reqParams {
		var pd *apb.ParameterDescriptor

		// Confirm the parameter actually exists on the plan
		if pd = toPlan.GetParameter(requestedParamKey); pd == nil {
			log.Warningf("Removing non-existent parameter %s, requested for update on instance %s, from request.", requestedParamKey, si.ID)
			delete(reqParams, requestedParamKey)
		} else if !pd.Updatable {
			log.Warningf("Removing non-updatable parameter %s, requested for update on instance %s, from request.", requestedParamKey, si.ID)
			delete(reqParams, requestedParamKey)
		}
	}
	return reqParams
}

// LastOperation - gets the last operation and status
func (a AnsibleBroker) LastOperation(instanceUUID uuid.UUID, req *LastOperationRequest,
) (*LastOperationResponse, error) {
	/*
		look up the resource in etcd the operation should match what was returned by provision
		take the status and return that.

		process:

		if async, provision: it should create a Job that calls apb.Provision. And write the output to etcd.
	*/
	log.Debugf("service_id: %s", req.ServiceID)
	log.Debugf("plan_id: %s", req.PlanID)
	log.Debugf("operation:  %s", req.Operation) // Operation is the job token id from the work_engine

	// TODO:validate the format to avoid some sort of injection hack
	jobstate, err := a.dao.GetState(instanceUUID.String(), req.Operation)
	if err != nil {
		// not sure what we do with the error if we can't find the state
		log.Error(fmt.Sprintf("problem reading job state: [%s]. error: [%v]", instanceUUID, err.Error()))
	}

	state := StateToLastOperation(jobstate.State)
	log.Debugf("state: %s", state)
	// Error can not be nil, so we're not checking for it.
	if jobstate.Error != "" {
		log.Debugf("job state has an error. Assuming that any error here is human readable. err - %v", jobstate.Error)
	}
	return &LastOperationResponse{State: state, Description: jobstate.Error}, err
}

// AddSpec - adding the spec to the catalog for local development
func (a AnsibleBroker) AddSpec(spec apb.Spec) (*CatalogResponse, error) {
	log.Debug("broker::AddSpec")
	spec.Image = spec.FQName
	addNameAndIDForSpec([]*apb.Spec{&spec}, apbPushRegName)
	log.Debugf("Generated name for pushed APB: [%s], ID: [%s]", spec.FQName, spec.ID)
	for _, p := range spec.Plans {
		a.dao.SetPlanName(p.ID, p.Name)
	}

	if err := a.dao.SetSpec(spec.ID, &spec); err != nil {
		return nil, err
	}
	apb.AddSecretsFor(&spec)
	service, err := SpecToService(&spec)
	if err != nil {
		log.Debugf("spec was not added due to issue with transformation to serivce - %v", err)
		return nil, err
	}
	metrics.SpecsLoaded(apbPushRegName, 1)
	return &CatalogResponse{Services: []Service{service}}, nil
}

// RemoveSpec - remove the spec specified from the catalog/etcd
func (a AnsibleBroker) RemoveSpec(specID string) error {
	spec, err := a.dao.GetSpec(specID)
	if client.IsKeyNotFound(err) {
		return ErrorNotFound
	}
	if err != nil {
		log.Error("Something went real bad trying to retrieve spec for deletion... - %v", err)
		return err
	}
	err = a.dao.DeleteSpec(spec.ID)
	if err != nil {
		log.Error("Something went real bad trying to delete spec... - %v", err)
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
		log.Error("Something went real bad trying to retrieve batch specs for deletion... - %v", err)
		return err
	}
	err = a.dao.BatchDeleteSpecs(specs)
	if err != nil {
		log.Error("Something went real bad trying to delete batch specs... - %v", err)
		return err
	}
	metrics.SpecsLoadedReset()
	return nil
}
