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
	"bytes"
	"errors"
	"fmt"
	"io/ioutil"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"sync"

	"github.com/automationbroker/bundle-lib/bundle"
	"github.com/automationbroker/bundle-lib/clients"
	"github.com/automationbroker/bundle-lib/registries/adapters"
	log "github.com/sirupsen/logrus"

	yaml "gopkg.in/yaml.v1"
)

var regex = regexp.MustCompile(`[a-z0-9]([-a-z0-9]*[a-z0-9])?(\.[a-z0-9]([-a-z0-9]*[a-z0-9])?)*`)

// Config - Configuration for the registry
type Config struct {
	URL string
	// AuthType is an optional way to declare where credentials for the registry are stored.
	//   Valid options: `secret`, `file`
	// AuthName is used to define the location of the credentials
	//   Valid options: `<secret-name>`, `<file_location>`
	AuthType   string `yaml:"auth_type"`
	AuthName   string `yaml:"auth_name"`
	User       string
	Pass       string
	Org        string
	Tag        string
	Type       string
	Name       string
	Runner     string
	Images     []string
	Namespaces []string
	// Fail will tell the registry that it is ok to fail the bootstrap if
	// just this registry has failed.
	Fail      bool     `yaml:"fail_on_error"`
	WhiteList []string `yaml:"white_list"`
	BlackList []string `yaml:"black_list"`
}

// Validate - makes sure the registry config is valid.
func (c Config) Validate() bool {
	if c.Name == "" {
		return false
	}
	switch c.AuthType {
	case "file":
		if c.AuthName == "" {
			return false
		}
	case "secret":
		if c.AuthName == "" {
			return false
		}
	case "config":
		if c.User == "" || c.Pass == "" {
			return false
		}
	case "":
		if c.AuthName != "" {
			return false
		}
	default:
		return false
	}

	m := regex.FindString(c.Name)
	return m == c.Name
}

// Registry - manages an adapter to retrieve and manage images to specs.
type Registry struct {
	adapter adapters.Adapter
	filter  Filter
	config  Config
}

// LoadSpecs - Load the specs for the registry.
func (r Registry) LoadSpecs() ([]*bundle.Spec, int, error) {
	imageNames, err := r.adapter.GetImageNames()
	if err != nil {
		log.Errorf("unable to retrieve image names for registry %v - %v",
			r.config.Name, err)
		return []*bundle.Spec{}, 0, err
	}
	validNames, filteredNames := r.filter.Run(imageNames)

	log.Debugf("Filter applied against registry: %s", r.config.Name)

	if len(validNames) != 0 {
		log.Debugf("APBs passing white/blacklist filter:")
		for _, name := range validNames {
			log.Debugf("-> %s", name)
		}
	}

	if len(filteredNames) != 0 {
		go func() {
			var buffer bytes.Buffer
			buffer.WriteString("APBs filtered by white/blacklist filter:")
			for _, name := range filteredNames {
				buffer.WriteString(fmt.Sprintf("-> %s", name))
			}
			log.Infof(buffer.String())
		}()
	}

	// Debug output filtered out names.
	specs, err := r.adapter.FetchSpecs(validNames)
	if err != nil {
		log.Errorf("unable to fetch specs for registry %v - %v",
			r.config.Name, err)
		return []*bundle.Spec{}, 0, err
	}

	log.Infof("Validating specs...")
	validatedSpecs := validateSpecs(specs)
	failedSpecsCount := len(specs) - len(validatedSpecs)

	if failedSpecsCount != 0 {
		log.Warningf(
			"%d specs of %d discovered specs failed validation from registry: %s",
			failedSpecsCount, len(specs), r.adapter.RegistryName())
	} else {
		log.Infof("All specs passed validation!")
	}

	return validatedSpecs, len(imageNames), nil
}

// Fail - will determine if the registry should cause a failure.
func (r Registry) Fail(err error) bool {
	if r.config.Fail {
		return true
	}
	return false
}

// RegistryName - retrieve the registry name to allow namespacing.
func (r Registry) RegistryName() string {
	return r.config.Name
}

// NewCustomRegistry - Create a new registry from the registry config.
func NewCustomRegistry(configuration Config, adapter adapters.Adapter, asbNamespace string) (Registry, error) {
	if !configuration.Validate() {
		return Registry{}, errors.New("unable to validate registry name")
	}

	// Retrieve registry auth if defined.
	configuration, err := retrieveRegistryAuth(configuration, asbNamespace)
	if err != nil {
		log.Errorf("Unable to retrieve registry auth: %v", err)
		return Registry{}, err
	}

	log.Info("== REGISTRY CX == ")
	log.Info(fmt.Sprintf("Name: %s", configuration.Name))
	log.Info(fmt.Sprintf("Type: %s", configuration.Type))
	log.Info(fmt.Sprintf("Url: %s", configuration.URL))
	// Validate URL
	u, err := url.Parse(configuration.URL)
	if err != nil {
		log.Errorf("url is not valid: %v", configuration.URL)
		// Default url, allow the registry to fail gracefully or un gracefully.
		u = &url.URL{}
	}
	if u.Scheme == "" {
		u.Scheme = "http"
	}

	if adapter == nil {
		c := adapters.Configuration{
			URL:        u,
			User:       configuration.User,
			Pass:       configuration.Pass,
			Org:        configuration.Org,
			Runner:     configuration.Runner,
			Images:     configuration.Images,
			Namespaces: configuration.Namespaces,
			Tag:        configuration.Tag,
		}

		switch strings.ToLower(configuration.Type) {
		case "rhcc":
			adapter = &adapters.RHCCAdapter{Config: c}
		case "dockerhub":
			adapter = &adapters.DockerHubAdapter{Config: c}
		case "mock":
			adapter = &adapters.MockAdapter{Config: c}
		case "openshift":
			adapter = &adapters.OpenShiftAdapter{Config: c}
		case "local_openshift":
			adapter = &adapters.LocalOpenShiftAdapter{Config: c}
		case "helm":
			adapter = &adapters.HelmAdapter{Config: c}
		default:
			panic("Unknown registry")
		}
	} else {
		log.Infof("Using custom adapter, %v", adapter.RegistryName())
	}

	return Registry{
		adapter: adapter,
		filter:  createFilter(configuration),
		config:  configuration,
	}, nil
}

// NewRegistry - Create a new registry from the registry config.
func NewRegistry(configuration Config, asbNamespace string) (Registry, error) {
	return NewCustomRegistry(configuration, nil, asbNamespace)
}

func createFilter(config Config) Filter {
	log.Debugf("Creating filter for registry: %s", config.Name)
	log.Debugf("whitelist: %v", config.WhiteList)
	log.Debugf("blacklist: %v", config.BlackList)

	filter := Filter{
		whitelist: config.WhiteList,
		blacklist: config.BlackList,
	}

	filter.Init()
	if len(filter.failedWhiteRegexp) != 0 {
		log.Warningf("Some whitelist regex failed for registry: %s", config.Name)
		for _, failed := range filter.failedWhiteRegexp {
			log.Warning(failed.regex)
			log.Warning(failed.err.Error())
		}
	}

	if len(filter.failedBlackRegexp) != 0 {
		log.Warningf("Some blacklist regex failed for registry: %s", config.Name)
		for _, failed := range filter.failedBlackRegexp {
			log.Warning(failed.regex)
			log.Warning(failed.err.Error())
		}
	}

	return filter
}

func validateSpecs(inSpecs []*bundle.Spec) []*bundle.Spec {
	var wg sync.WaitGroup
	wg.Add(len(inSpecs))

	type resultT struct {
		ok         bool
		spec       *bundle.Spec
		failReason string
	}

	out := make(chan resultT)
	for _, spec := range inSpecs {
		go func(s *bundle.Spec) {
			defer wg.Done()
			ok, failReason := validateSpecFormat(s)
			out <- resultT{ok, s, failReason}
		}(spec)
	}

	go func() {
		wg.Wait()
		close(out)
	}()

	validSpecs := make([]*bundle.Spec, 0, len(inSpecs))
	for result := range out {
		if result.ok {
			validSpecs = append(validSpecs, result.spec)
		} else {
			log.Warningf(
				"Spec [ %s ] failed validation for the following reason: [ %s ]. "+
					"It will not be made available.",
				result.spec.FQName, result.failReason,
			)
		}
	}

	return validSpecs
}

func validateSpecFormat(spec *bundle.Spec) (bool, string) {
	// Specs must have compatible version
	if !isCompatibleVersion(spec.Version, "1.0", "1.0") {
		return false, fmt.Sprintf(
			"APB Spec version [%v] out of bounds %v <= %v",
			spec.Version,
			"1.0",
			"1.0",
		)
	}

	// Specs must have compatible runtime version
	if !isCompatibleRuntime(spec.Runtime, 1, 2) {
		return false, fmt.Sprintf(
			"APB Runtime version [%v] out of bounds %v <= %v",
			spec.Runtime,
			1,
			2,
		)
	}

	// Specs must have at least one plan
	if !(len(spec.Plans) > 0) {
		return false, "Specs must have at least one plan"
	}

	dupes := make(map[string]bool)
	for _, plan := range spec.Plans {
		if _, contains := dupes[plan.Name]; contains {
			reason := fmt.Sprintf("%s: %s",
				"Plans within a spec must not contain duplicate value", plan.Name)

			return false, reason
		}
		dupes[plan.Name] = true
	}

	return true, ""
}

func isCompatibleVersion(specVersion string, minVersion string, maxVersion string) bool {
	if len(strings.Split(specVersion, ".")) != 2 || len(strings.Split(minVersion, ".")) != 2 || len(strings.Split(maxVersion, ".")) != 2 {
		return false
	}
	specMajorVersion, err := strconv.Atoi(strings.Split(specVersion, ".")[0])
	if err != nil {
		return false
	}
	minMajorVersion, err := strconv.Atoi(strings.Split(minVersion, ".")[0])
	if err != nil {
		return false
	}
	maxMajorVersion, err := strconv.Atoi(strings.Split(maxVersion, ".")[0])
	if err != nil {
		return false
	}

	if specMajorVersion >= minMajorVersion && specMajorVersion <= maxMajorVersion {
		return true
	}
	return false
}

func isCompatibleRuntime(specRuntime int, minVersion int, maxVersion int) bool {
	return specRuntime >= minVersion && specRuntime <= maxVersion
}

func retrieveRegistryAuth(reg Config, asbNamespace string) (Config, error) {
	var username, password string
	var err error
	switch reg.AuthType {
	case "secret":
		username, password, err = readSecret(reg.AuthName, asbNamespace)
		if err != nil {
			return Config{}, err
		}
	case "file":
		username, password, err = readFile(reg.AuthName)
		if err != nil {
			return Config{}, err
		}
	case "config":
		if reg.User == "" || reg.Pass == "" {
			return Config{}, fmt.Errorf("Failed to find registry credentials in config")
		}
		return reg, nil
	case "":
		// Assuming that the user has either no credentials or defined them in the config
		username = reg.User
		password = reg.Pass
	default:
		return Config{}, fmt.Errorf("Unrecognized registry AuthType: %s", reg.AuthType)
	}
	reg.User = username
	reg.Pass = password
	return reg, nil
}

func readFile(fileName string) (string, string, error) {
	regCred := struct {
		Username string
		Password string
	}{}

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
