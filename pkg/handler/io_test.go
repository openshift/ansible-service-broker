package handler

import (
	"io"
	"net/http/httptest"
	"testing"

	"github.com/fusor/ansible-service-broker/pkg/broker"
	ft "github.com/fusor/ansible-service-broker/pkg/fusortest"
)

type TestRequestReader struct {
	Msg  string
	done bool
}

func (r TestRequestReader) Read(p []byte) (n int, err error) {

	if r.done {
		return 0, io.EOF
	}
	for i, b := range []byte(r.Msg) {
		p[i] = b
	}
	r.done = true
	return len(r.Msg), nil
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

func TestReadRequest(t *testing.T) {
	var req *broker.ProvisionRequest

	trr := TestRequestReader{Msg: "{\"plan_id\": \"4c10ff43-be89-420a-9bab-27a9bef9aed8\",\"service_id\": \"f32de3bc-3225-429a-b23b-cef47ca1d25b\", \"parameters\": { \"MYSQL_USER\": \"username\"}}"}

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
