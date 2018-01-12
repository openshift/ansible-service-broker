//
// Copyright (c) 2017 Red Hat, Inc.
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
// Red Hat trademarks are not licensed under Apache License, Version 2.
// No permission is granted to use or replicate Red Hat trademarks that
// are incorporated in this software or its documentation.
//

package app

import (
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	"github.com/openshift/ansible-service-broker/pkg/apb"
	"github.com/openshift/ansible-service-broker/pkg/broker"
	"github.com/openshift/ansible-service-broker/pkg/clients"
	"github.com/openshift/ansible-service-broker/pkg/dao"
	"github.com/openshift/ansible-service-broker/pkg/registries"
	yaml "gopkg.in/yaml.v2"
)

// Config - The base config for the pieces of the applcation
type Config struct {
	Registry   []registries.Config
	Dao        dao.Config
	Log        LogConfig
	Openshift  apb.ClusterConfig
	ConfigFile string
	Broker     broker.Config
	Secrets    []apb.SecretsConfig
}

type regCreds struct {
	Username string
	Password string
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

	if config.Openshift.Namespace == "" {
		if dat, err = ioutil.ReadFile("/var/run/secrets/kubernetes.io/serviceaccount/namespace"); err != nil {
			return Config{}, err
		}
		config.Openshift.Namespace = string(dat)
	}

	for regCount, reg := range config.Registry {
		var username, password string
		switch reg.AuthType {
		case "secret":
			username, password, err = readSecret(reg.AuthName, config.Openshift.Namespace)
			if err != nil {
				return Config{}, err
			}
		case "file":
			username, password, err = readFile(reg.AuthName)
			if err != nil {
				return Config{}, err
			}
		case "config":
			if config.Registry[regCount].User == "" || config.Registry[regCount].Pass == "" {
				return Config{}, fmt.Errorf("Failed to find registry credentials in config")
			}
			username = config.Registry[regCount].User
			password = config.Registry[regCount].Pass
		case "":
			// Assuming that the user has either no credentials or defined them in the config
			username = reg.User
			password = reg.Pass
		default:
			return Config{}, fmt.Errorf("Unrecognized registry AuthType: %s", reg.AuthType)
		}

		config.Registry[regCount].User = username
		config.Registry[regCount].Pass = password
	}

	if err = validateConfig(config); err != nil {
		return Config{}, err
	}

	return config, nil
}

func readFile(fileName string) (string, string, error) {
	regCred := regCreds{}

	dat, err := ioutil.ReadFile(fileName)
	if err != nil {
		return "", "", fmt.Errorf("Failed to read registry credentials from file: %s", fileName)
	}
	err = yaml.Unmarshal(dat, &regCred)
	if err != nil {
		return "", "", fmt.Errorf("Failed to unmarshal registry credentials from file: %s", fileName)
	}
	return regCred.Username, regCred.Password, nil
}

func readSecret(secretName string, namespace string) (string, string, error) {
	data, err := clients.GetSecretData(secretName, namespace)
	if err != nil {
		return "", "", fmt.Errorf("Failed to find Dockerhub credentials in secret: %s", secretName)
	}
	var username = strings.TrimSpace(string(data["username"]))
	var password = strings.TrimSpace(string(data["password"]))

	if username == "" || password == "" {
		return username, password, fmt.Errorf("Secret: %s did not contain username/password credentials", secretName)
	}

	return username, password, nil
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

	for _, sc := range c.Secrets {
		if !sc.Validate() {
			// TODO: Terrible error message
			return fmt.Errorf("secrets config is not valid - %#v", sc)
		}

	}
	return nil
}
