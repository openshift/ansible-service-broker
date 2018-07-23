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
	"context"
	"encoding/json"
	"fmt"
	"github.com/automationbroker/bundle-lib/bundle"
	log "github.com/sirupsen/logrus"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
)

const (
	defaultRunner     = "ansibleplaybookbundle/apb-base:latest"
	defaultURL        = "https://galaxy.ansible.com"
	galaxyName        = "galaxy"
	galaxySearchURL   = "%v/api/v1/content/?content_type__name=apb"
	galaxyNSSearchURL = "%v/api/v1/content/?content_type__name=apb&namespace__name=%v"
	galaxyRoleURL     = "%v/api/v1/content/%v/"
	galaxyAPIURL      = "%v/api/v1%v"
)

// GalaxyAdapter - Galaxy Adapter
type GalaxyAdapter struct {
	Config Configuration
}

// GalaxyRole - Role from Ansible Galaxy.
type GalaxyRole struct {
	Name    string            `json:"name"`
	RoleID  int               `json:"id"`
	Summary GalaxyRoleSummary `json:"summary_fields"`
}

// GalaxyRoleResponse - Role Response from Ansible Galaxy.
type GalaxyRoleResponse struct {
	Name     string             `json:"name"`
	Metadata GalaxyRoleMetadata `json:"metadata"`
	Summary  GalaxyRoleSummary  `json:"summary_fields"`
}

// GalaxyRoleMetadata - Role Metadata obtained from Role Response.
type GalaxyRoleMetadata struct {
	Spec bundle.Spec `json:"apb_metadata"`
}

// GalaxyRoleSummary - Role Summary obtained from Role Response.
type GalaxyRoleSummary struct {
	Namespace GalaxyRoleNamespace `json:"namespace"`
}

// GalaxyRoleNamespace - Role Namespace obtained from Role Response Summary.
type GalaxyRoleNamespace struct {
	Name string `json:"name"`
}

// GalaxySearchResponse - Search response for Galaxy.
type GalaxySearchResponse struct {
	Count   int           `json:"count"`
	Results []*GalaxyRole `json:"results"`
	Next    string        `json:"next"`
}

// RegistryName - Retrieve the registry name
func (r GalaxyAdapter) RegistryName() string {
	return galaxyName
}

// GetImageNames - retrieve the images
func (r *GalaxyAdapter) GetImageNames() ([]string, error) {
	log.Debug("GalaxyAdapter::GetImages")
	log.Debugf("BundleSpecLabel: %s", BundleSpecLabel)
	log.Debug("Loading role list with tag: [apb]")

	// default galaxy url
	if r.Config.URL.Host == "" {
		log.Debugf("Using default galaxy url: %v", defaultURL)
		r.Config.URL, _ = url.Parse(defaultURL)
	}

	// generate search url
	searchURL := fmt.Sprintf(galaxySearchURL, r.Config.URL.String())
	if len(r.Config.Org) != 0 {
		log.Debugf("Using namespaced galaxy search")
		searchURL = fmt.Sprintf(galaxyNSSearchURL, r.Config.URL.String(), r.Config.Org)
	}

	channel := make(chan string)
	ctx, cancelFunc := context.WithCancel(context.Background())
	defer cancelFunc()

	var imageResp *GalaxySearchResponse
	var err error

	// Initial call to getNextImages this will fan out to retrieve all the values.
	imageResp, err = r.getNextImages(ctx, searchURL, channel, cancelFunc)

	// if there was an issue with the first call, return the error
	if err != nil {
		return nil, err
	}

	// If no results in the first call then close the channel as nothing will get loaded.
	if len(imageResp.Results) == 0 {
		log.Info("canceled retrieval as no items in org")
		close(channel)
	}
	var apbData []string
	counter := 1
	for imageData := range channel {
		apbData = append(apbData, imageData)
		if counter < imageResp.Count {
			counter++
		} else {
			close(channel)
		}
	}
	// check to see if the context had an error
	if ctx.Err() != nil {
		log.Errorf("encountered an error while loading images, we may not have all the apb in the catalog - %v", ctx.Err())
		return apbData, ctx.Err()
	}

	return apbData, nil
}

// FetchSpecs - retrieve the spec for the image names.
func (r GalaxyAdapter) FetchSpecs(imageNames []string) ([]*bundle.Spec, error) {

	// default galaxy url
	if r.Config.URL.Host == "" {
		log.Debugf("Using default galaxy url: %v", defaultURL)
		r.Config.URL, _ = url.Parse(defaultURL)
	}

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

// getNextImages - will follow the next URL using go routines.
func (r GalaxyAdapter) getNextImages(ctx context.Context,
	url string, ch chan<- string,
	cancelFunc context.CancelFunc) (*GalaxySearchResponse, error) {

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		log.Errorf("unable to get next roles for url: %v - %v", url, err)
		cancelFunc()
		close(ch)
		return nil, err
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Errorf("unable to get next roles for url: %v - %v", url, err)
		cancelFunc()
		close(ch)
		return nil, err
	}
	defer resp.Body.Close()

	imageList, err := ioutil.ReadAll(resp.Body)

	iResp := GalaxySearchResponse{}
	err = json.Unmarshal(imageList, &iResp)
	if err != nil {
		log.Errorf("unable to get next images for url: %v - %v", url, err)
		cancelFunc()
		close(ch)
		return &iResp, err
	}
	// Keep getting the images
	if iResp.Next != "" {
		log.Debugf("getting next page of results - %v", iResp.Next)
		searchURL := fmt.Sprintf(galaxyAPIURL, r.Config.URL.String(), iResp.Next)
		// Fan out calls to get the next images.
		go r.getNextImages(ctx, searchURL, ch, cancelFunc)
	}
	for _, imageName := range iResp.Results {
		log.Debugf("Trying to load %v.%v", imageName.Summary.Namespace.Name, imageName.Name)
		go func(image *GalaxyRole) {
			select {
			case <-ctx.Done():
				log.Debugf(
					"loading images failed due to context err - %v name - %v",
					ctx.Err(), image.Name)
				return
			default:
				ch <- fmt.Sprintf("%v.%v#%v", image.Summary.Namespace.Name, image.Name, image.RoleID)
			}
		}(imageName)
	}
	return &iResp, nil
}

func (r GalaxyAdapter) loadSpec(imageName string) (*bundle.Spec, error) {

	imageSplit := strings.Split(imageName, "#")
	if len(imageSplit) < 2 {
		log.Debugf("Expected image [%v] to be of the form roleName#roleID", imageName)
		return nil, nil
	}
	roleName := imageSplit[0]
	roleID := imageSplit[1]

	req, err := http.NewRequest("GET", fmt.Sprintf(galaxyRoleURL, r.Config.URL.String(), roleID), nil)
	if err != nil {
		return nil, err
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	role, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	roleResp := GalaxyRoleResponse{}
	err = json.Unmarshal(role, &roleResp)
	if err != nil {
		return nil, err
	}

	spec := roleResp.Metadata.Spec
	spec.Runtime = 2
	spec.Image = defaultRunner
	if len(r.Config.Runner) != 0 {
		spec.Image = r.Config.Runner
	}
	// Override the name, or else APBs from galaxy will be indistinguishable from dockerhub
	spec.Metadata["displayName"] = fmt.Sprintf("%s (galaxy)", roleName)

	roleParam := bundle.ParameterDescriptor{
		Name:      "role_name",
		Title:     "Galaxy Role Name",
		Type:      "string",
		Updatable: false,
		Required:  true,
		Default:   roleResp.Name,
		Pattern:   fmt.Sprintf("^%s$", roleResp.Name),
	}
	namespaceParam := bundle.ParameterDescriptor{
		Name:      "role_namespace",
		Title:     "Galaxy Role Namespace",
		Type:      "string",
		Updatable: false,
		Required:  true,
		Default:   roleResp.Summary.Namespace.Name,
		Pattern:   fmt.Sprintf("^%s$", roleResp.Summary.Namespace.Name),
	}
	for key, plan := range spec.Plans {
		plan.Parameters = append([]bundle.ParameterDescriptor{roleParam, namespaceParam}, plan.Parameters...)
		spec.Plans[key].Parameters = plan.Parameters
	}

	return &spec, nil
}
