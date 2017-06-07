package apb

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"reflect"
	"testing"

	ft "github.com/openshift/ansible-service-broker/pkg/fusortest"
)

// array can't be const
var SpecTags = []string{"latest", "old-release"}

const SpecId = "ab094014-b740-495e-b178-946d5aa97ebf"
const SpecName = "etherpad-apb"
const SpecImage = "fusor/etherpad-apb"
const SpecBindable = false
const SpecAsync = "optional"
const SpecDescription = "A note taking webapp"
const SpecParameters = `
	[
		{"name": "hostport", "description": "The host TCP port as the external endpoint", "default": 9001, "type": "foo", "required": true},
		{"name": "db_user", "description": "Database User", "default": "db_user", "type": "", "required": true},
		{"name": "db_pass", "description": "Database Password", "default": "db_pass", "type": "", "required": true},
		{"name": "db_name", "description": "Database Name", "default": "db_name", "type": "", "required": true},
		{"name": "db_host", "description": "Database service hostname/ip", "default": "mariadb", "type": "", "required": true},
		{"name": "db_port", "description": "Database service port", "default": 3306, "type": "", "required": true}
	]
`

var expectedSpecParameters = ""

/*
var expectedSpecParameters = map[string][]*ParameterDescriptor{
	"hostportkey": &ParameterDescriptor{Title: "hostport", Description: "The host TCP port as the external endpoint", Default: float64(9001), Type: "foo", Required: true},
	"db_userkey":  &ParameterDescriptor{Title: "db_user", Description: "Database User", Default: "db_user", Type: "", Required: true},
	"db_passkey":  &ParameterDescriptor{Title: "db_pass", Description: "Database Password", Default: "db_pass", Type: "", Required: true},
	"db_namekey":  &ParameterDescriptor{Title: "db_name", Description: "Database Name", Default: "db_name", Type: "", Required: true},
	"db_hostkey":  &ParameterDescriptor{Title: "db_host", Description: "Database service hostname/ip", Default: "mariadb", Type: "", Required: true},
	"db_portkey":  &ParameterDescriptor{Title: "db_port", Description: "Database service port", Default: float64(3306), Type: "", Required: true},
}
*/

var convertedSpecTags, _ = json.Marshal(SpecTags)

var SpecJSON = fmt.Sprintf(`
{
	"id": "%s",
	"description": "%s",
	"name": "%s",
	"image": "%s",
	"tags": %s,
	"bindable": %t,
	"async": "%s",
	"parameters": %s
}
`, SpecId, SpecDescription, SpecName, SpecImage, convertedSpecTags, SpecBindable, SpecAsync, SpecParameters)

func TestSpecLoadJSON(t *testing.T) {

	t.Skip("FIX ME WHEN YOU FINISH PARAMETER SCHEMA")
	s := Spec{}
	err := LoadJSON(SpecJSON, &s)
	if err != nil {
		panic(err)
	}

	ft.AssertEqual(t, s.Id, SpecId)
	ft.AssertEqual(t, s.Description, SpecDescription)
	ft.AssertEqual(t, s.Name, SpecName)
	ft.AssertEqual(t, s.Image, SpecImage)
	ft.AssertEqual(t, s.Bindable, SpecBindable)
	ft.AssertEqual(t, s.Async, SpecAsync)
	ft.AssertTrue(t, reflect.DeepEqual(s.Parameters, expectedSpecParameters))

}

func TestSpecDumpJSON(t *testing.T) {
	s := Spec{
		Id:          SpecId,
		Description: SpecDescription,
		Name:        SpecName,
		Image:       SpecImage,
		Tags:        SpecTags,
		Bindable:    SpecBindable,
		Async:       SpecAsync,
		//Parameters:  expectedSpecParameters,
	}

	var knownMap interface{}
	var subjectMap interface{}

	raw, err := DumpJSON(&s)
	if err != nil {
		panic(err)
	}
	json.Unmarshal([]byte(SpecJSON), &knownMap)
	json.Unmarshal([]byte(raw), &subjectMap)

	ft.AssertTrue(t, reflect.DeepEqual(knownMap, subjectMap))
}

func TestSpecLabel(t *testing.T) {
	ft.AssertEqual(t, BundleSpecLabel, "com.redhat.apb.spec", "spec label does not match dockerhub")
}

func TestFoobar(t *testing.T) {
	encodedstring :=
		`aWQ6IDU1YzUzYTVkLTY1YTYtNGMyNy04OGZjLWUwMjc0MTBiMTMzNwpuYW1lOiBtZWRpYXdpa2kx
MjMtYXBiCmltYWdlOiBhbnNpYmxlcGxheWJvb2tidW5kbGUvbWVkaWF3aWtpMTIzLWFwYgpkZXNj
cmlwdGlvbjogIk1lZGlhd2lraTEyMyBhcGIgaW1wbGVtZW50YXRpb24iCmJpbmRhYmxlOiBmYWxz
ZQphc3luYzogb3B0aW9uYWwKbWV0YWRhdGE6CiAgZGlzcGxheW5hbWU6ICJSZWQgSGF0IE1lZGlh
d2lraSIKICBsb25nRGVzY3JpcHRpb246ICJBbiBhcGIgdGhhdCBkZXBsb3lzIE1lZGlhd2lraSAx
LjIzIgogIGltYWdlVVJMOiAiaHR0cHM6Ly91cGxvYWQud2lraW1lZGlhLm9yZy93aWtpcGVkaWEv
Y29tbW9ucy8wLzAxL01lZGlhV2lraS1zbWFsbGVyLWxvZ28ucG5nIgogIGRvY3VtZW50YXRpb25V
Ukw6ICJodHRwczovL3d3dy5tZWRpYXdpa2kub3JnL3dpa2kvRG9jdW1lbnRhdGlvbiIKcGFyYW1l
dGVyczoKICAtIG1lZGlhd2lraV9kYl9zY2hlbWE6CiAgICAtIHRpdGxlOiBNZWRpYXdpa2kgREIg
U2NoZW1hCiAgICAgIHR5cGU6IHN0cmluZwogICAgICBkZWZhdWx0OiBtZWRpYXdpa2kKICAtIG1l
ZGlhd2lraV9zaXRlX25hbWU6CiAgICAtIHRpdGxlOiBNZWRpYXdpa2kgU2l0ZSBOYW1lCiAgICAg
IHR5cGU6IHN0cmluZwogICAgICBkZWZhdWx0OiBNZWRpYVdpa2kKICAtIG1lZGlhd2lraV9zaXRl
X2xhbmc6CiAgICAtIHRpdGxlOiBNZWRpYXdpa2kgU2l0ZSBMYW5ndWFnZQogICAgICB0eXBlOiBz
dHJpbmcKICAgICAgZGVmYXVsdDogZW4KICAtIG1lZGlhd2lraV9hZG1pbl91c2VyOgogICAgLSB0
aXRsZTogTWVkaWF3aWtpIEFkbWluIFVzZXIKICAgICAgdHlwZTogc3RyaW5nCiAgICAgIGRlZmF1
bHQ6IGFkbWluCiAgLSBtZWRpYXdpa2lfYWRtaW5fcGFzczoKICAgIC0gdGl0bGU6IE1lZGlhd2lr
aSBBZG1pbiBVc2VyIFBhc3N3b3JkCiAgICAgIHR5cGU6IHN0cmluZwpyZXF1aXJlZDoKICAtIG1l
ZGlhd2lraV9kYl9zY2hlbWEKICAtIG1lZGlhd2lraV9zaXRlX25hbWUKICAtIG1lZGlhd2lraV9z
aXRlX2xhbmcKICAtIG1lZGlhd2lraV9hZG1pbl91c2VyCiAgLSBtZWRpYXdpa2lfYWRtaW5fcGFz
cwo=`

	//fmt.Println("[" + encodedstring + "]")
	decodedyaml, err := base64.StdEncoding.DecodeString(encodedstring)
	if err != nil {
		t.Fatal(err)
	}
	t.Log(string(decodedyaml))
	spec := &Spec{}
	if err = LoadYAML(string(decodedyaml), spec); err != nil {
		t.Fatal(err)
	}
	t.Log(spec.Name)
	t.Log(len(spec.Parameters))
	t.Log(spec.Required)
	for _, pm := range spec.Parameters {
		for k, pd := range pm {
			t.Log(k)
			t.Log(len(pd))
			t.Log(fmt.Sprintf("\tTitle: %s", pd[0].Title))
			t.Log(fmt.Sprintf("\tType: %s", pd[0].Type))
			t.Log(fmt.Sprintf("\tDescription: %s", pd[0].Description))
			t.Log(fmt.Sprintf("\tDefault: %v", pd[0].Default))
			t.Log(fmt.Sprintf("\tMaxlength: %d", pd[0].Maxlength))
			t.Log(fmt.Sprintf("\tPattern: %s", pd[0].Pattern))
		}
	}

}
