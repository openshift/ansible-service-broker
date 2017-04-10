package apb

import (
        "fmt"
        "io/ioutil"

        logging "github.com/op/go-logging"
        yaml "gopkg.in/yaml.v1"
)

type MockRegistry struct {
        config RegistryConfig
        log    *logging.Logger
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
        
        specYaml, err := ioutil.ReadFile("/etc/ansible-service-broker/mock-registry-data.yml")
        r.log.Debug(fmt.Sprintf("%s",specYaml))
        if err != nil {
                r.log.Debug(fmt.Sprintf("Failed to read registry data from /etc/ansible-service-broker/mock-registry-data.yml"))
        }
        
        var parsedData struct {
                Apps []*Spec `yaml:"apps"`
        }
        
        err = yaml.Unmarshal(specYaml, &parsedData)
        if err != nil {
                r.log.Debug(fmt.Sprintf("Failed to ummarshal yaml file"))
        }
        
        r.log.Debug(fmt.Sprintf("Loaded Specs: %v", parsedData))
        
        r.log.Info(fmt.Sprintf("Loaded [ %d ] specs from %s registry", len(parsedData.Apps), r.config.Name))
        
        for _, spec := range parsedData.Apps {
                r.log.Debug(fmt.Sprintf("ID: %s", spec.Id))
        }
        
        return parsedData.Apps, nil
}
