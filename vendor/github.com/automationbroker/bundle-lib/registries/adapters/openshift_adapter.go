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
	"net/http"
	"strings"

	"github.com/automationbroker/bundle-lib/bundle"
	log "github.com/sirupsen/logrus"
)

const openShiftManifestURL = "%v/v2/%v/manifests/%v"

// OpenShiftAdapter - Docker Hub Adapter
type OpenShiftAdapter struct {
	Config Configuration
}

// OpenShiftImage - Image from a OpenShift registry.
type OpenShiftImage struct {
	Name      string `json:"name"`
	Namespace string `json:"namespace"`
}

// RegistryName - Retrieve the registry name
func (r OpenShiftAdapter) RegistryName() string {
	return strings.TrimPrefix(r.Config.URL.String(), "https://")
}

// GetImageNames - retrieve the images
func (r OpenShiftAdapter) GetImageNames() ([]string, error) {
	log.Debug("OpenShiftAdapter::GetImageNames")
	log.Debugf("BundleSpecLabel: %s", BundleSpecLabel)

	images := r.Config.Images
	log.Debugf("Configured to use images: %v", images)

	return images, nil
}

// FetchSpecs - retrieve the spec for the image names.
func (r OpenShiftAdapter) FetchSpecs(imageNames []string) ([]*bundle.Spec, error) {
	log.Debug("OpenShiftAdapter::FetchSpecs")
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

// getOpenShiftToken - will retrieve the docker hub token.
func (r OpenShiftAdapter) getOpenShiftAuthToken() (string, error) {
	type TokenResponse struct {
		Token string `json:"token"`
	}
	username := r.Config.User
	password := r.Config.Pass

	req, err := http.NewRequest("GET", fmt.Sprintf("%v/v2/", r.Config.URL), nil)
	if err != nil {
		return "", err
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	// Ensure that response holds data we expect
	if resp.Header.Get("Www-Authenticate") == "" {
		return "", errors.New("failed to find www-authenticate header from response")
	}

	authChallenge := resp.Header.Get("Www-Authenticate")
	if strings.Index(authChallenge, "realm=\"") == -1 {
		return "", errors.New("failed to find realm in www-authenticate header")
	}
	authOptions := ""
	if strings.Index(authChallenge, ",") != -1 {
		authOptions = strings.Split(authChallenge, ",")[1]
	}
	authRealm := strings.Split(strings.Split(authChallenge, "realm=\"")[1], "\"")[0]
	authURL := fmt.Sprintf("%v?%v", authRealm, authOptions)
	// Replace any quotes that exist in header from authOptions
	authURL = strings.Replace(authURL, "\"", "", -1)

	req, err = http.NewRequest("GET", authURL, nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")
	req.SetBasicAuth(username, password)

	resp, err = http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	tokenResp := TokenResponse{}
	err = json.NewDecoder(resp.Body).Decode(&tokenResp)
	if err != nil {
		return "", err
	}
	return tokenResp.Token, nil
}

func (r OpenShiftAdapter) loadSpec(imageName string) (*bundle.Spec, error) {
	log.Debug("OpenShiftAdapter::LoadSpec")
	if r.Config.Tag == "" {
		r.Config.Tag = "latest"
	}
	req, err := http.NewRequest("GET", fmt.Sprintf(openShiftManifestURL, r.Config.URL, imageName, r.Config.Tag), nil)
	if err != nil {
		return nil, err
	}
	token, err := r.getOpenShiftAuthToken()
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %v", token))
	req.Header.Add("Accept", "application/json")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	body, err := registryResponseHandler(resp)
	if err != nil {
		return nil, fmt.Errorf("OpenShiftAdapter::error handling openshift registery response %s", err)
	}
	return imageToSpec(body, fmt.Sprintf("%s/%s:%s", r.RegistryName(), imageName, r.Config.Tag))
}
