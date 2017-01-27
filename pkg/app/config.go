package app

import (
	"github.com/fusor/ansible-service-broker/pkg/ansibleapp"
	"github.com/fusor/ansible-service-broker/pkg/dao"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"os"
)

type Config struct {
	Registry   ansibleapp.RegistryConfig
	Dao        dao.Config
	Log        LogConfig
	ConfigFile string
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

	return config, nil
}

func validateConfig() error {
	// TODO: Config validation!
	return nil
}
