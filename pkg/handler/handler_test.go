package handler

import (
	"net/http/httptest"
	"os"
	"testing"

	"github.com/fusor/ansible-service-broker/pkg/broker"
	ft "github.com/fusor/ansible-service-broker/pkg/fusortest"
	"github.com/gorilla/mux"
	logging "github.com/op/go-logging"
	"github.com/pborman/uuid"
)

type MockBroker struct {
	Name   string
	Verify map[string]bool
	Err    error
}

func (m *MockBroker) called(method string, called bool) {
	if m.Verify == nil {
		m.Verify = make(map[string]bool)
	}
	m.Verify[method] = called
}

func (m MockBroker) Bootstrap() (*broker.BootstrapResponse, error) {
	m.called("bootstrap", true)
	return &broker.BootstrapResponse{10}, m.Err
}

func (m MockBroker) Catalog() (*broker.CatalogResponse, error) {
	m.called("catalog", true)
	return nil, m.Err
}
func (m MockBroker) Provision(uuid.UUID, *broker.ProvisionRequest, bool) (*broker.ProvisionResponse, error) {
	m.called("provision", true)
	return &broker.ProvisionResponse{Operation: "successful"}, m.Err
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
	//t.Fatal("lastoperation", true)
	state := broker.LastOperationStateInProgress
	return &broker.LastOperationResponse{State: state, Description: ""}, nil
}

var b MockBroker
var dahandler handler
var log = logging.MustGetLogger("handler")

func init() {
	// setup logging
	colorFormatter := logging.MustStringFormatter(
		"%{color}[%{time}] [%{level}] %{message}%{color:reset}",
	)
	backend := logging.NewLogBackend(os.Stdout, "", 1)
	backendFormatter := logging.NewBackendFormatter(backend, colorFormatter)
	logging.SetBackend(backend, backendFormatter)

	// setup the broker and handler
	b = MockBroker{Name: "testbroker"}
	dahandler = handler{*mux.NewRouter(), b, log}
}

func TestNewHandler(t *testing.T) {

	testb := MockBroker{Name: "testbroker"}
	testhandler := NewHandler(testb, log)
	ft.AssertNotNil(t, testhandler, "handler wasn't created")
}

func TestBootstrap(t *testing.T) {
	ft.AssertNotNil(t, dahandler, "")

	trr := TestRequest{Msg: "hello world", done: true}
	r := httptest.NewRequest("POST", "/v2/bootstrap", trr)
	w := httptest.NewRecorder()
	dahandler.bootstrap(w, r, nil)
	ft.AssertEqual(t, w.Code, 200, "code not equal")
}

func TestCatalog(t *testing.T) {
	ft.AssertNotNil(t, dahandler, "")

	trr := TestRequest{Msg: "hello world", done: true}
	r := httptest.NewRequest("GET", "/v2/catalog", trr)
	w := httptest.NewRecorder()
	dahandler.catalog(w, r, nil)
	ft.AssertEqual(t, w.Code, 200, "code not equal")
}

func TestProvisionCreate(t *testing.T) {
	trr := TestRequest{Msg: "{\"name\": \"hello world\"}"}
	r := httptest.NewRequest("PUT", "/v2/service_instance/688eea24-9cf9-43e3-9942-d1863b2a16af", trr)
	r.Header.Add("Content-Type", "application/json")
	w := httptest.NewRecorder()
	params := map[string]string{
		"instance_uuid": "688eea24-9cf9-43e3-9942-d1863b2a16af",
	}
	dahandler.provision(w, r, params)
	t.Log(w.Body)
	ft.AssertEqual(t, w.Code, 201, "provision not created")
}

func TestProvisionInvalidUUID(t *testing.T) {
	r := httptest.NewRequest("PUT", "/v2/service_instance/invaliduuid", nil)
	r.Header.Add("Content-Type", "application/json")
	w := httptest.NewRecorder()
	params := map[string]string{
		"instance_uuid": "invaliduuid",
	}
	dahandler.provision(w, r, params)
	ft.AssertEqual(t, w.Code, 400, "provision not created")
	ft.AssertError(t, w.Body, "invalid instance_uuid")
}

func TestProvisionCouldnotReadRequest(t *testing.T) {
	r := httptest.NewRequest("PUT", "/v2/service_instance/invaliduuid", nil)
	r.Header.Add("Content-Type", "application/json")
	w := httptest.NewRecorder()
	params := map[string]string{
		"instance_uuid": "688eea24-9cf9-43e3-9942-d1863b2a16af",
	}
	dahandler.provision(w, r, params)
	ft.AssertEqual(t, w.Code, 400, "provision not created")
	ft.AssertError(t, w.Body, "could not read request: EOF")
}

func TestProvisionConflict(t *testing.T) {
	t.SkipNow()
	trr := TestRequest{Msg: "{\"name\": \"hello world\"}"}
	r := httptest.NewRequest("PUT", "/v2/service_instance/688eea24-9cf9-43e3-9942-d1863b2a16af", trr)
	r.Header.Add("Content-Type", "application/json")
	w := httptest.NewRecorder()
	params := map[string]string{
		"instance_uuid": "688eea24-9cf9-43e3-9942-d1863b2a16af",
	}
	dahandler.provision(w, r, params)
	t.Log(w.Body)
	ft.AssertEqual(t, w.Code, 409, "provision not conflicted")
	// TODO: need to figure out how to provide the proper error that satisfies
	// IsAlreadyExists() on the handler.
	ft.AssertError(t, w.Body, "could not read request: EOF")
}

func TestUpdate(t *testing.T) {
}

func TestDeprovision(t *testing.T) {
}

func TestBind(t *testing.T) {
}

func TestUnbind(t *testing.T) {
}

func TestBindInvalidInstance(t *testing.T) {
	trr := TestRequest{Msg: "hello world", done: true}
	r := httptest.NewRequest("PUT", "/v2/service_instance/foo/service_bindings/bar", trr)
	w := httptest.NewRecorder()
	dahandler.bind(w, r, nil)
	ft.AssertEqual(t, w.Code, 400, "code not equal")
}

func TestInvalidLastOperation(t *testing.T) {
	t.Skip("Skipping because ultimately last_operation should expect the operation query param")
	r := httptest.NewRequest("GET", "/v2/service_instance/688eea24-9cf9-43e3-9942-d1863b2a16af/last_operation", nil)
	w := httptest.NewRecorder()
	params := map[string]string{
		"instance_uuid": "688eea24-9cf9-43e3-9942-d1863b2a16af",
	}
	dahandler.lastoperation(w, r, params)
	ft.AssertEqual(t, w.Code, 400, "invalid operation")
	ft.AssertError(t, w.Body, "invalid operation")
}

func TestMissingOperation(t *testing.T) {
	r := httptest.NewRequest("GET", "/v2/service_instance/688eea24-9cf9-43e3-9942-d1863b2a16af/last_operation", nil)
	w := httptest.NewRecorder()
	params := map[string]string{
		"instance_uuid": "688eea24-9cf9-43e3-9942-d1863b2a16af",
	}
	dahandler.lastoperation(w, r, params)
	ft.AssertEqual(t, w.Code, 200, "invalid error code")
	ft.AssertState(t, w.Body, "in progress")
}

func TestLastOperation(t *testing.T) {
	r := httptest.NewRequest("GET", "/v2/service_instance/688eea24-9cf9-43e3-9942-d1863b2a16af/last_operation?operation=abcd", nil)
	r.Header.Add("Content-Type", "application/json")
	w := httptest.NewRecorder()
	params := map[string]string{
		"instance_uuid": "688eea24-9cf9-43e3-9942-d1863b2a16af",
	}
	dahandler.lastoperation(w, r, params)
	ft.AssertEqual(t, w.Code, 200, "lastoperation should've returned 200")
	ft.AssertState(t, w.Body, "in progress")
}
