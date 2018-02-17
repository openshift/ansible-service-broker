//
// Copyright (c) 2018 Red Hat, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//    http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//
// Red Hat trademarks are not licensed under Apache License, Version 2.
// No permission is granted to use or replicate Red Hat trademarks that
// are incorporated in this software or its documentation.
//

package broker_test

import (
	"testing"

	"time"

	"fmt"

	"github.com/openshift/ansible-service-broker/pkg/apb"
	"github.com/openshift/ansible-service-broker/pkg/broker"
	"github.com/pborman/uuid"
)

func commonJobMsgValidation(expectedFinalState apb.State, expectedMethod apb.JobMethod, msgs []broker.JobMsg) error {
	if len(msgs) < 2 {
		return fmt.Errorf("expected 2 msgs but only got %v", len(msgs))
	}
	for i, msg := range msgs {
		if msg.State.Method != expectedMethod {
			return fmt.Errorf("expected job msg method to be %v but it was %v", expectedMethod, msg.State.Method)
		}
		if i == 0 && msg.State.State != apb.StateInProgress {
			return fmt.Errorf("expected job msg state to be %v but it was %v", apb.StateInProgress, msg.State.State)
		}

		if i == len(msgs)-1 && msg.State.State != expectedFinalState {
			return fmt.Errorf("expected job msg state to be %v but it was %v", expectedFinalState, msg.State.State)
		}
	}
	return nil
}

func TestDeprovisionJob_Run(t *testing.T) {
	var uid = uuid.NewRandom()
	var serviceInstance = &apb.ServiceInstance{
		ID: uid,
		Spec: &apb.Spec{
			ID: "test",
		},
	}

	cases := []struct {
		Name             string
		Deprovision      apb.Deprovisioner
		SkipAPBExecution bool
		Validate         func(msgs []broker.JobMsg) error
	}{
		{
			Name: "expect a success msg when no error occurs",
			Deprovision: func(si *apb.ServiceInstance, statusUpdate chan<- apb.JobState) (string, error) {
				return "somepod", nil
			},
			Validate: func(msgs []broker.JobMsg) error {
				return commonJobMsgValidation(apb.StateSucceeded, apb.JobMethodDeprovision, msgs)
			},
		},
		{
			Name: "expect a success msg when skip apb execution",
			Deprovision: func(si *apb.ServiceInstance, statusUpdate chan<- apb.JobState) (string, error) {
				return "", nil
			},
			SkipAPBExecution: true,
			Validate: func(msgs []broker.JobMsg) error {
				return commonJobMsgValidation(apb.StateSucceeded, apb.JobMethodDeprovision, msgs)
			},
		},
		{
			Name: "expect a generic failure msg when an unknown error occurs",
			Deprovision: func(si *apb.ServiceInstance, statusUpdate chan<- apb.JobState) (string, error) {
				return "", fmt.Errorf("some error")
			},
			Validate: func(msgs []broker.JobMsg) error {
				if err := commonJobMsgValidation(apb.StateFailed, apb.JobMethodDeprovision, msgs); err != nil {
					return err
				}
				last := msgs[len(msgs)-1]
				if last.State.Description == "some error" {
					return fmt.Errorf("expected a generic error but got %s", last.State.Description)
				}
				return nil
			},
		},
		{
			Name: "expect an in progress msg with last operation description",
			Deprovision: func(si *apb.ServiceInstance, statusUpdate chan<- apb.JobState) (string, error) {
				statusUpdate <- apb.JobState{State: apb.StateInProgress, Description: "doing something", Method: apb.JobMethodDeprovision, Podname: "somepod"}
				return "", nil
			},
			Validate: func(msgs []broker.JobMsg) error {
				if err := commonJobMsgValidation(apb.StateSucceeded, apb.JobMethodDeprovision, msgs); err != nil {
					return err
				}
				foundMessage := false
				for _, msg := range msgs {
					if msg.State.State == apb.StateInProgress && msg.State.Description == "doing something" {
						foundMessage = true
					}
				}
				if !foundMessage {
					return fmt.Errorf("expected to find a last operation description stating: doing something but found none")
				}
				return nil
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.Name, func(t *testing.T) {
			deprovisionJob := broker.NewDeprovisionJob(serviceInstance, tc.SkipAPBExecution, tc.Deprovision)
			receiver := make(chan broker.JobMsg, 2)
			// give some time to allow msgs to be sent as we are not actually provisioning this should be plenty
			time.AfterFunc(200*time.Millisecond, func() {
				close(receiver)
			})

			go deprovisionJob.Run("", receiver)
			var msgs []broker.JobMsg
			for msg := range receiver {
				msgs = append(msgs, msg)
			}
			if err := tc.Validate(msgs); err != nil {
				t.Fatal("failed to validate the jobmsgs ", err)
			}
		})
	}
}
