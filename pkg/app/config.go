package app

import (
	"fmt"
	"io/ioutil"
	"os"

	yaml "gopkg.in/yaml.v2"

	"github.com/openshift/ansible-service-broker/pkg/apb"
	"github.com/openshift/ansible-service-broker/pkg/broker"
	"github.com/openshift/ansible-service-broker/pkg/dao"
	"github.com/openshift/ansible-service-broker/pkg/registries"
)

// Config - The base config for the pieces of the applcation
type Config struct {
	Registry   []registries.Config
	Dao        dao.Config
	Log        LogConfig
	Openshift  apb.ClusterConfig
	ConfigFile string
	Broker     broker.Config
}

// CreateConfig - Read config file and create the Config struct
func CreateConfig(configFile string) (Config, error) {
	var err error

	// Confirm file is valid
	if _, err := os.Stat(configFile); err != nil {
		return Config{}, err
	}

	config := Config{
		ConfigFile: configFile,
	}

	// Load struct
	var dat []byte
	if dat, err = ioutil.ReadFile(configFile); err != nil {
		return Config{}, err
	}

	if err = yaml.Unmarshal(dat, &config); err != nil {
		return Config{}, err
	}

	if err = validateConfig(config); err != nil {
		return Config{}, err
	}

	return config, nil
}

func validateConfig(c Config) error {
	// TODO: Config validation!
	registryName := map[string]bool{}
	for _, rc := range c.Registry {
		if !rc.Validate() {
			return fmt.Errorf("registry config is not valid - %v", rc.Name)
		}
		if _, ok := registryName[rc.Name]; ok {
			return fmt.Errorf("registry name must be unique")
		}
		registryName[rc.Name] = true
	}
	return nil
}
