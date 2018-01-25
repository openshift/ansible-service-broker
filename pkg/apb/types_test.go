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
	"encoding/base64"
	"encoding/json"
	"fmt"
	"reflect"
	"testing"

	ft "github.com/openshift/ansible-service-broker/pkg/fusortest"
	"github.com/pborman/uuid"
	yaml "gopkg.in/yaml.v2"
)

const PlanName = "dev"
const PlanDescription = "Mediawiki123 apb implementation"

var PlanMetadata = map[string]interface{}{
	"displayName":     "Development",
	"longDescription": "Basic development plan",
	"cost":            "$0.00",
}

const PlanFree = true
const PlanBindable = true

var PlanUpdatesTo = []string{"foo"}

var expectedPlanParameters = []ParameterDescriptor{
	ParameterDescriptor{
		Name:      "mediawiki_db_schema",
		Title:     "Mediawiki DB Schema",
		Type:      "string",
		Default:   "mediawiki",
		Updatable: false,
		Required:  true},
	ParameterDescriptor{
		Name:      "mediawiki_site_name",
		Title:     "Mediawiki Site Name",
		Type:      "string",
		Default:   "MediaWiki",
		Updatable: true,
		Required:  true},
	ParameterDescriptor{
		Name:      "mediawiki_site_lang",
		Title:     "Mediawiki Site Language",
		Type:      "string",
		Default:   "en",
		Updatable: false,
		Required:  true},
	ParameterDescriptor{
		Name:      "mediawiki_admin_user",
		Title:     "Mediawiki Admin User",
		Type:      "string",
		Default:   "admin",
		Updatable: false,
		Required:  true},
	ParameterDescriptor{
		Name:      "mediawiki_admin_pass",
		Title:     "Mediawiki Admin User Password",
		Type:      "string",
		Updatable: false,
		Required:  true},
}

var p = Plan{
	ID:          "",
	Name:        PlanName,
	Description: PlanDescription,
	Metadata:    PlanMetadata,
	Free:        PlanFree,
	Bindable:    PlanBindable,
	Parameters:  expectedPlanParameters,
	UpdatesTo:   PlanUpdatesTo,
}

const SpecVersion = "1.0"
const SpecRuntime = 1
const SpecName = "mediawiki123-apb"
const SpecImage = "ansibleplaybookbundle/mediawiki123-apb"
const SpecBindable = false
const SpecAsync = "optional"
const SpecDescription = "Mediawiki123 apb implementation"
const SpecPlans = `
[
	{
		"id": "",
		"name": "dev",
		"description": "Mediawiki123 apb implementation",
		"free": true,
		"bindable": true,
		"metadata": {
			"displayName": "Development",
			"longDescription": "Basic development plan",
			"cost": "$0.00"
		},
        "updates_to": ["foo"],
		"parameters": [
		{
			"name": "mediawiki_db_schema",
			"title": "Mediawiki DB Schema",
			"type": "string",
			"default": "mediawiki",
            "updatable": false,
			"required": true
		},
		{
			"name": "mediawiki_site_name",
			"title": "Mediawiki Site Name",
			"type": "string",
			"default": "MediaWiki",
            "updatable": true,
			"required": true
		},
		{
			"name": "mediawiki_site_lang",
			"title": "Mediawiki Site Language",
			"type": "string",
			"default": "en",
            "updatable": false,
			"required": true
		},
		{
			"name": "mediawiki_admin_user",
			"title": "Mediawiki Admin User",
			"type": "string",
			"default": "admin",
            "updatable": false,
			"required": true
		},
		{
			"name": "mediawiki_admin_pass",
			"title": "Mediawiki Admin User Password",
			"type": "string",
            "updatable": false,
			"required": true
		}
		]
	}
]
`

var SpecJSON = fmt.Sprintf(`
{
	"id": "",
	"tags": null,
	"description": "%s",
	"version": "%s",
	"runtime": %d,
	"name": "%s",
	"image": "%s",
	"bindable": %t,
	"async": "%s",
	"plans": %s
}
`, SpecDescription, SpecVersion, SpecRuntime, SpecName, SpecImage, SpecBindable, SpecAsync, SpecPlans)

func TestSpecLoadJSON(t *testing.T) {

	s := Spec{}
	err := LoadJSON(SpecJSON, &s)
	if err != nil {
		panic(err)
	}

	ft.AssertEqual(t, s.Description, SpecDescription)
	ft.AssertEqual(t, s.FQName, SpecName)
	ft.AssertEqual(t, s.Version, SpecVersion)
	ft.AssertEqual(t, s.Runtime, SpecRuntime)
	ft.AssertEqual(t, s.Image, SpecImage)
	ft.AssertEqual(t, s.Bindable, SpecBindable)
	ft.AssertEqual(t, s.Async, SpecAsync)
	ft.AssertTrue(t, reflect.DeepEqual(s.Plans[0].Parameters, expectedPlanParameters))
}

func TestSpecDumpJSON(t *testing.T) {
	s := Spec{
		Description: SpecDescription,
		Runtime:     SpecRuntime,
		Version:     SpecVersion,
		FQName:      SpecName,
		Image:       SpecImage,
		Bindable:    SpecBindable,
		Async:       SpecAsync,
		Plans:       []Plan{p},
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
	decodedyaml, err := base64.StdEncoding.DecodeString(ft.EncodedApb())
	if err != nil {
		t.Fatal(err)
	}

	spec := &Spec{}
	if err = yaml.Unmarshal(decodedyaml, spec); err != nil {
		t.Fatal(err)
	}
	fmt.Printf("%#v", spec)
	ft.AssertEqual(t, spec.FQName, "mediawiki123-apb")
	ft.AssertEqual(t, len(spec.Plans[0].Parameters), 5)

	// picking something other than the first one
	sitelang := spec.Plans[0].Parameters[2] // mediawiki_site_lang

	ft.AssertEqual(t, sitelang.Name, "mediawiki_site_lang")
	ft.AssertEqual(t, sitelang.Title, "Mediawiki Site Language")
	ft.AssertEqual(t, sitelang.Type, "string")
	ft.AssertEqual(t, sitelang.Description, "")
	ft.AssertEqual(t, sitelang.Default, "en")
	ft.AssertEqual(t, sitelang.DeprecatedMaxlength, 0)
	ft.AssertEqual(t, sitelang.Pattern, "")
	ft.AssertEqual(t, len(sitelang.Enum), 0)
}

func TestBindInstanceUserParamsNil(t *testing.T) {
	a := BindInstance{
		ID:        uuid.NewUUID(),
		ServiceID: uuid.NewUUID(),
	}
	up := a.UserParameters()
	ft.AssertTrue(t, up == nil)
}

func TestBindInstanceUserParams(t *testing.T) {
	a := BindInstance{
		ID:        uuid.NewUUID(),
		ServiceID: uuid.NewUUID(),
	}
	a.Parameters = &Parameters{
		"foo":                  "bar",
		"cluster":              "mycluster",
		"namespace":            "mynamespace",
		"_apb_provision_creds": "letmein",
	}

	up := a.UserParameters()

	// Make sure the "foo" key is still included
	ft.AssertTrue(t, up["foo"] == "bar")

	// Make sure all of these got filtered out
	for _, key := range []string{"cluster", "namespace", "_apb_provision_creds"} {
		_, ok := up[key]
		ft.AssertFalse(t, ok)
	}

}

func TestBindInstanceEqual(t *testing.T) {
	a := BindInstance{
		ID:         uuid.NewUUID(),
		ServiceID:  uuid.NewUUID(),
		Parameters: &Parameters{"foo": "bar"},
	}
	b := BindInstance{
		ID:         a.ID,
		ServiceID:  a.ServiceID,
		Parameters: &Parameters{"foo": "bar"},
	}
	ft.AssertTrue(t, a.IsEqual(&b))
	ft.AssertTrue(t, b.IsEqual(&a))
}

func TestBindInstanceNotEqual(t *testing.T) {

	a := BindInstance{
		ID:         uuid.Parse(uuid.New()),
		ServiceID:  uuid.Parse(uuid.New()),
		Parameters: &Parameters{"foo": "bar"},
	}

	data := map[string]BindInstance{
		"different parameters": BindInstance{
			ID:         a.ID,
			ServiceID:  a.ServiceID,
			Parameters: &Parameters{"foo": "notbar"},
		},
		"different ID": BindInstance{
			ID:         uuid.Parse(uuid.New()),
			ServiceID:  a.ServiceID,
			Parameters: &Parameters{"foo": "bar"},
		},
		"different ServiceID": BindInstance{
			ID:         a.ID,
			ServiceID:  uuid.Parse(uuid.New()),
			Parameters: &Parameters{"foo": "bar"},
		},
		"no parameters": BindInstance{
			ID:        a.ID,
			ServiceID: a.ServiceID,
		},
	}

	for key, binding := range data {
		if a.IsEqual(&binding) {
			t.Errorf("bindings were equal for case: %s", key)
		}
		if binding.IsEqual(&a) {
			t.Errorf("bindings were equal for case: %s", key)
		}
	}
}
