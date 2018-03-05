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
	"fmt"
	"testing"

	ft "github.com/stretchr/testify/assert"
)

func TestBuildExtractedCredentials(t *testing.T) {
	output := []byte("eyJkYiI6ICJmdXNvcl9ndWVzdGJvb2tfZGIiLCAidXNlciI6ICJkdWRlcl90d28iLCAicGFzcyI6ICJkb2c4dHdvIn0=")
	decoded, err := decodeOutput(output)
	if err != nil {
		t.Log(err.Error())
	}

	bd, _ := buildExtractedCredentials(decoded)
	ft.NotNil(t, bd, "credential is nil")
	ft.Equal(t, bd.Credentials["db"], "fusor_guestbook_db", "db is not fusor_guestbook_db")
	ft.Equal(t, bd.Credentials["user"], "duder_two", "user is not duder_two")
	ft.Equal(t, bd.Credentials["pass"], "dog8two", "password is not dog8two")
}

func TestExitGracefully(t *testing.T) {
	output := []byte("eyJkYiI6ICJmdXNvcl9ndWVzdGJvb2tfZGIiLCAidXNlciI6ICJkdWRlcl90d28iLCAicGFzcyI6ICJkb2c4dHdvIn0=")

	_, err := decodeOutput(output)
	ft.Equal(t, err, nil)
}

func TestInt(t *testing.T) {
	output := []byte("eyJEQl9OQU1FIjogImZvb2JhciIsICJEQl9QQVNTV09SRCI6ICJzdXBlcnNlY3JldCIsICJEQl9UWVBFIjogIm15c3FsIiwgIkRCX1BPUlQiOiAzMzA2LCAiREJfVVNFUiI6ICJkdWRlciIsICJEQl9IT1NUIjogIm15aW5zdGFuY2UuMTIzNDU2Nzg5MDEyLnVzLWVhc3QtMS5yZHMuYW1hem9uYXdzLmNvbSJ9")

	decoded, err := decodeOutput(output)
	if err != nil {
		t.Log(err.Error())
	}

	do := make(map[string]interface{})
	json.Unmarshal(decoded, &do)
	ft.Equal(t, do["DB_NAME"], "foobar", "name does not match")
	ft.Equal(t, do["DB_PASSWORD"], "supersecret", "password does not match")
	ft.Equal(t, do["DB_TYPE"], "mysql", "type does not match")
	ft.Equal(t, do["DB_PORT"], float64(3306), "port does not match")
	ft.Equal(t, do["DB_USER"], "duder", "user does not match")
	ft.Equal(t, do["DB_HOST"], "myinstance.123456789012.us-east-1.rds.amazonaws.com", "invalid hostname")
}

// didn't think this was generic enough to go in ft.
func assertError(t *testing.T, err error, verifystr string) {
	if err != nil {
		ft.Equal(t, err.Error(), verifystr, "error output didn't match expected output")
	} else {
		t.Fatal(fmt.Sprintf("method should return '%s' error", verifystr))
	}
}
