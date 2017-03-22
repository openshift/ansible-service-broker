package apb

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	logging "github.com/op/go-logging"
)

const AppsPath = "/bundles"

type DevRegistry struct {
	config RegistryConfig
	log    *logging.Logger
}

func (r *DevRegistry) Init(config RegistryConfig, log *logging.Logger) error {
	log.Debug("DevRegistry::Init")
	r.config = config
	r.log = log
	return nil
}

func (r *DevRegistry) LoadSpecs() ([]*Spec, error) {
	r.log.Debug("DevRegistry::LoadSpecs")

	appsUrl := r.fullAppsPath()

	r.log.Debug(fmt.Sprintf("Getting hardcoded specs from: %s", appsUrl))

	res, err := http.Get(appsUrl)
	if err != nil {
		return []*Spec{}, err
	}

	defer res.Body.Close()
	buf := new(bytes.Buffer)
	buf.ReadFrom(res.Body)

	r.log.Debug(strings.Replace(fmt.Sprintf("Loaded apps response: %s", buf.String()), "\n", "", -1))
	specs := loadSpecs(buf.Bytes())
	r.log.Debug(fmt.Sprintf("Loaded Specs: %v", specs))

	r.log.Info(fmt.Sprintf("Loaded [ %d ] specs from %s registry", len(specs), r.config.Name))

	for _, spec := range specs {
		r.log.Debug(fmt.Sprintf("ID: %s", spec.Id))
	}

	return specs, nil
}

func (r *DevRegistry) fullAppsPath() string {
	return fmt.Sprintf("%s%s", r.config.Url, AppsPath)
}

func loadSpecs(rawPayload []byte) []*Spec {
	var specs []*Spec
	json.Unmarshal(rawPayload, &specs)
	return specs
}
