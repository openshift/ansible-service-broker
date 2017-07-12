package apb

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"reflect"
	"testing"

	ft "github.com/openshift/ansible-service-broker/pkg/fusortest"
	yaml "gopkg.in/yaml.v2"
)

// array can't be const
var SpecTags = []string{"latest", "old-release"}

const SpecID = "ab094014-b740-495e-b178-946d5aa97ebf"
const SpecName = "etherpad-apb"
const SpecImage = "fusor/etherpad-apb"
const SpecBindable = false
const SpecAsync = "optional"
const SpecDescription = "A note taking webapp"
const SpecRegistryName = "test"
const SpecParameters = `
	[
		{ "postgresql_database": { "default": "admin", "type": "string", "title": "PostgreSQL Database Name" } },
		{ "postgresql_password": { "default": "admin", "type": "string", "description": "A random alphanumeric string if left blank", "title": "PostgreSQL Password" } },
		{ "postgresql_user": { "default": "admin", "title": "PostgreSQL User", "type": "string", "maxlength": 63 } },
		{ "postgresql_version": { "default": 9.5, "enum": [ "9.5", "9.4" ], "type": "enum", "title": "PostgreSQL Version" } },
		{ "postgresql_email": { "pattern": "\u201c^\\\\S+@\\\\S+$\u201d", "type": "string", "description": "email address", "title": "email" } }
	]
`

var expectedSpecParameters = []map[string]*ParameterDescriptor{
	map[string]*ParameterDescriptor{
		"postgresql_database": &ParameterDescriptor{
			Default: "admin",
			Type:    "string",
			Title:   "PostgreSQL Database Name"}},
	map[string]*ParameterDescriptor{
		"postgresql_password": &ParameterDescriptor{
			Default:     "admin",
			Type:        "string",
			Description: "A random alphanumeric string if left blank",
			Title:       "PostgreSQL Password"}},
	map[string]*ParameterDescriptor{
		"postgresql_user": &ParameterDescriptor{
			Default:   "admin",
			Title:     "PostgreSQL User",
			Type:      "string",
			Maxlength: 63}},
	map[string]*ParameterDescriptor{
		"postgresql_version": &ParameterDescriptor{
			Default: 9.5,
			Enum:    []string{"9.5", "9.4"},
			Type:    "enum",
			Title:   "PostgreSQL Version"}},
	map[string]*ParameterDescriptor{
		"postgresql_email": &ParameterDescriptor{
			Pattern:     "\u201c^\\\\S+@\\\\S+$\u201d",
			Type:        "string",
			Description: "email address",
			Title:       "email"}},
}

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
`, SpecID, SpecDescription, SpecName, SpecImage, convertedSpecTags, SpecBindable, SpecAsync, SpecParameters)

func TestSpecLoadJSON(t *testing.T) {

	s := Spec{}
	err := LoadJSON(SpecJSON, &s)
	if err != nil {
		panic(err)
	}

	ft.AssertEqual(t, s.ID, SpecID)
	ft.AssertEqual(t, s.Description, SpecDescription)
	ft.AssertEqual(t, s.FQName, SpecName)
	ft.AssertEqual(t, s.Image, SpecImage)
	ft.AssertEqual(t, s.Bindable, SpecBindable)
	ft.AssertEqual(t, s.Async, SpecAsync)
	ft.AssertTrue(t, reflect.DeepEqual(s.Parameters, expectedSpecParameters))

}

func TestSpecDumpJSON(t *testing.T) {
	s := Spec{
		ID:          SpecID,
		Description: SpecDescription,
		FQName:      SpecName,
		Image:       SpecImage,
		Tags:        SpecTags,
		Bindable:    SpecBindable,
		Async:       SpecAsync,
		Parameters:  expectedSpecParameters,
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

func TestEncodedParameters(t *testing.T) {
	encodedstring :=
		`
aWQ6IDU1YzUzYTVkLTY1YTYtNGMyNy04OGZjLWUwMjc0MTBiMTMzNwpuYW1lOiBtZWRpYXdpa2kx
MjMtYXBiCmltYWdlOiBhbnNpYmxlcGxheWJvb2tidW5kbGUvbWVkaWF3aWtpMTIzLWFwYgpkZXNj
cmlwdGlvbjogIk1lZGlhd2lraTEyMyBhcGIgaW1wbGVtZW50YXRpb24iCmJpbmRhYmxlOiBmYWxz
ZQphc3luYzogb3B0aW9uYWwKbWV0YWRhdGE6CiAgZGlzcGxheW5hbWU6ICJSZWQgSGF0IE1lZGlh
d2lraSIKICBsb25nRGVzY3JpcHRpb246ICJBbiBhcGIgdGhhdCBkZXBsb3lzIE1lZGlhd2lraSAx
LjIzIgogIGltYWdlVVJMOiAiaHR0cHM6Ly91cGxvYWQud2lraW1lZGlhLm9yZy93aWtpcGVkaWEv
Y29tbW9ucy8wLzAxL01lZGlhV2lraS1zbWFsbGVyLWxvZ28ucG5nIgogIGRvY3VtZW50YXRpb25V
Ukw6ICJodHRwczovL3d3dy5tZWRpYXdpa2kub3JnL3dpa2kvRG9jdW1lbnRhdGlvbiIKcGFyYW1l
dGVyczoKICAtIG1lZGlhd2lraV9kYl9zY2hlbWE6CiAgICAgIHRpdGxlOiBNZWRpYXdpa2kgREIg
U2NoZW1hCiAgICAgIHR5cGU6IHN0cmluZwogICAgICBkZWZhdWx0OiBtZWRpYXdpa2kKICAtIG1l
ZGlhd2lraV9zaXRlX25hbWU6CiAgICAgIHRpdGxlOiBNZWRpYXdpa2kgU2l0ZSBOYW1lCiAgICAg
IHR5cGU6IHN0cmluZwogICAgICBkZWZhdWx0OiBNZWRpYVdpa2kKICAtIG1lZGlhd2lraV9zaXRl
X2xhbmc6CiAgICAgIHRpdGxlOiBNZWRpYXdpa2kgU2l0ZSBMYW5ndWFnZQogICAgICB0eXBlOiBz
dHJpbmcKICAgICAgZGVmYXVsdDogZW4KICAtIG1lZGlhd2lraV9hZG1pbl91c2VyOgogICAgICB0
aXRsZTogTWVkaWF3aWtpIEFkbWluIFVzZXIKICAgICAgdHlwZTogc3RyaW5nCiAgICAgIGRlZmF1
bHQ6IGFkbWluCiAgLSBtZWRpYXdpa2lfYWRtaW5fcGFzczoKICAgICAgdGl0bGU6IE1lZGlhd2lr
aSBBZG1pbiBVc2VyIFBhc3N3b3JkCiAgICAgIHR5cGU6IHN0cmluZwpyZXF1aXJlZDoKICAtIG1l
ZGlhd2lraV9kYl9zY2hlbWEKICAtIG1lZGlhd2lraV9zaXRlX25hbWUKICAtIG1lZGlhd2lraV9z
aXRlX2xhbmcKICAtIG1lZGlhd2lraV9hZG1pbl91c2VyCiAgLSBtZWRpYXdpa2lfYWRtaW5fcGFz
cwo=`

	decodedyaml, err := base64.StdEncoding.DecodeString(encodedstring)
	if err != nil {
		t.Fatal(err)
	}

	spec := &Spec{}
	if err = yaml.Unmarshal(decodedyaml, spec); err != nil {
		t.Fatal(err)
	}
	fmt.Printf("%#v", spec)
	ft.AssertEqual(t, spec.FQName, "mediawiki123-apb")
	ft.AssertEqual(t, len(spec.Parameters), 5)
	ft.AssertNotNil(t, spec.Required)

	// picking something other than the first one
	sitelang := spec.Parameters[2]["mediawiki_site_lang"]

	ft.AssertEqual(t, sitelang.Title, "Mediawiki Site Language")
	ft.AssertEqual(t, sitelang.Type, "string")
	ft.AssertEqual(t, sitelang.Description, "")
	ft.AssertEqual(t, sitelang.Default, "en")
	ft.AssertEqual(t, sitelang.Maxlength, 0)
	ft.AssertEqual(t, sitelang.Pattern, "")
	ft.AssertEqual(t, len(sitelang.Enum), 0)

	// example of traversing the parameters
	// comment left on purpose
	/*
		for _, pm := range spec.Parameters {
			for k, pd := range pm {
				t.Log(k)
				t.Log(fmt.Sprintf("\tTitle: %s", pd.Title))
				t.Log(fmt.Sprintf("\tType: %s", pd.Type))
				t.Log(fmt.Sprintf("\tDescription: %s", pd.Description))
				t.Log(fmt.Sprintf("\tDefault: %v", pd.Default))
				t.Log(fmt.Sprintf("\tMaxlength: %d", pd.Maxlength))
				t.Log(fmt.Sprintf("\tPattern: %s", pd.Pattern))
				t.Log(fmt.Sprintf("\tEnum: %v", pd.Enum))
			}
		}
	*/
}
