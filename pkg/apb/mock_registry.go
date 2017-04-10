package apb

import (
	"io/ioutil"

	logging "github.com/op/go-logging"
	yaml "gopkg.in/yaml.v2"
)

var MockFile = "/etc/ansible-service-broker/mock-registry-data.yaml"

type MockRegistry struct {
	config     RegistryConfig
	log        *logging.Logger
	ScriptsDir string
}

func (r *MockRegistry) Init(config RegistryConfig, log *logging.Logger) error {
	log.Debug("MockRegistry::Init")
	r.config = config
	r.log = log
	return nil
}

func (r *MockRegistry) LoadSpecs() ([]*Spec, error) {
	r.log.Debug("MockRegistry::LoadSpecs")

	specYaml, err := ioutil.ReadFile(MockFile)
	if err != nil {
		r.log.Debug("Failed to read registry data from %s", MockFile)
	}

	var parsedData struct {
		Apps []*Spec `yaml:"apps"`
	}

	err = yaml.Unmarshal(specYaml, &parsedData)
	if err != nil {
		r.log.Error("Failed to ummarshal yaml file")
	}

	r.log.Debug("Loaded Specs: %v", parsedData)

	r.log.Info("Loaded [ %d ] specs from %s registry", len(parsedData.Apps), r.config.Name)

	for _, spec := range parsedData.Apps {
		r.log.Debug("ID: %s", spec.Id)
	}

	return parsedData.Apps, nil
}
