package handler

import (
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/openshift/ansible-service-broker/pkg/broker"
	ft "github.com/openshift/ansible-service-broker/pkg/fusortest"
	"github.com/gorilla/mux"
	logging "github.com/op/go-logging"
	"github.com/pborman/uuid"
)

type MockBroker struct {
	Name      string
	Verify    map[string]bool
	Err       error
	Operation string
}

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
func (m MockBroker) Update(uuid.UUID, *broker.UpdateRequest) (*broker.UpdateResponse, error) {
	m.called("update", true)
	return nil, m.Err
}
func (m MockBroker) Deprovision(uuid.UUID) (*broker.DeprovisionResponse, error) {
	m.called("deprovision", true)
	return nil, m.Err
}
func (m MockBroker) Bind(uuid.UUID, uuid.UUID, *broker.BindRequest) (*broker.BindResponse, error) {
	m.called("bind", true)
	return nil, m.Err
}
func (m MockBroker) Unbind(uuid.UUID, uuid.UUID) error {
	m.called("unbind", true)
	return m.Err
}
func (m MockBroker) LastOperation(uuid.UUID, *broker.LastOperationRequest) (*broker.LastOperationResponse, error) {
	state := broker.LastOperationStateInProgress
	return &broker.LastOperationResponse{State: state, Description: ""}, m.Err
}

var log = logging.MustGetLogger("handler")

func init() {
	// setup logging
	colorFormatter := logging.MustStringFormatter(
		"%{color}[%{time}] [%{level}] %{message}%{color:reset}",
	)
	backend := logging.NewLogBackend(os.Stdout, "", 1)
	backendFormatter := logging.NewBackendFormatter(backend, colorFormatter)
	logging.SetBackend(backend, backendFormatter)
}

func TestNewHandler(t *testing.T) {
	testb := MockBroker{Name: "testbroker"}
	testhandler := NewHandler(testb, log)
	ft.AssertNotNil(t, testhandler, "handler wasn't created")
}

func TestBootstrap(t *testing.T) {
	testhandler, w, r := BuildBootstrapHandler(nil)
	testhandler.bootstrap(w, r, nil)
	ft.AssertEqual(t, w.Code, 200, "code not equal")
}

func TestCatalog(t *testing.T) {
	testhandler, w, r := BuildCatalogHandler(nil)
	testhandler.catalog(w, r, nil)
	ft.AssertEqual(t, w.Code, 200, "code not equal")
}

func TestProvisionCreate(t *testing.T) {
	testhandler, w, r, params := BuildProvisionHandler(uuid.New(), nil, "")
	testhandler.provision(w, r, params)
	ft.AssertEqual(t, w.Code, 201, "provision not created")
	ft.AssertOperation(t, w.Body, "")
}

func TestProvisionInvalidUUID(t *testing.T) {
	testhandler, w, r, params := BuildProvisionHandler("invaliduuid", nil, "")
	testhandler.provision(w, r, params)
	ft.AssertEqual(t, w.Code, 400, "provision not created")
	ft.AssertError(t, w.Body, "invalid instance_uuid")
}

func TestProvisionCouldnotReadRequest(t *testing.T) {
	r := httptest.NewRequest("PUT", "/v2/service_instance/invaliduuid", nil)
	r.Header.Add("Content-Type", "application/json")

	testhandler, w, _, params := BuildProvisionHandler(uuid.New(), nil, "")
	testhandler.provision(w, r, params)
	ft.AssertEqual(t, w.Code, 400, "provision not created")
	ft.AssertError(t, w.Body, "could not read request: EOF")
}

func TestProvisionDuplicate(t *testing.T) {
	testhandler, w, r, params := BuildProvisionHandler(uuid.New(), broker.ErrorDuplicate, "")
	testhandler.provision(w, r, params)
	ft.AssertEqual(t, w.Code, 409, "should've been a conflict")
	ft.AssertOperation(t, w.Body, "")
}

func TestProvisionAlreadyProvisioned(t *testing.T) {
	testhandler, w, r, params := BuildProvisionHandler(uuid.New(), broker.ErrorAlreadyProvisioned, "")
	testhandler.provision(w, r, params)
	ft.AssertEqual(t, w.Code, 200, "should've been an OK ")
}

func TestProvisionNotFound(t *testing.T) {
	testhandler, w, r, params := BuildProvisionHandler(uuid.New(), broker.ErrorNotFound, "")
	testhandler.provision(w, r, params)
	ft.AssertEqual(t, w.Code, 400, "should've been a bad request for error not found")
	ft.AssertError(t, w.Body, "not found")
}

func TestProvisionOtherError(t *testing.T) {
	testhandler, w, r, params := BuildProvisionHandler(uuid.New(), errors.New("random error"), "")
	testhandler.provision(w, r, params)
	ft.AssertEqual(t, w.Code, 400, "should've been a bad request for error not found")
	ft.AssertError(t, w.Body, "random error")
}

func TestProvisionAccepted(t *testing.T) {
	testuuid := uuid.New()
	testhandler, w, r, params := BuildProvisionHandler(uuid.New(), nil, testuuid)
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

	testhandler, w, _, params := BuildBindHandler(testuuid, broker.ErrorDuplicate)
	testhandler.bind(w, r, params)
	ft.AssertEqual(t, w.Code, 500, "should've been an internal server error")
}

func TestBindDuplicate(t *testing.T) {
	testhandler, w, r, params := BuildBindHandler(uuid.New(), broker.ErrorDuplicate)
	testhandler.bind(w, r, params)
	ft.AssertEqual(t, w.Code, 409, "should've been a conflict")
	ft.AssertError(t, w.Body, "")
}

func TestBindAlreadyProvisioned(t *testing.T) {
	testhandler, w, r, params := BuildBindHandler(uuid.New(), broker.ErrorAlreadyProvisioned)
	testhandler.bind(w, r, params)
	ft.AssertEqual(t, w.Code, 200, "should've been an OK ")
}

func TestBindNotFound(t *testing.T) {
	testhandler, w, r, params := BuildBindHandler(uuid.New(), broker.ErrorNotFound)
	testhandler.bind(w, r, params)
	ft.AssertEqual(t, w.Code, 400, "should've been a bad request for error not found")
	ft.AssertError(t, w.Body, "not found")
}

func TestBindOtherError(t *testing.T) {
	testhandler, w, r, params := BuildBindHandler(uuid.New(), errors.New("random error"))
	testhandler.bind(w, r, params)
	ft.AssertEqual(t, w.Code, 400, "should've been a bad request for error not found")
	ft.AssertError(t, w.Body, "random error")
}

func TestBindCreated(t *testing.T) {
	testhandler, w, r, params := BuildBindHandler(uuid.New(), nil)
	testhandler.bind(w, r, params)
	ft.AssertEqual(t, w.Code, 201, "should've been a created")
	ft.AssertError(t, w.Body, "")
}

func TestBindInvalidInstance(t *testing.T) {
	testhandler, w, r, _ := BuildBindHandler(uuid.New(), nil)
	testhandler.bind(w, r, nil)
	ft.AssertEqual(t, w.Code, 400, "code not equal")
}

// LastOperation tests
func TestInvalidLastOperation(t *testing.T) {
	t.Skip("Skipping because ultimately last_operation should expect the operation query param")
	testuuid := uuid.New()
	r := httptest.NewRequest("GET", fmt.Sprintf("/v2/service_instance/%s/last_operation", testuuid), nil)

	testhandler, w, _, params := BuildLastOperationHandler(testuuid, nil)
	testhandler.lastoperation(w, r, params)
	ft.AssertEqual(t, w.Code, 400, "invalid operation")
	ft.AssertError(t, w.Body, "invalid operation")
}

func TestMissingOperation(t *testing.T) {
	testuuid := uuid.New()
	r := httptest.NewRequest("GET", fmt.Sprintf("/v2/service_instance/%s/last_operation", testuuid), nil)

	testhandler, w, _, params := BuildLastOperationHandler(testuuid, nil)
	testhandler.lastoperation(w, r, params)
	ft.AssertEqual(t, w.Code, 200, "invalid error code")
	ft.AssertState(t, w.Body, "in progress")
}

func TestLastOperation(t *testing.T) {
	testhandler, w, r, params := BuildLastOperationHandler(uuid.New(), nil)
	testhandler.lastoperation(w, r, params)
	ft.AssertEqual(t, w.Code, 200, "lastoperation should've returned 200")
	ft.AssertState(t, w.Body, "in progress")
}

// utility functions

func BuildBootstrapHandler(err error) (
	handler, *httptest.ResponseRecorder, *http.Request) {

	testb := MockBroker{Name: "testbroker", Err: err}
	testhandler := handler{*mux.NewRouter(), testb, log}

	r := httptest.NewRequest("POST", "/v2/bootstrap", nil)
	w := httptest.NewRecorder()
	return testhandler, w, r
}

func BuildCatalogHandler(err error) (
	handler, *httptest.ResponseRecorder, *http.Request) {

	testb := MockBroker{Name: "testbroker", Err: err}
	testhandler := handler{*mux.NewRouter(), testb, log}

	r := httptest.NewRequest("GET", "/v2/catalog", nil)
	w := httptest.NewRecorder()
	return testhandler, w, r
}

func BuildProvisionHandler(testuuid string, err error, operation string) (
	handler, *httptest.ResponseRecorder, *http.Request, map[string]string) {

	testb := MockBroker{Name: "testbroker", Err: err, Operation: operation}
	testhandler := handler{*mux.NewRouter(), testb, log}

	trr := TestRequest{Msg: fmt.Sprintf("{\"plan_id\": \"%s\",\"service_id\": \"%s\"}", testuuid, testuuid)}
	r := httptest.NewRequest("PUT", fmt.Sprintf("/v2/service_instance/%s", testuuid), trr)
	r.Header.Add("Content-Type", "application/json")
	w := httptest.NewRecorder()
	params := map[string]string{
		"instance_uuid":      testuuid,
		"accepts_incomplete": "true",
	}
	return testhandler, w, r, params
}

func BuildLastOperationHandler(testuuid string, err error) (
	handler, *httptest.ResponseRecorder, *http.Request, map[string]string) {

	testb := MockBroker{Name: "testbroker", Err: err}
	testhandler := handler{*mux.NewRouter(), testb, log}

	r := httptest.NewRequest("GET",
		fmt.Sprintf("/v2/service_instance/%s/last_operation?operation=%s", testuuid, testuuid), nil)
	r.Header.Add("Content-Type", "application/json")
	w := httptest.NewRecorder()
	params := map[string]string{
		"instance_uuid": testuuid,
	}

	return testhandler, w, r, params
}

func BuildBindHandler(testuuid string, err error) (
	handler, *httptest.ResponseRecorder, *http.Request, map[string]string) {

	testb := MockBroker{Name: "testbroker", Err: err}
	testhandler := handler{*mux.NewRouter(), testb, log}

	trr := TestRequest{Msg: fmt.Sprintf("{\"plan_id\": \"%s\",\"service_id\": \"%s\"}", testuuid, testuuid)}
	r := httptest.NewRequest("PUT",
		fmt.Sprintf("/v2/service_instance/%s/service_bindings/%s", testuuid, testuuid), trr)
	r.Header.Add("Content-Type", "application/json")
	w := httptest.NewRecorder()
	params := map[string]string{
		"instance_uuid": testuuid,
		"binding_uuid":  testuuid,
	}
	return testhandler, w, r, params
}
