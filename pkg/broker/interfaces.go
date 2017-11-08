package broker

import "github.com/openshift/ansible-service-broker/pkg/apb"

// JobStateDAO defines the actions that can be preformed on the JobState resource
type JobStateDAO interface {
	GetState(instanceUUID, operation string) (apb.JobState, error)
	SetState(id string, state apb.JobState) error
	GetSvcInstJobsByState(instanceID string, reqState apb.State) ([]apb.JobState, error)
}
