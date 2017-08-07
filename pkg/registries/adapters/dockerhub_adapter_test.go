package adapters

import (
	"testing"

	ft "github.com/openshift/ansible-service-broker/pkg/fusortest"
)

func TestDockerhubName(t *testing.T) {
	ft.AssertEqual(t, dockerhubName, "docker.io", "dockerhub name does not match docker.io")
}
