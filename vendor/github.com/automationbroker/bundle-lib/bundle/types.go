//
// Copyright (c) 2018 Red Hat, Inc.
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

package bundle

import (
	"encoding/json"
	"reflect"

	schema "github.com/lestrrat/go-jsschema"
	"github.com/pborman/uuid"
	log "github.com/sirupsen/logrus"
)

// Parameters - generic string to object or value parameter
type Parameters map[string]interface{}

// EnsureDefaults returns the sets any expected parameters that are not present with default values
func (p *Parameters) EnsureDefaults() {
	if _, ok := (*p)[ProvisionCredentialsKey]; !ok {
		(*p)[ProvisionCredentialsKey] = struct{}{}
	}
}

// Add adds a values to the parameters indexed by key
func (p *Parameters) Add(key string, val interface{}) {
	(*p)[key] = val
}

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

	Enum         []string     `json:"enum,omitempty"`
	Required     bool         `json:"required"`
	Updatable    bool         `json:"updatable"`
	DisplayType  string       `json:"displayType,omitempty" yaml:"display_type,omitempty"`
	DisplayGroup string       `json:"displayGroup,omitempty" yaml:"display_group,omitempty"`
	Dependencies []Dependency `json:"dependencies,omitempty" yaml:"dependencies,omitempty"`
}

// Dependency - a parameter dependency
type Dependency struct {
	Key   string      `json:"key,omitempty" yaml:"key,omitempty"`
	Value interface{} `json:"value,omitempty" yaml:"value,omitempty"`
}

// Schema  - Schema to be returned
// based on 2.13 of the open service broker api. https://github.com/avade/servicebroker/blob/cda8c57b6a4bb7eaee84be20bb52dc155269758a/spec.md
type Schema struct {
	ServiceInstance ServiceInstanceSchema `json:"service_instance"`
	ServiceBinding  ServiceBindingSchema  `json:"service_binding"`
}

// ServiceInstanceSchema - Schema definitions for creating and updating a service instance.
// Toyed with the idea of making an InputParameters
// that was a *schema.Schema
// based on 2.13 of the open service broker api. https://github.com/avade/servicebroker/blob/cda8c57b6a4bb7eaee84be20bb52dc155269758a/spec.md
type ServiceInstanceSchema struct {
	Create map[string]*schema.Schema `json:"create"`
	Update map[string]*schema.Schema `json:"update"`
}

// ServiceBindingSchema - Schema definitions for creating a service binding.
type ServiceBindingSchema struct {
	Create map[string]*schema.Schema `json:"create"`
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

// SchemaPlan - Plan object describing an APB deployment plan and associated parameters
type SchemaPlan struct {
	ID          string                 `json:"id" yaml:"-"`
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
	Free        bool                   `json:"free,omitempty"`
	Bindable    bool                   `json:"bindable,omitempty"`
	UpdatesTo   []string               `json:"updates_to,omitempty" yaml:"updates_to,omitempty"`
	Schemas     Schema                 `json:"schema,omitempty"`
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
	FQName      string                 `json:"name" yaml:"name" mapstructure:"name"`
	Image       string                 `json:"image" yaml:"-"`
	Tags        []string               `json:"tags"`
	Bindable    bool                   `json:"bindable"`
	Description string                 `json:"description"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
	Async       string                 `json:"async"`
	Plans       []Plan                 `json:"plans"`
	Alpha       map[string]interface{} `json:"alpha,omitempty"`
	Delete      bool                   `json:"delete"`
}

// GetPlan - retrieves a plan from a spec by name. Will return
// empty plan and false if the requested plan does not exist.
func (s *Spec) GetPlan(name string) (Plan, bool) {
	for _, plan := range s.Plans {
		if plan.Name == name {
			return plan, true
		}
	}
	return Plan{}, false
}

// GetPlanFromID - retrieves a plan from a spec by id. Will return
// empty plan and false if the requested plan does not exist.
func (s *Spec) GetPlanFromID(id string) (Plan, bool) {
	for _, plan := range s.Plans {
		if plan.ID == id {
			return plan, true
		}
	}
	return Plan{}, false
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

func buildExtractedCredentials(output []byte) (*ExtractedCredentials, error) {
	creds := make(map[string]interface{})
	err := json.Unmarshal(output, &creds)
	if err != nil {
		return nil, err
	}
	return &ExtractedCredentials{Credentials: creds}, nil
}

// State - Job State
type State string

// StatusMessage - Describes the latest known status of a running APB
type StatusMessage struct {
	State       State
	Description string
	Error       error
}

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
	Token       string    `json:"token"`
	State       State     `json:"state"`
	Podname     string    `json:"podname"`
	Method      JobMethod `json:"method"`
	Error       string    `json:"error"`
	Description string    `json:"description"`
}

// ClusterConfig - Configuration for the cluster.
type ClusterConfig struct {
	PullPolicy           string `yaml:"image_pull_policy"`
	SandboxRole          string `yaml:"sandbox_role"`
	Namespace            string `yaml:"namespace"`
	KeepNamespace        bool   `yaml:"keep_namespace"`
	KeepNamespaceOnError bool   `yaml:"keep_namespace_on_error"`
}

// ClusterConfiguration that should be used by the apb package.
var clusterConfig ClusterConfig

// InitializeClusterConfig - initialize the cluster config.
func InitializeClusterConfig(config ClusterConfig) {
	clusterConfig = config
}

const (
	// StateNotYetStarted - Executor has not yet started state.
	StateNotYetStarted State = "not yet started"
	// StateInProgress - APB is in progress state
	StateInProgress State = "in progress"
	// StateSucceeded - Succeeded state
	StateSucceeded State = "succeeded"
	// StateFailed - Failed state
	StateFailed State = "failed"

	// ApbContainerName - The name of the apb container
	ApbContainerName = "apb"

	// ProvisionCredentialsKey parameter name passed to APBs
	ProvisionCredentialsKey = "_apb_provision_creds"
	// BindCredentialsKey parameter name passed to APBs
	BindCredentialsKey = "_apb_bind_creds"
	// ClusterKey parameter name passed to APBs
	ClusterKey = "cluster"
	// NamespaceKey parameter name passed to APBs
	NamespaceKey = "namespace"
)

// SpecLogDump - log spec for debug
func SpecLogDump(spec *Spec) {
	log.Debug("============================================================")
	log.Debugf("Spec: %s", spec.ID)
	log.Debug("============================================================")
	log.Debugf("Name: %s", spec.FQName)
	log.Debugf("Image: %s", spec.Image)
	log.Debugf("Bindable: %t", spec.Bindable)
	log.Debugf("Description: %s", spec.Description)
	log.Debugf("Async: %s", spec.Async)

	for _, plan := range spec.Plans {
		log.Debugf("Plan: %s", plan.Name)
		for _, param := range plan.Parameters {
			log.Debugf("  Name: %#v", param.Name)
			log.Debugf("  Title: %s", param.Title)
			log.Debugf("  Type: %s", param.Type)
			log.Debugf("  Description: %s", param.Description)
			log.Debugf("  Default: %#v", param.Default)
			log.Debugf("  DeprecatedMaxlength: %d", param.DeprecatedMaxlength)
			log.Debugf("  MaxLength: %d", param.MaxLength)
			log.Debugf("  MinLength: %d", param.MinLength)
			log.Debugf("  Pattern: %s", param.Pattern)
			log.Debugf("  MultipleOf: %d", param.MultipleOf)
			log.Debugf("  Minimum: %#v", param.Minimum)
			log.Debugf("  Maximum: %#v", param.Maximum)
			log.Debugf("  ExclusiveMinimum: %#v", param.ExclusiveMinimum)
			log.Debugf("  ExclusiveMaximum: %#v", param.ExclusiveMaximum)
			log.Debugf("  Required: %s", param.Required)
			log.Debugf("  Enum: %v", param.Enum)
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
	ID           uuid.UUID       `json:"id"`
	Spec         *Spec           `json:"spec"`
	Context      *Context        `json:"context"`
	Parameters   *Parameters     `json:"parameters"`
	BindingIDs   map[string]bool `json:"binding_ids"`
	DashboardURL string          `json:"dashboard_url"`
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
		si.BindingIDs[bindingUUID.String()] = false
	}
}

// BindInstance - Binding Instance describes a completed binding
type BindInstance struct {
	ID           uuid.UUID   `json:"id"`
	ServiceID    uuid.UUID   `json:"service_id"`
	Parameters   *Parameters `json:"parameters"`
	CreateJobKey string
}

// UserParameters - returns the Parameters field with any keys and values
// removed that are typically added by the broker itself. The return value
// should represent what a OSB API client provides in a bind request.
func (bi *BindInstance) UserParameters() Parameters {
	if bi.Parameters == nil {
		return nil
	}
	userparams := make(Parameters)
	for key, value := range *bi.Parameters {
		switch key {
		// Do not copy keys that are generally added by the broker itself.
		case ClusterKey, NamespaceKey, ProvisionCredentialsKey:
			continue
		}
		userparams[key] = value
	}
	return userparams
}

// IsEqual - Determines if two BindInstances are equal, omitting any Parameters
// that generally get added by the broker itself.
func (bi *BindInstance) IsEqual(newbi *BindInstance) bool {
	if !uuid.Equal(bi.ID, newbi.ID) || !uuid.Equal(bi.ServiceID, newbi.ServiceID) {
		return false
	}
	return reflect.DeepEqual(bi.UserParameters(), newbi.UserParameters())
}

// LoadJSON - Generic function to unmarshal json
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

const (
	httpProxyEnvVar  = "HTTP_PROXY"
	httpsProxyEnvVar = "HTTPS_PROXY"
	noProxyEnvVar    = "NO_PROXY"
)
