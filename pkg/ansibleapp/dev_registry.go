package ansibleapp

import (
	"fmt"
)

type DevRegistry struct {
	config RegistryConfig
}

func (r *DevRegistry) Init(config RegistryConfig) error {
	r.config = config
	fmt.Printf("DevRegistry::Init with url -> [ %s ] \n", r.config.Url)
	return nil
}

func (r *DevRegistry) LoadApps() error {
	fmt.Println("DevRegistry::LoadApps ")
	return nil
}
