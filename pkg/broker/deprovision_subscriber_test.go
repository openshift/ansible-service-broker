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

func TestDeprovisionWorkSubscriber_Subscribe(t *testing.T) {
	instanceID := "id"
	cases := []struct {
		Name   string
		JobMsg broker.JobMsg
		DAO    func() (*mock.SubscriberDAO, map[string]int)
	}{
		{
			Name: "should set state and credentials when job is successful",
			JobMsg: broker.JobMsg{
				InstanceUUID: instanceID,
				State: apb.JobState{
					State:  apb.StateSucceeded,
					Method: apb.JobMethodProvision,
				},
				ExtractedCredentials: apb.ExtractedCredentials{
					Credentials: map[string]interface{}{"user": "test", "pass": "test"},
				},
			},
			DAO: func() (*mock.SubscriberDAO, map[string]int) {
				dao := mock.NewSubscriberDAO()
				dao.AssertOn["DeleteExtractedCredentials"] = func(args ...interface{}) error {
					id := args[0].(string)
					if id != "id" {
						return fmt.Errorf("epected the id to be : id but was %s", id)
					}
					return nil
				}
				dao.AssertOn["DeleteServiceInstance"] = func(args ...interface{}) error {
					id := args[0].(string)
					if id != "id" {
						return fmt.Errorf("epected the id to be : id but was %s", id)
					}
					return nil
				}
				dao.AssertOn["SetState"] = func(i ...interface{}) error {
					state := i[1].(apb.JobState)
					if state.State != apb.StateSucceeded {
						return fmt.Errorf("expected to the state to be %v but was %v", apb.StateSucceeded, state.State)
					}
					return nil
				}
				expectedCalls := map[string]int{
					"DeleteExtractedCredentials": 1,
					"DeleteServiceInstance":      1,
					"SetState":                   1,
				}
				return dao, expectedCalls
			},
		},
		{
			Name: "should set state to failed if clean up failed to delete serviceInstance",
			JobMsg: broker.JobMsg{
				InstanceUUID: instanceID,
				State: apb.JobState{
					State:  apb.StateSucceeded,
					Method: apb.JobMethodProvision,
				},
				ExtractedCredentials: apb.ExtractedCredentials{
					Credentials: map[string]interface{}{"user": "test", "pass": "test"},
				},
			},
			DAO: func() (*mock.SubscriberDAO, map[string]int) {
				dao := mock.NewSubscriberDAO()
				dao.AssertOn["DeleteExtractedCredentials"] = func(args ...interface{}) error {
					id := args[0].(string)
					if id != "id" {
						return fmt.Errorf("epected the id to be : id but was %s", id)
					}
					return nil
				}
				dao.AssertOn["DeleteServiceInstance"] = func(args ...interface{}) error {
					id := args[0].(string)
					if id != "id" {
						return fmt.Errorf("epected the id to be : id but was %s", id)
					}
					return nil
				}
				var states []apb.JobState
				dao.AssertOn["SetState"] = func(i ...interface{}) error {
					state := i[1].(apb.JobState)
					states = append(states, state)
					if states[0].State != apb.StateSucceeded {
						return fmt.Errorf("expected to the state to be %v but was %v", apb.StateSucceeded, states[0].State)
					}
					if len(states) == 2 && states[1].State != apb.StateFailed {
						return fmt.Errorf("expected to the state to be %v but was %v", apb.StateFailed, states[1].State)
					}
					return nil
				}
				dao.Errs["DeleteServiceInstance"] = fmt.Errorf("not there")
				expectedCalls := map[string]int{
					"DeleteExtractedCredentials": 1,
					"DeleteServiceInstance":      1,
					"SetState":                   2,
				}
				return dao, expectedCalls
			},
		},
		{
			Name: "should set state to failed if clean up failed to delete extractedCredentials",
			JobMsg: broker.JobMsg{
				InstanceUUID: instanceID,
				State: apb.JobState{
					State:  apb.StateSucceeded,
					Method: apb.JobMethodProvision,
				},
				ExtractedCredentials: apb.ExtractedCredentials{
					Credentials: map[string]interface{}{"user": "test", "pass": "test"},
				},
			},
			DAO: func() (*mock.SubscriberDAO, map[string]int) {
				dao := mock.NewSubscriberDAO()
				dao.AssertOn["DeleteExtractedCredentials"] = func(args ...interface{}) error {
					id := args[0].(string)
					if id != "id" {
						return fmt.Errorf("epected the id to be : id but was %s", id)
					}
					return nil
				}
				dao.AssertOn["DeleteServiceInstance"] = func(args ...interface{}) error {

					return fmt.Errorf("shouldn't have got to DeleteServiceInstance")

				}
				var states []apb.JobState
				dao.AssertOn["SetState"] = func(i ...interface{}) error {
					state := i[1].(apb.JobState)
					states = append(states, state)
					if states[0].State != apb.StateSucceeded {
						return fmt.Errorf("expected to the state to be %v but was %v", apb.StateSucceeded, states[0].State)
					}
					if len(states) == 2 && states[1].State != apb.StateFailed {
						return fmt.Errorf("expected to the state to be %v but was %v", apb.StateFailed, states[1].State)
					}
					return nil
				}
				dao.Errs["DeleteExtractedCredentials"] = fmt.Errorf("not there")
				expectedCalls := map[string]int{
					"DeleteExtractedCredentials": 1,
					"DeleteServiceInstance":      0,
					"SetState":                   2,
				}
				return dao, expectedCalls
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.Name, func(t *testing.T) {
			dao, expectedCalls := tc.DAO()
			sub := broker.NewDeprovisionWorkSubscriber(dao)
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
