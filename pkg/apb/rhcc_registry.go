package apb

import (
	b64 "encoding/base64"
	"encoding/json"
	logging "github.com/op/go-logging"
	"io/ioutil"
	"net/http"
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

	imageList, err := r.LoadImages("*-apb")
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

	req, err := http.NewRequest("GET", r.config.Url+"/v2/"+image.Name+"manifests/latest", nil)
	if err != nil {
		return nil, err
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	type label struct {
		Spec    string `json:"com.redhat.apb.spec"`
		Version string `json:"com.redhat.apb.version"`
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

	err = json.NewDecoder(resp.Body).Decode(&hist)
	if err != nil {
		r.log.Error("Error grabbing JSON body from response: %s", err)
		return nil, err
	}

	if hist.History == nil {
		r.log.Error("V1 Schema Manifest does not exist in registry")
		return nil, nil
	}

	err = json.Unmarshal([]byte(hist.History[0]["v1Compatibility"]), &conf)
	if err != nil {
		r.log.Error("Error unmarshalling intermediary JSON response: %s", err)
		return nil, err
	}

	r.log.Debug(r.config.Url)
	encodedSpec := conf.Config.Label.Spec
	decodedSpecYaml, err := b64.StdEncoding.DecodeString(encodedSpec)
	if err != nil {
		r.log.Error("Something went wrong decoding spec from label")
		return nil, err
	}

	if err = LoadYAML(string(decodedSpecYaml), _spec); err != nil {
		r.log.Error("Something went wrong loading decoded spec yaml, %s", err)
		return nil, err
	}
	r.log.Debug("Successfully converted RHCC Image %s into Spec", _spec.Name)

	return _spec, nil
}

func (r RHCCRegistry) LoadImages(Query string) (ImageResponse, error) {
	r.log.Debug("RHCCRegistry::LoadImages")
	r.log.Debug("Using " + r.config.Url + " to source APB images using query:" + Query)
	req, err := http.NewRequest("GET", r.config.Url+"/v1/search?q="+Query, nil)
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
