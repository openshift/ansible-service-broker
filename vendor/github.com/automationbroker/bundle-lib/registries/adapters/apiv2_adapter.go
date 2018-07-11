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
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/automationbroker/bundle-lib/bundle"
	"github.com/automationbroker/bundle-lib/registries/adapters/oauth"
	log "github.com/sirupsen/logrus"
)

const (
	apiV2ManifestURL = "%v/v2/%v/manifests/%v"
	apiV2CatalogURL  = "%v/v2/_catalog"
)

// OpenShiftAdapter - OpenShift Adapter
type OpenShiftAdapter struct {
	APIV2Adapter
}

// PartnerRhccAdapter - Partner RHCC Adapter
type PartnerRhccAdapter struct {
	APIV2Adapter
}

// APIV2Adapter - API V2 Adapter
type APIV2Adapter struct {
	config Configuration
	client *oauth.Client
}

// apiV2CatalogResponse - Catalog Response
type apiV2CatalogResponse struct {
	Repositories []string `json:"repositories"`
}

// NewOpenShiftAdapter - creates a new OpenShift Adapter
func NewOpenShiftAdapter(config Configuration) (OpenShiftAdapter, error) {
	apiV2, err := NewAPIV2Adapter(config)
	if err != nil {
		return OpenShiftAdapter{}, err
	}
	return OpenShiftAdapter{apiV2}, nil
}

// NewPartnerRhccAdapter - creates a new Partner RHCC Adapter
func NewPartnerRhccAdapter(config Configuration) (PartnerRhccAdapter, error) {
	apiV2, err := NewAPIV2Adapter(config)
	if err != nil {
		return PartnerRhccAdapter{}, err
	}
	return PartnerRhccAdapter{apiV2}, nil
}

// NewAPIV2Adapter - creates and returns a APIV2Adapter ready to use.
func NewAPIV2Adapter(config Configuration) (APIV2Adapter, error) {
	apiv2a := APIV2Adapter{
		config: config,
		client: oauth.NewClient(config.User, config.Pass, config.SkipVerifyTLS, config.URL),
	}

	// Authorization
	err := apiv2a.client.Getv2()
	if err != nil {
		log.Errorf("Failed to GET /v2 at %s - %v", config.URL, err)
		return APIV2Adapter{}, err
	}

	// set Tag to latest if empty
	if apiv2a.config.Tag == "" {
		apiv2a.config.Tag = "latest"
	}

	return apiv2a, nil
}

// RegistryName - Retrieve the registry name
func (r APIV2Adapter) RegistryName() string {
	if r.config.URL.Host == "" {
		return r.config.URL.Path
	}
	return r.config.URL.Host
}

// GetImageNames - retrieve the images
func (r APIV2Adapter) GetImageNames() ([]string, error) {
	log.Debugf("%s - GetImageNames", r.config.AdapterName)
	log.Debugf("BundleSpecLabel: %s", BundleSpecLabel)

	var imageList []string

	// check if we're configured for specific images
	if len(r.config.Images) > 0 {
		log.Debugf("Configured to use images: %v", r.config.Images)
		imageList = append(imageList, r.config.Images...)
	}

	// discover images from URL
	discoveredImages, err := r.discoverImages(fmt.Sprintf(apiV2CatalogURL, r.config.URL))
	if err != nil && len(imageList) == 0 {
		return nil, err
	}
	if len(discoveredImages) > 0 {
		log.Debugf("Discovered images: %v", discoveredImages)
		imageList = append(imageList, discoveredImages...)
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
func (r APIV2Adapter) FetchSpecs(imageNames []string) ([]*bundle.Spec, error) {
	log.Debugf("%s - FetchSpecs", r.config.AdapterName)
	specs := []*bundle.Spec{}
	for _, imageName := range imageNames {
		log.Debugf("%v", imageName)
		spec, err := r.loadSpec(imageName)
		if err != nil {
			log.Errorf("Failed to retrieve spec data for image %s:%s - %v", imageName, r.config.Tag, err)
		}
		if spec != nil {
			specs = append(specs, spec)
		}
	}
	return specs, nil
}

// discoverImages - Get Imagenames from the /v2/_catalog URL
func (r APIV2Adapter) discoverImages(url string) ([]string, error) {
	log.Debugf("%s - discoverImages", r.config.AdapterName)

	// Initial URL
	if len(url) == 0 {
		return nil, errors.New("url is empty")
	}

	// Get all Image Names until the 'Link' value is empty
	// https://docs.docker.com/registry/spec/api/#pagination
	var imageList []string
	for {
		images, linkStr, err := r.getNextImages(url)
		if err != nil {
			return imageList, err
		}
		log.Debugf("discovered images from - %s", url)
		imageList = append(imageList, images.Repositories...)

		// no more to get..
		if len(linkStr) == 0 {
			break
		}

		url = r.getNextImageURL(linkStr)
		if len(url) == 0 {
			return imageList, errors.New("invalid next image URL")
		}
	}
	return imageList, nil
}

// getNextImages - Get the next 'Link' URL.
func (r APIV2Adapter) getNextImages(url string) (*apiV2CatalogResponse, string, error) {
	req, err := r.client.NewRequest(url)
	if err != nil {
		return nil, "", err
	}

	resp, err := r.client.Do(req)
	if err != nil {
		return nil, "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		log.Errorf("Failed to fetch catalog response from '%s'. Expected a 200 status and got: %v", url, resp.Status)
		return nil, "", errors.New(resp.Status)
	}

	imageList := apiV2CatalogResponse{}
	err = json.NewDecoder(resp.Body).Decode(&imageList)
	if err != nil {
		return nil, "", err
	}
	log.Debug("Properly unmarshalled image response")

	return &imageList, resp.Header.Get("Link"), nil
}

// getNextImageURL - returns the next image URL created from the 'Link' in the header
func (r APIV2Adapter) getNextImageURL(link string) string {
	if len(link) == 0 {
		log.Errorf("'Link' value is empty")
		return ""
	}

	res := strings.Split(link, ";")
	if len(res[0]) == 0 {
		log.Errorf("Invalid Link value")
		return ""
	}

	var lvalue string
	lvalue = strings.TrimSpace(res[0])
	lvalue = strings.Trim(lvalue, "<>")
	return (r.config.URL.String() + lvalue)
}

func (r APIV2Adapter) loadSpec(imageName string) (*bundle.Spec, error) {
	log.Debugf("%s - LoadSpec", r.config.AdapterName)

	req, err := r.client.NewRequest(fmt.Sprintf(apiV2ManifestURL, r.config.URL, imageName, r.config.Tag))
	if err != nil {
		return nil, err
	}

	resp, err := r.client.Do(req)
	if err != nil {
		return nil, err
	}

	body, err := registryResponseHandler(resp)
	if err != nil {
		return nil, fmt.Errorf("%s - error handling registry response %s", r.config.AdapterName, err)
	}

	registryName := r.config.URL.Hostname()
	if r.config.URL.Port() != "" {
		registryName = fmt.Sprintf("%s:%s", r.config.URL.Hostname(), r.config.URL.Port())
	}

	schemaVersion, err := getSchemaVersion(body)
	if err != nil {
		return nil, err
	}

	switch schemaVersion {
	case 1:
		log.Debugf("manifest schema 1 for image [%s]", imageName)
		return responseToSpec(body, fmt.Sprintf("%s/%s:%s", registryName, imageName, r.config.Tag))
	case 2:
		log.Debugf("manifest schema 2 for image [%s]", imageName)
		mConf := manifestConfig{}
		rdr := bytes.NewReader(body)

		// get the digest
		err = json.NewDecoder(rdr).Decode(&mConf)
		if err != nil {
			log.Errorf("unable to get digest for image [%s]: %v", imageName, err)
			return nil, err
		}
		digest := mConf.Config.Digest

		// get response with digest
		req, err = r.client.NewRequest(fmt.Sprintf("%s/v2/%s/blobs/%s", r.config.URL, imageName, digest))
		if err != nil {
			return nil, err
		}
		resp, err = r.client.Do(req)
		if err != nil {
			return nil, err
		}
		body, err = registryResponseHandler(resp)
		if err != nil {
			return nil, fmt.Errorf("%s - error getting configuration object for image [%s] : %s", r.config.AdapterName, imageName, err)
		}
		return configToSpec(body, fmt.Sprintf("%s/%s:%s", registryName, imageName, r.config.Tag))
	default:
		return nil, errors.New("unsupported schema version")
	}
}
