package adapters

import (
	b64 "encoding/base64"
	"encoding/json"
	"net/http"

	logging "github.com/op/go-logging"
	"github.com/openshift/ansible-service-broker/pkg/apb"
	"github.com/pborman/uuid"
	yaml "gopkg.in/yaml.v1"
)

type Adapter interface {
	// RegistryName will return the registiry prefix for the adapter.
	// Example is docker.io for the dockerhub adapter.
	RegistryName() string
	// GetImages will return all the image names for the adapter configuration.
	GetImages() ([]string, error)
	//FetchSpecs will retrieve all the specs for the list of images names.
	FetchSpecs([]string) ([]*apb.Spec, error)
}

// BundleSpecLabel - label on the image that we should use to pull out the abp spec.
// TODO: needs to remain ansibleapp UNTIL we redo the apps in dockerhub
const BundleSpecLabel = "com.redhat.apb.spec"

// Configuration - Adapter configuration. Contains the info that the adapter
// would need to complete it's request to the images.
type Configuration struct {
	URL  string
	User string
	Pass string
	Org  string
}

// Retrieve the spec from a registry manifest request
func imageToSpec(log *logging.Logger, req *http.Request) (*apb.Spec, error) {
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
		log.Error("Error grabbing JSON body from response: %s", err)
		return nil, err
	}

	if hist.History == nil {
		log.Errorf("V1 Schema Manifest does not exist in registry")
		return nil, nil
	}

	err = json.Unmarshal([]byte(hist.History[0]["v1Compatibility"]), &conf)
	if err != nil {
		log.Error("Error unmarshalling intermediary JSON response: %s", err)
		return nil, err
	}
	if conf.Config.Label.Spec == "" {
		return nil, nil
	}
	encodedSpec := conf.Config.Label.Spec
	decodedSpecYaml, err := b64.StdEncoding.DecodeString(encodedSpec)
	if err != nil {
		log.Error("Something went wrong decoding spec from label")
		return nil, err
	}

	if err = yaml.Unmarshal(decodedSpecYaml, spec); err != nil {
		log.Error("Something went wrong loading decoded spec yaml, %s", err)
		return nil, err
	}
	log.Debug("Successfully converted Image %s into Spec", spec.Image)
	spec.ID = uuid.New()

	return spec, nil
}
