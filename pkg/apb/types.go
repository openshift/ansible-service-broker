//
// Copyright (c) 2017 Red Hat, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//    http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//

package apb

import (
	"encoding/json"

	"github.com/openshift/ansible-service-broker/pkg/config"
	logutil "github.com/openshift/ansible-service-broker/pkg/util/logging"
	"github.com/pborman/uuid"
)

var log = logutil.NewLog()

// Parameters - generic string to object or value parameter
type Parameters map[string]interface{}

// SpecManifest - Spec ID to Spec manifest
type SpecManifest map[string]*Spec

// NilableNumber - Number that could be nil (e.g. when omitted from json/yaml)
type NilableNumber float64

// ParameterDescriptor - a parameter to be used by the service catalog to get data.
type ParameterDescriptor struct {
	Name        string      `json:"name"`
	Title       string      `json:"title"`
	Type        string      `json:"type"`
	Description string      `json:"description,omitempty"`
	Default     interface{} `json:"default,omitempty"`

	// string validators
	DeprecatedMaxlength int    `json:"maxlength,omitempty" yaml:"maxlength,omitempty"` // backwards compatibility
	MaxLength           int    `json:"maxLength,omitempty" yaml:"max_length,omitempty"`
	MinLength           int    `json:"minLength,omitempty" yaml:"min_length,omitempty"`
	Pattern             string `json:"pattern,omitempty"`

	// number validators
	MultipleOf       float64        `json:"multipleOf,omitempty" yaml:"multiple_of,omitempty"`
	Maximum          *NilableNumber `json:"maximum,omitempty"`
	ExclusiveMaximum *NilableNumber `json:"exclusiveMaximum,omitempty" yaml:"exclusive_maximum,omitempty"`
	Minimum          *NilableNumber `json:"minimum,omitempty"`
	ExclusiveMinimum *NilableNumber `json:"exclusiveMinimum,omitempty" yaml:"exclusive_minimum,omitempty"`

	Enum         []string `json:"enum,omitempty"`
	Required     bool     `json:"required"`
	Updatable    bool     `json:"updatable"`
	DisplayType  string   `json:"displayType,omitempty" yaml:"display_type,omitempty"`
	DisplayGroup string   `json:"displayGroup,omitempty" yaml:"display_group,omitempty"`
}

// Plan - Plan object describing an APB deployment plan and associated parameters
type Plan struct {
	ID             string                 `json:"id" yaml:"-"`
	Name           string                 `json:"name"`
	Description    string                 `json:"description"`
	Metadata       map[string]interface{} `json:"metadata,omitempty"`
	Free           bool                   `json:"free,omitempty"`
	Bindable       bool                   `json:"bindable,omitempty"`
	Parameters     []ParameterDescriptor  `json:"parameters"`
	BindParameters []ParameterDescriptor  `json:"bind_parameters,omitempty" yaml:"bind_parameters,omitempty"`
	UpdatesTo      []string               `json:"updates_to,omitempty" yaml:"updates_to,omitempty"`
}

// GetParameter - retrieves a reference to a ParameterDescriptor from a plan by name. Will return
// nil if the requested ParameterDescriptor does not exist.
func (p *Plan) GetParameter(name string) *ParameterDescriptor {
	for i, pd := range p.Parameters {
		if pd.Name == name {
			return &p.Parameters[i]
		}
	}
	return nil
}

// Spec - A APB spec
type Spec struct {
	ID          string                 `json:"id"`
	Runtime     int                    `json:"runtime"`
	Version     string                 `json:"version"`
	FQName      string                 `json:"name" yaml:"name"`
	Image       string                 `json:"image" yaml:"-"`
	Tags        []string               `json:"tags"`
	Bindable    bool                   `json:"bindable"`
	Description string                 `json:"description"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
	Async       string                 `json:"async"`
	Plans       []Plan                 `json:"plans"`
}

// GetPlan - retrieves a reference to a plan from a spec by name. Will return
// nil if the requested plan does not exist.
func (s *Spec) GetPlan(name string) *Plan {
	for i, plan := range s.Plans {
		if plan.Name == name {
			return &s.Plans[i]
		}
	}
	return nil
}

// Context - Determines the context in which the service is running
type Context struct {
	Platform  string `json:"platform"`
	Namespace string `json:"namespace"`
}

// ExtractedCredentials - Credentials that are extracted from the pods
type ExtractedCredentials struct {
	Credentials map[string]interface{} `json:"credentials,omitempty"`
}

// State - Job State
type State string

// JobMethod - APB Method Type that the job was spawned from.
type JobMethod string

const (
	// JobMethodProvision - Provision MethodType const.
	JobMethodProvision JobMethod = "provision"

	// JobMethodDeprovision - Deprovision MethodType const.
	JobMethodDeprovision JobMethod = "deprovision"

	// JobMethodBind - Bind MethodType const.
	JobMethodBind JobMethod = "bind"

	// JobMethodUnbind - Unbind MethodType const.
	JobMethodUnbind JobMethod = "unbind"

	// JobMethodUpdate - Update MethodType const.
	JobMethodUpdate JobMethod = "update"
)

// JobState - The job state
type JobState struct {
	Token   string    `json:"token"`
	State   State     `json:"state"`
	Podname string    `json:"podname"`
	Method  JobMethod `json:"method"`
	Error   string    `json:"error"`
}

// ClusterConfig - Configuration for the cluster.
type ClusterConfig struct {
	PullPolicy           string `yaml:"image_pull_policy"`
	SandboxRole          string `yaml:"sandbox_role"`
	Namespace            string `yaml:"namespace"`
	KeepNamespace        bool   `yaml:"keep_namespace"`
	KeepNamespaceOnError bool   `yaml:"keep_namespace_on_error"`
}

// ClusterConfiguration that should be used by the abp package.
var clusterConfig ClusterConfig

// InitializeClusterConfig - initialize the cluster config.
func InitializeClusterConfig(config *config.Config) {
	clusterConfig = ClusterConfig{
		PullPolicy:           config.GetString("image_pull_policy"),
		SandboxRole:          config.GetString("sandbox_role"),
		Namespace:            config.GetString("namespace"),
		KeepNamespace:        config.GetBool("keep_namespace"),
		KeepNamespaceOnError: config.GetBool("keep_namespace_on_error"),
	}
}

const (
	// StateInProgress - In progress job state
	StateInProgress State = "in progress"
	// StateSucceeded - Succeeded job state
	StateSucceeded State = "succeeded"
	// StateFailed - Failed job state
	StateFailed State = "failed"

	// 5s x 7200 retries, 2 hours
	apbWatchInterval = 5
	apbWatchRetries  = 7200

	// ApbContainerName - The name of the apb container
	ApbContainerName = "apb"
)

// SpecLogDump - log spec for debug
func SpecLogDump(spec *Spec) {
	log.Debug("============================================================")
	log.Debug("Spec: %s", spec.ID)
	log.Debug("============================================================")
	log.Debug("Name: %s", spec.FQName)
	log.Debug("Image: %s", spec.Image)
	log.Debug("Bindable: %t", spec.Bindable)
	log.Debug("Description: %s", spec.Description)
	log.Debug("Async: %s", spec.Async)

	for _, plan := range spec.Plans {
		log.Debugf("Plan: %s", plan.Name)
		for _, param := range plan.Parameters {
			log.Debug("  Name: %#v", param.Name)
			log.Debug("  Title: %s", param.Title)
			log.Debug("  Type: %s", param.Type)
			log.Debug("  Description: %s", param.Description)
			log.Debug("  Default: %#v", param.Default)
			log.Debug("  DeprecatedMaxlength: %d", param.DeprecatedMaxlength)
			log.Debug("  MaxLength: %d", param.MaxLength)
			log.Debug("  MinLength: %d", param.MinLength)
			log.Debug("  Pattern: %s", param.Pattern)
			log.Debug("  MultipleOf: %d", param.MultipleOf)
			log.Debug("  Minimum: %#v", param.Minimum)
			log.Debug("  Maximum: %#v", param.Maximum)
			log.Debug("  ExclusiveMinimum: %#v", param.ExclusiveMinimum)
			log.Debug("  ExclusiveMaximum: %#v", param.ExclusiveMaximum)
			log.Debug("  Required: %s", param.Required)
			log.Debug("  Enum: %v", param.Enum)
		}
	}
}

// SpecsLogDump - log specs for debug
func SpecsLogDump(specs []*Spec) {
	for _, spec := range specs {
		SpecLogDump(spec)
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

// ExecutionContext - Contains the information necessary to track and clean up
// an APB run
type ExecutionContext struct {
	PodName        string
	Namespace      string
	ServiceAccount string
	Targets        []string
}
