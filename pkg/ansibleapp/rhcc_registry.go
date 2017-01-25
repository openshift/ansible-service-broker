package ansibleapp

import (
	"fmt"
)

type RHCCRegistry struct {
	config RegistryConfig
}

func (r *RHCCRegistry) Init(config RegistryConfig) error {
	r.config = config
	fmt.Printf("RHCCRegistry::Init with url -> [ %s ] \n", r.config.Url)
	return nil
}

func (r *RHCCRegistry) LoadApps() error {
	fmt.Println("RHCCRegistry::LoadApps ")
	return nil
}
