package broker

import (
	"github.com/fusor/ansible-service-broker/pkg/ansibleapp"
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
}

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
	OrganizationID    uuid.UUID             `json:"organization_guid"`
	PlanID            uuid.UUID             `json:"plan_id"`
	ServiceID         uuid.UUID             `json:"service_id"`
	SpaceID           uuid.UUID             `json:"space_guid"`
	Parameters        ansibleapp.Parameters `json:"parameters,omitempty"`
	AcceptsIncomplete bool                  `json:"accepts_incomplete,omitempty"`
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
	ServiceID    uuid.UUID `json:"service_id"`
	PlanID       uuid.UUID `json:"plan_id"`
	AppID        uuid.UUID `json:"app_guid,omitempty"`
	BindResource struct {
		AppID uuid.UUID `json:"app_guid,omitempty"`
		Route string    `json:"route,omitempty"`
	} `json:"bind_resource,omitempty"`
	Parameters map[string]string `json:"parameters,omitempty"`
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

// TODO: What belongs on this thing?
type BootstrapResponse struct {
	SpecCount int
}
