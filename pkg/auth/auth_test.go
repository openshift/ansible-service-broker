package auth

import (
	"io/ioutil"
	"net/http"
	"net/http/httptest"
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

// auth handler tests
func TestHandlerAuthorized(t *testing.T) {
	handlerCalled := false
	testhandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		handlerCalled = true
	})

	ba := NewBasicAuth(
		MockUserServiceAdapter{userdb: map[string]string{"admin": "password"}}, log)

	authhandler := Handler(testhandler, []Provider{ba}, log)

	w := httptest.NewRecorder()

	r, err := http.NewRequest(http.MethodPost, "/v2/bootstrap", nil)
	if err != nil {
		t.Fatal(err.Error())
	}
	r.SetBasicAuth("admin", "password")

	authhandler.ServeHTTP(w, r)

	ft.AssertTrue(t, handlerCalled, "handler not called")
	ft.AssertEqual(t, w.Code, http.StatusOK)
}

func TestHandlerRejected(t *testing.T) {
	handlerCalled := false
	testhandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		handlerCalled = true
	})

	ba := NewBasicAuth(
		MockUserServiceAdapter{userdb: map[string]string{"admin": "password"}}, log)

	authhandler := Handler(testhandler, []Provider{ba}, log)

	w := httptest.NewRecorder()

	r, err := http.NewRequest(http.MethodPost, "/v2/bootstrap", nil)
	if err != nil {
		t.Fatal(err.Error())
	}
	r.SetBasicAuth("admin", "invalid")

	authhandler.ServeHTTP(w, r)

	ft.AssertFalse(t, handlerCalled, "handler called")
	ft.AssertEqual(t, w.Code, http.StatusUnauthorized)
}
