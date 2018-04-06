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
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/automationbroker/bundle-lib/apb"
	log "github.com/sirupsen/logrus"
	"io/ioutil"
	"net/http"
	"regexp"
	"strings"
)

// CFMEAdapter - Red Hat Container Catalog Registry
type CFMEAdapter struct {
	Config Configuration
}

// CFMEImageResponse - CFME Registry Image Response returned for the CFME Catalog api
type CFMEImageResponse struct {
	NumResults int          `json:"count"`
	Results    []*CFMEImage `json:"resources"`
}

// CFMEImage - CFME Registry Image that is returned from the CFME Catalog api.
type CFMEImage struct {
	Href string `json:"href"`
}

type CFMEServiceTemplate struct {
	Id             string          `json:"id"`
	Name           string          `json:"name"`
	Description    string          `json:"description"`
	CFMEConfigInfo *CFMEConfigInfo `json:"config_info"`
}

type CFMEConfigInfo struct {
	CFMEProvision map[string]interface{} `json:"provision"`
}

type CFMEServiceDialog struct {
	Id                       string                      `json:"id"`
	Name                     string                      `json:"label"`
	Description              string                      `json:"description"`
	CFMEServiceDialogContent []*CFMEServiceDialogContent `json:"content"`
}

type CFMEServiceDialogContent struct {
	Id                    string                  `json:"id"`
	Name                  string                  `json:"label"`
	Description           string                  `json:"description"`
	CFMEServiceDialogTabs []*CFMEServiceDialogTab `json:"dialog_tabs"`
}

type CFMEServiceDialogTab struct {
	Id                      string                    `json:"id"`
	Name                    string                    `json:"label"`
	Description             string                    `json:"description"`
	CFMEServiceDialogGroups []*CFMEServiceDialogGroup `json:"dialog_groups"`
}

type CFMEServiceDialogGroup struct {
	Id                      string                    `json:"id"`
	Name                    string                    `json:"label"`
	Description             string                    `json:"description"`
	CFMEServiceDialogFields []*CFMEServiceDialogField `json:"dialog_fields"`
}

type CFMEServiceDialogField struct {
	Id          string `json:"id"`
	Name        string `json:"label"`
	Description string `json:"description"`
	Type        string `json:"type"`
	Required    bool   `json:"required"`
	Values      string `json:"values"`
	Default     string `json:"default_value"`
}

// RegistryName - retrieve the registry pr
func (r CFMEAdapter) RegistryName() string {
	if r.Config.URL.Host == "" {
		return r.Config.URL.Path
	}
	return r.Config.URL.Host
}

// GetImageNames - retrieve the images from the registry
func (r CFMEAdapter) GetImageNames() ([]string, error) {
	imageNames := []string{}
	imageList, err := r.loadImages()
	if err != nil {
		return imageNames, err
	}

	for _, image := range imageList.Results {
		imageNames = append(imageNames, image.Href)
	}

	return imageNames, nil
}

func (r CFMEAdapter) getServiceTemplates(imageNames []string) ([]CFMEServiceTemplate, error) {
	log.Debug("CFMERegistry::getServiceTemplates")
	serviceTemplates := []CFMEServiceTemplate{}
	for _, imageName := range imageNames {
		req, err := http.NewRequest("GET", string(imageName), nil)
		if err != nil {
			return []CFMEServiceTemplate{}, err
		}
		req.SetBasicAuth(r.Config.User, r.Config.Pass)

		transport := &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		}
		httpClient := &http.Client{Transport: transport}

		resp, err := httpClient.Do(req)
		if err != nil {
			return []CFMEServiceTemplate{}, err
		}
		defer resp.Body.Close()

		if resp.StatusCode != 200 {
			return []CFMEServiceTemplate{}, errors.New(resp.Status)
		}
		template, err := ioutil.ReadAll(resp.Body)

		templateResp := CFMEServiceTemplate{}
		err = json.Unmarshal(template, &templateResp)
		if err != nil {
			return []CFMEServiceTemplate{}, err
		}
		log.Debug("Properly unmarshalled image response")

		serviceTemplates = append(serviceTemplates, templateResp)
	}

	return serviceTemplates, nil
}

func (r CFMEAdapter) getServiceDialogs(dialogList []string) ([]CFMEServiceDialog, error) {
	log.Debug("CFMERegistry::getServiceDialogs")
	serviceDialogs := []CFMEServiceDialog{}
	for _, dialogId := range dialogList {
		req, err := http.NewRequest("GET", fmt.Sprintf("%v/api/service_dialogs/%v", r.Config.URL.String(), dialogId), nil)
		if err != nil {
			return []CFMEServiceDialog{}, err
		}
		req.SetBasicAuth(r.Config.User, r.Config.Pass)

		transport := &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		}
		httpClient := &http.Client{Transport: transport}

		resp, err := httpClient.Do(req)
		if err != nil {
			return []CFMEServiceDialog{}, err
		}
		defer resp.Body.Close()

		if resp.StatusCode != 200 {
			return []CFMEServiceDialog{}, errors.New(resp.Status)
		}
		dialog, err := ioutil.ReadAll(resp.Body)

		dialogResp := CFMEServiceDialog{}
		err = json.Unmarshal(dialog, &dialogResp)
		if err != nil {
			return []CFMEServiceDialog{}, err
		}
		log.Debug("Properly unmarshalled image response")

		serviceDialogs = append(serviceDialogs, dialogResp)
	}

	return serviceDialogs, nil
}

// FetchSpecs - retrieve the spec from the image names
func (r CFMEAdapter) FetchSpecs(imageNames []string) ([]*apb.Spec, error) {
	log.Debug("CFMEAdapter::FetchSpecs")
	var specs []*apb.Spec

	templates, err := r.getServiceTemplates(imageNames)
	if err != nil {
		log.Errorf("Failed to retrieve templates: %v", err)
	}

	for _, template := range templates {
		log.Debug("%v", template.Name)

		dataMap := map[string]string{"template_id": template.Id}

		var re = regexp.MustCompile(`[()_,. ]`)
		normalizedName := strings.ToLower(re.ReplaceAllString(template.Name, `$1-$2`))

		// Convert Service Template to Spec
		spec := &apb.Spec{
			Version:     "1.0",
			FQName:      normalizedName + "-apb",
			Async:       "optional",
			Bindable:    false,
			Image:       "https://docker.io/manageiq/manageiq-apb-runner:latest",
			Tags:        []string{"iaas"},
			Description: template.Description,
			Runtime:     2,
			Metadata: map[string]interface{}{
				"displayName":      template.Name + " (APB)",
				"documentationUrl": r.Config.URL.String(),
				"dependencies":     "https://docker.io/manageiq/manageiq-apb-runner:latest",
				"imageUrl":         "https://s3.amazonaws.com/fusor/2017demo/ManageIQ.png",
			},
			Plans: []apb.Plan{
				apb.Plan{
					Name:        "default",
					Description: "Default deployment plan for " + normalizedName + "-apb",
					Metadata: map[string]interface{}{
						"displayName":     "Default",
						"longDescription": template.Description,
						"cost":            "$0.0",
					},
					Parameters: []apb.ParameterDescriptor{
						apb.ParameterDescriptor{
							Name:         "cfme_user",
							Title:        "CFME Requestor",
							Type:         "string",
							Updatable:    false,
							Required:     true,
							DisplayGroup: "CloudForms Credentials",
						},
						apb.ParameterDescriptor{
							Name:         "cfme_password",
							Title:        "CFME Password",
							Type:         "string",
							Updatable:    false,
							Required:     true,
							DisplayType:  "password",
							DisplayGroup: "CloudForms Credentials",
						},
						apb.ParameterDescriptor{
							Name:         "cfme_url",
							Title:        "CFME URL",
							Type:         "string",
							Updatable:    false,
							Required:     true,
							Default:      r.Config.URL.String(),
							DisplayGroup: "CloudForms Credentials",
						},
					},
				},
			},
		}

		var dialogIds []string
		dialogObject := template.CFMEConfigInfo.CFMEProvision
		for key, value := range dialogObject {
			if key == "dialog_id" {
				dialogIds = append(dialogIds, value.(string))
			}
		}

		serviceDialogs, err := r.getServiceDialogs(dialogIds)
		if err != nil {
			log.Errorf("Failed to retrieve spec data for image %s - %v", template.Name, err)
		}

		for _, serviceDialog := range serviceDialogs {
			for _, content := range serviceDialog.CFMEServiceDialogContent {
				for _, tab := range content.CFMEServiceDialogTabs {
					for _, group := range tab.CFMEServiceDialogGroups {
						for _, field := range group.CFMEServiceDialogFields {
							param := apb.ParameterDescriptor{}
							param.Name = field.Name
							param.Title = field.Name
							param.DisplayGroup = tab.Name + "/" + group.Name
							if field.Required == true {
								param.Required = true
							}
							// FIXME: Cover Types a lot better
							if field.Type == "DialogFieldCheckBox" {
								param.Type = "bool"
								if field.Default == "t" {
									param.Default = true
								}
							} else if field.Type == "DialogFieldDropDownList" {
								param.Type = "enum"
								param.Default = field.Default

							} else if field.Type == "DialogFieldRadioButton" {
								param.Type = "enum"
								param.Default = field.Default
							} else {
								param.Type = "string"
							}
							spec.Plans[0].Parameters = append(spec.Plans[0].Parameters, param)
						}
					}
				}
			}
		}

		dataMapJson, err := json.Marshal(dataMap)
		if err != nil {
			log.Errorf("Failed to retrieve spec data for image %s - %v", template.Name, err)
		}

		dataMapParam := apb.ParameterDescriptor{
			Name:         "data_map",
			Title:        "Data Map",
			Type:         "string",
			Updatable:    false,
			Required:     true,
			Default:      string(dataMapJson),
			DisplayGroup: "CloudForms Credentials",
		}
		spec.Plans[0].Parameters = append(spec.Plans[0].Parameters, dataMapParam)

		specs = append(specs, spec)

	}

	return specs, nil
}

// LoadImages - Get all the images for a particular query
func (r CFMEAdapter) loadImages() (CFMEImageResponse, error) {
	log.Debug("CFMERegistry::LoadImages")
	log.Debug("Using " + r.Config.URL.String() + " to source APB images.")
	req, err := http.NewRequest("GET",
		fmt.Sprintf("%v/api/service_templates", r.Config.URL.String()), nil)
	if err != nil {
		return CFMEImageResponse{}, err
	}
	req.SetBasicAuth(r.Config.User, r.Config.Pass)

	transport := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	httpClient := &http.Client{Transport: transport}

	resp, err := httpClient.Do(req)
	if err != nil {
		return CFMEImageResponse{}, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return CFMEImageResponse{}, errors.New(resp.Status)
	}
	imageList, err := ioutil.ReadAll(resp.Body)

	imageResp := CFMEImageResponse{}
	err = json.Unmarshal(imageList, &imageResp)
	if err != nil {
		return CFMEImageResponse{}, err
	}
	log.Debug("Properly unmarshalled image response")

	return imageResp, nil
}

func (r CFMEAdapter) loadSpec(imageName string) (*apb.Spec, error) {
	log.Debug("CFMEAdapter::LoadSpec")
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
