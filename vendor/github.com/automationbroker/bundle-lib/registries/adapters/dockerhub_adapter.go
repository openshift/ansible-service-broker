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
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/automationbroker/bundle-lib/bundle"
	log "github.com/sirupsen/logrus"
)

var (
	dockerhubName        = "docker.io"
	dockerHubLoginURL    = "https://hub.docker.com/v2/users/login/"
	dockerHubRepoImages  = "https://hub.docker.com/v2/repositories/%v/?page_size=100"
	dockerHubManifestURL = "https://registry.hub.docker.com/v2/%v/manifests/%v"
	dockerBearerTokenURL = "https://auth.docker.io/token"
)

// DockerHubAdapter - Docker Hub Adapter
type DockerHubAdapter struct {
	Config Configuration
}

// DockerHubImage - Image from a dockerhub registry.
type DockerHubImage struct {
	Name      string `json:"name"`
	Namespace string `json:"namespace"`
}

// DockerHubImageResponse - Image response for dockerhub.
type DockerHubImageResponse struct {
	Count   int               `json:"count"`
	Results []*DockerHubImage `json:"results"`
	Next    string            `json:"next"`
}

// RegistryName - Retrieve the registry name
func (r DockerHubAdapter) RegistryName() string {
	return dockerhubName
}

// GetImageNames - retrieve the images
func (r DockerHubAdapter) GetImageNames() ([]string, error) {
	log.Debug("DockerHubAdapter::GetImages")
	log.Debugf("BundleSpecLabel: %s", BundleSpecLabel)
	log.Debugf("Loading image list for org: [ %s ]", r.Config.Org)

	token, err := r.getDockerHubToken()
	if err != nil {
		log.Errorf("unable to generate docker hub token - %v", err)
		return nil, err
	}

	channel := make(chan string)
	ctx, cancelFunc := context.WithCancel(context.Background())
	defer cancelFunc()

	// Initial call to getNextImages this will fan out to retrieve all the values.
	imageResp, err := r.getNextImages(ctx, r.Config.Org, token,
		fmt.Sprintf(dockerHubRepoImages, r.Config.Org),
		channel, cancelFunc)
	// if there was an issue with the first call, return the error
	if err != nil {
		return nil, err
	}
	// If no results in the fist call then close the channel as nothing will get loaded.
	if len(imageResp.Results) == 0 {
		log.Info("canceled retrieval as no items in org")
		close(channel)
	}
	var bundleData []string
	counter := 1
	for imageData := range channel {
		bundleData = append(bundleData, imageData)
		if counter < imageResp.Count {
			counter++
		} else {
			close(channel)
		}
	}
	// check to see if the context had an error
	if ctx.Err() != nil {
		log.Errorf("encountered an error while loading images, we may not have all the apb in the catalog - %v", ctx.Err())
		return bundleData, ctx.Err()
	}

	return bundleData, nil
}

// FetchSpecs - retrieve the spec for the image names.
func (r DockerHubAdapter) FetchSpecs(imageNames []string) ([]*bundle.Spec, error) {
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

// getDockerHubToken - will retrieve the docker hub token.
func (r DockerHubAdapter) getDockerHubToken() (string, error) {
	type TokenResponse struct {
		Token string `json:"token"`
	}

	payloadBytes, err := json.Marshal(map[string]string{
		"username": r.Config.User,
		"password": r.Config.Pass,
	})
	if err != nil {
		return "", err
	}
	body := bytes.NewReader(payloadBytes)

	req, err := http.NewRequest("POST", dockerHubLoginURL, body)
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	jsonToken, err := ioutil.ReadAll(resp.Body)

	tokenResp := TokenResponse{}
	err = json.Unmarshal(jsonToken, &tokenResp)
	if err != nil {
		return "", err
	}
	return tokenResp.Token, nil
}

// getNextImages - will follow the next URL using go routines.
func (r DockerHubAdapter) getNextImages(ctx context.Context,
	org, token, url string,
	ch chan<- string,
	cancelFunc context.CancelFunc) (*DockerHubImageResponse, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		log.Errorf("unable to get next images for url: %v - %v", url, err)
		cancelFunc()
		close(ch)
		return nil, err
	}

	req.Header.Set("Authorization", fmt.Sprintf("JWT %v", token))

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Errorf("unable to get next images for url: %v - %v", url, err)
		cancelFunc()
		close(ch)
		return nil, err
	}
	defer resp.Body.Close()

	imageList, err := ioutil.ReadAll(resp.Body)

	iResp := DockerHubImageResponse{}
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
		// Fan out calls to get the next images.
		go r.getNextImages(ctx, org, token, iResp.Next, ch, cancelFunc)
	}
	for _, imageName := range iResp.Results {
		log.Debugf("Trying to load %v/%v", imageName.Namespace, imageName.Name)
		go func(image *DockerHubImage) {
			select {
			case <-ctx.Done():
				log.Debugf(
					"loading images failed due to context err - %v name - %v",
					ctx.Err(), image.Name)
				return
			default:
				ch <- fmt.Sprintf("%v/%v", image.Namespace, image.Name)
			}
		}(imageName)
	}
	return &iResp, nil
}

func (r DockerHubAdapter) loadSpec(imageName string) (*bundle.Spec, error) {
	if r.Config.Tag == "" {
		r.Config.Tag = "latest"
	}

	token, err := r.getBearerToken(imageName)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("GET", fmt.Sprintf(dockerHubManifestURL, imageName, r.Config.Tag), nil)
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
		return nil, fmt.Errorf("DockerHubAdapter::error handling dockerhub registery response %s", err)
	}
	return responseToSpec(body, fmt.Sprintf("%s/%s:%s", r.RegistryName(), imageName, r.Config.Tag))
}

func (r DockerHubAdapter) getBearerToken(imageName string) (string, error) {
	var err error
	var req *http.Request
	if r.Config.User == "" {
		req, err = http.NewRequest("GET",
			fmt.Sprintf("%s?service=registry.docker.io&scope=repository:%v:pull",
				dockerBearerTokenURL, imageName), nil)
		if err != nil {
			return "", err
		}
	} else {
		req, err = http.NewRequest("GET",
			fmt.Sprintf("%s?grant_type=password&service=registry.docker.io&scope=repository:%v:pull",
				dockerBearerTokenURL, imageName), nil)
		if err != nil {
			return "", err
		}
		req.SetBasicAuth(r.Config.User, r.Config.Pass)
	}
	response, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	defer response.Body.Close()
	t := struct {
		Token string `json:"token"`
	}{}
	err = json.NewDecoder(response.Body).Decode(&t)
	if err != nil {
		return "", err
	}
	return t.Token, nil
}
