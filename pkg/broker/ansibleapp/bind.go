package ansibleapp

import (
	"github.com/fusor/ansible-service-broker/pkg/broker"
	"github.com/pborman/uuid"
)

func (b Broker) Bind(instanceUUID uuid.UUID, bindingUUID uuid.UUID, req *broker.BindRequest) (*broker.BindResponse, error) {
	return nil, notImplemented
}

func (b Broker) Unbind(instanceUUID uuid.UUID, bindingUUID uuid.UUID) error {
	return notImplemented
}
