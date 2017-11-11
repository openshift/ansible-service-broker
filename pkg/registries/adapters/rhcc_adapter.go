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

package adapters

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	logging "github.com/op/go-logging"
	"github.com/openshift/ansible-service-broker/pkg/apb"
)

// RHCCAdapter - Red Hat Container Catalog Registry
type RHCCAdapter struct {
	Config Configuration
	Log    *logging.Logger
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
	specs := []*apb.Spec{}
	if r.Config.Tag == "" {
		r.Config.Tag = "latest"
	}
	for _, imageName := range imageNames {
		req, err := http.NewRequest("GET",
			fmt.Sprintf("%v/v2/%v/manifests/%v", r.Config.URL.String(), imageName, r.Config.Tag), nil)
		if err != nil {
			return specs, err
		}
		spec, err := imageToSpec(r.Log, req, fmt.Sprintf("%s/%s:%s", r.RegistryName(), imageName, r.Config.Tag))
		if err != nil {
			return specs, err
		}
		if spec != nil {
			specs = append(specs, spec)
		}
	}
	return specs, nil
}

// LoadImages - Get all the images for a particular query
func (r RHCCAdapter) loadImages(Query string) (RHCCImageResponse, error) {
	r.Log.Debug("RHCCRegistry::LoadImages")
	r.Log.Debug("Using " + r.Config.URL.String() + " to source APB images using query:" + Query)
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

	r.Log.Debug("Got Image Response from RHCC")
	imageList, err := ioutil.ReadAll(resp.Body)

	imageResp := RHCCImageResponse{}
	err = json.Unmarshal(imageList, &imageResp)
	if err != nil {
		return RHCCImageResponse{}, err
	}
	r.Log.Debug("Properly unmarshalled image response")

	return imageResp, nil
}
