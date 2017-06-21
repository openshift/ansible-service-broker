package app

import (
	"io/ioutil"
	"os"

	yaml "gopkg.in/yaml.v2"

	"github.com/openshift/ansible-service-broker/pkg/apb"
	"github.com/openshift/ansible-service-broker/pkg/broker"
	"github.com/openshift/ansible-service-broker/pkg/clients"
)

type Config struct {
	Registry   apb.RegistryConfig
	Dao        clients.EtcdConfig
	Log        LogConfig
	Openshift  apb.ClusterConfig
	ConfigFile string
	Broker     broker.BrokerConfig
}

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

	if err = validateConfig(); err != nil {
		return Config{}, err
	}

	// If no target is provided assume we are inside a cluster
	config.Openshift.InCluster = config.Openshift.Target == ""

	return config, nil
}

func validateConfig() error {
	// TODO: Config validation!
	return nil
}
