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

package registries

import (
	"bytes"
	"errors"
	"fmt"
	"regexp"
	"strings"
	"sync"

	logging "github.com/op/go-logging"
	"github.com/openshift/ansible-service-broker/pkg/apb"
	"github.com/openshift/ansible-service-broker/pkg/config"
	"github.com/openshift/ansible-service-broker/pkg/registries/adapters"
)

var regex = regexp.MustCompile(`[a-z0-9]([-a-z0-9]*[a-z0-9])?(\.[a-z0-9]([-a-z0-9]*[a-z0-9])?)*`)

// Registry - manages an adapter to retrieve and manage images to specs.
type Registry struct {
	adapter     adapters.Adapter
	log         *logging.Logger
	filter      Filter
	Name        string
	FailOnError bool
}

// LoadSpecs - Load the specs for the registry.
func (r Registry) LoadSpecs() ([]*apb.Spec, int, error) {
	imageNames, err := r.adapter.GetImageNames()
	if err != nil {
		r.log.Errorf("unable to retrieve image names for registry %v - %v",
			r.Name, err)
		return []*apb.Spec{}, 0, err
	}
	// Registry will throw out all images that do not end in -apb
	imageNames = registryFilterImagesForAPBs(imageNames)
	validNames, filteredNames := r.filter.Run(imageNames)

	r.log.Debug("Filter applied against registry: %s", r.Name)

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
			r.Name, err)
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
	if r.FailOnError {
		return true
	}
	return false
}

// RegistryName - retrieve the registry name to allow namespacing.
func (r Registry) RegistryName() string {
	return r.Name
}

// NewRegistry - Create a new registry from the registry config.
func NewRegistry(con *config.Config, log *logging.Logger) (Registry, error) {
	var adapter adapters.Adapter

	if !validName(con.GetString("name")) {
		return Registry{}, errors.New("unable to validate registry name")
	}

	log.Info("== REGISTRY CX == ")
	log.Info(fmt.Sprintf("Name: %s", con.GetString("name")))
	log.Info(fmt.Sprintf("Type: %s", con.GetString("type")))
	log.Info(fmt.Sprintf("Url: %s", con.GetString("url")))

	switch strings.ToLower(con.GetString("type")) {
	case "rhcc":
		adapter = &adapters.RHCCAdapter{Config: con, Log: log}
	case "dockerhub":
		adapter = &adapters.DockerHubAdapter{Config: con, Log: log}
	case "mock":
		adapter = &adapters.MockAdapter{Config: con, Log: log}
	case "openshift":
		adapter = &adapters.OpenShiftAdapter{Config: con, Log: log}
	default:
		panic("Unknown registry")
	}

	return Registry{
		adapter: adapter,
		log:     log,
		filter:  createFilter(con, log),
	}, nil
}

func createFilter(config *config.Config, log *logging.Logger) Filter {
	log.Debug("Creating filter for registry: %s", config.GetString("name"))
	log.Debug("whitelist: %v", config.GetSliceOfStrings("white_list"))
	log.Debug("blacklist: %v", config.GetSliceOfStrings("black_list"))

	filter := Filter{
		whitelist: config.GetSliceOfStrings("white_list"),
		blacklist: config.GetSliceOfStrings("black_list"),
	}

	filter.Init()
	if len(filter.failedWhiteRegexp) != 0 {
		log.Warning("Some whitelist regex failed for registry: %s", config.GetString("name"))
		for _, failed := range filter.failedWhiteRegexp {
			log.Warning(failed.regex)
			log.Warning(failed.err.Error())
		}
	}

	if len(filter.failedBlackRegexp) != 0 {
		log.Warning("Some blacklist regex failed for registry: %s", config.GetString("name"))
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
			ok, failReason := validateSpecPlans(s)
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

func validateSpecPlans(spec *apb.Spec) (bool, string) {
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

// validateName - makes sure the registry name is valid.
func validName(name string) bool {
	if name == "" {
		return false
	}
	m := regex.FindString(name)
	return m == name
}
