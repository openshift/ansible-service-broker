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

package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// LocalObjectReference - reference to an object in the same namespace.
type LocalObjectReference struct {
	Name string `json:"name"`
}

// +genclient
// +genclient:noStatus
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// Bundle describes a apb spec.
type Bundle struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   BundleSpec   `json:"spec"`
	Status BundleStatus `json:"status,omitempty"`
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

// BundleSpec is the spec for an bundle resource
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
	Alpha  string
	Plans  []Plan `json:"plans"`
	Delete bool   `json:"delete"`
}

// Status - The status for the bundle
type Status string

const (
	// ErrorStatus - Bundle error status
	ErrorStatus Status = "error"
	// OKStatus - Bundle ok status
	OKStatus Status = "ok"
)

// BundleStatus is the status for a bundle resource
type BundleStatus struct {
	Status      Status `json:"status"`
	Description string `json:"description,omitempty"`
}

// Plan - a plan for a bundle spec
type Plan struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	// Store the metadata as a json encoded string to preserve the genericness
	Metadata       string      `json:"metadata"`
	Free           bool        `json:"free"`
	Bindable       bool        `json:"bindable"`
	UpdatesTo      []string    `json:"updates_to,omitempty"`
	Parameters     []Parameter `json:"parameters,omitempty"`
	BindParameters []Parameter `json:"bindParameters,omitempty"`
}

// NilableNumber - Number that could be nil (e.g. when omitted from json/yaml)
type NilableNumber float64

// Parameter - describe the parameters for a plan
type Parameter struct {
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

// State - State the job is in.
type State string

// JobMethod - Method that the job is running.
type JobMethod string

const (
	//StateNotYetStarted - has not yet started state
	StateNotYetStarted State = "not-started"
	// StateInProgress - bundle is in progress state
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

// Job is a particluar job that is running for a particular bundle type
type Job struct {
	Podname          string       `json:"podname"`
	State            State        `json:"state"`
	Description      string       `json:"description,omitempty"`
	Error            string       `json:"error,omitempty"`
	Method           JobMethod    `json:"method"`
	LastModifiedTime *metav1.Time `json:"lastModifiedTime,omitempty"`
}

// +genclient
// +genclient:noStatus
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// BundleBinding describes a service binding.
type BundleBinding struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   BundleBindingSpec   `json:"spec"`
	Status BundleBindingStatus `json:"status,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// BundleBindingList is a list of ServiceBinding resources
type BundleBindingList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata"`

	Items []BundleBinding `json:"items"`
}

// BundleBindingSpec  is a service binding.
type BundleBindingSpec struct {
	BundleInstance LocalObjectReference `json:"bundleInstance"`
	// Store the parameters as a json encoded string.
	Parameters string `json:"parameters"`
}

// BundleBindingStatus - status of the bundle
type BundleBindingStatus struct {
	State           State          `json:"state"`
	LastDescription string         `json:"lastDescription,omitempty"`
	Jobs            map[string]Job `json:"jobs"`
}

// +genclient
// +genclient:noStatus
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// BundleInstance describes a service binding.
type BundleInstance struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   BundleInstanceSpec   `json:"spec"`
	Status BundleInstanceStatus `json:"status,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// BundleInstanceList is a list of ServiceInstance resources
type BundleInstanceList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata"`

	Items []BundleInstance `json:"items"`
}

// BundleInstanceSpec  is a service instance.
type BundleInstanceSpec struct {
	Bundle  LocalObjectReference `json:"bundle"`
	Context Context              `json:"context"`
	// Store the parameters as json encoded strings.
	Parameters   string `json:"parameters"`
	DashboardURL string `json:"dashboardUrl"`
}

// Context is the context for the ServiceInstance.
type Context struct {
	Platform  string `json:"platform"`
	Namespace string `json:"namespace"`
}

// BundleInstanceStatus status is a service instance status.
type BundleInstanceStatus struct {
	Bindings        []LocalObjectReference `json:"bindings"`
	State           State                  `json:"state"`
	LastDescription string                 `json:"lastDescription,omitempty"`
	Jobs            map[string]Job         `json:"jobs"`
}
