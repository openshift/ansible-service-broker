package auth

import (
	"encoding/base64"
	"errors"
	"net/http/httptest"
	"testing"

	ft "github.com/openshift/ansible-service-broker/pkg/fusortest"
)

type MockUserServiceAdapter struct {
	userdb map[string]string
}

func (m MockUserServiceAdapter) FindByLogin(username string) (User, error) {
	if m.userdb[username] == "" {
		return User{}, errors.New("user not found")
	}

	return User{Username: username, Password: m.userdb[username]}, nil
}

func (m MockUserServiceAdapter) ValidateUser(username string, password string) bool {
	return m.userdb[username] == password
}

func TestGetPrincipalNoHeader(t *testing.T) {
	musa := MockUserServiceAdapter{}
	ba := NewBasicAuth(musa, log)
	r := httptest.NewRequest("POST", "/does/not/matter", nil)
	principal, err := ba.GetPrincipal(r)
	ft.AssertEqual(t, err.Error(), "user not found", "")
	ft.AssertTrue(t, principal == nil, "we should not have a principal")
}

func TestNewBasicAuth(t *testing.T) {
	musa := MockUserServiceAdapter{}
	ba := NewBasicAuth(musa, log)
	ft.AssertNotNil(t, ba, "new returned nil")
}

func TestValidAuth(t *testing.T) {
	musa := MockUserServiceAdapter{userdb: map[string]string{"admin": "password"}}
	ba := NewBasicAuth(musa, log)
	r := httptest.NewRequest("POST", "/does/not/matter", nil)
	auth := "Basic " + base64.StdEncoding.EncodeToString([]byte("admin:password"))
	r.Header.Add("Authorization", auth)
	principal, err := ba.GetPrincipal(r)
	if err != nil {
		t.Fatal(err)
	}
	ft.AssertNotNil(t, principal, "we should have a principal")
}

func TestInvalidAuth(t *testing.T) {
	musa := MockUserServiceAdapter{userdb: map[string]string{"admin": "invalid"}}
	ba := NewBasicAuth(musa, log)
	r := httptest.NewRequest("POST", "/does/not/matter", nil)
	auth := "Basic " + base64.StdEncoding.EncodeToString([]byte("admin:password"))
	r.Header.Add("Authorization", auth)
	principal, err := ba.GetPrincipal(r)
	ft.AssertTrue(t, principal == nil, "we should have a principal")
	ft.AssertNotNil(t, err, "we expected an error")
	ft.AssertEqual(t, err.Error(), "invalid credentials", "wrong error returned")
}

func TestCreatePrincipal(t *testing.T) {
	musa := MockUserServiceAdapter{userdb: map[string]string{"admin": "invalid"}}
	ba := NewBasicAuth(musa, log)
	p, err := ba.createPrincipal("admin")
	if err != nil {
		t.Fatal(err)
	}
	ft.AssertEqual(t, p.GetType(), "user", "did not get a user type")
	ft.AssertEqual(t, p.GetName(), "admin", "username didn't match")
}

func TestFailedCreatePrincipal(t *testing.T) {
	musa := MockUserServiceAdapter{}
	ba := NewBasicAuth(musa, log)
	p, err := ba.createPrincipal("admin")
	ft.AssertEqual(t, err.Error(), "user not found", "")
	ft.AssertTrue(t, p == nil, "principal is not nil")
}
