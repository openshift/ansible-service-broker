package ansibleapp

type RegistryConfig struct {
	Name string
	Url  string
}

type Registry interface {
	Init(RegistryConfig) error
	LoadApps() error
}

func CreateRegistry(config RegistryConfig) Registry {
	var reg Registry

	switch config.Name {
	case "dev":
		reg = &DevRegistry{}
	case "rhcc":
		reg = &RHCCRegistry{}
	default:
		panic("Unknown registry")
	}

	reg.Init(config)

	return reg
}
