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
	"strings"

	"github.com/automationbroker/bundle-lib/bundle"
	log "github.com/sirupsen/logrus"
)

const partnerName = "partner_rhcc"
const partnerManifestURL = "%v/v2/%v/manifests/%v"
const partnerCatalogURL = "%v/v2/_catalog"

// PartnerRhccAdapter - Partner RHCC Adapter
type PartnerRhccAdapter struct {
	Config Configuration
}

// PartnerCatalogResponse - Partner Catalog Response
type PartnerCatalogResponse struct {
	Repositories []string `json:"repositories"`
}

// RegistryName - Retrieve the registry name
func (r PartnerRhccAdapter) RegistryName() string {
	return partnerName
}

// GetImageNames - retrieve the images
func (r PartnerRhccAdapter) GetImageNames() ([]string, error) {
	log.Debug("PartnerRhccAdapter::GetImageNames")
	log.Debugf("BundleSpecLabel: %s", BundleSpecLabel)

	if r.Config.Images != nil {
		log.Debugf("Configured to use images: %v", r.Config.Images)
		return r.Config.Images, nil
	}
	log.Debugf("Did not find images in config, attempting to discover from %s/v2/_catalog", r.Config.URL)

	req, err := http.NewRequest("GET", fmt.Sprintf(partnerCatalogURL, r.Config.URL), nil)
	if err != nil {
		return nil, err
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Errorf("Failed to load catalog response at %s - %v", partnerCatalogURL, err)
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		log.Errorf("Failed to fetch catalog response. Expected a 200 status and got: %v", resp.Status)
		return nil, errors.New(resp.Status)
	}
	imageResp, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	imageList := PartnerCatalogResponse{}
	err = json.Unmarshal(imageResp, &imageList)
	if err != nil {
		return nil, err
	}

	return imageList.Repositories, nil
}

// FetchSpecs - retrieve the spec for the image names.
func (r PartnerRhccAdapter) FetchSpecs(imageNames []string) ([]*bundle.Spec, error) {
	log.Debug("PartnerRhccAdapter::FetchSpecs")
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

// getAuthToken - will retrieve the docker hub token.
func (r PartnerRhccAdapter) getAuthToken() (string, error) {
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

func (r PartnerRhccAdapter) loadSpec(imageName string) (*bundle.Spec, error) {
	log.Debug("PartnerRhccAdapter::LoadSpec")
	if r.Config.Tag == "" {
		r.Config.Tag = "latest"
	}
	req, err := http.NewRequest("GET", fmt.Sprintf(partnerManifestURL, r.Config.URL, imageName, r.Config.Tag), nil)
	if err != nil {
		return nil, err
	}
	token, err := r.getAuthToken()
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %v", token))
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	body, err := registryResponseHandler(resp)
	if err != nil {
		return nil, fmt.Errorf("PartnerRhccAdapter::error handling registry response %s", err)
	}
	registryName := r.Config.URL.Hostname()
	if r.Config.URL.Port() != "" {
		registryName = fmt.Sprintf("%s:%s", r.Config.URL.Hostname(), r.Config.URL.Port())
	}

	return imageToSpec(body, fmt.Sprintf("%s/%s:%s", registryName, imageName, r.Config.Tag))
}
