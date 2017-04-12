package broker

import (
	"os"
	"path"
	"testing"

	"github.com/fusor/ansible-service-broker/pkg/apb"
	ft "github.com/fusor/ansible-service-broker/pkg/fusortest"
	"github.com/pborman/uuid"
)

func TestSpecToService(t *testing.T) {

	param := []*apb.ParameterDescriptor{
		&apb.ParameterDescriptor{
			Name:        "hostport",
			Description: "The host TCP port as the external end point",
			Default:     float64(9001),
			Type:        "foo",
			Required:    true}}

	spec := apb.Spec{
		Id:          "50eb5637-6ffe-480d-a52e-a7e603a50fca",
		Name:        "testspec",
		Bindable:    false,
		Description: "test spec to be converted",
		Async:       "unsupported",
		Parameters:  param}

	descriptors := make(map[string]interface{})
	descriptors["parameters"] = param

	expectedsvc := Service{
		ID:          uuid.Parse("50eb5637-6ffe-480d-a52e-a7e603a50fca"),
		Name:        "testspec",
		Description: "test spec to be converted",
		Bindable:    false,
		Plans:       nil,
		Metadata:    descriptors,
	}
	svc := SpecToService(&spec)
	ft.AssertEqual(t, svc.Name, expectedsvc.Name, "name is not equal")
	ft.AssertEqual(t, svc.Description, expectedsvc.Description, "description is not equal")
	ft.AssertEqual(t, svc.Bindable, expectedsvc.Bindable, "bindable wrong")
}

func TestProjectRoot(t *testing.T) {
	gopath := os.Getenv("GOPATH")
	rootpath := path.Join(gopath, "src/github.com/fusor/ansible-service-broker")
	ft.AssertEqual(t, ProjectRoot(), rootpath, "paths not equal")
}

func TestState(t *testing.T) {
	state := StateToLastOperation(apb.StateInProgress)
	ft.AssertEqual(t, state, LastOperationStateInProgress, "should be in progress")
	state = StateToLastOperation(apb.StateSucceeded)
	ft.AssertEqual(t, state, LastOperationStateSucceeded, "should be succeeded")
	state = StateToLastOperation(apb.StateFailed)
	ft.AssertEqual(t, state, LastOperationStateFailed, "should be failed")
}
