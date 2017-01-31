package broker

import (
	"fmt"
	"github.com/fusor/ansible-service-broker/pkg/ansibleapp"
	"github.com/fusor/ansible-service-broker/pkg/dao"
	"github.com/op/go-logging"
	"github.com/pborman/uuid"
)

type Broker interface {
	Catalog() (*CatalogResponse, error)
	Provision(uuid.UUID, *ProvisionRequest) (*ProvisionResponse, error)
	Update(uuid.UUID, *UpdateRequest) (*UpdateResponse, error)
	Deprovision(uuid.UUID) (*DeprovisionResponse, error)
	Bind(uuid.UUID, uuid.UUID, *BindRequest) (*BindResponse, error)
	Unbind(uuid.UUID, uuid.UUID) error
}

type AnsibleBroker struct {
	dao      *dao.Dao
	log      *logging.Logger
	registry ansibleapp.Registry
}

func NewAnsibleBroker(
	dao *dao.Dao,
	log *logging.Logger,
	registry ansibleapp.Registry,
) (*AnsibleBroker, error) {

	broker := &AnsibleBroker{
		dao:      dao,
		log:      log,
		registry: registry,
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
	var specs []*ansibleapp.Spec

	if specs, err = a.registry.LoadSpecs(); err != nil {
		return nil, err
	}

	if err := a.dao.BatchSetSpecs(ansibleapp.NewSpecManifest(specs)); err != nil {
		return nil, err
	}

	return &BootstrapResponse{len(specs)}, nil
}

func (a AnsibleBroker) Catalog() (*CatalogResponse, error) {
	a.log.Info("AnsibleBroker::Catalog")

	var specs []*ansibleapp.Spec
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

func (a AnsibleBroker) Provision(instanceUUID uuid.UUID, req *ProvisionRequest) (*ProvisionResponse, error) {
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

	var spec *ansibleapp.Spec
	var err error

	// Retrieve requested spec
	specId := req.ServiceID.String()
	if spec, err = a.dao.GetSpec(specId); err != nil {
		// TODO: Handle unknown spec
		return nil, err
	}

	// HACK: Figure out map[string]string -> Parameters typecast, because this is terrible
	var JediMindTrick = func(in map[string]string) *ansibleapp.Parameters {
		out := make(ansibleapp.Parameters)
		for k, v := range in {
			out[k] = v
		}
		return &out
	}
	// These aren't the droids you're looking for...
	parameters := JediMindTrick(req.Parameters)

	// Build and persist record of service instance
	serviceInstance := &ansibleapp.ServiceInstance{
		Id:         instanceUUID,
		Spec:       spec,
		Parameters: parameters,
	}

	err = a.dao.SetServiceInstance(instanceUUID.String(), serviceInstance)
	if err != nil {
		return nil, err
	}

	// TODO: Async? Bring in WorkEngine.
	err = ansibleapp.Provision(spec, parameters, a.log)

	// TODO: What data needs to be sent back on a respone?
	// Not clear what dashboardURL means in an AnsibleApp context
	// Operation needs to be present if this is an async provisioning
	// 202 (Accepted), inprogress last_operation status
	// Will need to come with a "state" update in etcd on the ServiceInstance
	return &ProvisionResponse{}, nil
}

func (a AnsibleBroker) Deprovision(instanceUUID uuid.UUID) (*DeprovisionResponse, error) {
	////////////////////////////////////////////////////////////
	// Deprovision flow
	// -> Lookup bindings by instance ID; 400 if any are active, related issue:
	//    https://github.com/openservicebrokerapi/servicebroker/issues/127
	// -> Atomic deprovision and removal of service entry in etcd?
	//    * broker::Deprovision
	//    Arguments for this? What data do ansibleapps require to deprovision?
	//    Maybe just hand off a serialized ServiceInstance and let the ansibleapp
	//    decide what's important?
	//    * if noerror: delete serviceInstance entry with Dao
	////////////////////////////////////////////////////////////
	var err error
	var instance *ansibleapp.ServiceInstance
	instanceId := instanceUUID.String()

	if err = a.validateDeprovision(instanceId); err != nil {
		return nil, err
	}

	if instance, err = a.dao.GetServiceInstance(instanceId); err != nil {
		return nil, err
	}

	if err = ansibleapp.Deprovision(instance, a.log); err != nil {
		return nil, err
	}

	a.dao.DeleteServiceInstance(instanceId)

	return nil, notImplemented
}

func (a AnsibleBroker) validateDeprovision(id string) error {
	// TODO: Check if there are outstanding bindings; return typed errors indicating
	// *why* things can't be deprovisioned
	a.log.Debug(fmt.Sprintf("AnsibleBroker::validateDeprovision -> [ %s ]", id))
	return nil
}

func (a AnsibleBroker) Bind(instanceUUID uuid.UUID, bindingUUID uuid.UUID, req *BindRequest) (*BindResponse, error) {
	return nil, notImplemented
}

func (a AnsibleBroker) Unbind(instanceUUID uuid.UUID, bindingUUID uuid.UUID) error {
	return notImplemented
}

func (a AnsibleBroker) Update(instanceUUID uuid.UUID, req *UpdateRequest) (*UpdateResponse, error) {
	return nil, notImplemented
}
