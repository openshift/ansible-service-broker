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

func TestAddNameAndIDForSpecStripsTailingDash(t *testing.T) {
	spec1 := apb.Spec{Image: "1234567890123456789012345678901234567890-"}
	spec2 := apb.Spec{Image: "org/hello-world-apb"}
	spcs := []*apb.Spec{&spec1, &spec2}
	addNameAndIDForSpec(spcs, "h")
	ft.AssertEqual(t, "h-1234567890123456789012345678901234567890", spcs[0].FQName)
	ft.AssertEqual(t, "h-org-hello-world-apb", spcs[1].FQName)
}
