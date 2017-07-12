package registries

import (
	"fmt"
	"strings"

	logging "github.com/op/go-logging"
	"github.com/openshift/ansible-service-broker/pkg/apb"
	"github.com/openshift/ansible-service-broker/pkg/registries/adapters"
)

// Config - Configuration for the registry
type Config struct {
	URL  string
	User string
	Pass string
	Org  string
	Type string
	Name string
	// Fail will tell the registry that it is ok to fail the bootstrap if
	// just this registry has failed.
	Fail      bool     `yaml:"fail_on_error"`
	WhiteList []string `yaml:"white_list"`
	BlackList []string `yaml:"black_list"`
}

// Registry - manages an adapter to retrieve and manage images to specs.
type Registry struct {
	config  Config
	adapter adapters.Adapter
	log     *logging.Logger
	filter  Filter
}

// LoadSpecs - Load the specs for the registry.
func (r Registry) LoadSpecs() ([]*apb.Spec, int, error) {
	r.log.Infof("%#v", r.config)
	imageNames, err := r.adapter.GetImageNames()
	if err != nil {
		r.log.Errorf("unable to retrieve image names for registry %v - %v",
			r.config.Name, err)
		return []*apb.Spec{}, 0, err
	}
	validNames, filteredNames := r.filter.Run(imageNames)
	for _, name := range filteredNames {
		r.log.Debugf("registry %v filtered out image -%v", r.config.Name, name)
	}
	// Debug output filtered out names.
	specs, err := r.adapter.FetchSpecs(validNames)
	if err != nil {
		r.log.Errorf("unable to fetch specs for registry %v - %v",
			r.config.Name, err)
		return []*apb.Spec{}, 0, err
	}
	return specs, len(imageNames), nil
}

// Fail - should this registry and error cause a failure.
func (r Registry) Fail(err error) bool {
	if r.config.Fail {
		return true
	}
	return false
}

// RegistryName - retrieve the registry name to allow namespacing.
func (r Registry) RegistryName() string {
	return r.adapter.RegistryName()
}

// NewRegistry - Create a new registry from the registry config.
func NewRegistry(config Config, log *logging.Logger) (Registry, error) {
	var adapter adapters.Adapter

	log.Info("== REGISTRY CX == ")
	log.Info(fmt.Sprintf("Name: %s", config.Name))
	log.Info(fmt.Sprintf("Type: %s", config.Type))
	log.Info(fmt.Sprintf("Url: %s", config.URL))
	c := adapters.Configuration{URL: config.URL,
		User: config.User,
		Pass: config.Pass,
		Org:  config.Org}

	switch strings.ToLower(config.Type) {
	case "rhcc":
		adapter = &adapters.RHCCAdapter{Config: c, Log: log}
	case "dockerhub":
		adapter = &adapters.DockerHubAdapter{Config: c, Log: log}
	case "mock":
		adapter = &adapters.MockAdapter{Config: c, Log: log}
	default:
		panic("Unknown registry")
	}
	return Registry{config: config,
		adapter: adapter,
		log:     log,
		filter: Filter{config.WhiteList,
			config.BlackList},
	}, nil
}
