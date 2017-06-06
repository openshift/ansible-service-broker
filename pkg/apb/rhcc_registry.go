package apb

import (
	"encoding/json"
	logging "github.com/op/go-logging"
	"io/ioutil"
	"net/http"
	"strings"
)

// RHCCRegistry - Red Hat Container Catalog Registry
type RHCCRegistry struct {
	config RegistryConfig
	log    *logging.Logger
}

type Image struct {
	Description  string `json:"description"`
	IsOfficial   bool   `json:"is_official"`
	IsTrusted    bool   `json:"is_trusted"`
	Name         string `json:"name"`
	ShouldFilter bool   `json:"should_filter"`
	StarCount    int    `json:"star_count"`
}

type ImageResponse struct {
	NumResults int      `json:"num_results"`
	Query      string   `json:"query"`
	Results    []*Image `json:"results"`
}

// Init - Initialize the Red Hat Container Catalog
func (r *RHCCRegistry) Init(config RegistryConfig, log *logging.Logger) error {
	log.Debug("RHCCRegistry::Init")
	r.config = config
	r.log = log
	return nil
}

// LoadSpecs - Load Red Hat Container Catalog specs
func (r RHCCRegistry) LoadSpecs() ([]*Spec, int, error) {
	r.log.Debug("RHCCRegistry::LoadSpecs")
	var specs []*Spec

	imageList, err := r.LoadImages("apb")
	if err != nil {
		return []*Spec{}, 0, err
	}

	numResults := imageList.NumResults
	r.log.Debug("Found %d images in RHCC", numResults)
	for _, image := range imageList.Results {
		spec, err := r.imageToSpec(image)
		if err != nil {
			return []*Spec{}, 0, err
		}
		specs = append(specs, spec)
	}

	return specs, numResults, nil
}

func (r RHCCRegistry) imageToSpec(image *Image) (*Spec, error) {
	r.log.Debug("RHCCRegistry::imageToSpec")
	_spec := &Spec{}
	// Setting ID to foo because Dao expects non blank IDs
	_spec.Id = "foo_id"
	var name = strings.SplitAfter(image.Name, "/")
	_spec.Name = name[len(name)-1]
	_spec.Image = image.Name
	_spec.Description = image.Description
	return _spec, nil
}

func (r RHCCRegistry) LoadImages(Query string) (ImageResponse, error) {
	r.log.Debug("RHCCRegistry::LoadImages")
	r.log.Debug("Using registry.access.redhat.com to source APB images using query:" + Query)
	req, err := http.NewRequest("GET", "https://registry.access.redhat.com/v1/search?q="+Query, nil)
	if err != nil {
		return ImageResponse{}, err
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return ImageResponse{}, err
	}
	defer resp.Body.Close()

	r.log.Debug("Got Image Response from RHCC")
	imageList, err := ioutil.ReadAll(resp.Body)

	imageResp := ImageResponse{}
	err = json.Unmarshal(imageList, &imageResp)
	if err != nil {
		return ImageResponse{}, err
	}
	r.log.Debug("Properly unmarshalled image response")

	return imageResp, nil
}
