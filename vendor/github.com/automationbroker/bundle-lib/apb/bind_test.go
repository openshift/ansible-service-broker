//
// Copyright (c) 2018 Red Hat, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//    http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//

package apb

import (
	"encoding/json"
	"testing"

	ft "github.com/stretchr/testify/assert"
)

func TestBind(t *testing.T) {
	t.Skip("skipping bind until we can pass in a mock client")
	output := []byte(`
Login failed (401 Unauthorized)

PLAY [all] *********************************************************************

TASK [setup] *******************************************************************
ok: [localhost]

TASK [Bind] ********************************************************************
changed: [localhost]

TASK [debug] *******************************************************************
ok: [localhost] => {
    "msg": "eyJkYiI6ICJmdXNvcl9ndWVzdGJvb2tfZGIiLCAidXNlciI6ICJkdWRlcl90d28iLCAicGFzcyI6ICJkb2c4dHdvIn0="
}

PLAY RECAP *********************************************************************
localhost                  : ok=3    changed=1    unreachable=0    failed=0
`)
	decoded, err := decodeOutput(output)
	if err != nil {
		t.Fatal(err)
	}

	result := make(map[string]interface{})
	json.Unmarshal(decoded, &result)

	ft.NotNil(t, result, "result")
	ft.Equal(t, result["db"], "fusor_guestbook_db", "db is not fusor_guestbook_db")
	ft.Equal(t, result["user"], "duder_two", "user is not duder_two")
	ft.Equal(t, result["pass"], "dog8two", "password is not dog8two")
}
