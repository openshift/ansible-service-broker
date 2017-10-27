//
// Copyright (c) 2017 Red Hat, Inc.
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
// Red Hat trademarks are not licensed under Apache License, Version 2.
// No permission is granted to use or replicate Red Hat trademarks that
// are incorporated in this software or its documentation.
//

package fusortest

import (
	"bytes"
	"encoding/json"
	"fmt"
	"regexp"
	"runtime"
	"strings"
	"testing"
)

///////////////////////////////////////////////////////////////////////////////
/*
name: mediawiki123-apb
image: ansibleplaybookbundle/mediawiki123-apb
description: "Mediawiki123 apb implementation"
bindable: false
async: optional
metadata:
  displayname: "Red Hat Mediawiki"
  longDescription: "An apb that deploys Mediawiki 1.23"
  imageURL: "https://upload.wikimedia.org/wikipedia/commons/0/01/MediaWiki-smaller-logo.png"
  documentationURL: "https://www.mediawiki.org/wiki/Documentation"
plans:
  - name: dev
    description: "Mediawiki123 apb implementation"
    free: true
    bindable: true
    metadata:
      displayName: Development
      longDescription: Basic development plan
      cost: $0.00
    parameters:
      - name: mediawiki_db_schema
        title: Mediawiki DB Schema
        type: string
        default: mediawiki
        required: true
      - name: mediawiki_site_name
        title: Mediawiki Site Name
        type: string
        default: MediaWiki
        required: true
        updatable: true
      - name: mediawiki_site_lang
        title: Mediawiki Site Language
        type: string
        default: en
        required: true
      - name: mediawiki_admin_user
        title: Mediawiki Admin User
        type: string
        default: admin
        required: true
      - name: mediawiki_admin_pass
        title: Mediawiki Admin User Password
        type: string
        required: true
    bind_parameters:
      - name: bind_param_1
        title: Bind Param 1
        type: string
        required: true
      - name: bind_param_2
        title: Bind Param 2
        type: int
        required: true
      - name: bind_param_3
        title: Bind Param 3
        type: string
*/ //////////////////////////////////////////////////////////////////////////////

// EncodedApb - Returns a preincoded APB for repeated test usage
func EncodedApb() string {
	apb := `bmFtZTogbWVkaWF3aWtpMTIzLWFwYgppbWFnZTogYW5zaWJsZXBsYXlib29rYnVuZGxlL21lZGlh
d2lraTEyMy1hcGIKZGVzY3JpcHRpb246ICJNZWRpYXdpa2kxMjMgYXBiIGltcGxlbWVudGF0aW9u
IgpiaW5kYWJsZTogZmFsc2UKYXN5bmM6IG9wdGlvbmFsCm1ldGFkYXRhOgogIGRpc3BsYXluYW1l
OiAiUmVkIEhhdCBNZWRpYXdpa2kiCiAgbG9uZ0Rlc2NyaXB0aW9uOiAiQW4gYXBiIHRoYXQgZGVw
bG95cyBNZWRpYXdpa2kgMS4yMyIKICBpbWFnZVVSTDogImh0dHBzOi8vdXBsb2FkLndpa2ltZWRp
YS5vcmcvd2lraXBlZGlhL2NvbW1vbnMvMC8wMS9NZWRpYVdpa2ktc21hbGxlci1sb2dvLnBuZyIK
ICBkb2N1bWVudGF0aW9uVVJMOiAiaHR0cHM6Ly93d3cubWVkaWF3aWtpLm9yZy93aWtpL0RvY3Vt
ZW50YXRpb24iCnBsYW5zOgogIC0gbmFtZTogZGV2CiAgICBkZXNjcmlwdGlvbjogIk1lZGlhd2lr
aTEyMyBhcGIgaW1wbGVtZW50YXRpb24iCiAgICBmcmVlOiB0cnVlCiAgICBiaW5kYWJsZTogdHJ1
ZQogICAgbWV0YWRhdGE6CiAgICAgIGRpc3BsYXlOYW1lOiBEZXZlbG9wbWVudAogICAgICBsb25n
RGVzY3JpcHRpb246IEJhc2ljIGRldmVsb3BtZW50IHBsYW4KICAgICAgY29zdDogJDAuMDAKICAg
IHBhcmFtZXRlcnM6CiAgICAgIC0gbmFtZTogbWVkaWF3aWtpX2RiX3NjaGVtYQogICAgICAgIHRp
dGxlOiBNZWRpYXdpa2kgREIgU2NoZW1hCiAgICAgICAgdHlwZTogc3RyaW5nCiAgICAgICAgZGVm
YXVsdDogbWVkaWF3aWtpCiAgICAgICAgcmVxdWlyZWQ6IHRydWUKICAgICAgLSBuYW1lOiBtZWRp
YXdpa2lfc2l0ZV9uYW1lCiAgICAgICAgdGl0bGU6IE1lZGlhd2lraSBTaXRlIE5hbWUKICAgICAg
ICB0eXBlOiBzdHJpbmcKICAgICAgICBkZWZhdWx0OiBNZWRpYVdpa2kKICAgICAgICByZXF1aXJl
ZDogdHJ1ZQogICAgICAtIG5hbWU6IG1lZGlhd2lraV9zaXRlX2xhbmcKICAgICAgICB0aXRsZTog
TWVkaWF3aWtpIFNpdGUgTGFuZ3VhZ2UKICAgICAgICB0eXBlOiBzdHJpbmcKICAgICAgICBkZWZh
dWx0OiBlbgogICAgICAgIHJlcXVpcmVkOiB0cnVlCiAgICAgIC0gbmFtZTogbWVkaWF3aWtpX2Fk
bWluX3VzZXIKICAgICAgICB0aXRsZTogTWVkaWF3aWtpIEFkbWluIFVzZXIKICAgICAgICB0eXBl
OiBzdHJpbmcKICAgICAgICBkZWZhdWx0OiBhZG1pbgogICAgICAgIHJlcXVpcmVkOiB0cnVlCiAg
ICAgIC0gbmFtZTogbWVkaWF3aWtpX2FkbWluX3Bhc3MKICAgICAgICB0aXRsZTogTWVkaWF3aWtp
IEFkbWluIFVzZXIgUGFzc3dvcmQKICAgICAgICB0eXBlOiBzdHJpbmcKICAgICAgICByZXF1aXJl
ZDogdHJ1ZQogICAgYmluZF9wYXJhbWV0ZXJzOgogICAgICAtIG5hbWU6IGJpbmRfcGFyYW1fMQog
ICAgICAgIHRpdGxlOiBCaW5kIFBhcmFtIDEKICAgICAgICB0eXBlOiBzdHJpbmcKICAgICAgICBy
ZXF1aXJlZDogdHJ1ZQogICAgICAtIG5hbWU6IGJpbmRfcGFyYW1fMgogICAgICAgIHRpdGxlOiBC
aW5kIFBhcmFtIDIKICAgICAgICB0eXBlOiBpbnQKICAgICAgICByZXF1aXJlZDogdHJ1ZQogICAg
ICAtIG5hbWU6IGJpbmRfcGFyYW1fMwogICAgICAgIHRpdGxlOiBCaW5kIFBhcmFtIDMKICAgICAg
ICB0eXBlOiBzdHJpbmcKCg==`
	return apb
}

////////////////////////////////////////////////////////////////////////////////

// Condition - the function to use for testing.
type Condition func(a interface{}, b interface{}) bool

func assert(t *testing.T, a interface{}, b interface{}, message []string, test Condition) {
	var msg string
	if len(message) != 0 {
		msg = message[0]
	}

	if !test(a, b) {
		fail(t, a, b, msg)
	}
}

func fail(t *testing.T, a interface{}, b interface{}, message string) {
	if len(message) == 0 {
		message = fmt.Sprintf("%v != %v", a, b)
	}
	_, file, line, _ := runtime.Caller(3)
	t.Fatal(fmt.Sprintf("\n%v had error on line %v\nMessage: %v", file[strings.LastIndex(file, "/")+1:], line, message))
}

// AssertEqual - Assert that inputs are equivalent.
func AssertEqual(t *testing.T, a interface{}, b interface{}, message ...string) {
	assert(t, a, b, message, func(a interface{}, b interface{}) bool {
		return a == b
	})
}

// AssertNotEqual - Assert that inputs are not equivalent
func AssertNotEqual(t *testing.T, a interface{}, b interface{}, message ...string) {
	assert(t, a, b, message, func(a interface{}, b interface{}) bool {
		return a != b
	})
}

// AssertTrue - Assert that input is true
func AssertTrue(t *testing.T, a interface{}, message ...string) {
	if a == true {
		return
	}

	var msg string
	if len(message) != 0 {
		msg = message[0]
	}

	if len(message) == 0 {
		msg = fmt.Sprintf("%v is not true!", a)
	}
	t.Fatal(msg)
}

// AssertFalse - Assert that input is false
func AssertFalse(t *testing.T, a interface{}, message ...string) {
	if a == false {
		return
	}

	var msg string
	if len(message) != 0 {
		msg = message[0]
	}

	if len(message) == 0 {
		msg = fmt.Sprintf("%v is not true!", a)
	}
	t.Fatal(msg)
}

// AssertNotNil - Assert that input is not nil
func AssertNotNil(t *testing.T, a interface{}, message ...string) {
	if a != nil {
		return
	}

	var msg string
	if len(message) != 0 {
		msg = message[0]
	}

	if len(message) == 0 {
		msg = fmt.Sprintf("%v is nil!", a)
	}
	t.Fatal(msg)
}

// AssertNil - Assert input is nil
func AssertNil(t *testing.T, a interface{}, message ...string) {
	if a == nil {
		return
	}

	var msg string
	if len(message) != 0 {
		msg = message[0]
	}

	if len(message) == 0 {
		msg = fmt.Sprintf("%v is not nil!", a)
	}
	t.Fatal(msg)
}

// AssertError - Assert that input is an error response
func AssertError(t *testing.T, body *bytes.Buffer, msg string) {
	var errResp = make(map[string]string)

	if body == nil {
		t.Fatal("invalid response body")
	}

	json.Unmarshal(body.Bytes(), &errResp)
	if errResp["description"] != msg {
		t.Log(errResp["description"])
		t.Fatal("error message does not match")
	}
}

// AssertState - Assert that state contianed in the body is of a certain state.
func AssertState(t *testing.T, body *bytes.Buffer, state string) {
	var resp = make(map[string]string)

	if body == nil {
		t.Fatal("invalid response body")
	}

	json.Unmarshal(body.Bytes(), &resp)
	if resp["state"] != state {
		t.Log(resp["state"])
		t.Fatal("state does not match")
	}
}

// AssertOperation -  Assert that the operation contained in the body is of a certain operation.
func AssertOperation(t *testing.T, body *bytes.Buffer, op string) {
	var resp = make(map[string]string)

	if body == nil {
		t.Fatal("invalid response body")
	}

	json.Unmarshal(body.Bytes(), &resp)
	if resp["operation"] != op {
		t.Log(resp["operation"])
		t.Fatal("state does not match")
	}
}

// StripNewline - String all new lines from string.
func StripNewline(input string) string {
	re := regexp.MustCompile("\\n")
	return re.ReplaceAllString(input, "")
}

// MinifyJSON - Minify the json outputted.
func MinifyJSON(input string) string {
	var mm interface{}
	json.Unmarshal([]byte(input), &mm)
	fmt.Println("before")
	fmt.Printf("%v", mm)
	fmt.Println("after")
	dat, _ := json.Marshal(mm)
	return string(dat)
}
