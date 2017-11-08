//
// Copyright (c) 2017 Red Hat, Inc.
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

package broker

import (
	"os"
	"testing"

	logging "github.com/op/go-logging"
	"github.com/openshift/ansible-service-broker/pkg/apb"
	ft "github.com/openshift/ansible-service-broker/pkg/fusortest"
	"github.com/pborman/uuid"
)

var log = logging.MustGetLogger("handler")

func init() {
	colorFormatter := logging.MustStringFormatter(
		"%{color}[%{time}] [%{level}] %{message}%{color:reset}",
	)
	backend := logging.NewLogBackend(os.Stdout, "", 1)
	backendFormatter := logging.NewBackendFormatter(backend, colorFormatter)
	logging.SetBackend(backend, backendFormatter)
}

func TestAddNameAndIDForSpecStripsTailingDash(t *testing.T) {
	spec1 := apb.Spec{FQName: "1234567890123456789012345678901234567890-"}
	spec2 := apb.Spec{FQName: "org/hello-world-apb"}
	spcs := []*apb.Spec{&spec1, &spec2}
	addNameAndIDForSpec(spcs, "h")
	ft.AssertEqual(t, "h-1234567890123456789012345678901234567890", spcs[0].FQName)
	ft.AssertEqual(t, "h-org-hello-world-apb", spcs[1].FQName)
}

func TestAddIdForPlan(t *testing.T) {
	plan1 := apb.Plan{Name: "default"}
	plans := []apb.Plan{plan1}
	addIDForPlan(plans, "dh-sns-apb")
	ft.AssertNotEqual(t, plans[0].ID, "", "plan id not updated")
}

// note this would likely move to a mocks package
type mockJobStateDAO struct {
	state  apb.JobState
	states []apb.JobState
	err    error
	calls  map[string]int
}

func (mjs *mockJobStateDAO) GetState(instanceUUID, operation string) (apb.JobState, error) {
	mjs.calls["GetState"]++
	return mjs.state, mjs.err
}
func (mjs *mockJobStateDAO) SetState(id string, state apb.JobState) error {
	mjs.calls["SetState"]++
	return mjs.err
}
func (mjs *mockJobStateDAO) GetSvcInstJobsByState(instanceID string, reqState apb.State) ([]apb.JobState, error) {
	mjs.calls["GetSvcInstJobsByState"]++
	return mjs.states, mjs.err
}

func TestAnsibleBroker_LastOperation(t *testing.T) {
	cases := []struct {
		Name                 string
		JobStateDAO          func() JobStateDAO
		ExpectError          bool
		InstanceID           uuid.UUID
		LastOperationRequest *LastOperationRequest
		Validate             func(t *testing.T, resp *LastOperationResponse)
		ExpectedCalls        map[string]int
	}{
		{
			Name:       "test get last_operation successful",
			InstanceID: uuid.NewUUID(),
			LastOperationRequest: &LastOperationRequest{
				ServiceID: "someserviceID",
				Operation: "provision",
				PlanID:    "default",
			},
			JobStateDAO: func() JobStateDAO {
				return &mockJobStateDAO{
					state: apb.JobState{State: apb.StateInProgress},
					calls: map[string]int{},
				}
			},
			Validate: func(t *testing.T, resp *LastOperationResponse) {
				if nil == resp {
					t.Fatal("expected a lastOperationResponse but got none")
				}
				if string(resp.State) != string(apb.StateInProgress) {
					t.Fatalf("expected lastOperationResponse to have state %s but got %s ", apb.StateInProgress, resp.State)
				}
			},
			ExpectedCalls: map[string]int{
				"GetState": 1,
			},
		},
	}

	for _, tc := range cases {
		logger, err := logging.GetLogger("test")
		if err != nil {
			t.Fatal("unexpected error when setting up broker", err)
		}
		jobStateDao := tc.JobStateDAO()
		broker, err := NewAnsibleBroker(nil, jobStateDao, logger, apb.ClusterConfig{}, nil, WorkEngine{}, Config{})
		if err != nil {
			t.Fatal("unexpected error when setting up broker", err)
		}
		t.Run(tc.Name, func(t *testing.T) {
			resp, err := broker.LastOperation(tc.InstanceID, tc.LastOperationRequest)
			if tc.ExpectError && err == nil {
				t.Fatal("Expected an error but got none")
			}
			if !tc.ExpectError && err != nil {
				t.Fatal("Did not expect an error but got one : ", err)
			}
			if tc.Validate != nil {
				tc.Validate(t, resp)
			}
			mockDao := jobStateDao.(*mockJobStateDAO)
			for k, v := range tc.ExpectedCalls {
				if mockDao.calls[k] != v {
					t.Fatalf("expected %s to be called %v times but it was called %v times", k, v, mockDao.calls[k])
				}
			}
		})
	}
}
