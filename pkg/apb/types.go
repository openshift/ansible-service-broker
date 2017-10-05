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
// Red Hat trademarks are not licensed under Apache License, Version 2.
// No permission is granted to use or replicate Red Hat trademarks that
// are incorporated in this software or its documentation.
//

package apb

import (
	"encoding/json"

	logging "github.com/op/go-logging"
	"github.com/pborman/uuid"
)

// Parameters - generic string to object or value parameter
type Parameters map[string]interface{}

// SpecManifest - Spec ID to Spec manifest
type SpecManifest map[string]*Spec

// ParameterDescriptor - a parameter to be used by the service catalog to get data.
type ParameterDescriptor struct {
	Name         string      `json:"name"`
	Title        string      `json:"title"`
	Type         string      `json:"type"`
	Description  string      `json:"description,omitempty"`
	Default      interface{} `json:"default,omitempty"`
	Maxlength    int         `json:"maxlength,omitempty"`
	Pattern      string      `json:"pattern,omitempty"`
	Enum         []string    `json:"enum,omitempty"`
	Required     bool        `json:"required"`
	Updatable    bool        `json:"updatable"`
	DisplayType  string      `json:"display_type,omitempty" yaml:"display_type,omitempty"`
	DisplayGroup string      `json:"display_group,omitempty" yaml:"display_group,omitempty"`
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

// Spec - A APB spec
type Spec struct {
	ID          string                 `json:"id"`
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
)

// JobState - The job state
type JobState struct {
	Token   string    `json:"token"`
	State   State     `json:"state"`
	Podname string    `json:"podname"`
	Method  JobMethod `json:"method"`
}

// ClusterConfig - Configuration for the cluster.
type ClusterConfig struct {
	Host                 string `yaml:"host"`
	CAFile               string `yaml:"ca_file"`
	BearerTokenFile      string `yaml:"bearer_token_file"`
	PullPolicy           string `yaml:"image_pull_policy"`
	SandboxRole          string `yaml:"sandbox_role"`
	Namespace            string `yaml:"namespace"`
	KeepNamespace        bool   `yaml:"keep_namespace"`
	KeepNamespaceOnError bool   `yaml:"keep_namespace_on_error"`
}

const (
	// StateInProgress - In progress job state
	StateInProgress State = "in progress"
	// StateSucceeded - Succeeded job state
	StateSucceeded State = "succeeded"
	// StateFailed - Failed job state
	StateFailed State = "failed"

	// 5s x 7200 retries, 2 hours
	apbWatchInterval     = 5
	apbWatchRetries      = 7200
	gatherCredentialsCMD = "broker-bind-creds"
)

// SpecLogDump - log spec for debug
func SpecLogDump(spec *Spec, log *logging.Logger) {
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
			log.Debug("  Maxlength: %d", param.Maxlength)
			log.Debug("  Pattern: %s", param.Pattern)
			log.Debug("  Pattern: %s", param.Required)
			log.Debug("  Enum: %v", param.Enum)
		}
	}
}

// SpecsLogDump - log specs for debug
func SpecsLogDump(specs []*Spec, log *logging.Logger) {
	for _, spec := range specs {
		SpecLogDump(spec, log)
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
