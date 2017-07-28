package fusortest

import (
	"bytes"
	"encoding/json"
	"fmt"
	"regexp"
	"testing"
)

///////////////////////////////////////////////////////////////////////////////
/*
id: 8b1db903-f30a-4c05-bb60-ee2d73926d4c
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
    bindable: false
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
*/ //////////////////////////////////////////////////////////////////////////////

// EncodedApb - Returns a preincoded APB for repeated test usage
func EncodedApb() string {
	apb := `aWQ6IDhiMWRiOTAzLWYzMGEtNGMwNS1iYjYwLWVlMmQ3MzkyNmQ0YwpuYW1lOiBtZWRpYXdpa2kx
MjMtYXBiCmltYWdlOiBhbnNpYmxlcGxheWJvb2tidW5kbGUvbWVkaWF3aWtpMTIzLWFwYgpkZXNj
cmlwdGlvbjogIk1lZGlhd2lraTEyMyBhcGIgaW1wbGVtZW50YXRpb24iCmJpbmRhYmxlOiBmYWxz
ZQphc3luYzogb3B0aW9uYWwKbWV0YWRhdGE6CiAgZGlzcGxheW5hbWU6ICJSZWQgSGF0IE1lZGlh
d2lraSIKICBsb25nRGVzY3JpcHRpb246ICJBbiBhcGIgdGhhdCBkZXBsb3lzIE1lZGlhd2lraSAx
LjIzIgogIGltYWdlVVJMOiAiaHR0cHM6Ly91cGxvYWQud2lraW1lZGlhLm9yZy93aWtpcGVkaWEv
Y29tbW9ucy8wLzAxL01lZGlhV2lraS1zbWFsbGVyLWxvZ28ucG5nIgogIGRvY3VtZW50YXRpb25V
Ukw6ICJodHRwczovL3d3dy5tZWRpYXdpa2kub3JnL3dpa2kvRG9jdW1lbnRhdGlvbiIKcGxhbnM6
CiAgLSBuYW1lOiBkZXYKICAgIGRlc2NyaXB0aW9uOiAiTWVkaWF3aWtpMTIzIGFwYiBpbXBsZW1l
bnRhdGlvbiIKICAgIGZyZWU6IHRydWUKICAgIGJpbmRhYmxlOiBmYWxzZQogICAgbWV0YWRhdGE6
CiAgICAgIGRpc3BsYXlOYW1lOiBEZXZlbG9wbWVudAogICAgICBsb25nRGVzY3JpcHRpb246IEJh
c2ljIGRldmVsb3BtZW50IHBsYW4KICAgICAgY29zdDogJDAuMDAKICAgIHBhcmFtZXRlcnM6CiAg
ICAgIC0gbmFtZTogbWVkaWF3aWtpX2RiX3NjaGVtYQogICAgICAgIHRpdGxlOiBNZWRpYXdpa2kg
REIgU2NoZW1hCiAgICAgICAgdHlwZTogc3RyaW5nCiAgICAgICAgZGVmYXVsdDogbWVkaWF3aWtp
CiAgICAgICAgcmVxdWlyZWQ6IHRydWUKICAgICAgLSBuYW1lOiBtZWRpYXdpa2lfc2l0ZV9uYW1l
CiAgICAgICAgdGl0bGU6IE1lZGlhd2lraSBTaXRlIE5hbWUKICAgICAgICB0eXBlOiBzdHJpbmcK
ICAgICAgICBkZWZhdWx0OiBNZWRpYVdpa2kKICAgICAgICByZXF1aXJlZDogdHJ1ZQogICAgICAt
IG5hbWU6IG1lZGlhd2lraV9zaXRlX2xhbmcKICAgICAgICB0aXRsZTogTWVkaWF3aWtpIFNpdGUg
TGFuZ3VhZ2UKICAgICAgICB0eXBlOiBzdHJpbmcKICAgICAgICBkZWZhdWx0OiBlbgogICAgICAg
IHJlcXVpcmVkOiB0cnVlCiAgICAgIC0gbmFtZTogbWVkaWF3aWtpX2FkbWluX3VzZXIKICAgICAg
ICB0aXRsZTogTWVkaWF3aWtpIEFkbWluIFVzZXIKICAgICAgICB0eXBlOiBzdHJpbmcKICAgICAg
ICBkZWZhdWx0OiBhZG1pbgogICAgICAgIHJlcXVpcmVkOiB0cnVlCiAgICAgIC0gbmFtZTogbWVk
aWF3aWtpX2FkbWluX3Bhc3MKICAgICAgICB0aXRsZTogTWVkaWF3aWtpIEFkbWluIFVzZXIgUGFz
c3dvcmQKICAgICAgICB0eXBlOiBzdHJpbmcKICAgICAgICByZXF1aXJlZDogdHJ1ZQo=`

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
	t.Fatal(message)
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
