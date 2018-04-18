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
	"io/ioutil"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/pborman/uuid"
	ft "github.com/stretchr/testify/assert"
	yaml "gopkg.in/yaml.v2"
)

const alphaApbTestFile = "alpha_apb.yml"

func loadTestFile(t *testing.T, name string) []byte {
	path := filepath.Join("testdata", name)
	bytes, err := ioutil.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	return bytes
}

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

var SpecAlpha = map[string]interface{}{"dashboard_redirect": true}
var SpecAlphaStr = `
{
	"dashboard_redirect": true
}
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
	"plans": %s,
	"alpha": %s
}
`, SpecDescription, SpecVersion, SpecRuntime, SpecName, SpecImage, SpecBindable, SpecAsync, SpecPlans, SpecAlphaStr)

func TestSpecLoadJSON(t *testing.T) {
	s := Spec{}
	err := LoadJSON(SpecJSON, &s)
	if err != nil {
		panic(err)
	}

	ft.Equal(t, s.Description, SpecDescription)
	ft.Equal(t, s.FQName, SpecName)
	ft.Equal(t, s.Version, SpecVersion)
	ft.Equal(t, s.Runtime, SpecRuntime)
	ft.Equal(t, s.Image, SpecImage)
	ft.Equal(t, s.Bindable, SpecBindable)
	ft.Equal(t, s.Async, SpecAsync)
	ft.True(t, reflect.DeepEqual(s.Plans[0].Parameters, expectedPlanParameters))
	ft.True(t, reflect.DeepEqual(s.Alpha, SpecAlpha))
}

func EncodedApb() string {
	apb := `bmFtZTogbWVkaWF3aWtpMTIzLWFwYgppbWFnZTogYW5zaWJsZXBsYXlib29rYnVuZGxlL21lZGlhd2lraTEyMy1hcGIKZGVzY3JpcHRpb246ICJNZWRpYXdpa2kxMjMgYXBiIGltcGxlbWVudGF0aW9uIgpiaW5kYWJsZTogZmFsc2UKYXN5bmM6IG9wdGlvbmFsCm1ldGFkYXRhOgogIGRpc3BsYXluYW1lOiAiUmVkIEhhdCBNZWRpYXdpa2kiCiAgbG9uZ0Rlc2NyaXB0aW9uOiAiQW4gYXBiIHRoYXQgZGVwbG95cyBNZWRpYXdpa2kgMS4yMyIKICBpbWFnZVVSTDogImh0dHBzOi8vdXBsb2FkLndpa2ltZWRpYS5vcmcvd2lraXBlZGlhL2NvbW1vbnMvMC8wMS9NZWRpYVdpa2ktc21hbGxlci1sb2dvLnBuZyIKICBkb2N1bWVudGF0aW9uVVJMOiAiaHR0cHM6Ly93d3cubWVkaWF3aWtpLm9yZy93aWtpL0RvY3VtZW50YXRpb24iCnBsYW5zOgogIC0gbmFtZTogZGV2CiAgICBkZXNjcmlwdGlvbjogIk1lZGlhd2lraTEyMyBhcGIgaW1wbGVtZW50YXRpb24iCiAgICBmcmVlOiB0cnVlCiAgICBiaW5kYWJsZTogdHJ1ZQogICAgbWV0YWRhdGE6CiAgICAgIGRpc3BsYXlOYW1lOiBEZXZlbG9wbWVudAogICAgICBsb25nRGVzY3JpcHRpb246IEJhc2ljIGRldmVsb3BtZW50IHBsYW4KICAgICAgY29zdDogJDAuMDAKICAgIHBhcmFtZXRlcnM6CiAgICAgIC0gbmFtZTogbWVkaWF3aWtpX2RiX3NjaGVtYQogICAgICAgIHRpdGxlOiBNZWRpYXdpa2kgREIgU2NoZW1hCiAgICAgICAgdHlwZTogc3RyaW5nCiAgICAgICAgZGVmYXVsdDogbWVkaWF3aWtpCiAgICAgICAgcmVxdWlyZWQ6IHRydWUKICAgICAgLSBuYW1lOiBtZWRpYXdpa2lfc2l0ZV9uYW1lCiAgICAgICAgdGl0bGU6IE1lZGlhd2lraSBTaXRlIE5hbWUKICAgICAgICB0eXBlOiBzdHJpbmcKICAgICAgICBkZWZhdWx0OiBNZWRpYVdpa2kKICAgICAgICByZXF1aXJlZDogdHJ1ZQogICAgICAtIG5hbWU6IG1lZGlhd2lraV9zaXRlX2xhbmcKICAgICAgICB0aXRsZTogTWVkaWF3aWtpIFNpdGUgTGFuZ3VhZ2UKICAgICAgICB0eXBlOiBzdHJpbmcKICAgICAgICBkZWZhdWx0OiBlbgogICAgICAgIHJlcXVpcmVkOiB0cnVlCiAgICAgIC0gbmFtZTogbWVkaWF3aWtpX2FkbWluX3VzZXIKICAgICAgICB0aXRsZTogTWVkaWF3aWtpIEFkbWluIFVzZXIKICAgICAgICB0eXBlOiBzdHJpbmcKICAgICAgICBkZWZhdWx0OiBhZG1pbgogICAgICAgIHJlcXVpcmVkOiB0cnVlCiAgICAgIC0gbmFtZTogbWVkaWF3aWtpX2FkbWluX3Bhc3MKICAgICAgICB0aXRsZTogTWVkaWF3aWtpIEFkbWluIFVzZXIgUGFzc3dvcmQKICAgICAgICB0eXBlOiBzdHJpbmcKICAgICAgICByZXF1aXJlZDogdHJ1ZQogICAgYmluZF9wYXJhbWV0ZXJzOgogICAgICAtIG5hbWU6IGJpbmRfcGFyYW1fMQogICAgICAgIHRpdGxlOiBCaW5kIFBhcmFtIDEKICAgICAgICB0eXBlOiBzdHJpbmcKICAgICAgICByZXF1aXJlZDogdHJ1ZQogICAgICAtIG5hbWU6IGJpbmRfcGFyYW1fMgogICAgICAgIHRpdGxlOiBCaW5kIFBhcmFtIDIKICAgICAgICB0eXBlOiBpbnQKICAgICAgICByZXF1aXJlZDogdHJ1ZQogICAgICAtIG5hbWU6IGJpbmRfcGFyYW1fMwogICAgICAgIHRpdGxlOiBCaW5kIFBhcmFtIDMKICAgICAgICB0eXBlOiBzdHJpbmcKCg==`
	return apb
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
		Alpha:       SpecAlpha,
	}

	var knownMap interface{}
	var subjectMap interface{}

	raw, err := DumpJSON(&s)
	if err != nil {
		panic(err)
	}

	json.Unmarshal([]byte(SpecJSON), &knownMap)
	json.Unmarshal([]byte(raw), &subjectMap)
	ft.True(t, reflect.DeepEqual(knownMap, subjectMap))
}

func TestEncodedParameters(t *testing.T) {
	decodedyaml, err := base64.StdEncoding.DecodeString(EncodedApb())
	if err != nil {
		t.Fatal(err)
	}

	spec := &Spec{}
	if err = yaml.Unmarshal(decodedyaml, spec); err != nil {
		t.Fatal(err)
	}
	fmt.Printf("%#v", spec)
	ft.Equal(t, spec.FQName, "mediawiki123-apb")
	ft.Equal(t, len(spec.Plans[0].Parameters), 5)

	// picking something other than the first one
	sitelang := spec.Plans[0].Parameters[2] // mediawiki_site_lang

	ft.Equal(t, sitelang.Name, "mediawiki_site_lang")
	ft.Equal(t, sitelang.Title, "Mediawiki Site Language")
	ft.Equal(t, sitelang.Type, "string")
	ft.Equal(t, sitelang.Description, "")
	ft.Equal(t, sitelang.Default, "en")
	ft.Equal(t, sitelang.DeprecatedMaxlength, 0)
	ft.Equal(t, sitelang.Pattern, "")
	ft.Equal(t, len(sitelang.Enum), 0)
}

func TestBindInstanceUserParamsNil(t *testing.T) {
	a := BindInstance{
		ID:        uuid.NewUUID(),
		ServiceID: uuid.NewUUID(),
	}
	up := a.UserParameters()
	ft.True(t, up == nil)
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
	ft.True(t, up["foo"] == "bar")

	// Make sure all of these got filtered out
	for _, key := range []string{"cluster", "namespace", "_apb_provision_creds"} {
		_, ok := up[key]
		ft.False(t, ok)
	}

}

func TestEnsureDefaults(t *testing.T) {
	cases := []struct {
		Name           string
		ProvidedParams func() Parameters
		Validate       func(t *testing.T, params Parameters)
	}{
		{
			Name: "test defaults are set",
			ProvidedParams: func() Parameters {
				p := Parameters{}
				p.EnsureDefaults()
				return p
			},
			Validate: func(t *testing.T, actual Parameters) {
				if _, ok := actual[ProvisionCredentialsKey]; !ok {
					t.Fatalf("expected the key %s to be present but it was missing", ProvisionCredentialsKey)
				}
			},
		},
		{
			Name: "test existing key not overwritten",
			ProvidedParams: func() Parameters {
				p := Parameters{ProvisionCredentialsKey: "avalue"}
				p.EnsureDefaults()
				return p
			},
			Validate: func(t *testing.T, p Parameters) {
				if v, ok := p[ProvisionCredentialsKey]; ok {
					if v != "avalue" {
						t.Fatalf("expected the value for %s to be %v but got %v", ProvisionCredentialsKey, "avalue", v)
					}
					return
				}
				t.Fatalf("missing key %v from params", ProvisionCredentialsKey)
			},
		},
		{
			Name: "test default key set if other keys present",
			ProvidedParams: func() Parameters {
				p := Parameters{"somekey": "avalue"}
				p.EnsureDefaults()
				return p
			},
			Validate: func(t *testing.T, p Parameters) {
				if v, ok := p["somekey"]; ok {
					if v != "avalue" {
						t.Fatalf("expected somekey to be set to avalue but was %s", v)
					}
				}
				if v, ok := p[ProvisionCredentialsKey]; ok {
					if v != struct{}{} {
						t.Fatalf("expected the default value for %v to be %v but got %v", ProvisionCredentialsKey, struct{}{}, v)
					}
					return
				}
				t.Fatalf("expected key somekey to be set but it wasnt")
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.Name, func(t *testing.T) {
			tc.Validate(t, tc.ProvidedParams())
		})
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
	ft.True(t, a.IsEqual(&b))
	ft.True(t, b.IsEqual(&a))
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

func TestAlphaParser(t *testing.T) {
	spec := &Spec{}
	testYaml := loadTestFile(t, alphaApbTestFile)
	if err := yaml.Unmarshal(testYaml, spec); err != nil {
		t.Fatal(err)
	}

	if len(spec.Alpha) == 0 {
		t.Error("spec.Alpha should not be empty")
	}

	var val interface{}
	var dr, ok bool

	if val, ok = spec.Alpha["dashboard_redirect"]; !ok {
		t.Error("spec.Alpha should contain dashboard_redirect key")
	}

	if dr, ok = val.(bool); !ok {
		t.Error(`spec.Alpha["dashboard_redirect"] should assert to bool`)
	}

	ft.True(t, dr)
}
