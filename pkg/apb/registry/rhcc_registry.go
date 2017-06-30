package registry

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/url"

	logging "github.com/op/go-logging"
	"github.com/openshift/ansible-service-broker/pkg/apb"
)

// RHCCRegistry - Red Hat Container Catalog Registry
type RHCCRegistry struct {
	config Config
	log    *logging.Logger
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

// Init - Initialize the Red Hat Container Catalog
func (r *RHCCRegistry) Init(config Config, log *logging.Logger) error {
	log.Debug("RHCCRegistry::Init")
	r.log = log
	u, err := url.Parse(config.URL)
	if err != nil {
		r.log.Errorf("URL is to valid - %v", err)
		return err
	}
	if u.Scheme == "" {
		u.Scheme = "http"
	}
	config.URL = u.String()
	r.config = config
	return nil
}

// LoadSpecs - Load Red Hat Container Catalog specs
func (r RHCCRegistry) LoadSpecs() ([]*apb.Spec, int, error) {
	r.log.Debug("RHCCRegistry::LoadSpecs")
	var specs []*apb.Spec

	imageList, err := r.LoadImages("\"*-apb\"")
	if err != nil {
		return []*apb.Spec{}, 0, err
	}

	numResults := imageList.NumResults
	r.log.Debug("Found %d images in RHCC", numResults)
	for _, image := range imageList.Results {
		req, err := http.NewRequest("GET", r.config.URL+"/v2/"+image.Name+"/manifests/latest", nil)
		if err != nil {
			return []*apb.Spec{}, 0, err
		}
		spec, err := imageToSpec(r.log, req)
		if err != nil {
			return []*apb.Spec{}, 0, err
		}
		if spec != nil {
			spec.RegistryName = r.config.Name
			specs = append(specs, spec)
		}
	}
	return specs, numResults, nil
}

// Fail - will determine if this reqistry can cause a failure.
func (r RHCCRegistry) Fail(err error) bool {
	if r.config.Fail {
		return true
	}
	return false
}

// LoadImages - Get all the images for a particular query
func (r RHCCRegistry) LoadImages(Query string) (RHCCImageResponse, error) {
	r.log.Debug("RHCCRegistry::LoadImages")
	r.log.Debug("Using " + r.config.URL + " to source APB images using query:" + Query)
	req, err := http.NewRequest("GET", r.config.URL+"/v1/search?q="+Query, nil)
	if err != nil {
		return RHCCImageResponse{}, err
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return RHCCImageResponse{}, err
	}
	defer resp.Body.Close()

	r.log.Debug("Got Image Response from RHCC")
	imageList, err := ioutil.ReadAll(resp.Body)

	imageResp := RHCCImageResponse{}
	err = json.Unmarshal(imageList, &imageResp)
	if err != nil {
		return RHCCImageResponse{}, err
	}
	r.log.Debug("Properly unmarshalled image response")

	return imageResp, nil
}
