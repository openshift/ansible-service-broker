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

var expectedSpecParameters = []map[string]*apb.ParameterDescriptor{
	map[string]*apb.ParameterDescriptor{
		"postgresql_database": &apb.ParameterDescriptor{
			Default: "admin",
			Type:    "string",
			Title:   "PostgreSQL Database Name"}},
	map[string]*apb.ParameterDescriptor{
		"postgresql_password": &apb.ParameterDescriptor{
			Default:     "admin",
			Type:        "string",
			Description: "A random alphanumeric string if left blank",
			Title:       "PostgreSQL Password"}},
	map[string]*apb.ParameterDescriptor{
		"postgresql_user": &apb.ParameterDescriptor{
			Default:   "admin",
			Title:     "PostgreSQL User",
			Type:      "string",
			Maxlength: 63}},
	map[string]*apb.ParameterDescriptor{
		"postgresql_version": &apb.ParameterDescriptor{
			Default: 9.5,
			Enum:    []string{"9.5", "9.4"},
			Type:    "enum",
			Title:   "PostgreSQL Version"}},
	map[string]*apb.ParameterDescriptor{
		"postgresql_email": &apb.ParameterDescriptor{
			Pattern:     "\u201c^\\\\S+@\\\\S+$\u201d",
			Type:        "string",
			Description: "email address",
			Title:       "email"}},
}

var s = apb.Spec{
	ID:          SpecID,
	Description: SpecDescription,
	FQName:      SpecName,
	Image:       SpecImage,
	Tags:        SpecTags,
	Bindable:    SpecBindable,
	Async:       SpecAsync,
	Parameters:  expectedSpecParameters,
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
		Images: []string{"image1", "image2"},
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

func TestRegistryName(t *testing.T) {
	r := setUp()
	nameGot := r.RegistryName()
	ft.AssertTrue(t, a.Called["RegistryName"])
	ft.AssertEqual(t, nameGot, a.RegistryName())
}

func TestRegistryLoadSpecsNoError(t *testing.T) {
	r := setUp()
	specs, numImages, err := r.LoadSpecs()
	if err != nil {
		ft.AssertTrue(t, false)
	}
	ft.AssertTrue(t, a.Called["GetImageNames"])
	ft.AssertTrue(t, a.Called["FetchSpecs"])
	ft.AssertEqual(t, numImages, 2)
	ft.AssertEqual(t, len(specs), 1)
	ft.AssertEqual(t, specs[0], &s)
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
