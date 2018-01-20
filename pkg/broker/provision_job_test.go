package broker_test

import (
	"testing"

	"time"

	"fmt"

	"github.com/openshift/ansible-service-broker/pkg/apb"
	"github.com/openshift/ansible-service-broker/pkg/broker"
	"github.com/pborman/uuid"
)

func TestProvisionJob_Run(t *testing.T) {
	var uid = uuid.NewRandom()
	var serviceInstance = &apb.ServiceInstance{
		ID: uid,
		Spec: &apb.Spec{
			ID: "test",
		},
	}
	cases := []struct {
		Name      string
		Provision apb.Provisioner
		Validate  func(msgs []broker.JobMsg) error
	}{
		{
			Name: "expect a success msg with extracted credentials",
			Provision: func(si *apb.ServiceInstance) (string, *apb.ExtractedCredentials, error) {
				return "podName", &apb.ExtractedCredentials{Credentials: map[string]interface{}{
					"user": "test",
					"pass": "test",
				}}, nil
			},
			Validate: func(msgs []broker.JobMsg) error {
				if err := commonJobMsgValidation(apb.StateSucceeded, apb.JobMethodProvision, msgs); err != nil {
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
			Provision: func(si *apb.ServiceInstance) (string, *apb.ExtractedCredentials, error) {
				return "", nil, fmt.Errorf("should not see")
			},
			Validate: func(msgs []broker.JobMsg) error {
				if err := commonJobMsgValidation(apb.StateFailed, apb.JobMethodProvision, msgs); err != nil {
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
			Provision: func(si *apb.ServiceInstance) (string, *apb.ExtractedCredentials, error) {
				return "", nil, apb.ErrorPodPullErr
			},
			Validate: func(msgs []broker.JobMsg) error {
				if err := commonJobMsgValidation(apb.StateFailed, apb.JobMethodProvision, msgs); err != nil {
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
			provJob := broker.NewProvisionJob(serviceInstance, tc.Provision)
			receiver := make(chan broker.JobMsg)
			time.AfterFunc(1*time.Second, func() {
				close(receiver)
			})
			go provJob.Run("", receiver)
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
