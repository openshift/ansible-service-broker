package broker

import (
	"encoding/base64"
	"testing"

	schema "github.com/lestrrat/go-jsschema"
	"github.com/openshift/ansible-service-broker/pkg/apb"
	ft "github.com/openshift/ansible-service-broker/pkg/fusortest"
	yaml "gopkg.in/yaml.v2"
)

func TestEnumIsCopied(t *testing.T) {
	params := []map[string]*apb.ParameterDescriptor{
		map[string]*apb.ParameterDescriptor{
			"email_address": &apb.ParameterDescriptor{
				Title:       "Email Address",
				Type:        "enum",
				Description: "example enum parameter",
				Enum:        []string{"google@gmail.com", "redhat@redhat.com"},
				Default:     float64(9001)}}}

	schemaObj := ParametersToSchema(params, []string{})

	emailParam := schemaObj.ServiceInstance.Create["parameters"].Properties["email_address"]
	ft.AssertEqual(t, len(emailParam.Enum), 2, "enum mismatch")
	ft.AssertEqual(t, emailParam.Enum[0], "google@gmail.com")
	ft.AssertEqual(t, emailParam.Enum[1], "redhat@redhat.com")

}

func TestSpecToService(t *testing.T) {
	param := []map[string]*apb.ParameterDescriptor{
		map[string]*apb.ParameterDescriptor{
			"hostport": &apb.ParameterDescriptor{
				Title:       "Host Port",
				Type:        "int",
				Description: "The host TCP port as the external end point",
				Default:     float64(9001)}}}

	spec := apb.Spec{
		ID:          "50eb5637-6ffe-480d-a52e-a7e603a50fca",
		FQName:      "testspec",
		Bindable:    false,
		Description: "test spec to be converted",
		Async:       "unsupported",
		Parameters:  param}

	descriptors := make(map[string]interface{})
	descriptors["parameters"] = param

	expectedsvc := Service{
		ID:          "50eb5637-6ffe-480d-a52e-a7e603a50fca",
		Name:        "testspec",
		Description: "test spec to be converted",
		Bindable:    false,
		Plans:       nil,
		Metadata:    descriptors,
	}
	svc := SpecToService(&spec)
	ft.AssertEqual(t, svc.Name, expectedsvc.Name, "name is not equal")
	ft.AssertEqual(t, svc.Description, expectedsvc.Description, "description is not equal")
	ft.AssertEqual(t, svc.Bindable, expectedsvc.Bindable, "bindable wrong")
}

func TestParametersToSchema(t *testing.T) {
	// being lazy, easy way to create a spec with parameters
	encodedstring :=
		`aWQ6IDU1YzUzYTVkLTY1YTYtNGMyNy04OGZjLWUwMjc0MTBiMTMzNwpuYW1lOiBtZWRpYXdpa2kx
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

	spec := &apb.Spec{}
	if err = yaml.Unmarshal(decodedyaml, spec); err != nil {
		t.Fatal(err)
	}
	required := []string{"mediawiki_site_lang"}
	schemaObj := ParametersToSchema(spec.Parameters, required)

	found := false
	for k, p := range schemaObj.ServiceInstance.Create["parameters"].Properties {
		// let's verify the site language
		if k == "mediawiki_site_lang" {
			found = true
			ft.AssertEqual(t, p.Title, "Mediawiki Site Language", "title mismatch")
			ft.AssertTrue(t, p.Type.Contains(schema.StringType), "type mismatch")
			ft.AssertEqual(t, p.Description, "", "description mismatch")
			ft.AssertEqual(t, p.Default, "en", "default mismatch")
			ft.AssertEqual(t, p.MaxLength.Val, 0, "maxlength mismatch")
			ft.AssertFalse(t, p.MaxLength.Initialized, "maxlength initialized")
			ft.AssertEqual(t, len(p.Enum), 0, "enum mismatch")
		}
	}
	ft.AssertTrue(t, found, "no mediawiki_site_lang property found")
	ft.AssertEqual(t, len(schemaObj.ServiceInstance.Create["parameters"].Required),
		len(required), "required len mismatch")
	ft.AssertEqual(t, schemaObj.ServiceInstance.Create["parameters"].Required[0],
		required[0], "required mismatch")
}

func TestGetType(t *testing.T) {
	ft.AssertTrue(t, getType("string").Contains(schema.StringType), "no string type")
	ft.AssertTrue(t, getType("enum").Contains(schema.StringType), "no enum type")
	ft.AssertTrue(t, getType("int").Contains(schema.IntegerType), "no int type")
	ft.AssertTrue(t, getType("object").Contains(schema.ObjectType), "no object type")
	ft.AssertTrue(t, getType("array").Contains(schema.ArrayType), "no array type")
	ft.AssertTrue(t, getType("bool").Contains(schema.BooleanType), "no bool type")
	ft.AssertTrue(t, getType("boolean").Contains(schema.BooleanType), "no boolean type")
	ft.AssertTrue(t, getType("number").Contains(schema.NumberType), "no number type")
	ft.AssertTrue(t, getType("nil").Contains(schema.NullType), "no nil type")
	ft.AssertTrue(t, getType("null").Contains(schema.NullType), "no null type")
	ft.AssertTrue(t, getType("biteme").Contains(schema.UnspecifiedType), "biteme type returned a known type")
}

func TestState(t *testing.T) {
	state := StateToLastOperation(apb.StateInProgress)
	ft.AssertEqual(t, state, LastOperationStateInProgress, "should be in progress")
	state = StateToLastOperation(apb.StateSucceeded)
	ft.AssertEqual(t, state, LastOperationStateSucceeded, "should be succeeded")
	state = StateToLastOperation(apb.StateFailed)
	ft.AssertEqual(t, state, LastOperationStateFailed, "should be failed")
}
