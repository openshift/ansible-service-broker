package apb

import (
	"encoding/json"

	logging "github.com/op/go-logging"
	"github.com/pborman/uuid"
	yaml "gopkg.in/yaml.v2"
)

type Parameters map[string]interface{}
type SpecManifest map[string]*Spec

// TODO: needs to remain ansibleapp UNTIL we redo the apps in dockerhub
var BundleSpecLabel = "com.redhat.apb.spec"

type ImageData struct {
	Name             string
	Tag              string
	Labels           map[string]string
	Layers           []string
	IsPlaybookBundle bool
	Error            error
}

type ParameterDescriptor struct {
	Title       string      `json:"title"`
	Type        string      `json:"type"`
	Description string      `json:"description,omitempty"`
	Default     interface{} `json:"default,omitempty"`
	Maxlength   int         `json:"maxlength,omitempty"`
	Pattern     string      `json:"pattern,omitempty"`
	Enum        []string    `json:"enum,omitempty"`
}

/*
array of maps with an array of ParameterDescriptors
*/

type Spec struct {
	Id          string                 `json:"id"`
	Name        string                 `json:"name"`
	Image       string                 `json:"image"`
	Tags        []string               `json:"tags"`
	Bindable    bool                   `json:"bindable"`
	Description string                 `json:"description"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`

	// required, optional, unsupported
	Async      string                            `json:"async"`
	Parameters []map[string]*ParameterDescriptor `json:"parameters"`
	Required   []string                          `json:"required,omitempty"`
}

type Context struct {
	Platform  string `json:"platform"`
	Namespace string `json:"namespace"`
}

type ExtractedCredentials struct {
	Credentials map[string]interface{} `json:"credentials,omitempty"`
	// might be more one day
}

type State string

type JobState struct {
	Token string `json:"token"`
	State State  `json:"state"`
}

const (
	StateInProgress State = "in progress"
	StateSucceeded  State = "succeeded"
	StateFailed     State = "failed"
)

func specLogDump(spec *Spec, log *logging.Logger) {
	log.Debug("============================================================")
	log.Debug("Spec: %s", spec.Id)
	log.Debug("============================================================")
	log.Debug("Name: %s", spec.Name)
	log.Debug("Image: %s", spec.Image)
	log.Debug("Bindable: %t", spec.Bindable)
	log.Debug("Description: %s", spec.Description)
	log.Debug("Async: %s", spec.Async)

	for _, params := range spec.Parameters {
		log.Debug("ParameterDescriptor")
		for k, param := range params {
			log.Debug("  Name: %#v", k)
			log.Debug("  Title: %s", param.Title)
			log.Debug("  Type: %s", param.Type)
			log.Debug("  Description: %s", param.Description)
			log.Debug("  Default: %#v", param.Default)
			log.Debug("  Maxlength: %d", param.Maxlength)
			log.Debug("  Pattern: %s", param.Pattern)
			log.Debug("  Enum: %v", param.Enum)
		}
	}
}

func specsLogDump(specs []*Spec, log *logging.Logger) {
	for _, spec := range specs {
		specLogDump(spec, log)
	}
}

func NewSpecManifest(specs []*Spec) SpecManifest {
	manifest := make(map[string]*Spec)
	for _, spec := range specs {
		manifest[spec.Id] = spec
	}
	return manifest
}

type ServiceInstance struct {
	Id         uuid.UUID   `json:"id"`
	Spec       *Spec       `json:"spec"`
	Context    *Context    `json:"context"`
	Parameters *Parameters `json:"parameters"`
}

type BindInstance struct {
	Id         uuid.UUID   `json:"id"`
	ServiceId  uuid.UUID   `json:"service_id"`
	Parameters *Parameters `json:"parameters"`
}

func LoadJSON(payload string, obj interface{}) error {
	err := json.Unmarshal([]byte(payload), obj)
	if err != nil {
		return err
	}

	return nil
}

func DumpJSON(obj interface{}) (string, error) {
	payload, err := json.Marshal(obj)
	if err != nil {
		return "", err
	}

	return string(payload), nil
}

func LoadYAML(payload string, obj interface{}) error {
	var err error

	if err = yaml.Unmarshal([]byte(payload), obj); err != nil {
		return err
	}

	return nil
}
