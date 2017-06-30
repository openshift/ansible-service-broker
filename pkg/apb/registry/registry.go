package registry

import (
	b64 "encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"

	yaml "gopkg.in/yaml.v1"

	logging "github.com/op/go-logging"
	"github.com/openshift/ansible-service-broker/pkg/apb"
	"github.com/pborman/uuid"
)

// BundleSpecLabel - label on the image that we should use to pull out the abp spec.
// TODO: needs to remain ansibleapp UNTIL we redo the apps in dockerhub
var BundleSpecLabel = "com.redhat.apb.spec"

// Config - Configuration for the registry
type Config struct {
	Type string
	Name string
	URL  string
	User string
	Pass string
	Org  string // Target org to load playbook bundles from
	// Fail will tell the registry that it is ok to fail the bootstrap if
	// just this registry has failed.
	Fail bool `yaml:"fail_on_error"`
}

// Registry - Interface that wraps the methods need for a registry
type Registry interface {
	Init(Config, *logging.Logger) error
	LoadSpecs() ([]*apb.Spec, int, error)
	// Will attempt to decide from the error if the registry can fail loudly on LoadSpecs
	Fail(error) bool
}

// NewRegistry - Create a new registry from the registry config.
func NewRegistry(config Config, log *logging.Logger) (Registry, error) {
	var reg Registry

	log.Info("== REGISTRY CX == ")
	log.Info(fmt.Sprintf("Name: %s", config.Name))
	log.Info(fmt.Sprintf("Type: %s", config.Type))
	log.Info(fmt.Sprintf("Url: %s", config.URL))

	switch config.Type {
	case "rhcc":
		reg = &RHCCRegistry{}
	case "dockerhub":
		reg = &DockerHubRegistry{}
	case "mock":
		reg = &MockRegistry{}
	default:
		panic("Unknown registry")
	}

	err := reg.Init(config, log)
	if err != nil {
		return nil, err
	}

	return reg, err
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
	log.Debug("Successfully converted Image %s into Spec", spec.Name)
	spec.ID = uuid.New()

	return spec, nil
}
