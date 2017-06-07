package broker

import (
	"github.com/openshift/ansible-service-broker/pkg/apb"
	"github.com/pborman/uuid"
)

type Service struct {
	Name            string                 `json:"name"`
	ID              uuid.UUID              `json:"id"`
	Description     string                 `json:"description"`
	Tags            []string               `json:"tags,omitempty"`
	Requires        []string               `json:"requires,omitempty"`
	Bindable        bool                   `json:"bindable"`
	Metadata        map[string]interface{} `json:"metadata,omitempty"`
	DashboardClient *DashboardClient       `json:"dashboard_client,omitempty"`
	PlanUpdatable   bool                   `json:"plan_updateable,omitempty"`
	Plans           []Plan                 `json:"plans"`
}

type DashboardClient struct {
	ID          string `json:"id"`
	Secret      string `json:"secret"`
	RedirectURI string `json:"redirect_uri"`
}

type Plan struct {
	ID          uuid.UUID              `json:"id"`
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
	Free        bool                   `json:"free,omitempty"`
	Bindable    bool                   `json:"bindable,omitempty"`
	Schemas     Schema                 `json:"schemas,omitempty"`
}

type Schema struct {
	ServiceInstance ServiceInstance `json:"service_instance"`
	ServiceBinding  ServiceBinding  `json:"service_binding"`
}

type ServiceInstance struct {
	Create InputParameter `json:"create"`
	Update InputParameter `json:"update"`
}

type ServiceBinding struct {
	Create InputParameter `json:"create"`
}

type InputParameter struct {
	Parameters interface{} `json:"parameters"`
}

/*
type
title
properties:
	propertyname: {
		title
		type
		length
		default
	}
*/

type CatalogResponse struct {
	Services []Service `json:"services"`
}

type LastOperationRequest struct {
	ServiceID uuid.UUID `json:"service_id"`
	PlanID    uuid.UUID `json:"plan_id"`
	Operation string    `json:"operation"`
}

type LastOperationState string

const (
	LastOperationStateInProgress LastOperationState = "in progress"
	LastOperationStateSucceeded  LastOperationState = "succeeded"
	LastOperationStateFailed     LastOperationState = "failed"
)

type LastOperationResponse struct {
	State       LastOperationState `json:"state"`
	Description string             `json:"description,omitempty"`
}

type ProvisionRequest struct {
	OrganizationID    uuid.UUID      `json:"organization_guid"`
	PlanID            uuid.UUID      `json:"plan_id"`
	ServiceID         uuid.UUID      `json:"service_id"`
	SpaceID           uuid.UUID      `json:"space_guid"`
	Parameters        apb.Parameters `json:"parameters,omitempty"`
	AcceptsIncomplete bool           `json:"accepts_incomplete,omitempty"`
}

type ProvisionResponse struct {
	DashboardURL string `json:"dashboard_url,omitempty"`
	Operation    string `json:"operation,omitempty"`
}

type UpdateRequest struct {
	ServiceID      uuid.UUID         `json:"service_id"`
	PlanID         uuid.UUID         `json:"plan_id,omitempty"`
	Parameters     map[string]string `json:"parameters,omitempty"`
	PreviousValues struct {
		PlanID         uuid.UUID `json:"plan_id,omitempty"`
		ServiceID      uuid.UUID `json:"service_id,omitempty"`
		OrganizationID uuid.UUID `json:"organization_id,omitempty"`
		SpaceID        uuid.UUID `json:"space_id,omitempty"`
	} `json:"previous_values,omitempty"`
	AcceptsIncomplete bool `json:"accepts_incomplete,omitempty"`
}

type UpdateResponse struct {
	Operation string `json:"operation,omitempty"`
}

type BindRequest struct {
	ServiceID uuid.UUID `json:"service_id"`
	PlanID    uuid.UUID `json:"plan_id"`
	// Deprecated: AppID deprecated in favor of BindResource.AppID
	AppID uuid.UUID `json:"app_guid,omitempty"`

	BindResource struct {
		AppID uuid.UUID `json:"app_guid,omitempty"`
		Route string    `json:"route,omitempty"`
	} `json:"bind_resource,omitempty"`
	Parameters apb.Parameters `json:"parameters,omitempty"`
}

type BindResponse struct {
	Credentials     map[string]interface{} `json:"credentials,omitempty"`
	SyslogDrainURL  string                 `json:"syslog_drain_url,omitempty"`
	RouteServiceURL string                 `json:"route_service_url,omitempty"`
	VolumeMounts    []interface{}          `json:"volume_mounts,omitempty"`
}

type DeprovisionResponse struct {
	Operation string `json:"operation,omitempty"`
}

type ErrorResponse struct {
	Description string `json:"description"`
}

// BootstrapResponse - The response for a bootstrap request
// TODO: What belongs on this thing?
type BootstrapResponse struct {
	SpecCount  int `json:"spec_count"`
	ImageCount int `json:"image_count"`
}
