package apb

import (
	"testing"

	ft "github.com/openshift/ansible-service-broker/pkg/fusortest"
)

func TestCreateExtraVarsNilParametersRef(t *testing.T) {
	context := &Context{Platform: "kubernetes", Namespace: "testing-project"}

	t.Log("calling createExtraVars")
	value, err := createExtraVars(context, nil)
	if err != nil {
		t.Log("we have an error")
		t.Fatal(err)
	}

	t.Log(value)
	expected := "{\"namespace\":\"testing-project\"}"
	ft.AssertEqual(t, value, expected, "invalid value on nil parameters ref")
}

func TestCreateExtraVarsNilParameters(t *testing.T) {
	context := Context{Platform: "kubernetes", Namespace: "testing-project"}
	parameters := Parameters(nil)

	t.Log("calling createExtraVars")
	value, err := createExtraVars(&context, &parameters)
	if err != nil {
		t.Log("we have an error")
		t.Fatal(err)
	}

	t.Log(value)
	expected := "{\"namespace\":\"testing-project\"}"
	ft.AssertEqual(t, value, expected, "invalid value on nil parameters")
}

func TestCreateExtraVarsNilContextRef(t *testing.T) {
	parameters := &Parameters{"key": "param"}

	t.Log("calling createExtraVars")
	value, err := createExtraVars(nil, parameters)
	if err != nil {
		t.Log("we have an error")
		t.Fatal(err)
	}

	t.Log(value)
	expected := "{\"key\":\"param\"}"
	ft.AssertEqual(t, value, expected, "invalid value on empty context")
}

func TestCreateExtraVars(t *testing.T) {
	context := &Context{Platform: "kubernetes", Namespace: "testing-project"}
	parameters := &Parameters{"key": "param"}
	value, err := createExtraVars(context, parameters)
	if err != nil {
		t.Log("we have an error")
		t.Fatal(err)
	}

	expected := "{\"key\":\"param\",\"namespace\":\"testing-project\"}"
	ft.AssertEqual(t, value, expected, "extravars do not match")
}
