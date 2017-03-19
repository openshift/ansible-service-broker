package broker

import (
	"testing"

	"github.com/fusor/ansible-service-broker/pkg/ansibleapp"
	ft "github.com/fusor/ansible-service-broker/pkg/fusortest"
	"github.com/pborman/uuid"
)

func TestUpdate(t *testing.T) {
	broker, _ := NewAnsibleBroker(nil, nil, ansibleapp.ClusterConfig{}, nil)
	resp, err := broker.Update(uuid.NewUUID(), nil)
	if resp != nil {
		t.Fail()
	}
	ft.AssertEqual(t, err, notImplemented, "Update must have been implemented")
}

func TestUnbind(t *testing.T) {
	broker, _ := NewAnsibleBroker(nil, nil, ansibleapp.ClusterConfig{}, nil)
	err := broker.Unbind(uuid.NewUUID(), uuid.NewUUID())

	ft.AssertEqual(t, err, notImplemented, "Unbind must have been implemented")
}

/*
need a way to mock out the logger.
func TestValidateDeprovision(t *testing.T) {
	broker, _ := NewAnsibleBroker(nil, nil, ansibleapp.ClusterConfig{}, nil)
	err := broker.validateDeprovision(uuid.New())
	if err != nil {
		t.Fail()
	}
}
*/
