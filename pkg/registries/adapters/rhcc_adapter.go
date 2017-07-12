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
	return r.Config.URL
}

// GetImages - retrieve the images from the registry
func (r RHCCAdapter) GetImages() ([]string, error) {
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
	for _, imageName := range imageNames {
		req, err := http.NewRequest("GET",
			fmt.Sprintf("%v/v2/%v/manifests/latest", r.Config.URL, imageName), nil)
		if err != nil {
			return specs, err
		}
		spec, err := imageToSpec(r.Log, req)
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
	r.Log.Debug("Using " + r.Config.URL + " to source APB images using query:" + Query)
	req, err := http.NewRequest("GET",
		fmt.Sprintf("%v/v1/search?q=%v", r.Config.URL, Query), nil)
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
