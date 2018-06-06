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

package broker

import (
	"encoding/json"
	"errors"

	"github.com/automationbroker/bundle-lib/bundle"
	schema "github.com/lestrrat/go-jsschema"
	"github.com/pborman/uuid"
	log "github.com/sirupsen/logrus"
)

const (
	// Development tooling consts
	apbPushRegName = "apb-push"
)

const (
	planParameterKey      = "_apb_plan_id"
	serviceClassIDKey     = "_apb_service_class_id"
	serviceInstIDKey      = "_apb_service_instance_id"
	lastRequestingUserKey = "_apb_last_requesting_user"
	serviceBindingIDKey   = "_apb_service_binding_id"
)

// WorkTopic - Topic jobs can publish messages to, and subscribers can listen to
type WorkTopic string

// Work Engine messaging consts
const (
	ProvisionTopic   WorkTopic = "provision_topic"
	DeprovisionTopic WorkTopic = "deprovision_topic"
	UpdateTopic      WorkTopic = "update_topic"
	BindingTopic     WorkTopic = "binding_topic"
	UnbindingTopic   WorkTopic = "unbinding_topic"
)

var workTopicSet = map[WorkTopic]bool{
	ProvisionTopic:   true,
	DeprovisionTopic: true,
	UpdateTopic:      true,
	BindingTopic:     true,
	UnbindingTopic:   true,
}

// IsValidWorkTopic - Check if WorkTopic is part of acceptable set
func IsValidWorkTopic(topic WorkTopic) bool {
	_, ok := workTopicSet[topic]
	return ok
}

// Service - Service object to be returned.
// based on https://github.com/openservicebrokerapi/servicebroker/blob/master/spec.md#service-objects
type Service struct {
	Name                 string                 `json:"name"`
	ID                   string                 `json:"id"`
	Description          string                 `json:"description"`
	Tags                 []string               `json:"tags,omitempty"`
	Requires             []string               `json:"requires,omitempty"`
	Bindable             bool                   `json:"bindable"`
	Metadata             map[string]interface{} `json:"metadata,omitempty"`
	DashboardClient      *DashboardClient       `json:"dashboard_client,omitempty"`
	InstancesRetrievable bool                   `json:"instances_retrievable"`
	BindingsRetrievable  bool                   `json:"bindings_retrievable"`
	PlanUpdatable        bool                   `json:"plan_updateable,omitempty"`
	Plans                []Plan                 `json:"plans"`
}

// DashboardClient - Dashboard Client to be returned
// based on https://github.com/openservicebrokerapi/servicebroker/blob/master/spec.md#dashboard-client-object
type DashboardClient struct {
	ID          string `json:"id"`
	Secret      string `json:"secret"`
	RedirectURI string `json:"redirect_uri"`
}

// Plan - Plan to be returned
// based on https://github.com/openservicebrokerapi/servicebroker/blob/master/spec.md#plan-object
type Plan struct {
	ID          string                 `json:"id"`
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
	Free        bool                   `json:"free,omitempty"`
	Bindable    bool                   `json:"bindable,omitempty"`
	Schemas     Schema                 `json:"schemas,omitempty"`
	UpdatesTo   []string               `json:"updates_to,omitempty"`
}

// Schema  - Schema to be returned
// based on 2.13 of the open service broker api. https://github.com/avade/servicebroker/blob/cda8c57b6a4bb7eaee84be20bb52dc155269758a/spec.md
type Schema struct {
	ServiceInstance ServiceInstance `json:"service_instance"`
	ServiceBinding  ServiceBinding  `json:"service_binding"`
}

// ServiceInstance - Schema definitions for creating and updating a service instance.
// Toyed with the idea of making an InputParameters
// that was a *schema.Schema
// based on 2.13 of the open service broker api. https://github.com/avade/servicebroker/blob/cda8c57b6a4bb7eaee84be20bb52dc155269758a/spec.md
type ServiceInstance struct {
	Create map[string]*schema.Schema `json:"create"`
	Update map[string]*schema.Schema `json:"update"`
}

// ServiceBinding - Schema definitions for creating a service binding.
type ServiceBinding struct {
	Create map[string]*schema.Schema `json:"create"`
}

// CatalogResponse - Response for the catalog call.
// Defined here https://github.com/openservicebrokerapi/servicebroker/blob/v2.12/spec.md#response
type CatalogResponse struct {
	Services []Service `json:"services"`
}

// LastOperationRequest - Request to obtain state information about an action that was taken
// Defined in more detail here https://github.com/openservicebrokerapi/servicebroker/blob/v2.12/spec.md#polling-last-operation
type LastOperationRequest struct {
	ServiceID string `json:"service_id"`
	PlanID    string `json:"plan_id"`
	Operation string `json:"operation"`
}

// LastOperationState - State that the Last Operation is allowed to be.
// https://github.com/openservicebrokerapi/servicebroker/blob/v2.12/spec.md#body
type LastOperationState string

const (
	// LastOperationStateInProgress - In Progress state for last operation.
	LastOperationStateInProgress LastOperationState = "in progress"
	// LastOperationStateSucceeded - Succeeded state for the last operation.
	LastOperationStateSucceeded LastOperationState = "succeeded"
	// LastOperationStateFailed - Failed state for the last operation.
	LastOperationStateFailed LastOperationState = "failed"
)

// LastOperationResponse - Response for the last operation request.
// Defined here https://github.com/openservicebrokerapi/servicebroker/blob/v2.12/spec.md#response-1
type LastOperationResponse struct {
	State       LastOperationState `json:"state"`
	Description string             `json:"description,omitempty"`
}

// ProvisionRequest - Request for provision
// Defined here https://github.com/openservicebrokerapi/servicebroker/blob/v2.12/spec.md#request-2
type ProvisionRequest struct {
	OrganizationID    uuid.UUID         `json:"organization_guid"`
	PlanID            string            `json:"plan_id"`
	ServiceID         string            `json:"service_id"`
	SpaceID           uuid.UUID         `json:"space_guid"`
	Context           bundle.Context    `json:"context"`
	Parameters        bundle.Parameters `json:"parameters,omitempty"`
	AcceptsIncomplete bool              `json:"accepts_incomplete,omitempty"`
}

// ProvisionResponse - Response for provision
// Defined here https://github.com/openservicebrokerapi/servicebroker/blob/v2.12/spec.md#response-2
type ProvisionResponse struct {
	DashboardURL string `json:"dashboard_url,omitempty"`
	Operation    string `json:"operation,omitempty"`
}

// UpdateRequest - Request for an update for a service instance.
// Defined here https://github.com/openservicebrokerapi/servicebroker/blob/v2.12/spec.md#request-3
type UpdateRequest struct {
	ServiceID      string            `json:"service_id"`
	PlanID         string            `json:"plan_id,omitempty"`
	Parameters     map[string]string `json:"parameters,omitempty"`
	PreviousValues struct {
		PlanID         string    `json:"plan_id,omitempty"`
		ServiceID      string    `json:"service_id,omitempty"`
		OrganizationID uuid.UUID `json:"organization_id,omitempty"`
		SpaceID        uuid.UUID `json:"space_id,omitempty"`
	} `json:"previous_values,omitempty"`
	Context           bundle.Context `json:"context"`
	AcceptsIncomplete bool           `json:"accepts_incomplete,omitempty"`
}

// UpdateResponse - Response for an update for a service instance.
// Defined here https://github.com/openservicebrokerapi/servicebroker/blob/v2.12/spec.md#response-3
type UpdateResponse struct {
	Operation string `json:"operation,omitempty"`
}

// BindRequest - Request for a bind
// Defined here https://github.com/openservicebrokerapi/servicebroker/blob/v2.12/spec.md#request-4
type BindRequest struct {
	ServiceID string `json:"service_id"`
	PlanID    string `json:"plan_id"`
	// Deprecated: AppID deprecated in favor of BindResource.AppID
	AppID uuid.UUID `json:"app_guid,omitempty"`

	BindResource struct {
		AppID uuid.UUID `json:"app_guid,omitempty"`
		Route string    `json:"route,omitempty"`
	} `json:"bind_resource,omitempty"`
	Parameters bundle.Parameters `json:"parameters,omitempty"`
}

// BindResponse - Response for a bind
// Defined here https://github.com/openservicebrokerapi/servicebroker/blob/v2.12/spec.md#response-4
type BindResponse struct {
	Credentials     map[string]interface{} `json:"credentials,omitempty"`
	SyslogDrainURL  string                 `json:"syslog_drain_url,omitempty"`
	RouteServiceURL string                 `json:"route_service_url,omitempty"`
	VolumeMounts    []interface{}          `json:"volume_mounts,omitempty"`
	Operation       string                 `json:"operation,omitempty"`
}

// NewBindResponse - creates a BindResponse based on available credentials.
func NewBindResponse(pCreds, bCreds *bundle.ExtractedCredentials) (*BindResponse, error) {
	// Can't bind to anything if we have nothing to return to the catalog
	if pCreds == nil && bCreds == nil {
		log.Errorf("No extracted credentials found from provision or bind instance ID")
		return nil, errors.New("No credentials available")
	}

	if bCreds != nil {
		log.Debugf("bind creds: %v", bCreds.Credentials)
		return &BindResponse{Credentials: bCreds.Credentials}, nil
	}

	log.Debugf("provision bind creds: %v", pCreds.Credentials)
	return &BindResponse{Credentials: pCreds.Credentials}, nil
}

// DeprovisionResponse - Response for a deprovision
// Defined here https://github.com/openservicebrokerapi/servicebroker/blob/v2.12/spec.md#response-6
type DeprovisionResponse struct {
	Operation string `json:"operation,omitempty"`
}

// UnbindResponse - Response for unbinding
// Defined here https://github.com/openservicebrokerapi/servicebroker/blob/v2.12/spec.md#response-5
type UnbindResponse struct {
	Operation string `json:"operation,omitempty"`
}

// ErrorResponse - Error response for all broker errors
// Defined here https://github.com/openservicebrokerapi/servicebroker/blob/v2.12/spec.md#broker-errors
type ErrorResponse struct {
	Description string `json:"description"`
}

// BootstrapResponse - The response for a bootstrap request
// TODO: What belongs on this thing?
type BootstrapResponse struct {
	SpecCount  int `json:"spec_count"`
	ImageCount int `json:"image_count"`
}

// ServiceInstanceResponse - The response for a get service instance request
type ServiceInstanceResponse struct {
	ServiceID    string            `json:"service_id"`
	PlanID       string            `json:"plan_id"`
	DashboardURL string            `json:"dashboard_url,omitempty"`
	Parameters   bundle.Parameters `json:"parameters,omitempty"`
}

// UserInfo - holds information about the user that created a resource.
type UserInfo struct {
	Username string              `json:"username"`
	UID      string              `json:"uid"`
	Groups   []string            `json:"groups,omitempty"`
	Scopes   []string            `json:"scopes.authorization.openshift.io"`
	Extra    map[string][]string `json:"extra,omitempty"`
}

// JobMsg - Message to be returned from the various jobs
type JobMsg struct {
	InstanceUUID         string                      `json:"instance_uuid"`
	JobToken             string                      `json:"job_token"`
	SpecID               string                      `json:"spec_id"`
	PodName              string                      `json:"podname"`
	Msg                  string                      `json:"msg"`
	State                bundle.JobState             `json:"state"`
	ExtractedCredentials bundle.ExtractedCredentials `json:"extracted_credentials"`
	DashboardURL         string                      `json:"dashboard_url"`
	BindingUUID          string                      `json:"binding_uuid"`
	Error                string                      `json:"error"`
}

// Render - Display the job message.
func (jm JobMsg) Render() string {
	render, _ := json.Marshal(jm)
	return string(render)
}

// SubscriberDAO defines the interface subscribers use when persisting state
type SubscriberDAO interface {
	SetState(id string, state bundle.JobState) (string, error)
	GetServiceInstance(id string) (*bundle.ServiceInstance, error)
	DeleteServiceInstance(id string) error
	GetBindInstance(id string) (*bundle.BindInstance, error)
	DeleteBinding(bundle.BindInstance, bundle.ServiceInstance) error
	SetServiceInstance(id string, serviceInstance *bundle.ServiceInstance) error
}

// WorkSubscriber - Defines how a Subscriber can be notified of changes
type WorkSubscriber interface {
	ID() string
	Notify(msg JobMsg)
}
