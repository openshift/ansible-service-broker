package broker_test

import (
	"testing"

	"time"

	"fmt"

	"github.com/openshift/ansible-service-broker/pkg/apb"
	"github.com/openshift/ansible-service-broker/pkg/broker"
	"github.com/pborman/uuid"
)

func TestBindingJob_Run(t *testing.T) {
	instanceID := uuid.NewRandom()
	bindingID := uuid.NewRandom()
	serviceInstance := &apb.ServiceInstance{
		ID: instanceID,
		Spec: &apb.Spec{
			ID: "test",
		},
	}
	cases := []struct {
		Name       string
		Binder     apb.Binder
		BindParams *apb.Parameters
		Validate   func(msg broker.JobMsg) error
	}{
		{
			Name: "expect a success msg with extracted credentials",
			Binder: func(si *apb.ServiceInstance, params *apb.Parameters) (string, *apb.ExtractedCredentials, error) {
				return "podName", &apb.ExtractedCredentials{Credentials: map[string]interface{}{
					"user": "test",
					"pass": "test",
				}}, nil
			},
			Validate: func(msg broker.JobMsg) error {
				if msg.State.State != apb.StateSucceeded {
					return fmt.Errorf("expected the state to be %v but got %v", apb.StateSucceeded, msg.State.State)
				}
				if msg.State.Method != apb.JobMethodBind {
					return fmt.Errorf("expected job method to be %v but it was %v", apb.JobMethodBind, msg.State.Method)
				}
				if msg.PodName == "" {
					return fmt.Errorf("expected the podName to be set but it was empty")
				}
				credentials := msg.ExtractedCredentials.Credentials

				if _, ok := credentials["user"]; !ok {
					return fmt.Errorf("expected a user key in the credentials but it was missing")
				}
				if _, ok := credentials["pass"]; !ok {
					return fmt.Errorf("expected a pass key in the credentials but it was missing")
				}
				return nil
			},
		},
		{
			Name: "expect failure state and generic error when unknown error type",
			Binder: func(si *apb.ServiceInstance, params *apb.Parameters) (string, *apb.ExtractedCredentials, error) {
				return "", nil, fmt.Errorf("should not see")
			},
			Validate: func(msg broker.JobMsg) error {
				if msg.State.State != apb.StateFailed {
					return fmt.Errorf("expected the Job to be in state %v but was in %v ", apb.StateFailed, msg.State.State)
				}
				if msg.State.Method != apb.JobMethodBind {
					return fmt.Errorf("expected job method to be %v but it was %v", apb.JobMethodBind, msg.State.Method)
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
			Binder: func(si *apb.ServiceInstance, params *apb.Parameters) (string, *apb.ExtractedCredentials, error) {
				return "", nil, apb.ErrorPodPullErr
			},
			Validate: func(msg broker.JobMsg) error {
				if msg.State.State != apb.StateFailed {
					return fmt.Errorf("expected the Job to be in state %v but was in %v ", apb.StateFailed, msg.State.State)
				}
				if msg.State.Method != apb.JobMethodBind {
					return fmt.Errorf("expected job method to be %v but it was %v", apb.JobMethodBind, msg.State.Method)
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
			bindingJob := broker.NewBindingJob(serviceInstance, bindingID, tc.BindParams, tc.Binder)
			receiver := make(chan broker.JobMsg)
			timedOut := false
			time.AfterFunc(1*time.Second, func() {
				close(receiver)
				timedOut = true
			})
			go bindingJob.Run("", receiver)

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
