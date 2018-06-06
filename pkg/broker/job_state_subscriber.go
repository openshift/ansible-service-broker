package broker

import (
	"fmt"

	"github.com/automationbroker/bundle-lib/bundle"
	"github.com/automationbroker/bundle-lib/runtime"
	log "github.com/sirupsen/logrus"
)

// JobStateSubscriber is responsible for handling and persisting JobState changes
type JobStateSubscriber struct {
	dao SubscriberDAO
}

// NewJobStateSubscriber returns a newly initialized JobStateSubscriber
func NewJobStateSubscriber(dao SubscriberDAO) *JobStateSubscriber {
	return &JobStateSubscriber{
		dao: dao,
	}
}

func isBinding(msg JobMsg) bool {
	return msg.State.Method == bundle.JobMethodBind || msg.State.Method == bundle.JobMethodUnbind
}

// ID is used as an identifier for the type of subscriber
func (jss *JobStateSubscriber) ID() string {
	return "jobstate"
}

// Notify external API to notify this subscriber of a change in the Job
func (jss *JobStateSubscriber) Notify(msg JobMsg) {
	log.Debugf("JobStateSubscriber Notify : msg state %v ", msg.State)
	id := msg.InstanceUUID
	if isBinding(msg) {
		id = msg.BindingUUID
	}

	// Bug 1583064 - Should not launch multi unbind sandboxes frequently while
	// unbind failed.
	//
	// If the unbind fails we really want to let them know it failed because it
	// is probably something the APB developer needs to fix. But if the action
	// is non-existent we really need to prevent the multiple sandboxes.
	//
	// if we can't find the action, there is no point in continuing.
	// marking job as SUCCEEDED but setting the message to be the actual error
	// string. This will only happen for the optional actions: bind & unbind.
	if msg.State.Error == runtime.ErrorActionNotFound.Error() {
		log.Debug("We have an 'action not found' error. Looking to see if this is a bind or an unbind.")
		if msg.State.Method == bundle.JobMethodBind || msg.State.Method == bundle.JobMethodUnbind {
			log.Info("bind or unbind action not found, marking as succeeded.")
			msg.State.State = bundle.StateSucceeded
			// leave error message as not found
		}
	}

	if _, err := jss.dao.SetState(id, msg.State); err != nil {
		log.Errorf("Error JobStateSubscriber failed to set state after action %v completed with state %s err: %v", msg.State.Method, msg.State.State, err)
		return
	}
	if msg.State.State == bundle.StateSucceeded {
		if err := jss.handleSucceeded(msg); err != nil {
			log.Errorf("Error after job succeeded : %v", err)
			return
		}
	}
	//TODO: Need to get the service instance
	//TODO: NOTE: THIS NEEDS TO BREAK OUT TO OWN SUBSCRIBER
	// Update with dashboard URL.
	if msg.DashboardURL != "" {
		instance, err := jss.dao.GetServiceInstance(msg.InstanceUUID)
		if err != nil {
			log.Errorf("Error after job succeeded : %v", err)
			return
		}
		instance.DashboardURL = msg.DashboardURL
		err = jss.dao.SetServiceInstance(msg.InstanceUUID, instance)
		if err != nil {
			log.Errorf("Error after job succeeded : %v", err)
			return
		}
	}
	return
}

// handle specific logic for the succeeded state
func (jss *JobStateSubscriber) handleSucceeded(msg JobMsg) error {
	log.Debugf("JobStateSubscriber handleSucceeded : msg state %v ", msg.State)
	switch msg.State.Method {
	case bundle.JobMethodDeprovision:
		if err := jss.cleanupAfterDeprovision(msg); err != nil {
			return fmt.Errorf("Failed cleaning up deprovision after job succeeded, error: %v", err)
		}
	case bundle.JobMethodUnbind:
		if err := jss.cleanupAfterUnbind(msg); err != nil {
			return fmt.Errorf("Failed cleaning up unbinding after job succeeded, error: %v", err)
		}
	}
	return nil
}

func (jss *JobStateSubscriber) cleanupAfterDeprovision(msg JobMsg) error {
	log.Debugf("JobStateSubscriber cleanupAfterDeprovision : msg state %v ", msg.State)
	if deleteErr := jss.dao.DeleteServiceInstance(msg.InstanceUUID); deleteErr != nil {
		msg.State.State = bundle.StateFailed
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
		msg.State.State = bundle.StateFailed
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
	if err := jss.dao.DeleteBinding(*bindInstance, *svcInstance); err != nil {
		return setFailed(fmt.Errorf("Error cleaning up binding instance [ %s ] during cleanup of unbind job : %v", bindInstance.ID.String(), err))
	}
	log.Infof("Clean up of binding instance [ %s ] done. Unbinding successful", bindInstance.ID.String())
	return nil
}
