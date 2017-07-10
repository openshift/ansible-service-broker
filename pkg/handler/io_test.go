package handler

import (
	"errors"
	"io"
	"net/http/httptest"
	"testing"

	"github.com/openshift/ansible-service-broker/pkg/broker"
	ft "github.com/openshift/ansible-service-broker/pkg/fusortest"
)

// test object to marshall in the request
type Foo struct {
	Msg  string
	Code int
}

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

func TestReadRequest(t *testing.T) {
	var req *broker.ProvisionRequest

	trr := TestRequest{Msg: "{\"plan_id\": \"4c10ff43-be89-420a-9bab-27a9bef9aed8\",\"service_id\": \"f32de3bc-3225-429a-b23b-cef47ca1d25b\", \"parameters\": { \"MYSQL_USER\": \"username\"}}"}

	r := httptest.NewRequest("PUT", "/does/not/matter", trr)
	r.Header.Add("Content-Type", "application/json")

	err := readRequest(r, &req)
	if err != nil {
		t.Fatal(err)
	}

	ft.AssertNotNil(t, r, "r")
	ft.AssertNotNil(t, req, "req")
	ft.AssertEqual(t, req.PlanID.String(), "4c10ff43-be89-420a-9bab-27a9bef9aed8", "planid doesn't match")
	ft.AssertEqual(t, req.ServiceID.String(), "f32de3bc-3225-429a-b23b-cef47ca1d25b", "serviceid doesn't match")
	ft.AssertEqual(t, req.Parameters["MYSQL_USER"], "username", "parameters don't match")
}

func TestInvalidContentType(t *testing.T) {
	var req *broker.ProvisionRequest
	r := httptest.NewRequest("PUT", "/does/not/matter", nil)
	err := readRequest(r, &req)
	if err == nil {
		t.Fatal(err)
	}
	ft.AssertEqual(t, err.Error(), "error: invalid content-type", "expected error")

}

func TestWriteResponse(t *testing.T) {
	expected := `{
  "Msg": "hello world",
  "Code": 10
}
`
	w := httptest.NewRecorder()
	tobj := Foo{Msg: "hello world", Code: 10}
	err := writeResponse(w, 200, tobj)
	if err != nil {
		t.Fatal(err)
	}
	ft.AssertEqual(t, w.Code, 200, "code not equal")
	ft.AssertEqual(t, w.Body.String(), expected, "body not equal")
}

func TestErrorResponse(t *testing.T) {
	expected := `{
  "Msg": "hello world",
  "Code": 10
}
`
	w := httptest.NewRecorder()
	tobj := Foo{Msg: "hello world", Code: 10}
	err := writeResponse(w, 200, tobj)
	if err != nil {
		t.Fatal(err)
	}
	ft.AssertEqual(t, w.Code, 200, "code not equal")
	ft.AssertEqual(t, w.Body.String(), expected, "body not equal")
}

func TestServerError(t *testing.T) {
	expected := `{
  "description": "failure is not an option"
}
`
	w := httptest.NewRecorder()
	tobj := Foo{Msg: "hello world", Code: 10}
	daerr := errors.New("failure is not an option")
	err := writeDefaultResponse(w, 200, tobj, daerr)
	if err != nil {
		t.Fatal(err)
	}
	ft.AssertEqual(t, w.Code, 500, "code should be ISE")
	ft.AssertEqual(t, w.Body.String(), expected, "body not equal")
}

func TestWriteDefaultResponse(t *testing.T) {
	expected := `{
  "Msg": "hello world",
  "Code": 10
}
`
	w := httptest.NewRecorder()
	tobj := Foo{Msg: "hello world", Code: 10}
	err := writeDefaultResponse(w, 200, tobj, nil)
	if err != nil {
		t.Fatal(err)
	}
	ft.AssertEqual(t, w.Code, 200, "code not equal")
	ft.AssertEqual(t, w.Body.String(), expected, "body not equal")
}
