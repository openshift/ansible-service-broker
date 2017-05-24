package apb

import (
	"fmt"
	"testing"

	ft "github.com/openshift/ansible-service-broker/pkg/fusortest"
)

func TestDecodeOutput(t *testing.T) {
	output := []byte(`
Login failed (401 Unauthorized)

PLAY [all] *********************************************************************

TASK [setup] *******************************************************************
ok: [localhost]

TASK [Bind] ********************************************************************
changed: [localhost]

TASK [debug] *******************************************************************
ok: [localhost] => {
    "msg": "<BIND_CREDENTIALS>eyJkYiI6ICJmdXNvcl9ndWVzdGJvb2tfZGIiLCAidXNlciI6ICJkdWRlcl90d28iLCAicGFzcyI6ICJkb2c4dHdvIn0=</BIND_CREDENTIALS>"
}

PLAY RECAP *********************************************************************
localhost                  : ok=3    changed=1    unreachable=0    failed=0
`)
	result, err := decodeOutput(output)
	if err != nil {
		t.Fatal(err)
	}

	ft.AssertNotNil(t, result, "result")
	ft.AssertEqual(t, result["db"], "fusor_guestbook_db", "db is not fusor_guestbook_db")
	ft.AssertEqual(t, result["user"], "duder_two", "user is not duder_two")
	ft.AssertEqual(t, result["pass"], "dog8two", "password is not dog8two")
}

func TestImageCantbePulled(t *testing.T) {
	output := []byte(`
Error from server (BadRequest): container "aa-425ef090-5f6f-4a0a-87ed-b072881d944d" in pod "aa-425ef090-5f6f-4a0a-87ed
-b072881d944d" is waiting to start: image can't be pulled
`)
	result, err := decodeOutput(output)
	if err == nil {
		t.Fatal("decode should've returned an error")
	}

	if result != nil {
		t.Fatal("result should've been nil")
	}

	assertError(t, err, "image can't be pulled")
}

func TestFailedMessageDecodeOutput(t *testing.T) {
	output := []byte(`
PLAY [Deploy rds-apb to openshift] *********************************************
TASK [setup] *******************************************************************
ok: [localhost]
TASK [rds-apb-openshift : set_fact] ********************************************
ok: [localhost]
TASK [rds-apb-openshift : rds] *************************************************
fatal: [localhost]: FAILED! => {"changed": false, "failed": true, "msg": "Region not specified. Unable to determine re
gion from EC2_REGION."}
        to retry, use: --limit @/opt/ansibleapp/actions/provision.retry
PLAY RECAP *********************************************************************
localhost                  : ok=2    changed=0    unreachable=0    failed=1
`)
	result, err := decodeOutput(output)
	if err == nil {
		t.Fatal("decode should've returned an error")
	}

	if result != nil {
		t.Fatal("result should've been nil")
	}
	assertError(t, err, "provision failed, INSERT MESSAGE HERE")
	// need to better parse the FAILED json returned.
}

func TestBuildExtractedCredentialsError(t *testing.T) {
	output := []byte(`
Login failed (401 Unauthorized)

PLAY [all] *********************************************************************

TASK [setup] *******************************************************************
ok: [localhost]

TASK [Bind] ********************************************************************
changed: [localhost]

TASK [debug] *******************************************************************
ok: [localhost] => {
    "msg": "<BIND_CREDENTIALS>eyJkYiI6ICJmdXNvcl9ndWVzdGJvb2tfZGIiLCAidXNlciI6ICJkdWRlcl90d28iLCAicGFzcyI6ICJkb2c4dHdvIn0="
}

PLAY RECAP *********************************************************************
localhost                  : ok=3    changed=1    unreachable=0    failed=0
`)
	bd, _ := buildExtractedCredentials(output)
	ft.AssertNotNil(t, bd, "credential is nil")
}

func TestBuildExtractedCredentials(t *testing.T) {
	output := []byte(`
Login failed (401 Unauthorized)

PLAY [all] *********************************************************************

TASK [setup] *******************************************************************
ok: [localhost]

TASK [Bind] ********************************************************************
changed: [localhost]

TASK [debug] *******************************************************************
ok: [localhost] => {
    "msg": "<BIND_CREDENTIALS>eyJkYiI6ICJmdXNvcl9ndWVzdGJvb2tfZGIiLCAidXNlciI6ICJkdWRlcl90d28iLCAicGFzcyI6ICJkb2c4dHdvIn0=</BIND_CREDENTIALS>"
}

PLAY RECAP *********************************************************************
localhost                  : ok=3    changed=1    unreachable=0    failed=0
`)
	bd, _ := buildExtractedCredentials(output)
	ft.AssertNotNil(t, bd, "credential is nil")
	ft.AssertEqual(t, bd.Credentials["db"], "fusor_guestbook_db", "db is not fusor_guestbook_db")
	ft.AssertEqual(t, bd.Credentials["user"], "duder_two", "user is not duder_two")
	ft.AssertEqual(t, bd.Credentials["pass"], "dog8two", "password is not dog8two")
}

func TestErrorDecodeOutput(t *testing.T) {
	output := []byte(`
	error: dial tcp [::1]:8443: getsockopt: connection refused

PLAY [all] *********************************************************************

TASK [setup] *******************************************************************
ok: [localhost]

TASK [Bind] ********************************************************************
fatal: [localhost]: FAILED! => {"changed": true, "cmd": "./bind", "delta": "0:00:00.115091", "end": "2017-03-13 14:55:28.434412", "failed": true, "rc": 1, "start": "2017-03-13 14:55:28.319321", "stderr": "", "stdout": "<BIND_ERROR>Malformed parameter input</BIND_ERROR>", "stdout_lines": ["<BIND_ERROR>Malformed parameter input</BIND_ERROR>"], "warnings": []}
        to retry, use: --limit @/opt/apb/actions/bind.retry

PLAY RECAP *********************************************************************
localhost                  : ok=1    changed=0    unreachable=0    failed=1
`)
	_, err := decodeOutput(output)
	assertError(t, err, "Malformed parameter input")
}

func TestExitGracefully(t *testing.T) {
	output := []byte(`
	error: dial tcp [::1]:8443: getsockopt: connection refused

PLAY [all] *********************************************************************
`)
	_, err := decodeOutput(output)
	ft.AssertEqual(t, err, nil)
}

func TestHandleOpenErrorBracket(t *testing.T) {
	t.Skip("REVISIT when monitoring output is redone")
	output := []byte(`
TASK [Bind] ******************<BIND_ERROR>**************************************
`)
	_, err := decodeOutput(output)
	assertError(t, err, "Unable to parse output")
}

func TestHandleOpenCredentialsBracket(t *testing.T) {
	t.Skip("REVISIT when monitoring output is redone")
	output := []byte(`
TASK [Bind] ******************<BIND_CREDENTIALS>**************************************
`)
	_, err := decodeOutput(output)
	assertError(t, err, "Unable to parse output")
}

func TestHandleCloseCredsBrackets(t *testing.T) {
	t.Skip("REVISIT when monitoring output is redone")
	output := []byte(`
TASK [Bind] ******************</BIND_CREDENTIALS>**************************************
`)
	_, err := decodeOutput(output)
	assertError(t, err, "Unable to parse output")
}

func TestHandleCloseErrorBrackets(t *testing.T) {
	t.Skip("REVISIT when monitoring output is redone")
	output := []byte(`
TASK [Bind] ******************</BIND_ERROR>**************************************
`)
	_, err := decodeOutput(output)
	assertError(t, err, "Unable to parse output")
}

// didn't think this was generic enough to go in ft.
func assertError(t *testing.T, err error, verifystr string) {
	if err != nil {
		ft.AssertEqual(t, err.Error(), verifystr, "error output didn't match expected output")
	} else {
		t.Fatal(fmt.Sprintf("method should return '%s' error", verifystr))
	}
}

func TestGetPodName(t *testing.T) {
	output := []byte(`pod "aa-03be586b-25fc-4d68-8591-51deed8765a1" created`)

	podname, _ := GetPodName(output, nil)

	ft.AssertEqual(t, podname, "aa-03be586b-25fc-4d68-8591-51deed8765a1", "podname does not match")
}
