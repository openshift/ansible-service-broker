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

package adapters

import (
	"io/ioutil"

	"github.com/automationbroker/bundle-lib/apb"
	log "github.com/sirupsen/logrus"
	yaml "gopkg.in/yaml.v2"
)

// MockFile - Mock file contains fake regitry data
var MockFile = "/etc/automationbroker/bundle-lib/mock-registry-data.yaml"

// mockRegistryName - mock registry name
var mockRegistryName = "mock"

// MockAdapter - a adapter that is for mocking data
type MockAdapter struct {
	Config Configuration
	specs  map[string]*apb.Spec
}

// GetImageNames - retrieve the image names
func (r *MockAdapter) GetImageNames() ([]string, error) {
	log.Debug("MockRegistry::LoadSpecs")

	specYaml, err := ioutil.ReadFile(MockFile)
	if err != nil {
		log.Debug("Failed to read registry data from %s", MockFile)
		return nil, err
	}

	var parsedData struct {
		Apps []*apb.Spec `yaml:"apps"`
	}

	err = yaml.Unmarshal(specYaml, &parsedData)
	if err != nil {
		log.Error("Failed to ummarshal yaml file")
		return nil, err
	}

	log.Debug("Loaded Specs: %v", parsedData)

	log.Info("Loaded [ %d ] specs from %s registry", len(parsedData.Apps), "Mock")
	var names []string
	r.specs = map[string]*apb.Spec{}

	for _, spec := range parsedData.Apps {
		r.specs[spec.Image] = spec
		names = append(names, spec.Image)
	}
	return names, nil
}

// FetchSpecs - fetch the specs that were retrieved in the get images from the mock registry.
func (r MockAdapter) FetchSpecs(imageNames []string) ([]*apb.Spec, error) {
	specs := []*apb.Spec{}
	for _, name := range imageNames {
		spec, ok := r.specs[name]
		if ok {
			specs = append(specs, spec)
		}
	}
	return specs, nil
}

// RegistryName - retrieve the registry name
func (r MockAdapter) RegistryName() string {
	return mockRegistryName
}
