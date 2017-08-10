package registries

import (
	"bytes"
	"fmt"
	"net/url"
	"regexp"
	"strings"
	"sync"

	logging "github.com/op/go-logging"
	"github.com/openshift/ansible-service-broker/pkg/apb"
	"github.com/openshift/ansible-service-broker/pkg/registries/adapters"
)

var regex = regexp.MustCompile(`[a-z0-9]([-a-z0-9]*[a-z0-9])?(\.[a-z0-9]([-a-z0-9]*[a-z0-9])?)*`)

// Config - Configuration for the registry
type Config struct {
	URL    string
	User   string
	Pass   string
	Org    string
	Type   string
	Name   string
	Images []string
	// Fail will tell the registry that it is ok to fail the bootstrap if
	// just this registry has failed.
	Fail      bool     `yaml:"fail_on_error"`
	WhiteList []string `yaml:"white_list"`
	BlackList []string `yaml:"black_list"`
}

// Validate - makes sure the registry config is valid.
func (c Config) Validate() bool {
	if c.Name == "" {
		return false
	}
	m := regex.FindString(c.Name)
	return m == c.Name
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
	imageNames, err := r.adapter.GetImageNames()
	if err != nil {
		r.log.Errorf("unable to retrieve image names for registry %v - %v",
			r.config.Name, err)
		return []*apb.Spec{}, 0, err
	}
	// Registry will throw out all images that do not end in -apb
	imageNames = registryFilterImagesForAPBs(imageNames)
	validNames, filteredNames := r.filter.Run(imageNames)

	r.log.Debug("Filter applied against registry: %s", r.config.Name)

	if len(validNames) != 0 {
		r.log.Debugf("APBs passing white/blacklist filter:")
		for _, name := range validNames {
			r.log.Debugf("-> %s", name)
		}
	}

	if len(filteredNames) != 0 {
		go func() {
			var buffer bytes.Buffer
			buffer.WriteString("APBs filtered by white/blacklist filter:")
			for _, name := range filteredNames {
				buffer.WriteString(fmt.Sprintf("-> %s", name))
			}
			r.log.Infof(buffer.String())
		}()
	}

	// Debug output filtered out names.
	specs, err := r.adapter.FetchSpecs(validNames)
	if err != nil {
		r.log.Errorf("unable to fetch specs for registry %v - %v",
			r.config.Name, err)
		return []*apb.Spec{}, 0, err
	}

	r.log.Infof("Validating specs...")
	validatedSpecs := validateSpecs(r.log, specs)
	failedSpecsCount := len(specs) - len(validatedSpecs)

	if failedSpecsCount != 0 {
		r.log.Warningf(
			"%d specs of %d discovered specs failed validation from registry: %s",
			failedSpecsCount, len(specs), r.adapter.RegistryName())
	} else {
		r.log.Notice("All specs passed validation!")
	}

	return validatedSpecs, len(imageNames), nil
}

func registryFilterImagesForAPBs(imageNames []string) []string {
	newNames := []string{}
	for _, imagesName := range imageNames {
		if strings.HasSuffix(strings.ToLower(imagesName), "-apb") {
			newNames = append(newNames, imagesName)
		}
	}
	return newNames
}

// Fail - will determine if the registry should cause a failure.
func (r Registry) Fail(err error) bool {
	if r.config.Fail {
		return true
	}
	return false
}

// RegistryName - retrieve the registry name to allow namespacing.
func (r Registry) RegistryName() string {
	return r.config.Name
}

// NewRegistry - Create a new registry from the registry config.
func NewRegistry(config Config, log *logging.Logger) (Registry, error) {
	var adapter adapters.Adapter

	log.Info("== REGISTRY CX == ")
	log.Info(fmt.Sprintf("Name: %s", config.Name))
	log.Info(fmt.Sprintf("Type: %s", config.Type))
	log.Info(fmt.Sprintf("Url: %s", config.URL))
	//Validate URL
	u, err := url.Parse(config.URL)
	if err != nil {
		log.Errorf("url is not valid: %v", config.URL)
		// Default url, allow the registry to fail gracefully or un gracefully.
		u = &url.URL{}
	}
	if u.Scheme == "" {
		u.Scheme = "http"
	}
	c := adapters.Configuration{URL: u,
		User:   config.User,
		Pass:   config.Pass,
		Org:    config.Org,
		Images: config.Images}

	switch strings.ToLower(config.Type) {
	case "rhcc":
		adapter = &adapters.RHCCAdapter{Config: c, Log: log}
	case "dockerhub":
		adapter = &adapters.DockerHubAdapter{Config: c, Log: log}
	case "mock":
		adapter = &adapters.MockAdapter{Config: c, Log: log}
	case "openshift":
		adapter = &adapters.OpenShiftAdapter{Config: c, Log: log}
	default:
		panic("Unknown registry")
	}

	return Registry{config: config,
		adapter: adapter,
		log:     log,
		filter:  createFilter(config, log),
	}, nil
}

func createFilter(config Config, log *logging.Logger) Filter {
	log.Debug("Creating filter for registry: %s", config.Name)
	log.Debug("whitelist: %v", config.WhiteList)
	log.Debug("blacklist: %v", config.BlackList)

	filter := Filter{
		whitelist: config.WhiteList,
		blacklist: config.BlackList,
	}

	filter.Init()
	if len(filter.failedWhiteRegexp) != 0 {
		log.Warning("Some whitelist regex failed for registry: %s", config.Name)
		for _, failed := range filter.failedWhiteRegexp {
			log.Warning(failed.regex)
			log.Warning(failed.err.Error())
		}
	}

	if len(filter.failedBlackRegexp) != 0 {
		log.Warning("Some blacklist regex failed for registry: %s", config.Name)
		for _, failed := range filter.failedBlackRegexp {
			log.Warning(failed.regex)
			log.Warning(failed.err.Error())
		}
	}

	return filter
}

func validateSpecs(log *logging.Logger, inSpecs []*apb.Spec) []*apb.Spec {
	var wg sync.WaitGroup
	wg.Add(len(inSpecs))

	type resultT struct {
		ok         bool
		spec       *apb.Spec
		failReason string
	}

	out := make(chan resultT)
	for _, spec := range inSpecs {
		go func(s *apb.Spec) {
			defer wg.Done()
			ok, failReason := validateSpecPlans(s)
			out <- resultT{ok, s, failReason}
		}(spec)
	}

	go func() {
		wg.Wait()
		close(out)
	}()

	validSpecs := make([]*apb.Spec, 0, len(inSpecs))
	for result := range out {
		if result.ok {
			validSpecs = append(validSpecs, result.spec)
		} else {
			log.Warningf(
				"Spec [ %s ] failed validation for the following reason: [ %s ]. "+
					"It will not be made available.",
				result.spec.FQName, result.failReason,
			)
		}
	}

	return validSpecs
}

func validateSpecPlans(spec *apb.Spec) (bool, string) {
	// Specs must have at least one plan
	if !(len(spec.Plans) > 0) {
		return false, "Specs must have at least one plan"
	}

	dupes := make(map[string]bool)
	for _, plan := range spec.Plans {
		if _, contains := dupes[plan.Name]; contains {
			reason := fmt.Sprintf("%s: %s",
				"Plans within a spec must not contain duplicate value", plan.Name)

			return false, reason
		}
		dupes[plan.Name] = true
	}

	return true, ""
}
