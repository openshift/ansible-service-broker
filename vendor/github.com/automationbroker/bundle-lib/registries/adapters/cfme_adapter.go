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
	"net/http"

	"github.com/automationbroker/bundle-lib/apb"
	log "github.com/sirupsen/logrus"
)

// RHCCAdapter - Red Hat Container Catalog Registry
type RHCCAdapter struct {
	Config Configuration
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
func (r RHCCAdapter) FetchSpecs(imageNames []string) ([]*apb.Spec, error) {
	log.Debug("RHCCAdapter::FetchSpecs")
	specs := []*apb.Spec{}
	for _, imageName := range imageNames {
		log.Debug("%v", imageName)
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
func (r RHCCAdapter) loadImages(Query string) (RHCCImageResponse, error) {
	log.Debug("RHCCRegistry::LoadImages")
	log.Debug("Using " + r.Config.URL.String() + " to source APB images using query:" + Query)
	req, err := http.NewRequest("GET",
		fmt.Sprintf("%v/v1/search?q=%v", r.Config.URL.String(), Query), nil)
	if err != nil {
		return RHCCImageResponse{}, err
	}

	resp, err := http.DefaultClient.Do(req)
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

func (r RHCCAdapter) loadSpec(imageName string) (*apb.Spec, error) {
	log.Debug("RHCCAdapter::LoadSpec")
	if r.Config.Tag == "" {
		r.Config.Tag = "latest"
	}
	req, err := http.NewRequest("GET",
		fmt.Sprintf("%v/v2/%v/manifests/%v", r.Config.URL.String(), imageName, r.Config.Tag), nil)
	if err != nil {
		return nil, err
	}

	return imageToSpec(req, fmt.Sprintf("%s/%s:%s", r.RegistryName(), imageName, r.Config.Tag))
}
