package auth

import (
	"testing"

	ft "github.com/openshift/ansible-service-broker/pkg/fusortest"
)

func TestNewBasicAuth(t *testing.T) {
	usa := DefaultUserServiceAdapter{}
	ba := NewBasicAuth(usa)
	ft.AssertTrue(t, ba.GetPrincipal(nil) == nil, "")
	ft.AssertEqual(t, "foo", "foo", "")
}
