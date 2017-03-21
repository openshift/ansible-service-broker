package handler

import (
	"net/http/httptest"
	"testing"

	"github.com/fusor/ansible-service-broker/pkg/broker"
	ft "github.com/fusor/ansible-service-broker/pkg/fusortest"
	"github.com/gorilla/mux"
	"github.com/pborman/uuid"
)

/*
// request object
type TestRequest struct {
	Msg  string
	done bool
}

func (r TestRequest) Read(p []byte) (n int, err error) {

	if r.done {
		return 0, io.EOF
	}
	for i, b := range []byte(r.Msg) {
		p[i] = b
	}
	r.done = true
	return len(r.Msg), nil
}
*/

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
func (m MockBroker) Provision(uuid.UUID, *broker.ProvisionRequest) (*broker.ProvisionResponse, error) {
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

func TestNewHandler(t *testing.T) {

	b := MockBroker{Name: "testbroker"}
	handler := NewHandler(b)
	ft.AssertNotNil(t, handler, "handler wasn't created")
}

func TestBootstrap(t *testing.T) {
	// create handler for testing
	b := MockBroker{Name: "testbroker"}
	handler := handler{*mux.NewRouter(), b}
	ft.AssertNotNil(t, handler, "")

	trr := TestRequest{Msg: "hello world", done: true}
	r := httptest.NewRequest("POST", "/v2/bootstrap", trr)
	w := httptest.NewRecorder()
	handler.bootstrap(w, r, nil)
	ft.AssertEqual(t, w.Code, 200, "code not equal")
}

func TestCatalog(t *testing.T) {
	// create handler for testing
	b := MockBroker{Name: "testbroker"}
	handler := handler{*mux.NewRouter(), b}
	ft.AssertNotNil(t, handler, "")

	trr := TestRequest{Msg: "hello world", done: true}
	r := httptest.NewRequest("GET", "/v2/catalog", trr)
	w := httptest.NewRecorder()
	handler.catalog(w, r, nil)
	ft.AssertEqual(t, w.Code, 200, "code not equal")
}

func TestProvisionCreate(t *testing.T) {
	// create handler for testing
	b := MockBroker{Name: "testbroker"}
	handler := handler{*mux.NewRouter(), b}

	trr := TestRequest{Msg: "{\"name\": \"hello world\"}"}
	r := httptest.NewRequest("PUT", "/v2/service_instance/688eea24-9cf9-43e3-9942-d1863b2a16af", trr)
	r.Header.Add("Content-Type", "application/json")
	w := httptest.NewRecorder()
	params := map[string]string{
		"instance_uuid": "688eea24-9cf9-43e3-9942-d1863b2a16af",
	}
	handler.provision(w, r, params)
	t.Log(w.Body)
	ft.AssertEqual(t, w.Code, 201, "provision not created")
}

func TestProvisionInvalidUUID(t *testing.T) {
	// create handler for testing
	b := MockBroker{Name: "testbroker"}
	handler := handler{*mux.NewRouter(), b}

	r := httptest.NewRequest("PUT", "/v2/service_instance/invaliduuid", nil)
	r.Header.Add("Content-Type", "application/json")
	w := httptest.NewRecorder()
	params := map[string]string{
		"instance_uuid": "invaliduuid",
	}
	handler.provision(w, r, params)
	ft.AssertEqual(t, w.Code, 400, "provision not created")
	ft.AssertError(t, w.Body, "invalid instance_uuid")
}

func TestProvisionCouldnotReadRequest(t *testing.T) {
	// create handler for testing
	b := MockBroker{Name: "testbroker"}
	handler := handler{*mux.NewRouter(), b}

	r := httptest.NewRequest("PUT", "/v2/service_instance/invaliduuid", nil)
	r.Header.Add("Content-Type", "application/json")
	w := httptest.NewRecorder()
	params := map[string]string{
		"instance_uuid": "688eea24-9cf9-43e3-9942-d1863b2a16af",
	}
	handler.provision(w, r, params)
	ft.AssertEqual(t, w.Code, 400, "provision not created")
	ft.AssertError(t, w.Body, "could not read request: EOF")
}

func TestProvisionConflict(t *testing.T) {
	t.SkipNow()
	// create handler for testing
	b := MockBroker{Name: "testbroker"}
	handler := handler{*mux.NewRouter(), b}

	trr := TestRequest{Msg: "{\"name\": \"hello world\"}"}
	r := httptest.NewRequest("PUT", "/v2/service_instance/688eea24-9cf9-43e3-9942-d1863b2a16af", trr)
	r.Header.Add("Content-Type", "application/json")
	w := httptest.NewRecorder()
	params := map[string]string{
		"instance_uuid": "688eea24-9cf9-43e3-9942-d1863b2a16af",
	}
	handler.provision(w, r, params)
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
	// create handler for testing
	b := MockBroker{Name: "testbroker"}
	handler := handler{*mux.NewRouter(), b}

	trr := TestRequest{Msg: "hello world", done: true}
	r := httptest.NewRequest("PUT", "/v2/service_instance/foo/service_bindings/bar", trr)
	w := httptest.NewRecorder()
	handler.bind(w, r, nil)
	ft.AssertEqual(t, w.Code, 400, "code not equal")
}
