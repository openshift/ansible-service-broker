/*
Copyright (c) 2018 Red Hat, Inc.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package v1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +genclient
// +genclient:noStatus
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// Bundle describes a apb spec.
type Bundle struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec BundleSpec `json:"spec"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// BundleList is a list of Database resources
type BundleList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata"`

	Items []Bundle `json:"items"`
}

// AsyncType - the type of async value allowed
type AsyncType string

const (
	// OptionalAsync - async is optional for the bundle.
	OptionalAsync AsyncType = "optional"
	// RequiredAsync - async is required for the bundle.
	RequiredAsync AsyncType = "required"
	// Unsupported - async is unsuported for the bundle.
	Unsupported AsyncType = "unsupported"
)

// BundleSpec is the spec for an APB resource
type BundleSpec struct {
	Runtime     int       `json:"runtime"`
	Version     string    `json:"version"`
	FQName      string    `json:"fq_name"`
	Image       string    `json:"image"`
	Description string    `json:"description"`
	Tags        []string  `json:"tags,omitempty"`
	Bindable    bool      `json:"bindable"`
	Async       AsyncType `json:"async"`
	// Store the metadata as a json encoded string to preserve the genericness
	Metadata string `json:"metadata"`
	// Store the alpha map as a json encoded string to preserve the genericness
	Alpha string
	Plans []Plan `json:"plans"`
}

// Plan - a plan for a bundle spec
type Plan struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	// Store the metadata as a json encoded string to preserve the genericness
	Metadata       string       `json:"metadata"`
	Free           bool         `json:"free"`
	Bindable       bool         `json:"bindable"`
	UpdatesTo      []string     `json:"updates_to,omitempty"`
	Parameters     []Parameters `json:"parameters,omitempty"`
	BindParameters []Parameters `json:"bindParameters,omitempty"`
}

// NilableNumber - Number that could be nil (e.g. when omitted from json/yaml)
type NilableNumber float64

// Parameters - describe the parameters for a plan
type Parameters struct {
	Name        string `json:"name"`
	Title       string `json:"title"`
	Type        string `json:"type"`
	Description string `json:"description"`
	// Store the default value as a json encoded string
	Default             string         `json:"default"`
	DeprecatedMaxLength int            `json:"deprecatedMaxLength"`
	MaxLength           int            `json:"maxLength"`
	MinLength           int            `json:"minLength,omitempty" yaml:"min_length,omitempty"`
	Pattern             string         `json:"pattern,omitempty"`
	MultipleOf          float64        `json:"multipleOf,omitempty" yaml:"multiple_of,omitempty"`
	Maximum             *NilableNumber `json:"maximum,omitempty"`
	ExclusiveMaximum    *NilableNumber `json:"exclusiveMaximum,omitempty" yaml:"exclusive_maximum,omitempty"`
	Minimum             *NilableNumber `json:"minimum,omitempty"`
	ExclusiveMinimum    *NilableNumber `json:"exclusiveMinimum,omitempty" yaml:"exclusive_minimum,omitempty"`
	Enum                []string       `json:"enum,omitempty"`
	Required            bool           `json:"required"`
	Updatable           bool           `json:"updatable"`
	DisplayType         string         `json:"displayType,omitempty" yaml:"display_type,omitempty"`
	DisplayGroup        string         `json:"displayGroup,omitempty" yaml:"display_group,omitempty"`
}

// +genclient
// +genclient:noStatus
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// JobState describes a apb spec.
type JobState struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec JobStateSpec `json:"spec"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// JobStateList is a list of Database resources
type JobStateList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata"`

	Items []JobState `json:"items"`
}

// State - State the job is in.
type State string

// JobMethod - Method that the job is running.
type JobMethod string

const (
	//StateNotYetStarted - has not yet started state
	StateNotYetStarted State = "not-started"
	// StateInProgress - APB is in progress state
	StateInProgress State = "in-progress"
	// StateSucceeded - Succeeded state
	StateSucceeded State = "succeeded"
	// StateFailed - Failed state
	StateFailed State = "failed"
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

// JobStateSpec describes the job state for an async job
type JobStateSpec struct {
	State       State     `json:"state"`
	PodName     string    `json:"podName"`
	Method      JobMethod `json:"method"`
	Error       string    `json:"error"`
	Description string    `json:"description"`
}

// +genclient
// +genclient:noStatus
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// ServiceBinding describes a service binding.
type ServiceBinding struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec ServiceBindingSpec `json:"spec"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// ServiceBindingList is a list of ServiceBinding resources
type ServiceBindingList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata"`

	Items []ServiceBinding `json:"items"`
}

// ServiceBindingSpec  is a service binding.
type ServiceBindingSpec struct {
	ServiceInstanceID string `json:"serviceInstanceID"`
	// Store the parameters as a json encoded string.
	Parameters string `json:"parameters"`
	JobToken   string `json:"jobToken"`
}

// +genclient
// +genclient:noStatus
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// ServiceInstance describes a service binding.
type ServiceInstance struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec ServiceInstanceSpec `json:"spec"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// ServiceInstanceList is a list of ServiceInstance resources
type ServiceInstanceList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata"`

	Items []ServiceInstance `json:"items"`
}

// ServiceInstanceSpec  is a service instance.
type ServiceInstanceSpec struct {
	BundleID string  `json:"bundleID"`
	Context  Context `json:"context"`
	// Store the parameters as json encoded strings.
	Parameters   string   `json:"parameters"`
	DashboardURL string   `json:"dashboardUrl"`
	BindingIDs   []string `json:"bindingIDs"`
}

// Context is the context for the ServiceInstance.
type Context struct {
	Plateform string `json:"plateform"`
	Namespace string `json:"namespace"`
}
