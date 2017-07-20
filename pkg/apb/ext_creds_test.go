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

// didn't think this was generic enough to go in ft.
func assertError(t *testing.T, err error, verifystr string) {
	if err != nil {
		ft.AssertEqual(t, err.Error(), verifystr, "error output didn't match expected output")
	} else {
		t.Fatal(fmt.Sprintf("method should return '%s' error", verifystr))
	}
}
