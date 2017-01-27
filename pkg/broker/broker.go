package broker

import (
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
	return nil, notImplemented
}

func (a AnsibleBroker) Update(instanceUUID uuid.UUID, req *UpdateRequest) (*UpdateResponse, error) {
	return nil, notImplemented
}

func (a AnsibleBroker) Deprovision(instanceUUID uuid.UUID) (*DeprovisionResponse, error) {
	return nil, notImplemented
}

func (a AnsibleBroker) Bind(instanceUUID uuid.UUID, bindingUUID uuid.UUID, req *BindRequest) (*BindResponse, error) {
	return nil, notImplemented
}

func (a AnsibleBroker) Unbind(instanceUUID uuid.UUID, bindingUUID uuid.UUID) error {
	return notImplemented
}
