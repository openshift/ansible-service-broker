package broker

import (
	schema "github.com/lestrrat/go-jsschema"
	"github.com/openshift/ansible-service-broker/pkg/apb"
	"github.com/pborman/uuid"
)

// Service - Service object to be returned.
// based on https://github.com/openservicebrokerapi/servicebroker/blob/master/spec.md#service-objects
type Service struct {
	Name            string                 `json:"name"`
	ID              string                 `json:"id"`
	Description     string                 `json:"description"`
	Tags            []string               `json:"tags,omitempty"`
	Requires        []string               `json:"requires,omitempty"`
	Bindable        bool                   `json:"bindable"`
	Metadata        map[string]interface{} `json:"metadata,omitempty"`
	DashboardClient *DashboardClient       `json:"dashboard_client,omitempty"`
	PlanUpdatable   bool                   `json:"plan_updateable,omitempty"`
	Plans           []Plan                 `json:"plans"`
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
	ID          uuid.UUID              `json:"id"`
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
	Free        bool                   `json:"free,omitempty"`
	Bindable    bool                   `json:"bindable,omitempty"`
	Schemas     Schema                 `json:"schemas,omitempty"`
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

// ServiceBinding - Schema definitions for creating a service binidng.
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
	ServiceID string    `json:"service_id"`
	PlanID    uuid.UUID `json:"plan_id"`
	Operation string    `json:"operation"`
}

// LastOperationState - State that the Last Operation is allowed to be.
// https://github.com/openservicebrokerapi/servicebroker/blob/v2.12/spec.md#body
type LastOperationState string

const (
	//LastOperationStateInProgress - In Progress state for last operation.
	LastOperationStateInProgress LastOperationState = "in progress"
	//LastOperationStateSucceeded - Succeeded state for the last operation.
	LastOperationStateSucceeded LastOperationState = "succeeded"
	//LastOperationStateFailed - Failed state for the last operation.
	LastOperationStateFailed LastOperationState = "failed"
)

// LastOperationResponse - Response for the laster operation request.
// Defined here https://github.com/openservicebrokerapi/servicebroker/blob/v2.12/spec.md#response-1
type LastOperationResponse struct {
	State       LastOperationState `json:"state"`
	Description string             `json:"description,omitempty"`
}

// ProvisionRequest - Request for provision
// Defined here https://github.com/openservicebrokerapi/servicebroker/blob/v2.12/spec.md#request-2
type ProvisionRequest struct {
	OrganizationID    uuid.UUID      `json:"organization_guid"`
	PlanID            uuid.UUID      `json:"plan_id"`
	ServiceID         string         `json:"service_id"`
	SpaceID           uuid.UUID      `json:"space_guid"`
	Context           apb.Context    `json:"context"`
	Parameters        apb.Parameters `json:"parameters,omitempty"`
	AcceptsIncomplete bool           `json:"accepts_incomplete,omitempty"`
}

// ProvisionResponse - Response for provison
// Defined here https://github.com/openservicebrokerapi/servicebroker/blob/v2.12/spec.md#response-2
type ProvisionResponse struct {
	DashboardURL string `json:"dashboard_url,omitempty"`
	Operation    string `json:"operation,omitempty"`
}

// UpdateRequest - Request for an update for a service instance.
// Defined here https://github.com/openservicebrokerapi/servicebroker/blob/v2.12/spec.md#request-3
type UpdateRequest struct {
	ServiceID      string            `json:"service_id"`
	PlanID         uuid.UUID         `json:"plan_id,omitempty"`
	Parameters     map[string]string `json:"parameters,omitempty"`
	PreviousValues struct {
		PlanID         uuid.UUID `json:"plan_id,omitempty"`
		ServiceID      string    `json:"service_id,omitempty"`
		OrganizationID uuid.UUID `json:"organization_id,omitempty"`
		SpaceID        uuid.UUID `json:"space_id,omitempty"`
	} `json:"previous_values,omitempty"`
	AcceptsIncomplete bool `json:"accepts_incomplete,omitempty"`
}

// UpdateResponse - Response for an update for a service instance.
// Defined here https://github.com/openservicebrokerapi/servicebroker/blob/v2.12/spec.md#response-3
type UpdateResponse struct {
	Operation string `json:"operation,omitempty"`
}

// BindRequest - Request for a bind
// Defined here https://github.com/openservicebrokerapi/servicebroker/blob/v2.12/spec.md#request-4
type BindRequest struct {
	ServiceID string    `json:"service_id"`
	PlanID    uuid.UUID `json:"plan_id"`
	// Deprecated: AppID deprecated in favor of BindResource.AppID
	AppID uuid.UUID `json:"app_guid,omitempty"`

	BindResource struct {
		AppID uuid.UUID `json:"app_guid,omitempty"`
		Route string    `json:"route,omitempty"`
	} `json:"bind_resource,omitempty"`
	Parameters apb.Parameters `json:"parameters,omitempty"`
}

//BindResponse - Response for a bind
// Defined here https://github.com/openservicebrokerapi/servicebroker/blob/v2.12/spec.md#response-4
type BindResponse struct {
	Credentials     map[string]interface{} `json:"credentials,omitempty"`
	SyslogDrainURL  string                 `json:"syslog_drain_url,omitempty"`
	RouteServiceURL string                 `json:"route_service_url,omitempty"`
	VolumeMounts    []interface{}          `json:"volume_mounts,omitempty"`
}

// DeprovisionResponse - Response for a deprovision
//  Defined here https://github.com/openservicebrokerapi/servicebroker/blob/v2.12/spec.md#response-6
type DeprovisionResponse struct {
	Operation string `json:"operation,omitempty"`
}

// UnbindResponse - Response for unbinding
// Defined here https://github.com/openservicebrokerapi/servicebroker/blob/v2.12/spec.md#response-5
type UnbindResponse struct{}

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
