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

package registries

import (
	"bytes"
	"errors"
	"fmt"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"sync"

	logging "github.com/op/go-logging"
	"github.com/openshift/ansible-service-broker/pkg/apb"
	"github.com/openshift/ansible-service-broker/pkg/config"
	"github.com/openshift/ansible-service-broker/pkg/metrics"
	"github.com/openshift/ansible-service-broker/pkg/registries/adapters"
	"github.com/openshift/ansible-service-broker/pkg/version"
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
	log     *logging.Logger
	filter  Filter
	config  Config
}

// LoadSpecs - Load the specs for the registry.
func (r Registry) LoadSpecs() ([]*apb.Spec, int, error) {
	imageNames, err := r.adapter.GetImageNames()
	if err != nil {
		r.log.Errorf("unable to retrieve image names for registry %v - %v",
			r.config.Name, err)
		return []*apb.Spec{}, 0, err
	}
	// Registry will throw out all images that do not end in -apb
	imageNames = registryFilterImagesForAPBs(imageNames)
	validNames, filteredNames := r.filter.Run(imageNames)

	r.log.Debug("Filter applied against registry: %s", r.config.Name)

	if len(validNames) != 0 {
		r.log.Debugf("APBs passing white/blacklist filter:")
		for _, name := range validNames {
			r.log.Debugf("-> %s", name)
		}
	}

	if len(filteredNames) != 0 {
		go func() {
			var buffer bytes.Buffer
			buffer.WriteString("APBs filtered by white/blacklist filter:")
			for _, name := range filteredNames {
				buffer.WriteString(fmt.Sprintf("-> %s", name))
			}
			r.log.Infof(buffer.String())
		}()
	}

	// Debug output filtered out names.
	specs, err := r.adapter.FetchSpecs(validNames)
	if err != nil {
		r.log.Errorf("unable to fetch specs for registry %v - %v",
			r.config.Name, err)
		return []*apb.Spec{}, 0, err
	}

	r.log.Infof("Validating specs...")
	validatedSpecs := validateSpecs(r.log, specs)
	failedSpecsCount := len(specs) - len(validatedSpecs)

	if failedSpecsCount != 0 {
		r.log.Warningf(
			"%d specs of %d discovered specs failed validation from registry: %s",
			failedSpecsCount, len(specs), r.adapter.RegistryName())
	} else {
		r.log.Notice("All specs passed validation!")
	}
	metrics.SpecsLoaded(r.RegistryName(), len(validatedSpecs))

	return validatedSpecs, len(imageNames), nil
}

func registryFilterImagesForAPBs(imageNames []string) []string {
	newNames := []string{}
	for _, imagesName := range imageNames {
		if strings.HasSuffix(strings.ToLower(imagesName), "-apb") {
			newNames = append(newNames, imagesName)
		}
	}
	return newNames
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

// NewRegistry - Create a new registry from the registry config.
func NewRegistry(con *config.Config, log *logging.Logger) (Registry, error) {
	log.Debugf("configu - %#v", con)
	var adapter adapters.Adapter
	configuration := Config{
		URL:        con.GetString("url"),
		User:       con.GetString("user"),
		Pass:       con.GetString("pass"),
		Org:        con.GetString("org"),
		Tag:        con.GetString("tag"),
		Type:       con.GetString("type"),
		Name:       con.GetString("name"),
		Images:     con.GetSliceOfStrings("images"),
		Namespaces: con.GetSliceOfStrings("namespaces"),
		Fail:       con.GetBool("fail_on_error"),
		WhiteList:  con.GetSliceOfStrings("white_list"),
		BlackList:  con.GetSliceOfStrings("black_list"),
		AuthType:   con.GetString("auth_type"),
		AuthName:   con.GetString("auth_name"),
	}
	if !configuration.Validate() {
		return Registry{}, errors.New("unable to validate registry name")
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
	c := adapters.Configuration{URL: u,
		User:       configuration.User,
		Pass:       configuration.Pass,
		Org:        configuration.Org,
		Images:     configuration.Images,
		Namespaces: configuration.Namespaces,
		Tag:        configuration.Tag}

	switch strings.ToLower(configuration.Type) {
	case "rhcc":
		adapter = &adapters.RHCCAdapter{Config: c, Log: log}
	case "dockerhub":
		adapter = &adapters.DockerHubAdapter{Config: c, Log: log}
	case "mock":
		adapter = &adapters.MockAdapter{Config: c, Log: log}
	case "openshift":
		adapter = &adapters.OpenShiftAdapter{Config: c, Log: log}
	case "local_openshift":
		adapter = &adapters.LocalOpenShiftAdapter{Config: c, Log: log}
	default:
		panic("Unknown registry")
	}

	return Registry{
		adapter: adapter,
		log:     log,
		filter:  createFilter(configuration, log),
		config:  configuration,
	}, nil
}

func createFilter(config Config, log *logging.Logger) Filter {
	log.Debug("Creating filter for registry: %s", config.Name)
	log.Debug("whitelist: %v", config.WhiteList)
	log.Debug("blacklist: %v", config.BlackList)

	filter := Filter{
		whitelist: config.WhiteList,
		blacklist: config.BlackList,
	}

	filter.Init()
	if len(filter.failedWhiteRegexp) != 0 {
		log.Warning("Some whitelist regex failed for registry: %s", config.Name)
		for _, failed := range filter.failedWhiteRegexp {
			log.Warning(failed.regex)
			log.Warning(failed.err.Error())
		}
	}

	if len(filter.failedBlackRegexp) != 0 {
		log.Warning("Some blacklist regex failed for registry: %s", config.Name)
		for _, failed := range filter.failedBlackRegexp {
			log.Warning(failed.regex)
			log.Warning(failed.err.Error())
		}
	}

	return filter
}

func validateSpecs(log *logging.Logger, inSpecs []*apb.Spec) []*apb.Spec {
	var wg sync.WaitGroup
	wg.Add(len(inSpecs))

	type resultT struct {
		ok         bool
		spec       *apb.Spec
		failReason string
	}

	out := make(chan resultT)
	for _, spec := range inSpecs {
		go func(s *apb.Spec) {
			defer wg.Done()
			ok, failReason := validateSpecFormat(s)
			out <- resultT{ok, s, failReason}
		}(spec)
	}

	go func() {
		wg.Wait()
		close(out)
	}()

	validSpecs := make([]*apb.Spec, 0, len(inSpecs))
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

func validateSpecFormat(spec *apb.Spec) (bool, string) {
	// Specs must have compatible version
	if !isCompatibleVersion(spec.Version, version.MinAPBVersion, version.MaxAPBVersion) {
		return false, fmt.Sprintf(
			"APB Spec version [%v] out of bounds %v <= %v",
			spec.Version,
			version.MinAPBVersion,
			version.MaxAPBVersion,
		)
	}

	// Specs must have compatible runtime version
	if !isCompatibleRuntime(spec.Runtime, version.MinRuntimeVersion, version.MaxRuntimeVersion) {
		return false, fmt.Sprintf(
			"APB Runtime version [%v] out of bounds %v <= %v",
			spec.Runtime,
			version.MinRuntimeVersion,
			version.MaxRuntimeVersion,
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
