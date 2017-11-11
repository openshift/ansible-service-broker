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
