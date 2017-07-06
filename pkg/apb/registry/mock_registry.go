package registry

import (
	"io/ioutil"

	logging "github.com/op/go-logging"
	"github.com/openshift/ansible-service-broker/pkg/apb"
	yaml "gopkg.in/yaml.v2"
)

// MockFile - Mock file contains fake regitry data
var MockFile = "/etc/ansible-service-broker/mock-registry-data.yaml"

// MockRegistry - a registry that is for mocking data
type MockRegistry struct {
	config Config
	log    *logging.Logger
}

// Init - Initialize the mock registry
func (r *MockRegistry) Init(config Config, log *logging.Logger) error {
	log.Debug("MockRegistry::Init")
	r.config = config
	r.log = log
	return nil
}

// LoadSpecs - Load mock specs
func (r *MockRegistry) LoadSpecs() ([]*apb.Spec, int, error) {
	r.log.Debug("MockRegistry::LoadSpecs")

	specYaml, err := ioutil.ReadFile(MockFile)
	if err != nil {
		r.log.Debug("Failed to read registry data from %s", MockFile)
	}

	var parsedData struct {
		Apps []*apb.Spec `yaml:"apps"`
	}

	err = yaml.Unmarshal(specYaml, &parsedData)
	if err != nil {
		r.log.Error("Failed to ummarshal yaml file")
	}

	r.log.Debug("Loaded Specs: %v", parsedData)

	r.log.Info("Loaded [ %d ] specs from %s registry", len(parsedData.Apps), r.config.Name)

	for _, spec := range parsedData.Apps {
		r.log.Debug("ID: %s", spec.ID)
	}

	return parsedData.Apps, len(parsedData.Apps), nil
}

// Fail - will determine if this reqistry can cause a failure.
func (r MockRegistry) Fail(err error) bool {
	if r.config.Fail {
		return true
	}
	return false
}
