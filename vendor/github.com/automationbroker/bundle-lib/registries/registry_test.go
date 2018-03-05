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

	"github.com/automationbroker/bundle-lib/apb"
	"github.com/automationbroker/bundle-lib/registries/adapters"
	"github.com/automationbroker/config"
	ft "github.com/stretchr/testify/assert"
)

var SpecTags = []string{"latest", "old-release"}

const SpecID = "ab094014-b740-495e-b178-946d5aa97ebf"
const SpecBadVersion = "2.0"
const SpecVersion = "1.0"
const SpecRuntime = 1
const SpecBadRuntime = 0
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
		Name:                "postgresql_user",
		Default:             "admin",
		Title:               "PostgreSQL User",
		Type:                "string",
		DeprecatedMaxlength: 63},
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
	Version:     SpecVersion,
	Runtime:     SpecRuntime,
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

var noVersionSpec = apb.Spec{
	Runtime:     SpecRuntime,
	ID:          SpecID,
	Description: SpecDescription,
	FQName:      SpecName,
	Image:       SpecImage,
	Tags:        SpecTags,
	Bindable:    SpecBindable,
	Async:       SpecAsync,
	Plans:       []apb.Plan{p},
}

var badVersionSpec = apb.Spec{
	Version:     SpecBadVersion,
	Runtime:     SpecRuntime,
	ID:          SpecID,
	Description: SpecDescription,
	FQName:      SpecName,
	Image:       SpecImage,
	Tags:        SpecTags,
	Bindable:    SpecBindable,
	Async:       SpecAsync,
	Plans:       []apb.Plan{p},
}

var badRuntimeSpec = apb.Spec{
	Version:     SpecVersion,
	Runtime:     SpecBadRuntime,
	ID:          SpecID,
	Description: SpecDescription,
	FQName:      SpecName,
	Image:       SpecImage,
	Tags:        SpecTags,
	Bindable:    SpecBindable,
	Async:       SpecAsync,
	Plans:       []apb.Plan{p},
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
	r = Registry{config: c,
		adapter: a,
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
	r = Registry{config: c,
		adapter: a,
		filter:  filter}
	return r
}

func setUpNoVersion() Registry {
	a = &TestingAdapter{
		Name:   "testing",
		Images: []string{"image1-apb", "image2"},
		Specs:  []*apb.Spec{&noVersionSpec},
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
		Images: []string{"image1-apb", "image2"},
		Specs:  []*apb.Spec{&badVersionSpec},
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
		Images: []string{"image1-apb", "image2"},
		Specs:  []*apb.Spec{&badRuntimeSpec},
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
	c, err := config.CreateConfig("testdata/registry.yaml")
	if err != nil {
		ft.True(t, false)
	}
	reg, err := NewRegistry(c.GetSubConfig("registry.rhcc"), "")
	if err != nil {
		ft.True(t, false)
	}
	_, ok := reg.adapter.(*adapters.RHCCAdapter)
	ft.True(t, ok)
}

func TestNewRegistryDockerHub(t *testing.T) {
	c, err := config.CreateConfig("testdata/registry.yaml")
	if err != nil {
		ft.True(t, false)
	}
	reg, err := NewRegistry(c.GetSubConfig("registry.dh"), "")
	if err != nil {
		ft.True(t, false)
	}
	_, ok := reg.adapter.(*adapters.DockerHubAdapter)
	ft.True(t, ok)
}

func TestNewRegistryMock(t *testing.T) {
	c, err := config.CreateConfig("testdata/registry.yaml")
	if err != nil {
		ft.True(t, false)
	}
	reg, err := NewRegistry(c.GetSubConfig("registry.mock"), "")
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
	c, _ := config.CreateConfig("testdata/registry.yaml")
	r, err := NewRegistry(c.GetSubConfig("registry.makes-no-sense"), "")
	fmt.Printf("%#v\n\n %v\n", r, err)
}

func TestValidateName(t *testing.T) {
	c, _ := config.CreateConfig("testdata/registry.yaml")
	_, err := NewRegistry(c.GetSubConfig("registry.makes_no_sense"), "")
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
