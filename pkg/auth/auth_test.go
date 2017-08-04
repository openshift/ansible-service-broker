package auth

import (
	"io/ioutil"
	"os"
	"strings"
	"testing"

	logging "github.com/op/go-logging"
	ft "github.com/openshift/ansible-service-broker/pkg/fusortest"
)

var log = logging.MustGetLogger("auth")

func TestNewFusa(t *testing.T) {
	username := []byte("admin")
	password := []byte("admin")
	ioutil.WriteFile("/tmp/username", username, 0644)
	ioutil.WriteFile("/tmp/password", password, 0644)

	defer os.Remove("/tmp/username")
	defer os.Remove("/tmp/password")

	fusa, err := NewFileUserServiceAdapter("/tmp/", log)
	if err != nil {
		t.Fatal(err.Error())
	}
	adminuser, _ := fusa.FindByLogin("admin")
	ft.AssertEqual(t, adminuser.Username, "admin", "username does not match")
	ft.AssertEqual(t, adminuser.Password, "admin", "password does not match")
	ft.AssertTrue(t, fusa.ValidateUser("admin", "admin"), "validation failed")
	ft.AssertFalse(t, fusa.ValidateUser("notme", "admin"), "validation passed, expected failure")
	ft.AssertFalse(t, fusa.ValidateUser("", ""), "expected failure on empty string")
}

func TestErrorBuild(t *testing.T) {
	fusa, err := NewFileUserServiceAdapter("", log)
	if fusa != nil {
		t.Fatal("fusa is not nil")
	}
	ft.AssertNotNil(t, err, "expected an error")
	ft.AssertTrue(t, strings.Contains(err.Error(), "directory is empty,"))
}

func TestFusaError(t *testing.T) {
	_, err := NewFileUserServiceAdapter("/var/tmp", log)
	ft.AssertNotNil(t, err, "should have gotten an error")
	ft.AssertTrue(t, strings.Contains(err.Error(), "no such file or directory"), "mismatch error message")
}

func TestUser(t *testing.T) {
	user := User{Username: "admin", Password: "password"}
	ft.AssertEqual(t, user.GetType(), "user", "type doesn't match user")
	ft.AssertEqual(t, user.GetName(), user.Username, "get name and username do not match")
}
