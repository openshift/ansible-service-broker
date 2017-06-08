package broker

import (
	"encoding/base64"
	"fmt"
	"os"
	"path"
	"testing"

	schema "github.com/lestrrat/go-jsschema"
	"github.com/openshift/ansible-service-broker/pkg/apb"
	ft "github.com/openshift/ansible-service-broker/pkg/fusortest"
)

/*
func TestSpecToService(t *testing.T) {
	param := []*apb.ParameterDescriptor{
		&apb.ParameterDescriptor{
			Name:        "hostport",
			Description: "The host TCP port as the external end point",
			Default:     float64(9001),
			Type:        "foo",
			Required:    true}}

	spec := apb.Spec{
		Id:          "50eb5637-6ffe-480d-a52e-a7e603a50fca",
		Name:        "testspec",
		Bindable:    false,
		Description: "test spec to be converted",
		Async:       "unsupported",
		Parameters:  param}

	descriptors := make(map[string]interface{})
	descriptors["parameters"] = param

	expectedsvc := Service{
		ID:          uuid.Parse("50eb5637-6ffe-480d-a52e-a7e603a50fca"),
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
*/

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
	if err = apb.LoadYAML(string(decodedyaml), spec); err != nil {
		t.Fatal(err)
	}
	t.Log(fmt.Sprintf("%#v", spec.Parameters))

	schema := ParametersToSchema(spec.Parameters)
	t.Log(fmt.Sprintf("%#v", schema))
	t.Log(fmt.Sprintf("%#v", schema.ServiceInstance.Create["parameters"].Properties))
	t.Fatal("need to validate schema")
}

func TestGetType(t *testing.T) {
	// TODO: FIX TEST
	ft.AssertEqual(t, getType("string"), []schema.PrimitiveType{schema.StringType})
	ft.AssertEqual(t, getType("int"), []schema.PrimitiveType{schema.IntegerType})
	ft.AssertEqual(t, getType("object"), []schema.PrimitiveType{schema.ObjectType})
	ft.AssertEqual(t, getType("array"), []schema.PrimitiveType{schema.ArrayType})
	ft.AssertEqual(t, getType("enum"), []schema.PrimitiveType{schema.ArrayType})
	ft.AssertEqual(t, getType("bool"), []schema.PrimitiveType{schema.BooleanType})
	ft.AssertEqual(t, getType("boolean"), []schema.PrimitiveType{schema.BooleanType})
	ft.AssertEqual(t, getType("number"), []schema.PrimitiveType{schema.NumberType})
	ft.AssertEqual(t, getType("nil"), []schema.PrimitiveType{schema.NullType})
	ft.AssertEqual(t, getType("null"), []schema.PrimitiveType{schema.NullType})
	ft.AssertEqual(t, getType("biteme"), []schema.PrimitiveType{schema.UnspecifiedType})
}

func TestProjectRoot(t *testing.T) {
	gopath := os.Getenv("GOPATH")
	rootpath := path.Join(gopath, "src/github.com/openshift/ansible-service-broker")
	ft.AssertEqual(t, ProjectRoot(), rootpath, "paths not equal")
}

func TestState(t *testing.T) {
	state := StateToLastOperation(apb.StateInProgress)
	ft.AssertEqual(t, state, LastOperationStateInProgress, "should be in progress")
	state = StateToLastOperation(apb.StateSucceeded)
	ft.AssertEqual(t, state, LastOperationStateSucceeded, "should be succeeded")
	state = StateToLastOperation(apb.StateFailed)
	ft.AssertEqual(t, state, LastOperationStateFailed, "should be failed")
}
