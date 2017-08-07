package apb

import (
	"fmt"
	"testing"

	ft "github.com/openshift/ansible-service-broker/pkg/fusortest"
)

func TestBuildExtractedCredentials(t *testing.T) {
	output := []byte("eyJkYiI6ICJmdXNvcl9ndWVzdGJvb2tfZGIiLCAidXNlciI6ICJkdWRlcl90d28iLCAicGFzcyI6ICJkb2c4dHdvIn0=")

	bd, _ := buildExtractedCredentials(output)
	ft.AssertNotNil(t, bd, "credential is nil")
	ft.AssertEqual(t, bd.Credentials["db"], "fusor_guestbook_db", "db is not fusor_guestbook_db")
	ft.AssertEqual(t, bd.Credentials["user"], "duder_two", "user is not duder_two")
	ft.AssertEqual(t, bd.Credentials["pass"], "dog8two", "password is not dog8two")
}

func TestExitGracefully(t *testing.T) {
	output := []byte("eyJkYiI6ICJmdXNvcl9ndWVzdGJvb2tfZGIiLCAidXNlciI6ICJkdWRlcl90d28iLCAicGFzcyI6ICJkb2c4dHdvIn0=")

	_, err := decodeOutput(output)
	ft.AssertEqual(t, err, nil)
}

func TestInt(t *testing.T) {
	output := []byte("eyJEQl9OQU1FIjogImZvb2JhciIsICJEQl9QQVNTV09SRCI6ICJzdXBlcnNlY3JldCIsICJEQl9UWVBFIjogIm15c3FsIiwgIkRCX1BPUlQiOiAzMzA2LCAiREJfVVNFUiI6ICJkdWRlciIsICJEQl9IT1NUIjogIm15aW5zdGFuY2UuMTIzNDU2Nzg5MDEyLnVzLWVhc3QtMS5yZHMuYW1hem9uYXdzLmNvbSJ9")

	do, err := decodeOutput(output)
	if err != nil {
		t.Log(err.Error())
	}
	ft.AssertEqual(t, do["DB_NAME"], "foobar", "name does not match")
	ft.AssertEqual(t, do["DB_PASSWORD"], "supersecret", "password does not match")
	ft.AssertEqual(t, do["DB_TYPE"], "mysql", "type does not match")
	ft.AssertEqual(t, do["DB_PORT"], float64(3306), "port does not match")
	ft.AssertEqual(t, do["DB_USER"], "duder", "user does not match")
	ft.AssertEqual(t, do["DB_HOST"], "myinstance.123456789012.us-east-1.rds.amazonaws.com", "invalid hostname")
}

// didn't think this was generic enough to go in ft.
func assertError(t *testing.T, err error, verifystr string) {
	if err != nil {
		ft.AssertEqual(t, err.Error(), verifystr, "error output didn't match expected output")
	} else {
		t.Fatal(fmt.Sprintf("method should return '%s' error", verifystr))
	}
}
