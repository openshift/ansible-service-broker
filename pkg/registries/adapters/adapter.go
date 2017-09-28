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

package adapters

import (
	b64 "encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	logging "github.com/op/go-logging"
	"github.com/openshift/ansible-service-broker/pkg/apb"
	"github.com/openshift/ansible-service-broker/pkg/version"
	yaml "gopkg.in/yaml.v1"
)

// Adapter - Adapter will wrap the methods that a registry needs to fully manage images.
type Adapter interface {
	// RegistryName will return the registiry prefix for the adapter.
	// Example is docker.io for the dockerhub adapter.
	RegistryName() string
	// GetImageNames will return all the image names for the adapter configuration.
	GetImageNames() ([]string, error)
	// FetchSpecs will retrieve all the specs for the list of images names.
	FetchSpecs([]string) ([]*apb.Spec, error)
}

// BundleSpecLabel - label on the image that we should use to pull out the abp spec.
// TODO: needs to remain ansibleapp UNTIL we redo the apps in dockerhub
const BundleSpecLabel = "com.redhat.apb.spec"

// Configuration - Adapter configuration. Contains the info that the adapter
// would need to complete its request to the images.
type Configuration struct {
	URL    *url.URL
	User   string
	Pass   string
	Org    string
	Images []string
	Tag    string
}

// Retrieve the spec from a registry manifest request
func imageToSpec(log *logging.Logger, req *http.Request, apbtag string) (*apb.Spec, error) {
	log.Debug("Registry::imageToSpec")
	spec := &apb.Spec{}
	req.Header.Add("Accept", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	type label struct {
		Spec    string `json:"com.redhat.apb.spec"`
		Version string `json:"com.redhat.apb.version"`
	}

	type config struct {
		Label label `json:"Labels"`
	}

	hist := struct {
		History []map[string]string `json:"history"`
	}{}

	conf := struct {
		Config *config `json:"config"`
	}{}

	err = json.NewDecoder(resp.Body).Decode(&hist)
	if err != nil {
		log.Errorf("Error grabbing JSON body from response: %s", err)
		return nil, err
	}

	if hist.History == nil {
		log.Errorf("V1 Schema Manifest does not exist in registry")
		return nil, nil
	}

	err = json.Unmarshal([]byte(hist.History[0]["v1Compatibility"]), &conf)
	if err != nil {
		log.Errorf("Error unmarshalling intermediary JSON response: %s", err)
		return nil, err
	}
	if conf.Config == nil {
		log.Infof("Did not find v1 Manifest in image history. Skipping image")
		return nil, nil
	}
	if conf.Config.Label.Spec == "" || conf.Config.Label.Version == "" {
		log.Infof("Didn't find encoded Spec or version label. Assuming image is not APB and skiping")
		return nil, nil
	}
	if isCompatibleVersion(conf.Config.Label.Version, version.MinAPBVersion, version.MaxAPBVersion) != true {
		log.Infof("APB spec version was incompatible. Assuming image is incompatible and skipping")
		return nil, nil
	}

	encodedSpec := conf.Config.Label.Spec
	decodedSpecYaml, err := b64.StdEncoding.DecodeString(encodedSpec)
	if err != nil {
		log.Errorf("Something went wrong decoding spec from label")
		return nil, err
	}

	if err = yaml.Unmarshal(decodedSpecYaml, spec); err != nil {
		log.Errorf("Something went wrong loading decoded spec yaml, %s", err)
		return nil, err
	}

	spec.Image = fmt.Sprintf("%s:%s", spec.Image, apbtag)

	log.Debugf("adapter::imageToSpec -> Got plans %+v", spec.Plans)
	log.Debugf("Successfully converted Image %s into Spec", spec.Image)

	return spec, nil
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
