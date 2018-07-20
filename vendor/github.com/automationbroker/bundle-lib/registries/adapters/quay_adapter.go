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
	b64 "encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"github.com/automationbroker/bundle-lib/bundle"
	log "github.com/sirupsen/logrus"
	yaml "gopkg.in/yaml.v2"
)

const (
	quayName        = "quay.io"
	quayCatalogURL  = "%v/api/v1/repository?public=true&private=true&namespace=%v"
	quayDigestURL   = "%v/api/v1/repository/%v/%v"
	quayManifestURL = "%v/api/v1/repository/%v/%v/manifest/%v/labels"
)

// QuayAdapter - Quay Adapter
type QuayAdapter struct {
	config Configuration
}

type quayRepository struct {
	IsPublic    bool   `json:"is_public"`
	Kind        string `json:"kind"`
	Namespace   string `json:"namespace"`
	Name        string `json:"name"`
	Description string `json:"description"`
}

type quayImageResponse struct {
	Repositories []quayRepository `json:"repositories"`
}

// NewQuayAdapter - creates and returns a QuayAdapter ready to use.
func NewQuayAdapter(config Configuration) (QuayAdapter, error) {
	a := QuayAdapter{
		config: config,
	}

	// set Tag to latest if empty
	if a.config.Tag == "" {
		a.config.Tag = "latest"
	}

	return a, nil
}

// RegistryName - Retrieve the registry name
func (r QuayAdapter) RegistryName() string {
	return quayName
}

// GetImageNames - retrieve the images
func (r QuayAdapter) GetImageNames() ([]string, error) {
	log.Debug("QuayAdapter::GetImages")
	log.Debug("BundleSpecLabel: %s", BundleSpecLabel)
	log.Debug("Loading image list for quay.io Org: [ %v ]", r.config.Org)

	var imageList []string

	// check if we're configured for specific images
	if len(r.config.Images) > 0 {
		log.Debugf("Configured to use images: %v", r.config.Images)
		imageList = append(imageList, r.config.Images...)
	}

	// discover images
	req, err := http.NewRequest("GET", fmt.Sprintf(quayCatalogURL, r.config.URL, r.config.Org), nil)
	if err != nil {
		return nil, err
	}
	req.Header.Add("Accept", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %v", r.config.Token))

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Errorf("Failed to load catalog response at %s - %v", fmt.Sprintf(quayCatalogURL, r.config.URL, r.config.Org), err)
		return nil, err
	}
	defer resp.Body.Close()

	catalogResp := quayImageResponse{}
	err = json.NewDecoder(resp.Body).Decode(&catalogResp)
	if err != nil {
		log.Errorf("Failed to decode Catalog response from '%s'", fmt.Sprintf(quayCatalogURL, r.config.URL, r.config.Org))
		return nil, err
	}

	for _, repo := range catalogResp.Repositories {
		imageList = append(imageList, repo.Name)
	}

	if len(imageList) == 0 {
		log.Warn("image list is empty. No images were discovered")
		return imageList, nil
	}

	var uniqueList []string
	imageMap := make(map[string]struct{})
	for _, image := range imageList {
		imageMap[image] = struct{}{}
	}

	// create a unique image list
	for key := range imageMap {
		uniqueList = append(uniqueList, key)
	}
	return uniqueList, nil
}

// FetchSpecs - retrieve the spec for the image names.
func (r QuayAdapter) FetchSpecs(imageNames []string) ([]*bundle.Spec, error) {
	specs := []*bundle.Spec{}
	for _, imageName := range imageNames {
		spec, err := r.loadSpec(imageName)
		if err != nil {
			log.Errorf("Failed to retrieve spec data for image %s - %v", imageName, err)
		}
		if spec != nil {
			specs = append(specs, spec)
		}
	}
	return specs, nil
}

func (r QuayAdapter) loadSpec(imageName string) (*bundle.Spec, error) {
	digest, err := r.getDigest(imageName)
	if err != nil {
		return nil, err
	}
	return r.digestToSpec(digest, imageName)
}

func (r QuayAdapter) getDigest(imageName string) (string, error) {
	req, err := http.NewRequest("GET", fmt.Sprintf(quayDigestURL, r.config.URL, r.config.Org, imageName), nil)
	if err != nil {
		return "", err
	}
	req.Header.Add("Accept", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %v", r.config.Token))

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	type repoResponse struct {
		Tags map[string]interface{} `json:"tags"`
	}

	digestResp := repoResponse{}
	err = json.NewDecoder(resp.Body).Decode(&digestResp)
	if err != nil {
		log.Errorf("unable to get repository Info for image: %s - %v", imageName, err)
		return "", err
	}

	var digest string
	for key, item := range digestResp.Tags {
		if key == r.config.Tag {
			if tag, ok := item.(map[string]interface{}); ok {
				digest = tag["manifest_digest"].(string)
				break
			}
		}
	}

	if digest == "" {
		return "", errors.New("unable to get manifest_digest")
	}

	return digest, nil
}

func (r QuayAdapter) digestToSpec(digest string, imageName string) (*bundle.Spec, error) {
	if digest == "" {
		return nil, errors.New("digest is nil")
	}

	spec := &bundle.Spec{}
	req, err := http.NewRequest("GET", fmt.Sprintf(quayManifestURL, r.config.URL, r.config.Org, imageName, digest), nil)
	if err != nil {
		return nil, err
	}
	req.Header.Add("Accept", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %v", r.config.Token))

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	type label struct {
		Key   string `json:"key"`
		Value string `json:"value"`
	}

	type imageLabels struct {
		Label []label `json:"labels"`
	}

	manifestResp := imageLabels{}
	err = json.NewDecoder(resp.Body).Decode(&manifestResp)
	if err != nil {
		log.Errorf("Unable to get Spec for [%s]: - %v", imageName, err)
		return nil, err
	}

	var encodedSpec, runtime string
	for _, l := range manifestResp.Label {
		if l.Key == "com.redhat.apb.spec" {
			encodedSpec = l.Value
		} else if l.Key == "com.redhat.apb.runtime" {
			runtime = l.Value
		}
	}

	if encodedSpec == "" {
		return nil, errors.New("Spec not found")
	}

	decodedSpecYaml, err := b64.StdEncoding.DecodeString(encodedSpec)
	if err != nil {
		log.Errorf("Something went wrong decoding spec from label")
		return nil, err
	}

	if err = yaml.Unmarshal(decodedSpecYaml, spec); err != nil {
		log.Errorf("Something went wrong loading decoded spec yaml, %s", err)
		return nil, err
	}

	spec.Runtime, err = getAPBRuntimeVersion(runtime)
	if err != nil {
		return nil, err
	}

	registryName := r.config.URL.Hostname()
	if r.config.URL.Port() != "" {
		registryName = fmt.Sprintf("%s:%s", r.config.URL.Hostname(), r.config.URL.Port())
	}

	spec.Image = fmt.Sprintf("%s/%s/%s:%s", registryName, r.config.Org, imageName, r.config.Tag)

	log.Debugf("adapter::imageToSpec -> Got plans %+v", spec.Plans)
	log.Debugf("Successfully converted Image '%s' into Spec", spec.Image)
	return spec, nil
}
