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
// Red Hat trademarks are not licensed under Apache License, Version 2.
// No permission is granted to use or replicate Red Hat trademarks that
// are incorporated in this software or its documentation.
//

package registries

import (
	"fmt"
	"testing"

	logging "github.com/op/go-logging"
	"github.com/openshift/ansible-service-broker/pkg/apb"
	ft "github.com/openshift/ansible-service-broker/pkg/fusortest"
	"github.com/openshift/ansible-service-broker/pkg/registries/adapters"
)

var SpecTags = []string{"latest", "old-release"}

const SpecID = "ab094014-b740-495e-b178-946d5aa97ebf"
const SpecName = "etherpad-apb"
const SpecImage = "fusor/etherpad-apb"
const SpecBindable = false
const SpecAsync = "optional"
const SpecDescription = "A note taking webapp"
const SpecRegistryName = "test"

const PlanName = "dev"
const PlanDescription = "Basic development plan"

var PlanMetadata = map[string]interface{}{
	"displayName":     "Development",
	"longDescription": PlanDescription,
	"cost":            "$0.00",
}

const PlanFree = true
const PlanBindable = true

var expectedPlanParameters = []apb.ParameterDescriptor{
	apb.ParameterDescriptor{
		Name:    "postgresql_database",
		Default: "admin",
		Type:    "string",
		Title:   "PostgreSQL Database Name"},
	apb.ParameterDescriptor{
		Name:        "postgresql_password",
		Default:     "admin",
		Type:        "string",
		Description: "A random alphanumeric string if left blank",
		Title:       "PostgreSQL Password"},
	apb.ParameterDescriptor{
		Name:      "postgresql_user",
		Default:   "admin",
		Title:     "PostgreSQL User",
		Type:      "string",
		Maxlength: 63},
	apb.ParameterDescriptor{
		Name:    "postgresql_version",
		Default: 9.5,
		Enum:    []string{"9.5", "9.4"},
		Type:    "enum",
		Title:   "PostgreSQL Version"},
	apb.ParameterDescriptor{
		Name:        "postgresql_email",
		Pattern:     "\u201c^\\\\S+@\\\\S+$\u201d",
		Type:        "string",
		Description: "email address",
		Title:       "email"},
}

var p = apb.Plan{
	Name:        PlanName,
	Description: PlanDescription,
	Metadata:    PlanMetadata,
	Free:        PlanFree,
	Bindable:    PlanBindable,
	Parameters:  expectedPlanParameters,
}

var s = apb.Spec{
	ID:          SpecID,
	Description: SpecDescription,
	FQName:      SpecName,
	Image:       SpecImage,
	Tags:        SpecTags,
	Bindable:    SpecBindable,
	Async:       SpecAsync,
	Plans:       []apb.Plan{p},
}

var noPlansSpec = apb.Spec{
	ID:          SpecID,
	Description: SpecDescription,
	FQName:      SpecName,
	Image:       SpecImage,
	Tags:        SpecTags,
	Bindable:    SpecBindable,
	Async:       SpecAsync,
}

type TestingAdapter struct {
	Name   string
	Images []string
	Specs  []*apb.Spec
	Called map[string]bool
}

func (t TestingAdapter) GetImageNames() ([]string, error) {
	t.Called["GetImageNames"] = true
	return t.Images, nil
}

func (t TestingAdapter) FetchSpecs(images []string) ([]*apb.Spec, error) {
	t.Called["FetchSpecs"] = true
	return t.Specs, nil
}

func (t TestingAdapter) RegistryName() string {
	t.Called["RegistryName"] = true
	return t.Name
}

var a *TestingAdapter
var r Registry

func setUp() Registry {
	a = &TestingAdapter{
		Name:   "testing",
		Images: []string{"image1-apb", "image2"},
		Specs:  []*apb.Spec{&s},
		Called: map[string]bool{},
	}
	filter := Filter{}
	c := Config{}
	log := &logging.Logger{}
	r = Registry{config: c,
		adapter: a,
		log:     log,
		filter:  filter}
	return r
}

func setUpNoPlans() Registry {
	a = &TestingAdapter{
		Name:   "testing",
		Images: []string{"image1-apb", "image2"},
		Specs:  []*apb.Spec{&noPlansSpec},
		Called: map[string]bool{},
	}
	filter := Filter{}
	c := Config{}
	log := &logging.Logger{}
	r = Registry{config: c,
		adapter: a,
		log:     log,
		filter:  filter}
	return r
}

func TestRegistryLoadSpecsNoError(t *testing.T) {
	r := setUp()
	specs, numImages, err := r.LoadSpecs()
	if err != nil {
		ft.AssertTrue(t, false)
	}
	ft.AssertTrue(t, a.Called["GetImageNames"])
	ft.AssertTrue(t, a.Called["FetchSpecs"])
	ft.AssertEqual(t, numImages, 1)
	ft.AssertEqual(t, len(specs), 1)
	ft.AssertEqual(t, specs[0], &s)
}

func TestRegistryLoadSpecsNoPlans(t *testing.T) {
	r := setUpNoPlans()
	specs, _, err := r.LoadSpecs()
	if err != nil {
		ft.AssertTrue(t, false)
	}
	ft.AssertTrue(t, a.Called["GetImageNames"])
	ft.AssertTrue(t, a.Called["FetchSpecs"])
	ft.AssertEqual(t, len(specs), 0)
}

func TestFail(t *testing.T) {
	r := setUp()
	r.config.Fail = true

	fail := r.Fail(fmt.Errorf("new error"))
	ft.AssertTrue(t, fail)
}

func TestFailIsFalse(t *testing.T) {
	r := setUp()
	r.config.Fail = false

	fail := r.Fail(fmt.Errorf("new error"))
	ft.AssertFalse(t, fail)
}

func TestNewRegistryRHCC(t *testing.T) {
	c := Config{Type: "rhcc"}
	log := &logging.Logger{}
	reg, err := NewRegistry(c, log)
	if err != nil {
		ft.AssertTrue(t, false)
	}
	_, ok := reg.adapter.(*adapters.RHCCAdapter)
	ft.AssertTrue(t, ok)
}

func TestNewRegistryDockerHub(t *testing.T) {
	c := Config{Type: "dockerhub"}
	log := &logging.Logger{}
	reg, err := NewRegistry(c, log)
	if err != nil {
		ft.AssertTrue(t, false)
	}
	_, ok := reg.adapter.(*adapters.DockerHubAdapter)
	ft.AssertTrue(t, ok)
}

func TestNewRegistryMock(t *testing.T) {
	c := Config{Type: "mocK"}
	log := &logging.Logger{}
	reg, err := NewRegistry(c, log)
	if err != nil {
		ft.AssertTrue(t, false)
	}
	_, ok := reg.adapter.(*adapters.MockAdapter)
	ft.AssertTrue(t, ok)
}

func TestPanicOnUnknow(t *testing.T) {
	defer func() {
		r := recover()
		if r == nil {
			ft.AssertTrue(t, false)
		}
	}()
	c := Config{Type: "UnKOwn"}
	log := &logging.Logger{}
	NewRegistry(c, log)
}
