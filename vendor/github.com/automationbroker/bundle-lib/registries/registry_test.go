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

package registries

import (
	"fmt"
	"testing"

	"github.com/automationbroker/bundle-lib/bundle"
	"github.com/automationbroker/bundle-lib/registries/adapters"
	ft "github.com/stretchr/testify/assert"
)

var SpecTags = []string{"latest", "old-release"}

const SpecID = "ab094014-b740-495e-b178-946d5aa97ebf"
const SpecBadVersion = "2.0"
const SpecVersion = "1.0"
const SpecRuntime = 1
const SpecBadRuntime = 0
const SpecName = "etherpad-bundle"
const SpecImage = "fusor/etherpad-bundle"
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

var expectedPlanParameters = []bundle.ParameterDescriptor{
	bundle.ParameterDescriptor{
		Name:    "postgresql_database",
		Default: "admin",
		Type:    "string",
		Title:   "PostgreSQL Database Name"},
	bundle.ParameterDescriptor{
		Name:        "postgresql_password",
		Default:     "admin",
		Type:        "string",
		Description: "A random alphanumeric string if left blank",
		Title:       "PostgreSQL Password"},
	bundle.ParameterDescriptor{
		Name:                "postgresql_user",
		Default:             "admin",
		Title:               "PostgreSQL User",
		Type:                "string",
		DeprecatedMaxlength: 63},
	bundle.ParameterDescriptor{
		Name:    "postgresql_version",
		Default: 9.5,
		Enum:    []string{"9.5", "9.4"},
		Type:    "enum",
		Title:   "PostgreSQL Version"},
	bundle.ParameterDescriptor{
		Name:        "postgresql_email",
		Pattern:     "\u201c^\\\\S+@\\\\S+$\u201d",
		Type:        "string",
		Description: "email address",
		Title:       "email"},
}

var p = bundle.Plan{
	Name:        PlanName,
	Description: PlanDescription,
	Metadata:    PlanMetadata,
	Free:        PlanFree,
	Bindable:    PlanBindable,
	Parameters:  expectedPlanParameters,
}

var s = bundle.Spec{
	Version:     SpecVersion,
	Runtime:     SpecRuntime,
	ID:          SpecID,
	Description: SpecDescription,
	FQName:      SpecName,
	Image:       SpecImage,
	Tags:        SpecTags,
	Bindable:    SpecBindable,
	Async:       SpecAsync,
	Plans:       []bundle.Plan{p},
}

var noPlansSpec = bundle.Spec{
	Version:     SpecVersion,
	Runtime:     SpecRuntime,
	ID:          SpecID,
	Description: SpecDescription,
	FQName:      SpecName,
	Image:       SpecImage,
	Tags:        SpecTags,
	Bindable:    SpecBindable,
	Async:       SpecAsync,
}

var noVersionSpec = bundle.Spec{
	Runtime:     SpecRuntime,
	ID:          SpecID,
	Description: SpecDescription,
	FQName:      SpecName,
	Image:       SpecImage,
	Tags:        SpecTags,
	Bindable:    SpecBindable,
	Async:       SpecAsync,
	Plans:       []bundle.Plan{p},
}

var badVersionSpec = bundle.Spec{
	Version:     SpecBadVersion,
	Runtime:     SpecRuntime,
	ID:          SpecID,
	Description: SpecDescription,
	FQName:      SpecName,
	Image:       SpecImage,
	Tags:        SpecTags,
	Bindable:    SpecBindable,
	Async:       SpecAsync,
	Plans:       []bundle.Plan{p},
}

var badRuntimeSpec = bundle.Spec{
	Version:     SpecVersion,
	Runtime:     SpecBadRuntime,
	ID:          SpecID,
	Description: SpecDescription,
	FQName:      SpecName,
	Image:       SpecImage,
	Tags:        SpecTags,
	Bindable:    SpecBindable,
	Async:       SpecAsync,
	Plans:       []bundle.Plan{p},
}

type TestingAdapter struct {
	Name   string
	Images []string
	Specs  []*bundle.Spec
	Called map[string]bool
}

func (t TestingAdapter) GetImageNames() ([]string, error) {
	t.Called["GetImageNames"] = true
	return t.Images, nil
}

func (t TestingAdapter) FetchSpecs(images []string) ([]*bundle.Spec, error) {
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
		Images: []string{"image1-bundle", "image2"},
		Specs:  []*bundle.Spec{&s},
		Called: map[string]bool{},
	}
	filter := Filter{}
	c := Config{}
	r = Registry{config: c,
		adapter: a,
		filter:  filter}
	return r
}

func setUpNoPlans() Registry {
	a = &TestingAdapter{
		Name:   "testing",
		Images: []string{"image1-bundle", "image2"},
		Specs:  []*bundle.Spec{&noPlansSpec},
		Called: map[string]bool{},
	}
	filter := Filter{}
	c := Config{}
	r = Registry{config: c,
		adapter: a,
		filter:  filter}
	return r
}

func setUpNoVersion() Registry {
	a = &TestingAdapter{
		Name:   "testing",
		Images: []string{"image1-bundle", "image2"},
		Specs:  []*bundle.Spec{&noVersionSpec},
		Called: map[string]bool{},
	}
	filter := Filter{}
	c := Config{}
	r = Registry{config: c,
		adapter: a,
		filter:  filter}
	return r
}

func setUpBadVersion() Registry {
	a = &TestingAdapter{
		Name:   "testing",
		Images: []string{"image1-bundle", "image2"},
		Specs:  []*bundle.Spec{&badVersionSpec},
		Called: map[string]bool{},
	}
	filter := Filter{}
	c := Config{}
	r = Registry{config: c,
		adapter: a,
		filter:  filter}
	return r
}

func setUpBadRuntime() Registry {
	a = &TestingAdapter{
		Name:   "testing",
		Images: []string{"image1-bundle", "image2"},
		Specs:  []*bundle.Spec{&badRuntimeSpec},
		Called: map[string]bool{},
	}
	filter := Filter{}
	c := Config{}
	r = Registry{config: c,
		adapter: a,
		filter:  filter}
	return r
}

func TestRegistryLoadSpecsNoError(t *testing.T) {
	r := setUp()
	specs, numImages, err := r.LoadSpecs()
	if err != nil {
		ft.True(t, false)
	}
	ft.True(t, a.Called["GetImageNames"])
	ft.True(t, a.Called["FetchSpecs"])
	ft.Equal(t, numImages, 2)
	ft.Equal(t, len(specs), 1)
	ft.Equal(t, specs[0], &s)
}

func TestRegistryLoadSpecsNoPlans(t *testing.T) {
	r := setUpNoPlans()
	specs, _, err := r.LoadSpecs()
	if err != nil {
		ft.True(t, false)
	}
	ft.True(t, a.Called["GetImageNames"])
	ft.True(t, a.Called["FetchSpecs"])
	ft.Equal(t, len(specs), 0)
}

func TestRegistryLoadSpecsNoVersion(t *testing.T) {
	r := setUpNoVersion()
	specs, _, err := r.LoadSpecs()
	if err != nil {
		ft.True(t, false)
	}
	ft.True(t, a.Called["GetImageNames"])
	ft.True(t, a.Called["FetchSpecs"])
	ft.Equal(t, len(specs), 0)
}

func TestRegistryLoadSpecsBadVersion(t *testing.T) {
	r := setUpBadVersion()
	specs, _, err := r.LoadSpecs()
	if err != nil {
		ft.True(t, false)
	}
	ft.True(t, a.Called["GetImageNames"])
	ft.True(t, a.Called["FetchSpecs"])
	ft.Equal(t, len(specs), 0)
}

func TestRegistryLoadSpecsBadRuntime(t *testing.T) {
	r := setUpBadRuntime()
	specs, _, err := r.LoadSpecs()
	if err != nil {
		ft.True(t, false)
	}
	ft.True(t, a.Called["GetImageNames"])
	ft.True(t, a.Called["FetchSpecs"])
	ft.Equal(t, len(specs), 0)
}

func TestFail(t *testing.T) {
	r := setUp()
	r.config.Fail = true

	fail := r.Fail(fmt.Errorf("new error"))
	ft.True(t, fail)
}

func TestFailIsFalse(t *testing.T) {
	r := setUp()
	r.config.Fail = false

	fail := r.Fail(fmt.Errorf("new error"))
	ft.False(t, fail)
}

func TestNewRegistryRHCC(t *testing.T) {
	c := Config{
		Type: "rhcc",
		Name: "rhcc",
	}
	reg, err := NewRegistry(c, "")
	if err != nil {
		ft.True(t, false)
	}
	_, ok := reg.adapter.(*adapters.RHCCAdapter)
	ft.True(t, ok)
}

func TestNewRegistryDockerHub(t *testing.T) {
	c := Config{
		Type: "dockerhub",
		Name: "dh",
		URL:  "https://registry.hub.docker.com",
		User: "shurley",
		Org:  "shurley",
	}
	reg, err := NewRegistry(c, "")
	if err != nil {
		ft.True(t, false)
	}
	_, ok := reg.adapter.(*adapters.DockerHubAdapter)
	ft.True(t, ok)
}

func TestNewRegistryMock(t *testing.T) {
	c := Config{
		Type: "mock",
		Name: "mock",
	}

	reg, err := NewRegistry(c, "")
	if err != nil {
		ft.True(t, false)
	}
	_, ok := reg.adapter.(*adapters.MockAdapter)
	ft.True(t, ok)
}

func TestPanicOnUnknow(t *testing.T) {
	defer func() {
		r := recover()
		fmt.Printf("%v", r)
		if r == nil {
			ft.True(t, false)
		}
	}()
	c := Config{
		Type: "makes_no_sense",
		Name: "dh",
	}
	r, err := NewRegistry(c, "")
	fmt.Printf("%#v\n\n %v\n", r, err)
}

func TestValidateName(t *testing.T) {
	c := Config{
		Type: "dockerhub",
	}
	_, err := NewRegistry(c, "")
	if err == nil {
		ft.True(t, false)
	}
}

func TestVersionCheck(t *testing.T) {
	// Test equal versions
	ft.True(t, isCompatibleVersion("1.0", "1.0", "1.0"))
	// Test out of range by major version
	ft.False(t, isCompatibleVersion("2.0", "1.0", "1.0"))
	// Test out of range by minor version
	ft.True(t, isCompatibleVersion("1.10", "1.0", "1.0"))
	// Test out of range by major and minor version
	ft.True(t, isCompatibleVersion("2.4", "1.0", "2.0"))
	// Test in range with differing  major and minor version
	ft.True(t, isCompatibleVersion("1.10", "1.0", "2.0"))
	// Test out of range by major and minor version
	ft.False(t, isCompatibleVersion("0.6", "1.0", "2.0"))
	// Test out of range by major and minor version and invalid version
	ft.False(t, isCompatibleVersion("0.1.0", "1.0", "1.0"))
	// Test in range of long possible window
	ft.True(t, isCompatibleVersion("2.5", "1.0", "3.0"))
	// Test invalid version
	ft.False(t, isCompatibleVersion("1", "1.0", "3.0"))
	// Test invalid version
	ft.False(t, isCompatibleVersion("2.5", "3.0", "4.0"))
}

type fakeAdapter struct{}

func (f fakeAdapter) GetImageNames() ([]string, error) {
	return []string{}, nil
}

func (f fakeAdapter) FetchSpecs(names []string) ([]*bundle.Spec, error) {
	return []*bundle.Spec{}, nil
}

func (f fakeAdapter) RegistryName() string {
	return ""
}

func TestAdapterWithConfiguration(t *testing.T) {
	c := Config{
		Name: "nsa",
		Type: "custom",
	}

	f := fakeAdapter{}

	reg, err := NewCustomRegistry(c, f, "")
	if err != nil {
		t.Fatal(err.Error())
	}
	ft.Equal(t, reg.adapter, f, "registry uses wrong adapter")
	ft.Equal(t, reg.config, c, "registrying using wrong config")
}
