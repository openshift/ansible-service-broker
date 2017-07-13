package apb

import (
	"encoding/json"

	logging "github.com/op/go-logging"
	"github.com/pborman/uuid"
)

// Parameters - generic string to object or value parameter
type Parameters map[string]interface{}

//SpecManifest - Spec ID to Spec manifest
type SpecManifest map[string]*Spec

// BundleSpecLabel - label on the image that we should use to pull out the abp spec.
// TODO: needs to remain ansibleapp UNTIL we redo the apps in dockerhub
var BundleSpecLabel = "com.redhat.apb.spec"

// ImageData - APB Image data
type ImageData struct {
	Name             string
	Tag              string
	Labels           map[string]string
	Layers           []string
	IsPlaybookBundle bool
	Error            error
}

// ParameterDescriptor - a parameter to be used by the service catalog to get data.
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

// Spec - A APB spec
type Spec struct {
	ID          string                 `json:"id"`
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

// Context - Determines the context in which the service is running
type Context struct {
	Platform  string `json:"platform"`
	Namespace string `json:"namespace"`
}

// ExtractedCredentials - Credentials that are extracted from the pods
type ExtractedCredentials struct {
	Credentials map[string]interface{} `json:"credentials,omitempty"`
	// might be more one day
}

// State - Job State
type State string

// JobState - The job state
type JobState struct {
	Token   string `json:"token"`
	State   State  `json:"state"`
	Podname string `json:"podname"`
}

// ClusterConfig - Configuration for the cluster.
type ClusterConfig struct {
	Host            string `yaml:"host"`
	CAFile          string `yaml:"ca_file"`
	BearerTokenFile string `yaml:"bearer_token_file"`
	PullPolicy      string `yaml:"image_pull_policy"`
}

const (
	// StateInProgress - In progress job state
	StateInProgress State = "in progress"
	// StateSucceeded - Succeeded job state
	StateSucceeded State = "succeeded"
	// StateFailed - Failed job state
	StateFailed State = "failed"

	// ApbRole - role to be used.
	// NOTE: Applied to APB ServiceAccount via RoleBinding, not ClusterRoleBinding
	// cluster-admin scoped to the project the apb is operating in
	ApbRole = "cluster-admin"

	// Bind credential gathering constants
	// 5 seconds x 60 retries = 5 minute timout
	credentialExtInterval = 5
	credentialExtRetries  = 60
	gatherCredentialsCMD  = "broker-bind-creds"
)

func specLogDump(spec *Spec, log *logging.Logger) {
	log.Debug("============================================================")
	log.Debug("Spec: %s", spec.ID)
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

// NewSpecManifest - Creates Spec manifest
func NewSpecManifest(specs []*Spec) SpecManifest {
	manifest := make(map[string]*Spec)
	for _, spec := range specs {
		if spec == nil {
			return nil
		}
		manifest[spec.ID] = spec
	}
	return manifest
}

// ServiceInstance - Service Instance describes a running service.
type ServiceInstance struct {
	ID         uuid.UUID       `json:"id"`
	Spec       *Spec           `json:"spec"`
	Context    *Context        `json:"context"`
	Parameters *Parameters     `json:"parameters"`
	BindingIDs map[string]bool `json:"binding_ids"`
}

// AddBinding - Add binding ID to service instance
func (si *ServiceInstance) AddBinding(bindingUUID uuid.UUID) {
	if si.BindingIDs == nil {
		si.BindingIDs = make(map[string]bool)
	}
	si.BindingIDs[bindingUUID.String()] = true
}

// RemoveBinding - Remove binding ID from service instance
func (si *ServiceInstance) RemoveBinding(bindingUUID uuid.UUID) {
	if si.BindingIDs != nil {
		delete(si.BindingIDs, bindingUUID.String())
	}
}

// BindInstance - Binding Instance describes a completed binding
type BindInstance struct {
	ID         uuid.UUID   `json:"id"`
	ServiceID  uuid.UUID   `json:"service_id"`
	Parameters *Parameters `json:"parameters"`
}

// LoadJSON - Generic function to unmarshal json
// TODO: Remove in favor of calling the same method.
func LoadJSON(payload string, obj interface{}) error {
	err := json.Unmarshal([]byte(payload), obj)
	if err != nil {
		return err
	}

	return nil
}

// DumpJSON - Generic function to marshal obj to json string
func DumpJSON(obj interface{}) (string, error) {
	payload, err := json.Marshal(obj)
	if err != nil {
		return "", err
	}

	return string(payload), nil
}

// RecoverStatus - Status of the recovery.
type RecoverStatus struct {
	InstanceID uuid.UUID `json:"id"`
	State      JobState  `json:"state"`
}
