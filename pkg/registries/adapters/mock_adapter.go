package adapters

import (
	"io/ioutil"

	logging "github.com/op/go-logging"
	"github.com/openshift/ansible-service-broker/pkg/apb"
	yaml "gopkg.in/yaml.v2"
)

// MockFile - Mock file contains fake regitry data
var MockFile = "/etc/ansible-service-broker/mock-registry-data.yaml"

// mockRegistryName - mock registry name
var mockRegistryName = "mock"

// MockAdapter - a adapter that is for mocking data
type MockAdapter struct {
	Config Configuration
	Log    *logging.Logger
	specs  map[string]*apb.Spec
}

// GetImageNames - retrieve the image names
func (r *MockAdapter) GetImageNames() ([]string, error) {
	r.Log.Debug("MockRegistry::LoadSpecs")

	specYaml, err := ioutil.ReadFile(MockFile)
	if err != nil {
		r.Log.Debug("Failed to read registry data from %s", MockFile)
		return nil, err
	}

	var parsedData struct {
		Apps []*apb.Spec `yaml:"apps"`
	}

	err = yaml.Unmarshal(specYaml, &parsedData)
	if err != nil {
		r.Log.Error("Failed to ummarshal yaml file")
		return nil, err
	}

	r.Log.Debug("Loaded Specs: %v", parsedData)

	r.Log.Info("Loaded [ %d ] specs from %s registry", len(parsedData.Apps), "Mock")
	var names []string
	r.specs = map[string]*apb.Spec{}

	for _, spec := range parsedData.Apps {
		r.specs[spec.Image] = spec
		names = append(names, spec.Image)
	}
	return names, nil
}

// FetchSpecs - fetch the specs that were retrieved in the get images from the mock registry.
func (r MockAdapter) FetchSpecs(imageNames []string) ([]*apb.Spec, error) {
	specs := []*apb.Spec{}
	for _, name := range imageNames {
		spec, ok := r.specs[name]
		if ok {
			specs = append(specs, spec)
		}
	}
	return specs, nil
}

// RegistryName - retrieve the registry name
func (r MockAdapter) RegistryName() string {
	return mockRegistryName
}
