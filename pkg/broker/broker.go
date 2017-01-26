package broker

import (
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
}

func NewAnsibleBroker() (*AnsibleBroker, error) {
	return &AnsibleBroker{}, nil
}

func (a AnsibleBroker) Catalog() (*CatalogResponse, error) {
}

func (b Broker) Provision(instanceUUID uuid.UUID, req *ProvisionRequest) (*ProvisionResponse, error) {
	return nil, notImplemented
}

func (b Broker) Update(instanceUUID uuid.UUID, req *UpdateRequest) (*UpdateResponse, error) {
	return nil, notImplemented
}

func (b Broker) Deprovision(instanceUUID uuid.UUID) (*DeprovisionResponse, error) {
	return nil, notImplemented
}

func (b Broker) Bind(instanceUUID uuid.UUID, bindingUUID uuid.UUID, req *BindRequest) (*BindResponse, error) {
	return nil, notImplemented
}

func (b Broker) Unbind(instanceUUID uuid.UUID, bindingUUID uuid.UUID) error {
	return notImplemented
}
