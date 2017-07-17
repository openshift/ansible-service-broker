package adapters

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	b64 "encoding/base64"
	logging "github.com/op/go-logging"
	"github.com/openshift/ansible-service-broker/pkg/apb"
)

const openShiftName = "docker.io"
const openShiftAuthURL = "https://sso.redhat.com/auth/realms/rhc4tp/protocol/docker-v2/auth?service=docker-registry"
const openShiftManifestURL = "https://registry.connect.redhat.com/v2/%v/manifests/latest"

// OpenShiftAdapter - Docker Hub Adapter
type OpenShiftAdapter struct {
	Config Configuration
	Log    *logging.Logger
}

// OpenShiftImage - Image from a OpenShift registry.
type OpenShiftImage struct {
	Name      string `json:"name"`
	Namespace string `json:"namespace"`
}

// OpenShiftImageResponse - Image response for OpenShift.
type OpenShiftImageResponse struct {
	Count   int               `json:"count"`
	Results []*OpenShiftImage `json:"results"`
	Next    string            `json:"next"`
}

// RegistryName - Retrieve the registry name
func (r OpenShiftAdapter) RegistryName() string {
	return openShiftName
}

// GetImageNames - retrieve the images
func (r OpenShiftAdapter) GetImageNames() ([]string, error) {
	r.Log.Debug("OpenShiftAdapter::GetImages")
	r.Log.Debug("BundleSpecLabel: %s", BundleSpecLabel)

	images := r.Config.Images

	return images, nil
}

// FetchSpecs - retrieve the spec for the image names.
func (r OpenShiftAdapter) FetchSpecs(imageNames []string) ([]*apb.Spec, error) {
	specs := []*apb.Spec{}
	for _, imageName := range imageNames {
		spec, err := r.loadSpec(imageName)
		if err != nil {
			r.Log.Errorf("unable to retrieve spec data for image - %v", err)
			return specs, err
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
	var auth_string = username + ":" + password

	auth_string = b64.StdEncoding.EncodeToString([]byte(auth_string))

	req, err := http.NewRequest("GET", openShiftAuthURL, nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("Authorization", fmt.Sprintf("Basic %v", auth_string))

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

func (r OpenShiftAdapter) loadSpec(imageName string) (*apb.Spec, error) {
	req, err := http.NewRequest("GET", fmt.Sprintf(openShiftManifestURL, imageName), nil)
	if err != nil {
		return nil, err
	}
	token, err := r.getOpenShiftAuthToken()
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %v", token))
	return imageToSpec(r.Log, req)
}
