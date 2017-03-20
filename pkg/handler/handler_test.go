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
	Name string
}

func (m MockBroker) Bootstrap() (*broker.BootstrapResponse, error) {
	return &broker.BootstrapResponse{10}, nil
}

func (m MockBroker) Catalog() (*broker.CatalogResponse, error) {
	return nil, nil
}
func (m MockBroker) Provision(uuid.UUID, *broker.ProvisionRequest) (*broker.ProvisionResponse, error) {
	return nil, nil
}
func (m MockBroker) Update(uuid.UUID, *broker.UpdateRequest) (*broker.UpdateResponse, error) {
	return nil, nil
}
func (m MockBroker) Deprovision(uuid.UUID) (*broker.DeprovisionResponse, error) {
	return nil, nil
}
func (m MockBroker) Bind(uuid.UUID, uuid.UUID, *broker.BindRequest) (*broker.BindResponse, error) {
	return nil, nil
}
func (m MockBroker) Unbind(uuid.UUID, uuid.UUID) error {
	return nil
}

func TestNewHandler(t *testing.T) {

	b := MockBroker{"testbroker"}
	handler := NewHandler(b)
	ft.AssertNotNil(t, handler, "handler wasn't created")
}

func TestBootstrap(t *testing.T) {
	// create handler for testing
	b := MockBroker{"testbroker"}
	handler := handler{*mux.NewRouter(), b}
	ft.AssertNotNil(t, handler, "")

	trr := TestRequest{Msg: "hello world", done: true}
	r := httptest.NewRequest("POST", "/v2/bootstrap", trr)
	w := httptest.NewRecorder()
	handler.bootstrap(w, r)
	ft.AssertEqual(t, w.Code, 200, "code not equal")
}

func TestCatalog(t *testing.T) {
	// create handler for testing
	b := MockBroker{"testbroker"}
	handler := handler{*mux.NewRouter(), b}
	ft.AssertNotNil(t, handler, "")

	trr := TestRequest{Msg: "hello world", done: true}
	r := httptest.NewRequest("GET", "/v2/catalog", trr)
	w := httptest.NewRecorder()
	handler.catalog(w, r)
	ft.AssertEqual(t, w.Code, 200, "code not equal")
}

func TestBindInvalidInstance(t *testing.T) {
	// create handler for testing
	b := MockBroker{"testbroker"}
	handler := handler{*mux.NewRouter(), b}
	ft.AssertNotNil(t, handler, "")

	trr := TestRequest{Msg: "hello world", done: true}
	r := httptest.NewRequest("PUT", "/v2/service_instance/foo/service_bindings/bar", trr)
	w := httptest.NewRecorder()
	handler.bind(w, r)
	ft.AssertEqual(t, w.Code, 400, "code not equal")
}
