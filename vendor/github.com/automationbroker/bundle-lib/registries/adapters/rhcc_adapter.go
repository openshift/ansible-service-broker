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
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"

	"github.com/automationbroker/bundle-lib/bundle"
	"github.com/automationbroker/bundle-lib/registries/adapters/oauth"
	log "github.com/sirupsen/logrus"
)

// NewRHCCAdapter - creates and returns a *RHCCAdapter ready to use.
func NewRHCCAdapter(config Configuration) *RHCCAdapter {
	return &RHCCAdapter{
		Config: config,
		client: oauth.NewClient(config.User, config.Pass, config.SkipVerifyTLS, config.URL),
	}
}

// RHCCAdapter - Red Hat Container Catalog Registry
type RHCCAdapter struct {
	Config Configuration
	client *oauth.Client
}

// RHCCImage - RHCC Registry Image that is returned from the RHCC Catalog api.
type RHCCImage struct {
	Description  string `json:"description"`
	IsOfficial   bool   `json:"is_official"`
	IsTrusted    bool   `json:"is_trusted"`
	Name         string `json:"name"`
	ShouldFilter bool   `json:"should_filter"`
	StarCount    int    `json:"star_count"`
}

// RHCCImageResponse - RHCC Registry Image Response returned for the RHCC Catalog api
type RHCCImageResponse struct {
	NumResults int          `json:"num_results"`
	Query      string       `json:"query"`
	Results    []*RHCCImage `json:"results"`
}

// RegistryName - retrieve the registry prefix
func (r RHCCAdapter) RegistryName() string {
	if r.Config.URL.Host == "" {
		return r.Config.URL.Path
	}
	return r.Config.URL.Host
}

// GetImageNames - retrieve the images from the registry
func (r RHCCAdapter) GetImageNames() ([]string, error) {
	r.client.Getv2()
	imageList, err := r.loadImages("\"*-apb\"")
	if err != nil {
		return nil, err
	}
	imageNames := []string{}
	for _, image := range imageList.Results {
		imageNames = append(imageNames, image.Name)
	}
	return imageNames, nil
}

// FetchSpecs - retrieve the spec from the image names
func (r RHCCAdapter) FetchSpecs(imageNames []string) ([]*bundle.Spec, error) {
	log.Debug("RHCCAdapter::FetchSpecs")
	specs := []*bundle.Spec{}
	for _, imageName := range imageNames {
		log.Debugf("%v", imageName)
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

// LoadImages - Get all the images for a particular query
func (r RHCCAdapter) loadImages(query string) (RHCCImageResponse, error) {
	log.Debug("RHCCRegistry::LoadImages")
	req, err := r.client.NewRequest("/v1/search")
	if err != nil {
		return RHCCImageResponse{}, err
	}
	q := req.URL.Query()
	q.Set("q", query)
	req.URL.RawQuery = q.Encode()
	log.Debug("Using " + req.URL.String() + " to source APB images ")

	resp, err := r.client.Do(req)
	if err != nil {
		return RHCCImageResponse{}, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return RHCCImageResponse{}, errors.New(resp.Status)
	}
	imageList, err := ioutil.ReadAll(resp.Body)

	imageResp := RHCCImageResponse{}
	err = json.Unmarshal(imageList, &imageResp)
	if err != nil {
		return RHCCImageResponse{}, err
	}
	log.Debug("Properly unmarshalled image response")

	return imageResp, nil
}

func (r RHCCAdapter) loadSpec(imageName string) (*bundle.Spec, error) {
	log.Debug("RHCCAdapter::LoadSpec")
	if r.Config.Tag == "" {
		r.Config.Tag = "latest"
	}
	req, err := r.client.NewRequest(fmt.Sprintf("/v2/%v/manifests/%v", imageName, r.Config.Tag))
	if err != nil {
		return nil, err
	}
	req.Header.Add("Accept", "application/json")
	resp, err := r.client.Do(req)
	if err != nil {
		return nil, err
	}
	body, err := registryResponseHandler(resp)
	if err != nil {
		return nil, fmt.Errorf("RHCCAdapter::error handling openshift registery response %s", err)
	}

	return imageToSpec(body, fmt.Sprintf("%s/%s:%s", r.RegistryName(), imageName, r.Config.Tag))
}
