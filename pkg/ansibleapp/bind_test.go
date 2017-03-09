package ansibleapp

import (
	ft "github.com/fusor/ansible-service-broker/pkg/fusortest"
	"testing"
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

func TestBuildBindData(t *testing.T) {
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
	bd, _ := buildBindData(output)
	ft.AssertNotNil(t, bd, "binddata is nil")
	ft.AssertEqual(t, bd.Credentials["db"], "fusor_guestbook_db", "db is not fusor_guestbook_db")
	ft.AssertEqual(t, bd.Credentials["user"], "duder_two", "user is not duder_two")
	ft.AssertEqual(t, bd.Credentials["pass"], "dog8two", "password is not dog8two")
}
