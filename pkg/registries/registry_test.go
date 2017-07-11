package registries

import (
	"testing"

	ft "github.com/openshift/ansible-service-broker/pkg/fusortest"
)

func TestSpecLabel(t *testing.T) {
	ft.AssertEqual(t, BundleSpecLabel, "com.redhat.apb.spec", "spec label does not match dockerhub")
}
