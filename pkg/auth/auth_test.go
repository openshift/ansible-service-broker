package auth

import (
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	ft "github.com/openshift/ansible-service-broker/pkg/fusortest"
)

func TestNewFusa(t *testing.T) {
	data := []byte("admin\nadmin")
	ioutil.WriteFile("/tmp/foo", data, 0644)

	defer os.Remove("/tmp/foo")

	fusa := NewFileUserServiceAdapter("/tmp/foo")
	adminuser, _ := fusa.FindByLogin("admin")
	ft.AssertEqual(t, adminuser.Username, "admin", "username does not match")
	ft.AssertEqual(t, adminuser.Password, "admin", "password does not match")
	ft.AssertTrue(t, fusa.ValidateUser("admin", "admin"), "validation failed")
	ft.AssertFalse(t, fusa.ValidateUser("notme", "admin"), "validation passed, expected failure")
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
		MockUserServiceAdapter{userdb: map[string]string{"admin": "password"}})

	authhandler := Handler(testhandler, []Provider{ba})

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
		MockUserServiceAdapter{userdb: map[string]string{"admin": "password"}})

	authhandler := Handler(testhandler, []Provider{ba})

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
