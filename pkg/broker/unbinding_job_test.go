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
		Validate      func(msgs []broker.JobMsg) error
	}{
		{
			Name: "expect a success msg",
			UnBinder: func(si *apb.ServiceInstance, params *apb.Parameters) error {
				return nil
			},
			Validate: func(msgs []broker.JobMsg) error {
				return commonJobMsgValidation(apb.StateSucceeded, apb.JobMethodUnbind, msgs)
			},
		},
		{
			Name:          "expect a success msg when skipping apb execution",
			SkipExecution: true,
			UnBinder: func(si *apb.ServiceInstance, params *apb.Parameters) error {
				return nil
			},
			Validate: func(msgs []broker.JobMsg) error {
				return commonJobMsgValidation(apb.StateSucceeded, apb.JobMethodUnbind, msgs)
			},
		},
		{
			Name: "expect failure state and generic error when unknown error type",
			UnBinder: func(si *apb.ServiceInstance, params *apb.Parameters) error {
				return fmt.Errorf("should not see")
			},
			Validate: func(msgs []broker.JobMsg) error {
				if err := commonJobMsgValidation(apb.StateFailed, apb.JobMethodUnbind, msgs); err != nil {
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
			UnBinder: func(si *apb.ServiceInstance, params *apb.Parameters) error {
				return apb.ErrorPodPullErr
			},
			Validate: func(msgs []broker.JobMsg) error {
				if err := commonJobMsgValidation(apb.StateFailed, apb.JobMethodUnbind, msgs); err != nil {
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
			unbindJob := broker.NewUnbindingJob(serviceInstance, bindingInst, tc.UnBindParams, tc.UnBinder, tc.SkipExecution)
			receiver := make(chan broker.JobMsg)
			time.AfterFunc(1*time.Second, func() {
				close(receiver)
			})
			go unbindJob.Run("", receiver)

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
