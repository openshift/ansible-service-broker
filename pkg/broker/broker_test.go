package broker

import (
	"testing"

	"github.com/openshift/ansible-service-broker/pkg/apb"
	ft "github.com/openshift/ansible-service-broker/pkg/fusortest"
	"github.com/pborman/uuid"
)

func TestUpdate(t *testing.T) {
	broker, _ := NewAnsibleBroker(nil, nil, apb.ClusterConfig{}, nil, WorkEngine{})
	resp, err := broker.Update(uuid.NewUUID(), nil)
	if resp != nil {
		t.Fail()
	}
	ft.AssertEqual(t, err, notImplemented, "Update must have been implemented")
}

func TestUnbind(t *testing.T) {
	broker, _ := NewAnsibleBroker(nil, nil, apb.ClusterConfig{}, nil, WorkEngine{})
	err := broker.Unbind(uuid.NewUUID(), uuid.NewUUID())

	ft.AssertEqual(t, err, notImplemented, "Unbind must have been implemented")
}
