package broker

import (
	"encoding/base64"
	"testing"

	"strings"

	schema "github.com/lestrrat/go-jsschema"
	"github.com/openshift/ansible-service-broker/pkg/apb"
	ft "github.com/openshift/ansible-service-broker/pkg/fusortest"
	yaml "gopkg.in/yaml.v2"
)

const PlanName = "dev"
const PlanDescription = "Basic development plan"

var PlanMetadata = map[string]interface{}{
	"displayName":     "Development",
	"longDescription": PlanDescription,
	"cost":            "$0.00",
}

const PlanFree = true
const PlanBindable = true

var PlanParams = []apb.ParameterDescriptor{
	{
		Name:        "email_address",
		Title:       "Email Address",
		Type:        "enum",
		Description: "example enum parameter",
		Enum:        []string{"google@gmail.com", "redhat@redhat.com"},
		Default:     float64(9001),
	},
	{
		Name:        "password",
		Title:       "Password",
		Type:        "string",
		Description: "example string parameter with a display type",
		DisplayType: "password",
	},
	{
		Name:         "first_name",
		Title:        "First Name",
		Type:         "string",
		Description:  "example grouped string parameter",
		DisplayGroup: "User Information",
	},
	{
		Name:         "last_name",
		Title:        "Last Name",
		Type:         "string",
		Description:  "example grouped string parameter",
		DisplayGroup: "User Information",
	},
}

var p = apb.Plan{
	Name:        PlanName,
	Description: PlanDescription,
	Metadata:    PlanMetadata,
	Free:        PlanFree,
	Bindable:    PlanBindable,
	Parameters:  PlanParams,
}

func TestEnumIsCopied(t *testing.T) {

	schemaObj := parametersToSchema(PlanParams)

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
		Plans:       []apb.Plan{p}}

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

func TestUpdateMetadata(t *testing.T) {
	planMetadata := updateMetadata(PlanMetadata, PlanParams)
	ft.AssertNotNil(t, planMetadata, "plan metadata is empty")

	verifyFormDefinition(t, planMetadata, []string{"schemas", "service_instance", "create"})

	updateFormDefnMap := verifyMapPath(t, planMetadata, []string{"schemas", "service_instance", "update"})
	ft.AssertEqual(t, len(updateFormDefnMap), 0, "schemas.service_instance.update is not empty")

	verifyFormDefinition(t, planMetadata, []string{"schemas", "service_binding", "create"})
}

func verifyFormDefinition(t *testing.T, planMetadata map[string]interface{}, path []string) {

	formDefnMap := verifyMapPath(t, planMetadata, path)
	formDefnMetadata, correctType := formDefnMap["form_definition"].([]interface{})
	ft.AssertTrue(t, correctType, strings.Join(path, ".")+" Form definition is of the wrong type")
	ft.AssertNotNil(t, formDefnMetadata, "Form definition is nil")

	passwordParam, correctType := formDefnMetadata[1].(map[string]string)
	ft.AssertTrue(t, correctType, strings.Join(path, ".")+" Form definition password param is of the wrong type")
	ft.AssertNotNil(t, passwordParam)
	ft.AssertEqual(t, passwordParam["key"], PlanParams[1].Name, "Password parameter has the wrong name")
	ft.AssertEqual(t, passwordParam["type"], PlanParams[1].DisplayType, "Password parameter display type is incorrect")

	group, correctType := formDefnMetadata[2].(map[string]interface{})
	ft.AssertTrue(t, correctType, strings.Join(path, ".")+" Form definition parameter group is of the wrong type")
	ft.AssertNotNil(t, group, "Parameter group is empty")
	ft.AssertEqual(t, group["type"], "fieldset")

	groupedItems, correctType := group["items"].([]interface{})
	ft.AssertTrue(t, correctType, strings.Join(path, ".")+" Form definition parameter group items are the wrong type")
	ft.AssertNotNil(t, groupedItems, "Group missing parameter items")
	ft.AssertEqual(t, len(groupedItems), 2, "Incorrect number of parameters in group")

	firstNameParam, correctType := groupedItems[0].(string)
	ft.AssertTrue(t, correctType, "first_name is of the wrong type")
	ft.AssertEqual(t, firstNameParam, PlanParams[2].Name, "Incorrect name for first_name")

	lastNameParam, correctType := groupedItems[1].(string)
	ft.AssertTrue(t, correctType, "last_name is of the wrong type")
	ft.AssertEqual(t, lastNameParam, PlanParams[3].Name, "Incorrect name for last_name")
}

func verifyMapPath(t *testing.T, planMetadata map[string]interface{}, path []string) map[string]interface{} {
	currentMap := planMetadata
	var correctType bool
	for _, jsonKey := range path {
		currentMap, correctType = currentMap[jsonKey].(map[string]interface{})
		ft.AssertTrue(t, correctType, "incorrectly typed "+jsonKey+" metadata")
		ft.AssertNotNil(t, currentMap, jsonKey+" metadata empty")
	}

	return currentMap
}

func TestParametersToSchema(t *testing.T) {
	decodedyaml, err := base64.StdEncoding.DecodeString(ft.EncodedApb())
	if err != nil {
		t.Fatal(err)
	}

	spec := &apb.Spec{}
	if err = yaml.Unmarshal(decodedyaml, spec); err != nil {
		t.Fatal(err)
	}
	schemaObj := parametersToSchema(spec.Plans[0].Parameters)

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
