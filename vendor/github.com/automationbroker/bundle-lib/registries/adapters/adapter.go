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
	b64 "encoding/base64"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/url"
	"strconv"

	"fmt"

	"github.com/automationbroker/bundle-lib/apb"
	log "github.com/sirupsen/logrus"
	yaml "gopkg.in/yaml.v1"
)

// Adapter - Adapter will wrap the methods that a registry needs to fully manage images.
type Adapter interface {
	// RegistryName will return the registiry prefix for the adapter.
	// Example is docker.io for the dockerhub adapter.
	RegistryName() string
	// GetImageNames will return all the image names for the adapter configuration.
	GetImageNames() ([]string, error)
	// FetchSpecs will retrieve all the specs for the list of images names.
	FetchSpecs([]string) ([]*apb.Spec, error)
}

// BundleSpecLabel - label on the image that we should use to pull out the apb spec.
// TODO: needs to remain ansibleapp UNTIL we redo the apps in dockerhub
const BundleSpecLabel = "com.redhat.apb.spec"

// Configuration - Adapter configuration. Contains the info that the adapter
// would need to complete its request to the images.
type Configuration struct {
	URL           *url.URL
	User          string
	Pass          string
	Org           string
	Runner        string
	Images        []string
	Namespaces    []string
	Tag           string
	SkipVerifyTLS bool
}

type registryResponseError struct {
	code    int
	message string
}

type imageLabel struct {
	Spec          string `json:"com.redhat.apb.spec"`
	Runtime       string `json:"com.redhat.apb.runtime"`
	BundleRuntime string `json:"com.redhat.bundle.runtime"`
}

type config struct {
	Label  imageLabel `json:"Labels"`
	Digest string     `json:"digest"`
}

type manifestResponse struct {
	SchemaVersion int                 `json:"schemaVersion"`
	History       []map[string]string `json:"history"`
}

type manifestConfig struct {
	Config config `json:"config"`
}

func (rre *registryResponseError) Error() string {
	return fmt.Sprintf("unexpected registry response code: %v message: %v", rre.code, rre.message)
}

func registryResponseHandler(resp *http.Response) ([]byte, error) {
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode == http.StatusUnauthorized {
		return nil, &registryResponseError{code: resp.StatusCode, message: "Unable to authenticate to the registry, registry credentials could be invalid"}
	}

	if resp.StatusCode != http.StatusOK {
		return nil, &registryResponseError{code: resp.StatusCode, message: fmt.Sprintf("unexpected response code %v body %v", resp.StatusCode, string(body))}
	}
	return body, nil
}

/*
// Retrieve the spec from a registry manifest request
func imageToSpec(req *http.Request, image string) (*apb.Spec, error) {
	log.Debug("Registry::imageToSpec")
	spec := &apb.Spec{}
	req.Header.Add("Accept", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	type label struct {
		Spec    string `json:"com.redhat.apb.spec"`
		Runtime string `json:"com.redhat.apb.runtime"`
	}

	type config struct {
		Label label `json:"Labels"`
	}

	hist := struct {
		History []map[string]string `json:"history"`
	}{}

	conf := struct {
		Config *config `json:"config"`
	}{}

	if resp.StatusCode == http.StatusUnauthorized {
		log.Errorf("Unable to authenticate to the registry, registry credentials could be invalid.")
		return nil, nil
	}

	// resp.Body is an io.Reader, which are a one time use.  Save the
	// contents to a byte[] for debugging, then remake the io.Reader for the
	// JSON decoder.
	debug, _ := ioutil.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		log.Errorf("Image '%s' may not exist in registry.", image)
		log.Error(string(debug))
		return nil, nil
	}

	r := bytes.NewReader(debug)
	err = json.NewDecoder(r).Decode(&hist)
	if err != nil {
		log.Errorf("Error grabbing JSON body from response: %s", err)
		return nil, err
	}

	if hist.History == nil {
		log.Errorf("V1 Schema Manifest does not exist in registry")
		return nil, nil
	}

	err = json.Unmarshal([]byte(hist.History[0]["v1Compatibility"]), &conf)
	if err != nil {
		log.Errorf("Error unmarshalling intermediary JSON response: %s", err)
		return nil, err
	}
	if conf.Config == nil {
		log.Infof("Did not find v1 Manifest in image history. Skipping image")
		return nil, nil
	}
	if conf.Config.Label.Spec == "" {
		log.Infof("Didn't find encoded Spec label. Assuming image is not APB and skiping")
		return nil, nil
	}

	encodedSpec := conf.Config.Label.Spec
	decodedSpecYaml, err := b64.StdEncoding.DecodeString(encodedSpec)
	if err != nil {
		log.Errorf("Something went wrong decoding spec from label")
		return nil, err
	}

	if err = yaml.Unmarshal(decodedSpecYaml, spec); err != nil {
		log.Errorf("Something went wrong loading decoded spec yaml, %s", err)
		return nil, err
	}

	spec.Runtime, err = getAPBRuntimeVersion(conf.Config.Label.Runtime)
	if err != nil {
		return nil, err
	}

	spec.Image = image

	log.Debugf("adapter::imageToSpec -> Got plans %+v", spec.Plans)
	log.Debugf("Successfully converted Image %s into Spec", spec.Image)

	return spec, nil
}
*/

// Retrieve the spec from a manifest response
func responseToSpec(response []byte, image string) (*apb.Spec, error) {
	mResp := manifestResponse{}

	r := bytes.NewReader(response)
	if err := json.NewDecoder(r).Decode(&mResp); err != nil {
		log.Errorf("Error grabbing JSON body from manifest response: %s", err)
		return nil, err
	}
	return configToSpec([]byte(mResp.History[0]["v1Compatibility"]), image)
}

// Retrieve the spec from manifest config
func configToSpec(config []byte, image string) (*apb.Spec, error) {
	spec := &apb.Spec{}
	mConf := manifestConfig{}

	r := bytes.NewReader(config)
	err := json.NewDecoder(r).Decode(&mConf)
	if err != nil {
		log.Errorf("Failed to unmarshal config object for image [%s]: %s", image, err)
		return nil, err
	}

	// encoded spec
	if mConf.Config.Label.Spec == "" {
		log.Infof("Didn't find encoded Spec label. Assuming image is not APB and skipping")
		return nil, nil
	}

	decodedSpecYaml, err := b64.StdEncoding.DecodeString(mConf.Config.Label.Spec)
	if err != nil {
		log.Errorf("Something went wrong decoding spec from label for '%s' : %s", image, err)
		return nil, err
	}
	if err = yaml.Unmarshal(decodedSpecYaml, spec); err != nil {
		log.Errorf("Something went wrong loading decoded spec yaml for '%s' : %s", image, err)
		return nil, err
	}

	// prefer bundle runtime
	version := mConf.Config.Label.Runtime
	if mConf.Config.Label.BundleRuntime != "" {
		log.Debugf("bundle runtime present using this over apb runtime")
		version = mConf.Config.Label.BundleRuntime
	}
	spec.Runtime, err = getAPBRuntimeVersion(version)
	if err != nil {
		return nil, err
	}

	// image name to be pulled during provision
	spec.Image = image

	log.Debugf("Successfully converted Image %s into Spec", spec.Image)
	log.Infof("adapter::configToSpec -> Image %s runtime is %d", spec.Image, spec.Runtime)

	return spec, nil
}

func getAPBRuntimeVersion(version string) (int, error) {

	if version == "" {
		log.Infof("No runtime label found. Set runtime=1. Will use 'exec' to gather bind credentials")
		return 1, nil
	}

	runtime, err := strconv.Atoi(version)
	if err != nil {
		log.Errorf("Unable to parse APB runtime version - %v", err)
		return 0, err
	}
	return runtime, nil
}
