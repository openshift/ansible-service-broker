package broker_test

import (
	"testing"

	"time"

	"fmt"

	"github.com/openshift/ansible-service-broker/pkg/apb"
	"github.com/openshift/ansible-service-broker/pkg/broker"
	"github.com/pborman/uuid"
)

func TestUnBindingJob_Run(t *testing.T) {
	instanceID := uuid.NewRandom()
	bindingInst := &apb.BindInstance{ID: uuid.NewRandom()}
	serviceInstance := &apb.ServiceInstance{
		ID: instanceID,
		Spec: &apb.Spec{
			ID: "test",
		},
	}
	cases := []struct {
		Name          string
		UnBinder      apb.UnBinder
		UnBindParams  *apb.Parameters
		SkipExecution bool
		Validate      func(msg broker.JobMsg) error
	}{
		{
			Name: "expect a success msg",
			UnBinder: func(si *apb.ServiceInstance, params *apb.Parameters) error {
				return nil
			},
			Validate: func(msg broker.JobMsg) error {
				if msg.State.State != apb.StateSucceeded {
					return fmt.Errorf("expected the state to be %v but got %v", apb.StateSucceeded, msg.State.State)
				}
				if msg.State.Method != apb.JobMethodUnbind {
					return fmt.Errorf("expected job method to be %v but it was %v", apb.JobMethodUnbind, msg.State.Method)
				}
				return nil
			},
		},
		{
			Name:          "expect a success msg when skipping apb execution",
			SkipExecution: true,
			UnBinder: func(si *apb.ServiceInstance, params *apb.Parameters) error {
				return nil
			},
			Validate: func(msg broker.JobMsg) error {
				if msg.State.State != apb.StateSucceeded {
					return fmt.Errorf("expected the state to be %v but got %v", apb.StateSucceeded, msg.State.State)
				}
				if msg.State.Method != apb.JobMethodUnbind {
					return fmt.Errorf("expected job method to be %v but it was %v", apb.JobMethodUnbind, msg.State.Method)
				}
				return nil
			},
		},
		{
			Name: "expect failure state and generic error when unknown error type",
			UnBinder: func(si *apb.ServiceInstance, params *apb.Parameters) error {
				return fmt.Errorf("should not see")
			},
			Validate: func(msg broker.JobMsg) error {
				if msg.State.State != apb.StateFailed {
					return fmt.Errorf("expected the Job to be in state %v but was in %v ", apb.StateFailed, msg.State.State)
				}
				if msg.State.Method != apb.JobMethodUnbind {
					return fmt.Errorf("expected job method to be %v but it was %v", apb.JobMethodUnbind, msg.State.Method)
				}
				if msg.State.Error == "" {
					return fmt.Errorf("expected an error in the job state but got none")
				}
				if msg.State.Error == "should not see" {
					return fmt.Errorf("expected not to see the error msg %s it should have been replaced with a generic error ", msg.State.Error)
				}
				return nil
			},
		},
		{
			Name: "expect failure state and full error when known error type",
			UnBinder: func(si *apb.ServiceInstance, params *apb.Parameters) error {
				return apb.ErrorPodPullErr
			},
			Validate: func(msg broker.JobMsg) error {
				if msg.State.State != apb.StateFailed {
					return fmt.Errorf("expected the Job to be in state %v but was in %v ", apb.StateFailed, msg.State.State)
				}
				if msg.State.Method != apb.JobMethodUnbind {
					return fmt.Errorf("expected job method to be %v but it was %v", apb.JobMethodUnbind, msg.State.Method)
				}
				if msg.State.Error == "" {
					return fmt.Errorf("expected an error in the job state but got none")
				}
				if msg.State.Error != apb.ErrorPodPullErr.Error() {
					return fmt.Errorf("expected to see the error msg %s but got %s ", apb.ErrorPodPullErr, msg.State.Error)
				}
				return nil
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.Name, func(t *testing.T) {
			unbindJob := broker.NewUnbindingJob(serviceInstance, bindingInst, tc.UnBindParams, tc.UnBinder, tc.SkipExecution)
			receiver := make(chan broker.JobMsg)
			timedOut := false
			time.AfterFunc(1*time.Second, func() {
				close(receiver)
				timedOut = true
			})
			go unbindJob.Run("", receiver)

			msg := <-receiver
			if timedOut {
				t.Fatal("timed out waiting for a msg from the Job")
			}
			if err := tc.Validate(msg); err != nil {
				t.Fatal("failed to validate the jobmsg ", err)
			}

		})
	}
}
