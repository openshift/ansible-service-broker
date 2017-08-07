package adapters

import (
	"testing"

	ft "github.com/openshift/ansible-service-broker/pkg/fusortest"
)

func TestDockerhubName(t *testing.T) {
	dha := DockerHubAdapter{}
	ft.AssertEqual(t, dha.RegistryName(), "docker.io", "dockerhub name does not match docker.io")
}
