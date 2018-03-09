package broker

import (
	"fmt"

	"github.com/openshift/ansible-service-broker/pkg/apb"
)

// JobStateSubscriber is responsible for handling and persisiting JobState changes
type JobStateSubscriber struct {
	dao SubscriberDAO
}

// NewJobStateSubscriber returns a newly initialised JobStateSubscriber
func NewJobStateSubscriber(dao SubscriberDAO) *JobStateSubscriber {
	return &JobStateSubscriber{
		dao: dao,
	}
}

func isBinding(msg JobMsg) bool {
	return msg.State.Method == apb.JobMethodBind || msg.State.Method == apb.JobMethodUnbind
}

// Notify external API to notify this subscriber of a change in the Job
func (jss *JobStateSubscriber) Notify(msg JobMsg) {
	log.Debugf("JobStateSubscriber Notify : msg state %v ", msg.State)
	id := msg.InstanceUUID
	if isBinding(msg) {
		id = msg.BindingUUID
	}
	if _, err := jss.dao.SetState(id, msg.State); err != nil {
		log.Errorf("Error JobStateSubscriber failed to set state after action %v completed with state %s err: %v", msg.State.Method, msg.State.State, err)
		return
	}
	if msg.State.State == apb.StateSucceeded {
		if err := jss.handleSucceeded(msg); err != nil {
			log.Errorf("Error after job succeeded : %v", err)
			return
		}
	}
}

// handle specific logic for the succeeded state
func (jss *JobStateSubscriber) handleSucceeded(msg JobMsg) error {
	log.Debugf("JobStateSubscriber handleSucceeded : msg state %v ", msg.State)
	switch msg.State.Method {
	case apb.JobMethodDeprovision:
		if err := jss.cleanupAfterDeprovision(msg); err != nil {
			return fmt.Errorf("Failed cleaning up deprovision after job succeeded, error: %v", err)
		}
	case apb.JobMethodUnbind:
		if err := jss.cleanupAfterUnbind(msg); err != nil {
			return fmt.Errorf("Failed cleaning up unbinding after job succeeded, error: %v", err)
		}
	}
	return nil
}

func (jss *JobStateSubscriber) cleanupAfterDeprovision(msg JobMsg) error {
	log.Debugf("JobStateSubscriber cleanupAfterDeprovision : msg state %v ", msg.State)
	if deleteErr := jss.dao.DeleteServiceInstance(msg.InstanceUUID); deleteErr != nil {
		msg.State.State = apb.StateFailed
		if _, err := jss.dao.SetState(msg.InstanceUUID, msg.State); err != nil {
			return fmt.Errorf("Error setting failed state after error : %s deleting service instance : %s", deleteErr, err)
		}
		return deleteErr
	}
	return nil
}

func (jss *JobStateSubscriber) cleanupAfterUnbind(msg JobMsg) error {
	log.Debugf("JobStateSubscriber cleanupAfterUnbind : msg state %v ", msg.State)
	// util function to set the state to failed and ensure no error information is lost
	var setFailed = func(failureErr error) error {
		msg.State.State = apb.StateFailed
		if _, err := jss.dao.SetState(msg.InstanceUUID, msg.State); err != nil {
			return fmt.Errorf("Error setting unbind state to failed after error %v occurred : %v during cleanup of unbind", failureErr, err)
		}
		return failureErr
	}
	svcInstance, err := jss.dao.GetServiceInstance(msg.InstanceUUID)
	if err != nil {
		return setFailed(fmt.Errorf("Error getting service instance [ %s ] during cleanup of unbind job : %v", msg.InstanceUUID, err))
	}
	bindInstance, err := jss.dao.GetBindInstance(msg.BindingUUID)
	if err != nil {
		return setFailed(fmt.Errorf("Error getting bind instance [ %s ] during cleanup of unbind job : %v", msg.BindingUUID, err))
	}
	id := bindInstance.ID.String()
	if err := jss.dao.DeleteBindInstance(id); err != nil {
		return setFailed(fmt.Errorf("Error deleting bind instance [ %s ] during cleanup of unbind job : %v", id, err))
	}
	svcInstance.RemoveBinding(bindInstance.ID)
	if err := jss.dao.SetServiceInstance(svcInstance.ID.String(), svcInstance); err != nil {
		return setFailed(fmt.Errorf("Error setting service instance [ %s ] during clean up of unbind of [ %s ]", svcInstance.ID.String(), id))
	}
	log.Info("Clean up of binding instance [ %s ] done. Unbinding successful", id)
	return nil
}
