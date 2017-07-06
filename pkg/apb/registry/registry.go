package registry

import (
	"fmt"

	logging "github.com/op/go-logging"
	"github.com/openshift/ansible-service-broker/pkg/apb"
)

// BundleSpecLabel - label on the image that we should use to pull out the abp spec.
// TODO: needs to remain ansibleapp UNTIL we redo the apps in dockerhub
var BundleSpecLabel = "com.redhat.apb.spec"

// ImageData - APB Image data
type ImageData struct {
	Name             string
	Tag              string
	Labels           map[string]string
	Layers           []string
	IsPlaybookBundle bool
	Error            error
}

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
