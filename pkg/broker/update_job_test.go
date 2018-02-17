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

func TestUpdateJob_Run(t *testing.T) {
	var uid = uuid.NewRandom()
	var serviceInstance = &apb.ServiceInstance{
		ID: uid,
		Spec: &apb.Spec{
			ID: "test",
		},
	}
	var commonMsgValidation = func(expectedFinalState apb.State, msgs []broker.JobMsg) error {
		if len(msgs) < 2 {
			return fmt.Errorf("expected 2 msgs but only got %v", len(msgs))
		}
		for i, msg := range msgs {
			if msg.State.Method != apb.JobMethodUpdate {
				return fmt.Errorf("expected job msg method to be %v but it was %v", apb.JobMethodUpdate, msg.State.Method)
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
	cases := []struct {
		Name     string
		Update   apb.Updater
		Validate func(msg []broker.JobMsg) error
	}{
		{
			Name: "expect a success msg with extracted credentials when no error occurs",
			Update: func(si *apb.ServiceInstance, statusUpdate chan<- apb.JobState) (string, *apb.ExtractedCredentials, error) {
				return "podName", &apb.ExtractedCredentials{Credentials: map[string]interface{}{
					"user": "test",
					"pass": "test",
				}}, nil
			},
			Validate: func(msgs []broker.JobMsg) error {
				return commonMsgValidation(apb.StateSucceeded, msgs)
			},
		},
		{
			Name: "expect failure state and generic error when unknown error type",
			Update: func(si *apb.ServiceInstance, statusUpdates chan<- apb.JobState) (string, *apb.ExtractedCredentials, error) {
				return "", nil, fmt.Errorf("should not see")
			},
			Validate: func(msgs []broker.JobMsg) error {
				if err := commonMsgValidation(apb.StateFailed, msgs); err != nil {
					return err
				}
				finalMsg := msgs[len(msgs)-1]
				if finalMsg.State.Error == "" {
					return fmt.Errorf("expected an error in the job state but got none")
				}
				if finalMsg.State.Description == "should not see" {
					return fmt.Errorf("expected not to see the error msg %s it should have been replaced with a generic error ", finalMsg.State.Error)
				}
				return nil
			},
		},
		{
			Name: "expect failure state and full error when known error type",
			Update: func(si *apb.ServiceInstance, statusUpdates chan<- apb.JobState) (string, *apb.ExtractedCredentials, error) {
				return "", nil, apb.ErrorPodPullErr
			},
			Validate: func(msgs []broker.JobMsg) error {
				if err := commonMsgValidation(apb.StateFailed, msgs); err != nil {
					return err
				}
				finalMsg := msgs[len(msgs)-1]
				if finalMsg.State.Error == "" {
					return fmt.Errorf("expected an error in the job state but got none")
				}
				if finalMsg.State.Error != apb.ErrorPodPullErr.Error() {
					return fmt.Errorf("expected to see the error failed %s but got %s ", apb.ErrorPodPullErr, finalMsg.State.Error)
				}
				return nil
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.Name, func(t *testing.T) {
			update := broker.NewUpdateJob(serviceInstance, tc.Update)
			receiver := make(chan broker.JobMsg)
			// give some time to allow msgs to be sent as we are not actually provisioning this should be plenty
			time.AfterFunc(200*time.Millisecond, func() {
				close(receiver)
			})
			go update.Run("", receiver)
			var msgs []broker.JobMsg
			for msg := range receiver {
				msgs = append(msgs, msg)
			}
			if err := tc.Validate(msgs); err != nil {
				t.Fatal("failed to validate the jobmsg ", err)
			}

		})
	}
}
