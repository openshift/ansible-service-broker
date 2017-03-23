package apb

import (
	"fmt"

	logging "github.com/op/go-logging"
)

type RegistryConfig struct {
	Name string
	Url  string
	User string
	Pass string
	Org  string // Target org to load playbook bundles from
}

type Registry interface {
	Init(RegistryConfig, *logging.Logger) error
	LoadSpecs() ([]*Spec, error)
}

func NewRegistry(config RegistryConfig, log *logging.Logger) (Registry, error) {
	var reg Registry

	log.Info("== REGISTRY CX == ")
	log.Info(fmt.Sprintf("Name: %s", config.Name))
	log.Info(fmt.Sprintf("Url: %s", config.Url))

	switch config.Name {
	case "dev":
		reg = &DevRegistry{}
	case "rhcc":
		reg = &RHCCRegistry{}
	case "dockerhub":
		reg = &DockerHubRegistry{}
	default:
		panic("Unknown registry")
	}

	err := reg.Init(config, log)
	if err != nil {
		return nil, err
	}

	return reg, err
}
