package broker

import (
	"errors"
	"fmt"
	"reflect"
	"strings"

	"github.com/openshift/ansible-service-broker/pkg/apb"
	"github.com/openshift/ansible-service-broker/pkg/dao"
	logging "github.com/op/go-logging"
	"github.com/pborman/uuid"
)

var (
	ErrorAlreadyProvisioned = errors.New("already provisioned")
	ErrorDuplicate          = errors.New("duplicate instance")
	ErrorNotFound           = errors.New("not found")
)

type Broker interface {
	Bootstrap() (*BootstrapResponse, error)
	Catalog() (*CatalogResponse, error)
	Provision(uuid.UUID, *ProvisionRequest, bool) (*ProvisionResponse, error)
	Update(uuid.UUID, *UpdateRequest) (*UpdateResponse, error)
	Deprovision(uuid.UUID) (*DeprovisionResponse, error)
	Bind(uuid.UUID, uuid.UUID, *BindRequest) (*BindResponse, error)
	Unbind(uuid.UUID, uuid.UUID) error
	LastOperation(uuid.UUID, *LastOperationRequest) (*LastOperationResponse, error)
}

type AnsibleBroker struct {
	dao           *dao.Dao
	log           *logging.Logger
	clusterConfig apb.ClusterConfig
	registry      apb.Registry
	engine        *WorkEngine
}

func NewAnsibleBroker(
	dao *dao.Dao,
	log *logging.Logger,
	clusterConfig apb.ClusterConfig,
	registry apb.Registry,
	engine WorkEngine,
) (*AnsibleBroker, error) {

	broker := &AnsibleBroker{
		dao:           dao,
		log:           log,
		clusterConfig: clusterConfig,
		registry:      registry,
		engine:        &engine,
	}

	return broker, nil
}

// Loads all known specs from a registry into local storage for reference
// Potentially a large download; on the order of 10s of thousands
// TODO: Response here? Async?
// TODO: How do we handle a large amount of data on this side as well? Pagination?
func (a AnsibleBroker) Bootstrap() (*BootstrapResponse, error) {
	a.log.Info("AnsibleBroker::Bootstrap")
	var err error
	var specs []*apb.Spec

	if specs, err = a.registry.LoadSpecs(); err != nil {
		return nil, err
	}

	if err := a.dao.BatchSetSpecs(apb.NewSpecManifest(specs)); err != nil {
		return nil, err
	}

	return &BootstrapResponse{len(specs)}, nil
}

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

func (a AnsibleBroker) Provision(instanceUUID uuid.UUID, req *ProvisionRequest, async bool) (*ProvisionResponse, error) {
	////////////////////////////////////////////////////////////
	//type ProvisionRequest struct {

	//-> OrganizationID    uuid.UUID
	//-> SpaceID           uuid.UUID
	// Used for determining where this service should be provisioned. Analagous to
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
	specId := req.ServiceID.String()
	if spec, err = a.dao.GetSpec(specId); err != nil {
		// etcd return not found i.e. code 100
		if strings.HasPrefix(err.Error(), "100") {
			return nil, ErrorNotFound
		}
		// otherwise unknown error bubble it up
		return nil, err
	}

	parameters := &req.Parameters

	// Build and persist record of service instance
	serviceInstance := &apb.ServiceInstance{
		Id:         instanceUUID,
		Spec:       spec,
		Parameters: parameters,
	}

	// Verify we're not reprovisioning the same instance
	// if err is nil, there is an instance. Let's compare it to the instance
	// we're being asked to provision.
	//
	// if err is not nil, we will just bubble that up
	if si, err := a.dao.GetServiceInstance(instanceUUID.String()); err == nil {
		if si.Id.String() == serviceInstance.Id.String() &&
			reflect.DeepEqual(si.Parameters, serviceInstance.Parameters) {

			a.log.Debug("already have this instance returning 200")
			return &ProvisionResponse{}, ErrorAlreadyProvisioned
		} else if si.Id.String() == serviceInstance.Id.String() &&
			!reflect.DeepEqual(si.Parameters, serviceInstance.Parameters) {

			// TODO: remove these debug statements at some point
			a.log.Debug("Existing parameters")
			for k, v := range map[string]interface{}(*si.Parameters) {
				a.log.Debug("%s = %s", k, v)
			}
			a.log.Debug("Incoming parameters")
			for k, v := range map[string]interface{}(*serviceInstance.Parameters) {
				a.log.Debug(fmt.Sprintf("%s = %s", k, v))
			}

			a.log.Info("we have a duplicate instance with identical parameters, returning 409 conflict")
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
		pjob := NewProvisionJob(instanceUUID, spec, parameters, a.clusterConfig, a.log)

		token = a.engine.StartNewJob(pjob)

		// HACK: there might be a delay between the first time the state in etcd
		// is set and the job was already started. But I need the token.
		a.dao.SetState(instanceUUID.String(), apb.JobState{Token: token, State: apb.StateInProgress})
	} else {
		// TODO: do we want to do synchronous provisioning?
		a.log.Info("reverting to synchronous provisioning in progress")
		extCreds, err := apb.Provision(spec, parameters, a.clusterConfig, a.log)
		if err != nil {
			a.log.Error("broker::Provision error occurred.")
			a.log.Error("%s", err.Error())
			return nil, err
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

	// TODO: What data needs to be sent back on a respone?
	// Not clear what dashboardURL means in an AnsibleApp context
	// operation should be the task id from the work_engine
	return &ProvisionResponse{Operation: token}, nil
}

func (a AnsibleBroker) Deprovision(instanceUUID uuid.UUID) (*DeprovisionResponse, error) {
	////////////////////////////////////////////////////////////
	// Deprovision flow
	// -> Lookup bindings by instance ID; 400 if any are active, related issue:
	//    https://github.com/openservicebrokerapi/servicebroker/issues/127
	// -> Atomic deprovision and removal of service entry in etcd?
	//    * broker::Deprovision
	//    Arguments for this? What data do apbs require to deprovision?
	//    Maybe just hand off a serialized ServiceInstance and let the apb
	//    decide what's important?
	//    * if noerror: delete serviceInstance entry with Dao
	////////////////////////////////////////////////////////////
	var err error
	var instance *apb.ServiceInstance
	instanceId := instanceUUID.String()

	if err = a.validateDeprovision(instanceId); err != nil {
		return nil, err
	}

	if instance, err = a.dao.GetServiceInstance(instanceId); err != nil {
		return nil, err
	}

	if err = apb.Deprovision(instance, a.log); err != nil {
		return nil, err
	}

	a.dao.DeleteServiceInstance(instanceId)

	return &DeprovisionResponse{Operation: "successful"}, nil
}

func (a AnsibleBroker) validateDeprovision(id string) error {
	// TODO: Check if there are outstanding bindings; return typed errors indicating
	// *why* things can't be deprovisioned
	a.log.Debug(fmt.Sprintf("AnsibleBroker::validateDeprovision -> [ %s ]", id))
	return nil
}

func (a AnsibleBroker) Bind(instanceUUID uuid.UUID, bindingUUID uuid.UUID, req *BindRequest) (*BindResponse, error) {

	// binding_id is the id of the binding.
	// the instanceUUID is the previously provisioned service id.
	//
	// See if the service instance still exists, if not send back a badrequest.

	instance, err := a.dao.GetServiceInstance(instanceUUID.String())
	if err != nil {
		a.log.Error("Couldn't find a service instance: ", err)
		// TODO: need to figure out how find out if an instance exists or not
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
	params["provision_params"] = *instance.Parameters
	params["bind_params"] = req.Parameters

	//
	// Create a BindingInstance with a reference to the serviceinstance.
	//

	bindingInstance := &apb.BindInstance{
		Id:         bindingUUID,
		ServiceId:  instanceUUID,
		Parameters: &params,
	}

	// if binding instance exists, and the parameters are the same return: 200.
	// if binding instance exists, and the parameters are different return: 409.
	//
	// return 201 when we're done.
	//
	// once we create the binding instance, we call apb.Bind

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

	////////////////////////////////////////////////////////////
	// HACK: NOOP the broker pod, we expect to find the credentials from
	// the provision run.
	//bindExtCreds, err := apb.Bind(instance, &params, a.clusterConfig, a.log)
	//if err != nil {
	//return nil, err
	//}
	a.log.Notice("NOOP: Bind run")
	var bindExtCreds *apb.ExtractedCredentials = nil
	err = nil
	////////////////////////////////////////////////////////////

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

func mergeCredentials(
	provExtCreds *apb.ExtractedCredentials, bindExtCreds *apb.ExtractedCredentials,
) map[string]interface{} {
	// TODO: Implement, need to handle case where either are empty
	return provExtCreds.Credentials
}

func (a AnsibleBroker) Unbind(instanceUUID uuid.UUID, bindingUUID uuid.UUID) error {
	return notImplemented
}

func (a AnsibleBroker) Update(instanceUUID uuid.UUID, req *UpdateRequest) (*UpdateResponse, error) {
	return nil, notImplemented
}

func (a AnsibleBroker) LastOperation(instanceUUID uuid.UUID, req *LastOperationRequest) (*LastOperationResponse, error) {
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
