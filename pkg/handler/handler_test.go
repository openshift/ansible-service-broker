//
// Copyright (c) 2017 Red Hat, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//    http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//
// Red Hat trademarks are not licensed under Apache License, Version 2.
// No permission is granted to use or replicate Red Hat trademarks that
// are incorporated in this software or its documentation.
//

package handler

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"testing"

	"github.com/gorilla/mux"
	logging "github.com/op/go-logging"
	"github.com/openshift/ansible-service-broker/pkg/apb"
	"github.com/openshift/ansible-service-broker/pkg/auth"
	"github.com/openshift/ansible-service-broker/pkg/broker"
	ft "github.com/openshift/ansible-service-broker/pkg/fusortest"
	"github.com/pborman/uuid"
)

type MockBroker struct {
	Name      string
	Verify    map[string]bool
	Err       error
	Operation string
}

type MockUserServiceAdapter struct {
	userdb map[string]string
}

func (m MockUserServiceAdapter) FindByLogin(username string) (auth.User, error) {
	if m.userdb[username] == "" {
		return auth.User{}, errors.New("user not found")
	}

	return auth.User{Username: username, Password: m.userdb[username]}, nil
}

func (m MockUserServiceAdapter) ValidateUser(username string, password string) bool {
	if m.userdb[username] == "" {
		return false
	}
	return m.userdb[username] == password
}

const base64TestSpec = "aWQ6IDU1YzUzYTVkLTY1YTYtNGMyNy04OGZjLWUwMjc0MTBiMTMzNw0KbmFtZTogbWVkaWF3aWtpMTIzLWFwYg0KaW1hZ2U6IGFuc2libGVwbGF5Ym9va2J1bmRsZS9tZWRpYXdpa2kxMjMtYXBiDQpkZXNjcmlwdGlvbjogIk1lZGlhd2lraTEyMyBhcGIgaW1wbGVtZW50YXRpb24iDQpiaW5kYWJsZTogZmFsc2UNCmFzeW5jOiBvcHRpb25hbA0KbWV0YWRhdGE6DQogIGRpc3BsYXluYW1lOiAiUmVkIEhhdCBNZWRpYXdpa2kiDQogIGxvbmdEZXNjcmlwdGlvbjogIkFuIGFwYiB0aGF0IGRlcGxveXMgTWVkaWF3aWtpIDEuMjMiDQogIGltYWdlVVJMOiAiaHR0cHM6Ly91cGxvYWQud2lraW1lZGlhLm9yZy93aWtpcGVkaWEvY29tbW9ucy8wLzAxL01lZGlhV2lraS1zbWFsbGVyLWxvZ28ucG5nIg0KICBkb2N1bWVudGF0aW9uVVJMOiAiaHR0cHM6Ly93d3cubWVkaWF3aWtpLm9yZy93aWtpL0RvY3VtZW50YXRpb24iDQpwYXJhbWV0ZXJzOg0KICAtIG1lZGlhd2lraV9kYl9zY2hlbWE6DQogICAgLSB0aXRsZTogTWVkaWF3aWtpIERCIFNjaGVtYQ0KICAgICAgdHlwZTogc3RyaW5nDQogICAgICBkZWZhdWx0OiBtZWRpYXdpa2kNCiAgLSBtZWRpYXdpa2lfc2l0ZV9uYW1lOg0KICAgIC0gdGl0bGU6IE1lZGlhd2lraSBTaXRlIE5hbWUNCiAgICAgIHR5cGU6IHN0cmluZw0KICAgICAgZGVmYXVsdDogTWVkaWFXaWtpDQogIC0gbWVkaWF3aWtpX3NpdGVfbGFuZzoNCiAgICAtIHRpdGxlOiBNZWRpYXdpa2kgU2l0ZSBMYW5ndWFnZQ0KICAgICAgdHlwZTogc3RyaW5nDQogICAgICBkZWZhdWx0OiBlbg0KICAtIG1lZGlhd2lraV9hZG1pbl91c2VyOg0KICAgIC0gdGl0bGU6IE1lZGlhd2lraSBBZG1pbiBVc2VyDQogICAgICB0eXBlOiBzdHJpbmcNCiAgICAgIGRlZmF1bHQ6IGFkbWluDQogIC0gbWVkaWF3aWtpX2FkbWluX3Bhc3M6DQogICAgLSB0aXRsZTogTWVkaWF3aWtpIEFkbWluIFVzZXIgUGFzc3dvcmQNCiAgICAgIHR5cGU6IHN0cmluZw0KcmVxdWlyZWQ6DQogIC0gbWVkaWF3aWtpX2RiX3NjaGVtYQ0KICAtIG1lZGlhd2lraV9zaXRlX25hbWUNCiAgLSBtZWRpYXdpa2lfc2l0ZV9sYW5nDQogIC0gbWVkaWF3aWtpX2FkbWluX3VzZXINCiAgLSBtZWRpYXdpa2lfYWRtaW5fcGFzcw0K"

func (m *MockBroker) called(method string, called bool) {
	if m.Verify == nil {
		m.Verify = make(map[string]bool)
	}
	m.Verify[method] = called
}

func (m MockBroker) Bootstrap() (*broker.BootstrapResponse, error) {
	m.called("bootstrap", true)
	return &broker.BootstrapResponse{SpecCount: 10, ImageCount: 10}, m.Err
}

func (m MockBroker) Catalog() (*broker.CatalogResponse, error) {
	m.called("catalog", true)
	return nil, m.Err
}
func (m MockBroker) Provision(uuid.UUID, *broker.ProvisionRequest, bool) (*broker.ProvisionResponse, error) {
	m.called("provision", true)
	fmt.Println("provision called")
	fmt.Println(m.Operation)
	return &broker.ProvisionResponse{Operation: m.Operation}, m.Err
}
func (m MockBroker) Update(uuid.UUID, *broker.UpdateRequest, bool) (*broker.UpdateResponse, error) {
	m.called("update", true)
	return nil, m.Err
}
func (m MockBroker) Deprovision(apb.ServiceInstance, string, bool) (*broker.DeprovisionResponse, error) {
	m.called("deprovision", true)
	return nil, m.Err
}
func (m MockBroker) Bind(apb.ServiceInstance, uuid.UUID, *broker.BindRequest) (*broker.BindResponse, error) {
	m.called("bind", true)
	return nil, m.Err
}
func (m MockBroker) Unbind(apb.ServiceInstance, uuid.UUID, string) (*broker.UnbindResponse, error) {
	m.called("unbind", true)
	return nil, m.Err
}
func (m MockBroker) LastOperation(uuid.UUID, *broker.LastOperationRequest) (*broker.LastOperationResponse, error) {
	state := broker.LastOperationStateInProgress
	return &broker.LastOperationResponse{State: state, Description: ""}, m.Err
}

func (m MockBroker) GetServiceInstance(uuid.UUID) (apb.ServiceInstance, error) {
	return apb.ServiceInstance{}, nil
}

func (m MockBroker) Recover() (string, error) {
	return "recover", nil
}

func (m MockBroker) AddSpec(spec apb.Spec) (*broker.CatalogResponse, error) {
	m.called("addSpec", true)
	return &broker.CatalogResponse{Services: []broker.Service{}}, nil
}

func (m MockBroker) RemoveSpec(specID string) error {
	m.called("removeSpec", true)
	return nil
}

func (m MockBroker) RemoveSpecs() error {
	m.called("removeSpecs", true)
	return nil
}

var log = logging.MustGetLogger("handler")
var brokerConfig broker.Config

func init() {
	// setup logging
	colorFormatter := logging.MustStringFormatter(
		"%{color}[%{time}] [%{level}] %{message}%{color:reset}",
	)
	backend := logging.NewLogBackend(os.Stdout, "", 1)
	backendFormatter := logging.NewBackendFormatter(backend, colorFormatter)
	logging.SetBackend(backend, backendFormatter)
	// will default to all config values to false
	brokerConfig = broker.Config{}
	brokerConfig.AutoEscalate = true
}

func TestNewHandler(t *testing.T) {
	testb := MockBroker{Name: "testbroker"}
	testhandler := NewHandler(testb, log, brokerConfig, "", nil, nil)
	ft.AssertNotNil(t, testhandler, "handler wasn't created")
}

func TestNewHandlerDoesNotHaveAPBRoute(t *testing.T) {
	testb := MockBroker{Name: "testbroker"}
	brokerConfig.DevBroker = false
	testhandler := NewHandler(testb, log, brokerConfig, "", nil, nil)
	req, err := http.NewRequest(http.MethodPost, "/apb/spec", nil)
	if err != nil {
		ft.AssertTrue(t, false, err.Error())
	}
	form := url.Values{}
	form.Add("apbSpec", base64TestSpec)
	req.PostForm = form
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	w := httptest.NewRecorder()
	testhandler.ServeHTTP(w, req)
	ft.AssertEqual(t, w.Result().StatusCode, http.StatusNotFound, fmt.Sprintf("resulting status was not 404 - %v", w.Result().Status))
	ft.AssertNotNil(t, testhandler, "handler wasn't created")
}

func TestDevHandlerDoesHaveAPBRoute(t *testing.T) {
	testb := MockBroker{Name: "testbroker"}
	brokerConfig.DevBroker = true
	testhandler := NewHandler(testb, log, brokerConfig, "/test-prefix", nil, nil)
	req, err := http.NewRequest(http.MethodPost, "/test-prefix/apb/spec", nil)
	if err != nil {
		ft.AssertTrue(t, false, err.Error())
	}
	form := url.Values{}
	form.Add("apbSpec", base64TestSpec)
	req.PostForm = form
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	w := httptest.NewRecorder()
	testhandler.ServeHTTP(w, req)
	ft.AssertEqual(t, w.Result().StatusCode, http.StatusOK, fmt.Sprintf("resulting status was not 404 - %v", w.Result().Status))
	ft.AssertNotNil(t, testhandler, "handler wasn't created")
}

func TestNewHandlerDoesNotHaveAPBSpecDeleteRoute(t *testing.T) {
	testb := MockBroker{Name: "testbroker"}
	brokerConfig.DevBroker = false
	testhandler := NewHandler(testb, log, brokerConfig, "", nil, nil)
	req, err := http.NewRequest(http.MethodDelete, "/apb/spec", nil)
	if err != nil {
		ft.AssertTrue(t, false, err.Error())
	}
	w := httptest.NewRecorder()
	testhandler.ServeHTTP(w, req)
	ft.AssertEqual(t, w.Result().StatusCode, http.StatusNotFound, fmt.Sprintf("resulting status was not 404 - %v", w.Result().Status))
	ft.AssertNotNil(t, testhandler, "handler wasn't created")
}

func TestDevHandlerDoesHaveAPBSpecDeleteRoute(t *testing.T) {
	testb := MockBroker{Name: "testbroker"}
	brokerConfig.DevBroker = true
	testhandler := NewHandler(testb, log, brokerConfig, "", nil, nil)
	req, err := http.NewRequest(http.MethodDelete, "/apb/spec", nil)
	if err != nil {
		ft.AssertTrue(t, false, err.Error())
	}
	w := httptest.NewRecorder()
	testhandler.ServeHTTP(w, req)
	ft.AssertEqual(t, w.Result().StatusCode, http.StatusNoContent, fmt.Sprintf("resulting status was not 204 - %v", w.Result().Status))
	ft.AssertNotNil(t, testhandler, "handler wasn't created")
}

func TestNewHandlerDoesNotHaveAPBSpecsDeleteRoute(t *testing.T) {
	testb := MockBroker{Name: "testbroker"}
	brokerConfig.DevBroker = false
	testhandler := NewHandler(testb, log, brokerConfig, "", nil, nil)
	req, err := http.NewRequest(http.MethodDelete, "/apb/spec", nil)
	if err != nil {
		ft.AssertTrue(t, false, err.Error())
	}
	w := httptest.NewRecorder()
	testhandler.ServeHTTP(w, req)
	ft.AssertEqual(t, w.Result().StatusCode, http.StatusNotFound, fmt.Sprintf("resulting status was not 404 - %v", w.Result().Status))
	ft.AssertNotNil(t, testhandler, "handler wasn't created")
}

func TestDevHandlerDoesHaveAPBSpecsDeleteRoute(t *testing.T) {
	testb := MockBroker{Name: "testbroker"}
	brokerConfig.DevBroker = true
	testhandler := NewHandler(testb, log, brokerConfig, "", nil, nil)
	req, err := http.NewRequest(http.MethodDelete, "/apb/spec", nil)
	if err != nil {
		ft.AssertTrue(t, false, err.Error())
	}
	w := httptest.NewRecorder()
	testhandler.ServeHTTP(w, req)
	ft.AssertEqual(t, w.Result().StatusCode, http.StatusNoContent, fmt.Sprintf("resulting status was not 204 - %v", w.Result().Status))
	ft.AssertNotNil(t, testhandler, "handler wasn't created")
}

func TestBootstrap(t *testing.T) {
	testhandler, w, r := buildBootstrapHandler(nil)
	testhandler.bootstrap(w, r, nil)
	ft.AssertEqual(t, w.Code, 200, "code not equal")
}

func TestCatalog(t *testing.T) {
	testhandler, w, r := buildCatalogHandler(nil)
	testhandler.catalog(w, r, nil)
	ft.AssertEqual(t, w.Code, 200, "code not equal")
}

func TestProvisionCreate(t *testing.T) {
	testhandler, w, r, params := buildProvisionHandler(uuid.New(), nil, "")
	testhandler.provision(w, r, params)
	ft.AssertEqual(t, w.Code, 201, "provision not created")
	ft.AssertOperation(t, w.Body, "")
}

func TestProvisionInvalidUUID(t *testing.T) {
	testhandler, w, r, params := buildProvisionHandler("invaliduuid", nil, "")
	testhandler.provision(w, r, params)
	ft.AssertEqual(t, w.Code, 400, "provision not created")
	ft.AssertError(t, w.Body, "invalid instance_uuid")
}

func TestProvisionCouldnotReadRequest(t *testing.T) {
	r := httptest.NewRequest("PUT", "/v2/service_instance/invaliduuid", nil)
	r.Header.Add("Content-Type", "application/json")

	testhandler, w, _, params := buildProvisionHandler(uuid.New(), nil, "")
	testhandler.provision(w, r, params)
	ft.AssertEqual(t, w.Code, 400, "provision not created")
	ft.AssertError(t, w.Body, "could not read request: EOF")
}

func TestProvisionDuplicate(t *testing.T) {
	testhandler, w, r, params := buildProvisionHandler(uuid.New(), broker.ErrorDuplicate, "")
	testhandler.provision(w, r, params)
	ft.AssertEqual(t, w.Code, 409, "should've been a conflict")
	ft.AssertOperation(t, w.Body, "")
}

func TestProvisionAlreadyProvisioned(t *testing.T) {
	testhandler, w, r, params := buildProvisionHandler(uuid.New(), broker.ErrorAlreadyProvisioned, "")
	testhandler.provision(w, r, params)
	ft.AssertEqual(t, w.Code, 200, "should've been an OK ")
}

func TestProvisionNotFound(t *testing.T) {
	testhandler, w, r, params := buildProvisionHandler(uuid.New(), broker.ErrorNotFound, "")
	testhandler.provision(w, r, params)
	ft.AssertEqual(t, w.Code, 400, "should've been a bad request for error not found")
	ft.AssertError(t, w.Body, "not found")
}

func TestProvisionOtherError(t *testing.T) {
	testhandler, w, r, params := buildProvisionHandler(uuid.New(), errors.New("random error"), "")
	testhandler.provision(w, r, params)
	ft.AssertEqual(t, w.Code, 400, "should've been a bad request for error not found")
	ft.AssertError(t, w.Body, "random error")
}

func TestProvisionAccepted(t *testing.T) {
	testuuid := uuid.New()
	testhandler, w, r, params := buildProvisionHandler(uuid.New(), nil, testuuid)
	testhandler.provision(w, r, params)
	ft.AssertEqual(t, w.Code, 201, "should've been 201 accepted")
	ft.AssertOperation(t, w.Body, testuuid)
}
func TestUpdate(t *testing.T) {
}

func TestDeprovision(t *testing.T) {
}

func TestBind(t *testing.T) {
}

func TestUnbind(t *testing.T) {
}

// Bind Tests
func TestBindBadBindRequest(t *testing.T) {
	testuuid := uuid.New()
	r := httptest.NewRequest("PUT",
		fmt.Sprintf("/v2/service_instance/%s/service_bindings/%s", testuuid, testuuid), nil)
	r.Header.Add("Content-Type", "application/json")

	testhandler, w, _, params := buildBindHandler(testuuid, broker.ErrorDuplicate)
	testhandler.bind(w, r, params)
	ft.AssertEqual(t, w.Code, 500, "should've been an internal server error")
}

func TestBindDuplicate(t *testing.T) {
	testhandler, w, r, params := buildBindHandler(uuid.New(), broker.ErrorDuplicate)
	testhandler.bind(w, r, params)
	ft.AssertEqual(t, w.Code, 409, "should've been a conflict")
	ft.AssertError(t, w.Body, "")
}

func TestBindAlreadyProvisioned(t *testing.T) {
	testhandler, w, r, params := buildBindHandler(uuid.New(), broker.ErrorAlreadyProvisioned)
	testhandler.bind(w, r, params)
	ft.AssertEqual(t, w.Code, 200, "should've been an OK ")
}

func TestBindNotFound(t *testing.T) {
	testhandler, w, r, params := buildBindHandler(uuid.New(), broker.ErrorNotFound)
	testhandler.bind(w, r, params)
	ft.AssertEqual(t, w.Code, 400, "should've been a bad request for error not found")
	ft.AssertError(t, w.Body, "not found")
}

func TestBindOtherError(t *testing.T) {
	testhandler, w, r, params := buildBindHandler(uuid.New(), errors.New("random error"))
	testhandler.bind(w, r, params)
	ft.AssertEqual(t, w.Code, 400, "should've been a bad request for error not found")
	ft.AssertError(t, w.Body, "random error")
}

func TestBindCreated(t *testing.T) {
	testhandler, w, r, params := buildBindHandler(uuid.New(), nil)
	testhandler.bind(w, r, params)
	ft.AssertEqual(t, w.Code, 201, "should've been a created")
	ft.AssertError(t, w.Body, "")
}

func TestBindInvalidInstance(t *testing.T) {
	testhandler, w, r, _ := buildBindHandler(uuid.New(), nil)
	testhandler.bind(w, r, nil)
	ft.AssertEqual(t, w.Code, 400, "code not equal")
}

// LastOperation tests
func TestInvalidLastOperation(t *testing.T) {
	t.Skip("Skipping because ultimately last_operation should expect the operation query param")
	testuuid := uuid.New()
	r := httptest.NewRequest("GET", fmt.Sprintf("/v2/service_instance/%s/last_operation", testuuid), nil)

	testhandler, w, _, params := buildLastOperationHandler(testuuid, nil)
	testhandler.lastoperation(w, r, params)
	ft.AssertEqual(t, w.Code, 400, "invalid operation")
	ft.AssertError(t, w.Body, "invalid operation")
}

func TestMissingOperation(t *testing.T) {
	testuuid := uuid.New()
	r := httptest.NewRequest("GET", fmt.Sprintf("/v2/service_instance/%s/last_operation", testuuid), nil)

	testhandler, w, _, params := buildLastOperationHandler(testuuid, nil)
	testhandler.lastoperation(w, r, params)
	ft.AssertEqual(t, w.Code, 200, "invalid error code")
	ft.AssertState(t, w.Body, "in progress")
}

func TestLastOperation(t *testing.T) {
	testhandler, w, r, params := buildLastOperationHandler(uuid.New(), nil)
	testhandler.lastoperation(w, r, params)
	ft.AssertEqual(t, w.Code, 200, "lastoperation should've returned 200")
	ft.AssertState(t, w.Body, "in progress")
}

// utility functions

func buildBootstrapHandler(err error) (handler, *httptest.ResponseRecorder, *http.Request) {

	testb := MockBroker{Name: "testbroker", Err: err}
	testhandler := handler{*mux.NewRouter(), testb, log, brokerConfig, nil}

	r := httptest.NewRequest("POST", "/v2/bootstrap", nil)
	w := httptest.NewRecorder()
	return testhandler, w, r
}

func buildCatalogHandler(err error) (handler, *httptest.ResponseRecorder, *http.Request) {

	testb := MockBroker{Name: "testbroker", Err: err}
	testhandler := handler{*mux.NewRouter(), testb, log, brokerConfig, nil}

	r := httptest.NewRequest("GET", "/v2/catalog", nil)
	w := httptest.NewRecorder()
	return testhandler, w, r
}

func buildProvisionHandler(testuuid string, err error, operation string) (handler, *httptest.ResponseRecorder, *http.Request, map[string]string) {

	testb := MockBroker{Name: "testbroker", Err: err, Operation: operation}
	testhandler := handler{*mux.NewRouter(), testb, log, brokerConfig, nil}

	trr := TestRequest{Msg: fmt.Sprintf("{\"plan_id\": \"%s\",\"service_id\": \"%s\"}", testuuid, testuuid)}
	r := httptest.NewRequest("PUT", fmt.Sprintf("/v2/service_instance/%s", testuuid), trr)
	r.Header.Add("Content-Type", "application/json")
	r = r.WithContext(context.WithValue(r.Context(), UserInfoContext, broker.UserInfo{Username: "admin"}))
	w := httptest.NewRecorder()
	params := map[string]string{
		"instance_uuid":      testuuid,
		"accepts_incomplete": "true",
	}
	return testhandler, w, r, params
}

func buildLastOperationHandler(testuuid string, err error) (handler, *httptest.ResponseRecorder, *http.Request, map[string]string) {

	testb := MockBroker{Name: "testbroker", Err: err}
	testhandler := handler{*mux.NewRouter(), testb, log, brokerConfig, nil}

	r := httptest.NewRequest("GET",
		fmt.Sprintf("/v2/service_instance/%s/last_operation?operation=%s", testuuid, testuuid), nil)
	r.Header.Add("Content-Type", "application/json")
	w := httptest.NewRecorder()
	params := map[string]string{
		"instance_uuid": testuuid,
	}

	return testhandler, w, r, params
}

func buildBindHandler(testuuid string, err error) (handler, *httptest.ResponseRecorder, *http.Request, map[string]string) {

	testb := MockBroker{Name: "testbroker", Err: err}
	testhandler := handler{*mux.NewRouter(), testb, log, brokerConfig, nil}

	trr := TestRequest{Msg: fmt.Sprintf("{\"plan_id\": \"%s\",\"service_id\": \"%s\"}", testuuid, testuuid)}
	r := httptest.NewRequest("PUT",
		fmt.Sprintf("/v2/service_instance/%s/service_bindings/%s", testuuid, testuuid), trr)
	r.Header.Add("Content-Type", "application/json")
	r = r.WithContext(context.WithValue(r.Context(), UserInfoContext, broker.UserInfo{Username: "admin"}))
	w := httptest.NewRecorder()
	params := map[string]string{
		"instance_uuid": testuuid,
		"binding_uuid":  testuuid,
	}
	return testhandler, w, r, params
}

// auth handler tests
func TestHandlerAuthorized(t *testing.T) {
	handlerCalled := false
	testhandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		handlerCalled = true
	})

	ba := auth.NewBasicAuth(
		MockUserServiceAdapter{userdb: map[string]string{"admin": "password"}}, log)

	authhandler := authHandler(testhandler, []auth.Provider{ba}, log)

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

	ba := auth.NewBasicAuth(
		MockUserServiceAdapter{userdb: map[string]string{"admin": "password"}}, log)

	authhandler := authHandler(testhandler, []auth.Provider{ba}, log)

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
