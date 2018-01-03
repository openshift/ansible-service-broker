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

	"fmt"

	"time"

	"sync"

	"github.com/openshift/ansible-service-broker/pkg/apb"
	"github.com/openshift/ansible-service-broker/pkg/broker"
	"github.com/openshift/ansible-service-broker/pkg/mock"
)

func TestUpdateWorkSubscriber_Subscribe(t *testing.T) {
	cases := []struct {
		Name   string
		JobMsg broker.JobMsg
		DAO    func() (*mock.SubscriberDAO, map[string]int)
	}{
		{
			Name: "should set state and credentials when job is successful",
			JobMsg: broker.JobMsg{
				State: apb.JobState{
					State:  apb.StateSucceeded,
					Method: apb.JobMethodUpdate,
				},
				ExtractedCredentials: apb.ExtractedCredentials{
					Credentials: map[string]interface{}{"user": "test", "pass": "test"},
				},
			},
			DAO: func() (*mock.SubscriberDAO, map[string]int) {
				dao := mock.NewSubscriberDAO()
				dao.AssertOn["SetExtractedCredentials"] = func(args ...interface{}) error {
					cred := args[1]
					if nil == cred {
						return fmt.Errorf("expected credentials to passed")
					}
					creds := cred.(*apb.ExtractedCredentials)
					if _, ok := creds.Credentials["user"]; !ok {
						return fmt.Errorf("expected there to be a user field in the credentials")
					}
					if _, ok := creds.Credentials["pass"]; !ok {
						return fmt.Errorf("expected there to be a pass field in the credentials")
					}
					return nil
				}
				dao.AssertOn["SetState"] = func(args ...interface{}) error {
					state := args[1].(apb.JobState)
					if state.Method != apb.JobMethodUpdate {
						return fmt.Errorf("expected to have a provision job state")
					}
					if state.State != apb.StateSucceeded {
						return fmt.Errorf("expected the job state to be %v but got %v", apb.StateSucceeded, state.State)
					}
					return nil

				}
				expectedCalls := map[string]int{
					"SetExtractedCredentials": 1,
					"SetState":                1,
				}
				return dao, expectedCalls
			},
		},
		{
			Name: "should set state but not credentials when failed",
			JobMsg: broker.JobMsg{
				State: apb.JobState{
					State:  apb.StateFailed,
					Method: apb.JobMethodUpdate,
				},
			},
			DAO: func() (*mock.SubscriberDAO, map[string]int) {
				dao := mock.NewSubscriberDAO()
				dao.AssertOn["SetState"] = func(args ...interface{}) error {
					state := args[1].(apb.JobState)
					if state.Method != apb.JobMethodUpdate {
						fmt.Println(state)
						return fmt.Errorf("expected to have a provision job state but was %v", state.Method)
					}
					if state.State != apb.StateFailed {
						return fmt.Errorf("expected the job state to be %v but got %v", apb.StateSucceeded, state.State)
					}
					return nil
				}
				expectedCalls := map[string]int{
					"SetExtractedCredentials": 0,
					"SetState":                1,
				}
				return dao, expectedCalls
			},
		},
		{
			Name: "should set state but not credentials when in progress",
			JobMsg: broker.JobMsg{
				State: apb.JobState{
					State:  apb.StateInProgress,
					Method: apb.JobMethodUpdate,
				},
			},
			DAO: func() (*mock.SubscriberDAO, map[string]int) {
				dao := mock.NewSubscriberDAO()
				dao.AssertOn["SetState"] = func(args ...interface{}) error {
					state := args[1].(apb.JobState)
					if state.Method != apb.JobMethodUpdate {
						return fmt.Errorf("expected to have a provision job state")
					}
					if state.State != apb.StateInProgress {
						return fmt.Errorf("expected the job state to be %v but got %v", apb.StateSucceeded, state.State)
					}
					return nil
				}
				expectedCalls := map[string]int{
					"SetExtractedCredentials": 0,
					"SetState":                1,
				}
				return dao, expectedCalls
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.Name, func(t *testing.T) {
			dao, expectedCalls := tc.DAO()
			sub := broker.NewUpdateWorkSubscriber(dao)
			wait := sync.WaitGroup{}
			wait.Add(1)
			// this is a bit gross but hard to test the subscribe method as it has a constant for loop
			// so we give it 100ms to process the message and then do our checks
			sender := make(chan broker.JobMsg)
			sub.Subscribe(sender)
			time.AfterFunc(100*time.Millisecond, func() {
				close(sender)
				wait.Done()
			})
			sender <- tc.JobMsg
			wait.Wait()
			if len(dao.AssertErrors()) != 0 {
				t.Fatal("unexpected error during data assertions ", dao.AssertErrors())
			}
			if err := dao.CheckCalls(expectedCalls); err != nil {
				t.Fatal("unexpected error checking calls ", err)
			}
		})
	}
}
