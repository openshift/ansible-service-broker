package auth

import (
	"io/ioutil"
	"testing"

	ft "github.com/openshift/ansible-service-broker/pkg/fusortest"
)

func TestNewFusa(t *testing.T) {
	data := []byte("admin\nadmin")
	ioutil.WriteFile("/tmp/foo", data, 0644)

	fusa := NewFileUserServiceAdapter("/tmp/foo")
	adminuser, _ := fusa.FindByLogin("admin")
	ft.AssertEqual(t, adminuser.Username, "admin", "username does not match")
	ft.AssertEqual(t, adminuser.Password, "admin", "password does not match")
}
