package broker

import (
	"os"
	"testing"

	logging "github.com/op/go-logging"
	"github.com/openshift/ansible-service-broker/pkg/apb"
	ft "github.com/openshift/ansible-service-broker/pkg/fusortest"
	"github.com/pborman/uuid"
)

var log = logging.MustGetLogger("handler")

func init() {
	colorFormatter := logging.MustStringFormatter(
		"%{color}[%{time}] [%{level}] %{message}%{color:reset}",
	)
	backend := logging.NewLogBackend(os.Stdout, "", 1)
	backendFormatter := logging.NewBackendFormatter(backend, colorFormatter)
	logging.SetBackend(backend, backendFormatter)
}

func TestUpdate(t *testing.T) {
	brokerConfig := new(Config)
	brokerConfig.DevBroker = true
	brokerConfig.LaunchApbOnBind = false
	broker, _ := NewAnsibleBroker(nil, log, apb.ClusterConfig{}, nil, WorkEngine{}, *brokerConfig)
	resp, err := broker.Update(uuid.NewUUID(), nil)
	if resp != nil {
		t.Fail()
	}
	ft.AssertEqual(t, err, notImplemented, "Update must have been implemented")
}
