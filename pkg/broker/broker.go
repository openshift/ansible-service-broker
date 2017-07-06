package broker

import (
	"errors"
	"fmt"
	"io/ioutil"
	"reflect"

	"github.com/coreos/etcd/client"
	logging "github.com/op/go-logging"
	"github.com/openshift/ansible-service-broker/pkg/apb"
	"github.com/openshift/ansible-service-broker/pkg/apb/registry"
	"github.com/openshift/ansible-service-broker/pkg/dao"
	"github.com/pborman/uuid"
	k8srestclient "k8s.io/client-go/rest"
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
)

// Broker - A broker is used to to compelete all the tasks that a broker must be able to do.
type Broker interface {
	Bootstrap() (*BootstrapResponse, error)
	Catalog() (*CatalogResponse, error)
	Provision(uuid.UUID, *ProvisionRequest, bool) (*ProvisionResponse, error)
	Update(uuid.UUID, *UpdateRequest) (*UpdateResponse, error)
	Deprovision(uuid.UUID, bool) (*DeprovisionResponse, error)
	Bind(uuid.UUID, uuid.UUID, *BindRequest) (*BindResponse, error)
	Unbind(uuid.UUID, uuid.UUID) (*UnbindResponse, error)
	LastOperation(uuid.UUID, *LastOperationRequest) (*LastOperationResponse, error)
	// TODO: consider returning a struct + error
	Recover() (string, error)
}

// Config - Configuration for the broker.
type Config struct {
	DevBroker       bool `yaml:"dev_broker"`
	LaunchApbOnBind bool `yaml:"launch_apb_on_bind"`
	Recovery        bool `yaml:"recovery"`
	OutputRequest   bool `yaml:"output_request"`
}

// DevBroker - Interface for the development broker.
type DevBroker interface {
	AddSpec(spec apb.Spec) (*CatalogResponse, error)
	RemoveSpec(specID string) error
	RemoveSpecs() error
}

// AnsibleBroker - Broker using ansible and images to interact with oc/kubernetes/etcd
type AnsibleBroker struct {
	dao           *dao.Dao
	log           *logging.Logger
	clusterConfig apb.ClusterConfig
	registry      []registry.Registry
	engine        *WorkEngine
	brokerConfig  Config
}

// NewAnsibleBroker - Creates a new ansible broker
func NewAnsibleBroker(dao *dao.Dao, log *logging.Logger, clusterConfig apb.ClusterConfig,
	registry []registry.Registry, engine WorkEngine, brokerConfig Config,
) (*AnsibleBroker, error) {
	broker := &AnsibleBroker{
		dao:           dao,
		log:           log,
		clusterConfig: clusterConfig,
		registry:      registry,
		engine:        &engine,
		brokerConfig:  brokerConfig,
	}

	err := broker.Login()
	if err != nil {
		return broker, err
	}

	return broker, nil
}

func (a AnsibleBroker) getServiceInstance(instanceUUID uuid.UUID) (*apb.ServiceInstance, error) {
	instance, err := a.dao.GetServiceInstance(instanceUUID.String())
	if err != nil {
		if client.IsKeyNotFound(err) {
			a.log.Errorf("Could not find a service instance in dao - %v", err)
			return nil, ErrorNotFound
		}
		a.log.Error("Couldn't find a service instance: ", err)
		return nil, err
	}
	return instance, nil

}

//Login - Will login the openshift user.
func (a AnsibleBroker) Login() error {
	config, err := a.getLoginDetails()
	if err != nil {
		return err
	}

	if config.CAFile != "" {
		err = ocLogin(a.log, config.Host,
			"--token", config.BearerToken,
			"--certificate-authority", config.CAFile,
		)
	} else {
		err = ocLogin(a.log, config.Host,
			"--token", config.BearerToken,
			"--insecure-skip-tls-verify=false",
		)
	}

	return err
}

type loginDetails struct {
	Host        string
	CAFile      string
	BearerToken string
}

func (a AnsibleBroker) getLoginDetails() (loginDetails, error) {
	config := loginDetails{}

	// If overrides are passed into the config map, Host and BearerTokenFile
	// values *must* be provided, else we'll default to the k8srestclient details
	if a.clusterConfig.Host != "" && a.clusterConfig.BearerTokenFile != "" {
		a.log.Info("ClusterConfig Host and BearerToken provided, preferring configurable overrides")
		a.log.Info("Host: [ %s ]", a.clusterConfig.Host)
		a.log.Info("BearerTokenFile: [ %s ]", a.clusterConfig.BearerTokenFile)

		token, err := ioutil.ReadFile(a.clusterConfig.BearerTokenFile)
		if err != nil {
			return config, err
		}

		config.Host = a.clusterConfig.Host
		config.BearerToken = string(token)
		config.CAFile = a.clusterConfig.CAFile
	} else {
		a.log.Info("No cluster credential overrides provided, using k8s InClusterConfig")
		k8sConfig, err := k8srestclient.InClusterConfig()
		if err != nil {
			a.log.Error("Cluster host & bearer_token_file missing from config, and failed to retrieve InClusterConfig")
			a.log.Error("Be sure you have configured a cluster host and service account credentials if" +
				" you are running the broker outside of a cluster Pod")
			return config, err
		}

		config.Host = k8sConfig.Host
		config.CAFile = k8sConfig.CAFile
		config.BearerToken = k8sConfig.BearerToken
	}

	return config, nil
}

// Bootstrap - Loads all known specs from a registry into local storage for reference
// Potentially a large download; on the order of 10s of thousands
// TODO: Response here? Async?
// TODO: How do we handle a large amount of data on this side as well? Pagination?
func (a AnsibleBroker) Bootstrap() (*BootstrapResponse, error) {
	a.log.Info("AnsibleBroker::Bootstrap")
	var err error
	var specs []*apb.Spec
	var imageCount int

	//Remove all specs that have been saved.
	dir := "/spec"
	specs, err = a.dao.BatchGetSpecs(dir)
	if err != nil {
		a.log.Error("Something went real bad trying to retrieve batch specs for deletion... - %v", err)
		return nil, err
	}
	err = a.dao.BatchDeleteSpecs(specs)
	if err != nil {
		a.log.Error("Something went real bad trying to delete batch specs... - %v", err)
		return nil, err
	}
	specs = []*apb.Spec{}

	//Load Specs for each registry
	registryErrors := []error{}
	for _, r := range a.registry {
		s, i, err := r.LoadSpecs()
		if err != nil && r.Fail(err) {
			a.log.Errorf("registry caused bootstrap failure - %v", err)
			return nil, err
		}
		if err != nil {
			registryErrors = append(registryErrors, err)
		}
		imageCount += i
		specs = append(specs, s...)
	}
	if len(registryErrors) == len(a.registry) {
		return nil, errors.New("all registries failed on bootstrap")
	}
	specIDToAdded := map[string]*apb.Spec{}
	for _, s := range specs {
		if val, ok := specIDToAdded[s.ID]; ok {
			a.log.Warningf("spec id collision - %v is overwriting %v", s.Name, val.Name)
		}
		specIDToAdded[s.ID] = s
	}

	if err := a.dao.BatchSetSpecs(specIDToAdded); err != nil {
		return nil, err
	}
	a.log.Debugf("specs -> %v", specs)

	return &BootstrapResponse{SpecCount: len(specs), ImageCount: imageCount}, nil
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
			a.log.Info("No jobs to recover")
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

			a.log.Info(fmt.Sprintf("No podname. Attempting to restart job: %s", instanceID))

			a.log.Debug(fmt.Sprintf("%v", instance))

			// Handle bad write of service instance
			if instance.Spec == nil || instance.Parameters == nil {
				a.dao.SetState(instanceID, apb.JobState{Token: rs.State.Token, State: apb.StateFailed})
				a.dao.DeleteServiceInstance(instance.ID.String())
				a.log.Warning(fmt.Sprintf("incomplete ServiceInstance [%s] record, marking job as failed", instance.ID))
				// skip to the next item
				continue
			}

			pjob := NewProvisionJob(instance, a.clusterConfig, a.log)

			// Need to use the same token as before, since that's what the
			// catalog will try to ping.
			a.engine.StartNewJob(rs.State.Token, pjob)

			// HACK: there might be a delay between the first time the state in etcd
			// is set and the job was already started. But I need the token.
			a.dao.SetState(instanceID, apb.JobState{Token: rs.State.Token, State: apb.StateInProgress})
		} else {
			// YES, we have a podname
			a.log.Info(fmt.Sprintf("We have a pod to recover: %s", rs.State.Podname))

			// did the pod finish?
			extCreds, extErr := apb.ExtractCredentials(rs.State.Podname, instance.Context.Namespace, a.log)

			// NO, pod failed.
			// TODO: do we restart the job or mark it as failed?
			if extErr != nil {
				a.log.Error("broker::Recover error occurred.")
				a.log.Error("%s", extErr.Error())
				return "", extErr
			}

			// YES, pod finished we have creds
			if extCreds != nil {
				a.log.Debug("broker::Recover, got ExtractedCredentials!")
				a.dao.SetState(instanceID, apb.JobState{Token: rs.State.Token,
					State: apb.StateSucceeded, Podname: rs.State.Podname})
				err = a.dao.SetExtractedCredentials(instanceID, extCreds)
				if err != nil {
					a.log.Error("Could not persist extracted credentials")
					a.log.Error("%s", err.Error())
					return "", err
				}
			}
		}
	}

	// if no pods, do we restart? or just return failed?

	//binding

	a.log.Info("Recovery complete")
	return "recover called", nil
}

// Catalog - returns the catalog of services defined
func (a AnsibleBroker) Catalog() (*CatalogResponse, error) {
	a.log.Info("AnsibleBroker::Catalog")

	var specs []*apb.Spec
	var err error
	var services []Service
	dir := "/spec"

	if specs, err = a.dao.BatchGetSpecs(dir); err != nil {
		a.log.Error("Something went real bad trying to retrieve batch specs...")
		return nil, err
	}

	services = make([]Service, len(specs))
	for i, spec := range specs {
		services[i] = SpecToService(spec)
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
	// Unclear how this is relevant

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

	// Retrieve requested spec
	specID := req.ServiceID.String()
	if spec, err = a.dao.GetSpec(specID); err != nil {
		// etcd return not found i.e. code 100
		if client.IsKeyNotFound(err) {
			return nil, ErrorNotFound
		}
		// otherwise unknown error bubble it up
		return nil, err
	}

	context := &req.Context
	parameters := &req.Parameters

	// Build and persist record of service instance
	serviceInstance := &apb.ServiceInstance{
		ID:         instanceUUID,
		Spec:       spec,
		Context:    context,
		Parameters: parameters,
	}

	// Verify we're not reprovisioning the same instance
	// if err is nil, there is an instance. Let's compare it to the instance
	// we're being asked to provision.
	//
	// if err is not nil, we will just bubble that up

	if si, err := a.dao.GetServiceInstance(instanceUUID.String()); err == nil {
		//This will use the package to make sure that if the type is changed away from []byte it can still be evaluated.
		if uuid.Equal(si.ID, serviceInstance.ID) {
			if reflect.DeepEqual(si.Parameters, serviceInstance.Parameters) {
				a.log.Debug("already have this instance returning 200")
				return &ProvisionResponse{}, ErrorAlreadyProvisioned
			}
			a.log.Info("we have a duplicate instance with parameters that differ, returning 409 conflict")
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
		a.log.Info("ASYNC provisioning in progress")
		// asyncronously provision and return the token for the lastoperation
		pjob := NewProvisionJob(serviceInstance, a.clusterConfig, a.log)

		token = a.engine.StartNewJob("", pjob)

		// HACK: there might be a delay between the first time the state in etcd
		// is set and the job was already started. But I need the token.
		a.dao.SetState(instanceUUID.String(), apb.JobState{Token: token, State: apb.StateInProgress})
	} else {
		// TODO: do we want to do synchronous provisioning?
		a.log.Info("reverting to synchronous provisioning in progress")
		podName, extCreds, err := apb.Provision(serviceInstance, a.clusterConfig, a.log)

		sm := apb.NewServiceAccountManager(a.log)
		a.log.Info("Destroying APB sandbox...")
		sm.DestroyApbSandbox(podName, context.Namespace)
		if err != nil {
			a.log.Error("broker::Provision error occurred.")
			a.log.Error("%s", err.Error())
			return nil, err
		}

		// TODO: do we need podname for synchronous provisions?
		extCreds, extErr := apb.ExtractCredentials(podName, context.Namespace, a.log)
		if extErr != nil {
			a.log.Error("broker::Provision error occurred.")
			a.log.Error("%s", extErr.Error())
			return nil, extErr
		}

		if extCreds != nil {
			a.log.Debug("broker::Provision, got ExtractedCredentials!")
			err = a.dao.SetExtractedCredentials(instanceUUID.String(), extCreds)
			if err != nil {
				a.log.Error("Could not persist extracted credentials")
				a.log.Error("%s", err.Error())
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
func (a AnsibleBroker) Deprovision(instanceUUID uuid.UUID, async bool,
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
	instance, err := a.getServiceInstance(instanceUUID)
	if err != nil {
		return nil, err
	}

	if err := a.validateDeprovision(instance); err != nil {
		return nil, err
	}

	var token string

	if async {
		a.log.Info("ASYNC deprovision in progress")
		// asynchronously provision and return the token for the lastoperation
		dpjob := NewDeprovisionJob(instance, a.clusterConfig, a.dao, a.log)

		token = a.engine.StartNewJob("", dpjob)

		// HACK: there might be a delay between the first time the state in etcd
		// is set and the job was already started. But I need the token.
		a.dao.SetState(instanceUUID.String(), apb.JobState{Token: token, State: apb.StateInProgress})
		return &DeprovisionResponse{Operation: token}, nil
	}
	// TODO: do we want to do synchronous deprovisioning?
	a.log.Info("Synchronous deprovision in progress")
	podName, err := apb.Deprovision(instance, a.clusterConfig, a.log)
	err = cleanupDeprovision(err, podName, instance, a.dao, a.log)
	if err != nil {
		return nil, err
	}
	return &DeprovisionResponse{}, nil
}

func cleanupDeprovision(err error, podName string, instance *apb.ServiceInstance, dao *dao.Dao,
	log *logging.Logger,
) error {
	instanceID := instance.ID.String()
	sm := apb.NewServiceAccountManager(log)
	log.Info("Destroying APB sandbox...")
	sm.DestroyApbSandbox(podName, instance.Context.Namespace)

	// bubble up error.
	if err != nil {
		log.Error("error from deprovision - %#v", err)
		return err
	}

	if err := dao.DeleteExtractedCredentials(instanceID); err != nil {
		log.Error("ERROR - failed to delete extracted credentials - %#v", err)
	}

	if err := dao.DeleteServiceInstance(instanceID); err != nil {
		log.Error("ERROR - failed to delete service instance - %#v", err)
		return err
	}
	return nil

}

func (a AnsibleBroker) validateDeprovision(instance *apb.ServiceInstance) error {
	// -> Lookup bindings by instance ID; 400 if any are active, related issue:
	//    https://github.com/openservicebrokerapi/servicebroker/issues/127
	if len(instance.BindingIDs) > 0 {
		a.log.Debugf("Found bindings with ids: %v", instance.BindingIDs)
		return ErrorBindingExists
	}
	// TODO WHAT TO DO IF ASYNC BIND/PROVISION IN PROGRESS
	return nil
}

// Bind - will create a binding between a service.
func (a AnsibleBroker) Bind(instanceUUID uuid.UUID, bindingUUID uuid.UUID, req *BindRequest,
) (*BindResponse, error) {
	// binding_id is the id of the binding.
	// the instanceUUID is the previously provisioned service id.
	//
	// See if the service instance still exists, if not send back a badrequest.

	instance, err := a.getServiceInstance(instanceUUID)
	if err != nil {
		return nil, err
	}

	// GET SERVICE get provision parameters

	// build bind parameters args:
	// {
	//     provision_params: {} same as what was stored in etcd
	//	   bind_params: {}
	// }
	// asbcli passes in user: aone, which bind passes to apb
	params := make(apb.Parameters)
	if instance.Parameters != nil {
		params["provision_params"] = *instance.Parameters
	}
	params["bind_params"] = req.Parameters

	//
	// Create a BindingInstance with a reference to the serviceinstance.
	//

	bindingInstance := &apb.BindInstance{
		ID:         bindingUUID,
		ServiceID:  instanceUUID,
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
	if bi, err := a.dao.GetBindInstance(bindingUUID.String()); err == nil {
		if uuid.Equal(bi.ID, bindingInstance.ID) {
			if reflect.DeepEqual(bi.Parameters, bindingInstance.Parameters) {
				a.log.Debug("already have this binding instance, returning 200")
				return &BindResponse{}, ErrorAlreadyProvisioned
			}

			// parameters are different
			a.log.Info("duplicate binding instance diff params, returning 409 conflict")
			return nil, ErrorDuplicate
		}
	}

	if err := a.dao.SetBindInstance(bindingUUID.String(), bindingInstance); err != nil {
		return nil, err
	}

	/*
		NOTE:

		type BindResponse struct {
		    Credentials     map[string]interface{} `json:"credentials,omitempty"`
		    SyslogDrainURL  string                 `json:"syslog_drain_url,omitempty"`
		    RouteServiceURL string                 `json:"route_service_url,omitempty"`
		    VolumeMounts    []interface{}          `json:"volume_mounts,omitempty"`
		}
	*/

	// NOTE: Design here is very WIP
	// Potentially have data from provision stashed away, and bind may also
	// produce new binding data. Take both sets and merge?
	provExtCreds, err := a.dao.GetExtractedCredentials(instanceUUID.String())
	if err != nil {
		a.log.Debug("provExtCreds a miss!")
		a.log.Debug("%s", err.Error())
	} else {
		a.log.Debug("Got provExtCreds hit!")
		a.log.Debug("%+v", provExtCreds)
	}

	// NOTE: We are currently disabling running an APB on bind via 'LaunchApbOnBind'
	// of the broker config, due to lack of async support of bind in Open Service Broker API
	// Currently, the 'launchapbonbind' is set to false in the 'config' ConfigMap
	bindExtCreds := &apb.ExtractedCredentials{Credentials: make(map[string]interface{})}
	var podName string
	if a.brokerConfig.LaunchApbOnBind {
		a.log.Info("Broker configured to run APB bind")
		a.log.Info("Starting APB bind...")
		podName, bindExtCreds, err = apb.Bind(instance, &params, a.clusterConfig, a.log)

		sm := apb.NewServiceAccountManager(a.log)
		a.log.Info("Destroying APB sandbox...")
		sm.DestroyApbSandbox(podName, instance.Context.Namespace)

		if err != nil {
			return nil, err
		}
	} else {
		a.log.Warning("Broker configured to *NOT* launch and run APB bind")
	}
	instance.AddBinding(bindingUUID)
	if err := a.dao.SetServiceInstance(instanceUUID.String(), instance); err != nil {
		return nil, err
	}
	// Can't bind to anything if we have nothing to return to the catalog
	if provExtCreds == nil && bindExtCreds == nil {
		a.log.Error("No extracted credentials found from provision or bind")
		a.log.Error("Instance ID: %s", instanceUUID.String())
		return nil, errors.New("No credentials available")
	}

	returnCreds := mergeCredentials(provExtCreds, bindExtCreds)
	// TODO: Insert merged credentials into etcd? Separate into bind/provision
	// so none are overwritten?

	return &BindResponse{Credentials: returnCreds}, nil
}

func mergeCredentials(provExtCreds *apb.ExtractedCredentials,
	bindExtCreds *apb.ExtractedCredentials,
) map[string]interface{} {
	// TODO: Implement, need to handle case where either are empty
	return provExtCreds.Credentials
}

// Unbind - unbind a services previous binding
func (a AnsibleBroker) Unbind(instanceUUID uuid.UUID, bindingUUID uuid.UUID,
) (*UnbindResponse, error) {
	if _, err := a.dao.GetBindInstance(bindingUUID.String()); err != nil {
		return nil, ErrorNotFound
	}

	serviceInstance, err := a.getServiceInstance(instanceUUID)
	if err != nil {
		a.log.Debugf("Service instance with id %s does not exist", instanceUUID.String())
	}

	err = apb.Unbind(serviceInstance, a.clusterConfig, a.log)
	if err != nil {
		return nil, err
	}

	err = a.dao.DeleteBindInstance(bindingUUID.String())
	if err != nil {
		return nil, err
	}

	serviceInstance.RemoveBinding(bindingUUID)
	err = a.dao.SetServiceInstance(instanceUUID.String(), serviceInstance)
	if err != nil {
		return nil, err
	}

	return &UnbindResponse{}, nil
}

// Update - update a service NOTE: not implemented
func (a AnsibleBroker) Update(instanceUUID uuid.UUID, req *UpdateRequest,
) (*UpdateResponse, error) {
	return nil, notImplemented
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
	a.log.Debug(fmt.Sprintf("service_id: %s", req.ServiceID.String())) // optional
	a.log.Debug(fmt.Sprintf("plan_id: %s", req.PlanID.String()))       // optional
	a.log.Debug(fmt.Sprintf("operation:  %s", req.Operation))          // this is provided with the provision. task id from the work_engine

	// TODO:validate the format to avoid some sort of injection hack
	jobstate, err := a.dao.GetState(instanceUUID.String(), req.Operation)
	if err != nil {
		// not sure what we do with the error if we can't find the state
		a.log.Error(fmt.Sprintf("problem reading job state: [%s]. error: [%v]", instanceUUID, err.Error()))
	}

	state := StateToLastOperation(jobstate.State)
	return &LastOperationResponse{State: state, Description: ""}, err
}

//AddSpec - adding the spec to the catalog for local development
func (a AnsibleBroker) AddSpec(spec apb.Spec) (*CatalogResponse, error) {
	if err := a.dao.SetSpec(spec.ID, &spec); err != nil {
		return nil, err
	}
	service := SpecToService(&spec)
	return &CatalogResponse{Services: []Service{service}}, nil
}

// RemoveSpec - remove the spec specified from the catalog/etcd
func (a AnsibleBroker) RemoveSpec(specID string) error {
	spec, err := a.dao.GetSpec(specID)
	if client.IsKeyNotFound(err) {
		return ErrorNotFound
	}
	if err != nil {
		a.log.Error("Something went real bad trying to retrieve spec for deletion... - %v", err)
		return err
	}
	err = a.dao.DeleteSpec(spec.ID)
	if err != nil {
		a.log.Error("Something went real bad trying to delete spec... - %v", err)
		return err
	}
	return nil
}

// RemoveSpecs - remove all the specs from the catalog/etcd
func (a AnsibleBroker) RemoveSpecs() error {
	dir := "/spec"
	specs, err := a.dao.BatchGetSpecs(dir)
	if err != nil {
		a.log.Error("Something went real bad trying to retrieve batch specs for deletion... - %v", err)
		return err
	}
	err = a.dao.BatchDeleteSpecs(specs)
	if err != nil {
		a.log.Error("Something went real bad trying to delete batch specs... - %v", err)
		return err
	}
	return nil
}

func ocLogin(log *logging.Logger, args ...string) error {
	log.Debug("Logging into openshift...")

	fullArgs := append([]string{"login"}, args...)

	output, err := apb.RunCommand("oc", fullArgs...)
	log.Debug("Login output:")
	log.Debug(string(output))

	if err != nil {
		log.Debug(string(output))
		return err
	}
	return nil
}
