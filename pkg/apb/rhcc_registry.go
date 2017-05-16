package apb

import logging "github.com/op/go-logging"

// RHCCRegistry - Red Hat Container Catalog Registry
type RHCCRegistry struct {
	config RegistryConfig
	log    *logging.Logger
}

// Init - Initialize the Red Hat Container Catalog
func (r *RHCCRegistry) Init(config RegistryConfig, log *logging.Logger) error {
	log.Debug("RHCCRegistry::Init")
	r.config = config
	r.log = log
	return nil
}

// LoadSpecs - Load Red Hat Container Catalog specs
func (r *RHCCRegistry) LoadSpecs() ([]*Spec, int, error) {
	r.log.Debug("RHCCRegistry::LoadSpecs")
	return []*Spec{}, 0, nil
}
