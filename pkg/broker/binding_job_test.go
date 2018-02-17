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
		Validate   func(msgs []broker.JobMsg) error
	}{
		{
			Name: "expect a success msg with extracted credentials",
			Binder: func(si *apb.ServiceInstance, params *apb.Parameters, status chan<- apb.JobState) (string, *apb.ExtractedCredentials, error) {
				return "podName", &apb.ExtractedCredentials{Credentials: map[string]interface{}{
					"user": "test",
					"pass": "test",
				}}, nil
			},
			Validate: func(msgs []broker.JobMsg) error {
				if err := commonJobMsgValidation(apb.StateSucceeded, apb.JobMethodBind, msgs); err != nil {
					return err
				}
				lastMsg := msgs[len(msgs)-1]
				if lastMsg.PodName == "" {
					return fmt.Errorf("expected the podName to be set but it was empty")
				}
				credentials := lastMsg.ExtractedCredentials.Credentials

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
			Binder: func(si *apb.ServiceInstance, params *apb.Parameters, status chan<- apb.JobState) (string, *apb.ExtractedCredentials, error) {
				return "", nil, fmt.Errorf("should not see")
			},
			Validate: func(msgs []broker.JobMsg) error {
				if err := commonJobMsgValidation(apb.StateFailed, apb.JobMethodBind, msgs); err != nil {
					return err
				}
				lastMsg := msgs[len(msgs)-1]
				if lastMsg.State.Error == "" {
					return fmt.Errorf("expected an error in the job state but got none")
				}
				if lastMsg.State.Error == "should not see" {
					return fmt.Errorf("expected not to see the error msg %s it should have been replaced with a generic error ", lastMsg.State.Error)
				}
				return nil
			},
		},
		{
			Name: "expect failure state and full error when known error type",
			Binder: func(si *apb.ServiceInstance, params *apb.Parameters, status chan<- apb.JobState) (string, *apb.ExtractedCredentials, error) {
				return "", nil, apb.ErrorPodPullErr
			},
			Validate: func(msgs []broker.JobMsg) error {
				if err := commonJobMsgValidation(apb.StateFailed, apb.JobMethodBind, msgs); err != nil {
					return err
				}
				lastMsg := msgs[len(msgs)-1]
				if lastMsg.State.Error == "" {
					return fmt.Errorf("expected an error in the job state but got none")
				}
				if lastMsg.State.Error != apb.ErrorPodPullErr.Error() {
					return fmt.Errorf("expected to see the error msg %s but got %s ", apb.ErrorPodPullErr, lastMsg.State.Error)
				}
				return nil
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.Name, func(t *testing.T) {
			bindingJob := broker.NewBindingJob(serviceInstance, bindingID, tc.BindParams, tc.Binder)
			receiver := make(chan broker.JobMsg)
			time.AfterFunc(1*time.Second, func() {
				close(receiver)
			})
			go bindingJob.Run("", receiver)
			var msgs []broker.JobMsg
			for m := range receiver {
				msgs = append(msgs, m)
			}
			if err := tc.Validate(msgs); err != nil {
				t.Fatal("failed to validate the jobmsg ", err)
			}

		})
	}
}
