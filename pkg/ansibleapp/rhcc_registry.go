package ansibleapp

import (
	"github.com/op/go-logging"
)

type RHCCRegistry struct {
	config RegistryConfig
	log    *logging.Logger
}

func (r *RHCCRegistry) Init(config RegistryConfig, log *logging.Logger) error {
	log.Debug("RHCCRegistry::Init")
	r.config = config
	r.log = log
	return nil
}

func (r *RHCCRegistry) LoadSpecs() ([]*Spec, error) {
	r.log.Debug("RHCCRegistry::LoadSpecs")
	return []*Spec{}, nil
}
