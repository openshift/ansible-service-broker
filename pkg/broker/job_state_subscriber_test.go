package broker_test

import (
	"errors"
	"fmt"
	"testing"

	apb "github.com/automationbroker/bundle-lib/bundle"
	"github.com/openshift/ansible-service-broker/pkg/broker"
	"github.com/openshift/ansible-service-broker/pkg/mock"
	"github.com/pborman/uuid"
)

type mockCredentialsDelete struct {
	called bool
	err    error
	assert func(string) error
}

func (mc mockCredentialsDelete) delteCreds(id string) error {
	mc.called = true
	if nil != mc.assert {
		if err := mc.assert(id); err != nil {
			return err
		}
	}
	return mc.err
}

func TestJobStateSubscriber(t *testing.T) {
	//instanceID := "id"
	uID := uuid.NewUUID()
	jobStates := func(state apb.State) []broker.JobMsg {
		return []broker.JobMsg{broker.JobMsg{
			State: apb.JobState{
				State:  state,
				Method: apb.JobMethodProvision,
			},
		}, broker.JobMsg{
			State: apb.JobState{
				State:  state,
				Method: apb.JobMethodUpdate,
			},
		}, broker.JobMsg{
			State: apb.JobState{
				State:  state,
				Method: apb.JobMethodDeprovision,
			},
		}, broker.JobMsg{
			State: apb.JobState{
				State:  state,
				Method: apb.JobMethodBind,
			},
		}, broker.JobMsg{
			State: apb.JobState{
				State:  state,
				Method: apb.JobMethodUnbind,
			},
		}}
	}

	allStates := func() []broker.JobMsg {
		js := jobStates(apb.StateFailed)
		js = append(js, jobStates(apb.StateInProgress)...)
		js = append(js, jobStates(apb.StateSucceeded)...)
		js = append(js, jobStates(apb.StateNotYetStarted)...)
		return js
	}

	cases := []struct {
		Name   string
		JobMsg []broker.JobMsg
		DAO    func() (*mock.SubscriberDAO, map[string]int)
	}{
		{
			Name:   "job state subscriber should always set state for job msg",
			JobMsg: allStates(),
			DAO: func() (*mock.SubscriberDAO, map[string]int) {
				dao := mock.NewSubscriberDAO()
				dao.Object["GetServiceInstance"] = &apb.ServiceInstance{}
				dao.Object["GetBindInstance"] = &apb.BindInstance{ID: uID}
				expectedCalls := map[string]int{
					"SetState": 1,
				}
				return dao, expectedCalls
			},
		},
		{
			Name: "after successful unbind binding instance should be removed",
			JobMsg: []broker.JobMsg{
				{
					BindingUUID: uID.String(),
					State: apb.JobState{
						State:  apb.StateSucceeded,
						Method: apb.JobMethodUnbind,
					}},
			},
			DAO: func() (*mock.SubscriberDAO, map[string]int) {
				dao := mock.NewSubscriberDAO()
				dao.Object["GetServiceInstance"] = &apb.ServiceInstance{ID: uID}
				dao.Object["GetBindInstance"] = &apb.BindInstance{ID: uID}
				dao.AssertOn["DeleteBinding"] = func(args ...interface{}) error {

					bi := args[0].(apb.BindInstance)
					si := args[1].(apb.ServiceInstance)
					if si.ID.String() != uID.String() || bi.ID.String() != uID.String() {
						return fmt.Errorf("expected the service instance to have the id %s ", uID.String())
					}
					return nil
				}
				expectedCalls := map[string]int{
					"SetState":      1,
					"DeleteBinding": 1,
				}
				return dao, expectedCalls
			},
		},
		{
			Name: "after successful deprovision the service instance should be removed",
			JobMsg: []broker.JobMsg{{
				State: apb.JobState{
					State:  apb.StateSucceeded,
					Method: apb.JobMethodDeprovision,
				},
			}},
			DAO: func() (*mock.SubscriberDAO, map[string]int) {
				dao := mock.NewSubscriberDAO()
				dao.AssertOn["SetState"] = func(args ...interface{}) error {
					state := args[1].(apb.JobState)
					if state.Method != apb.JobMethodDeprovision {
						return fmt.Errorf("expected to have a provision job state")
					}
					if state.State != apb.StateSucceeded {
						return fmt.Errorf("expected the job state to be %v but got %v", apb.StateSucceeded, state.State)
					}
					return nil
				}
				expectedCalls := map[string]int{
					"SetState":              1,
					"DeleteServiceInstance": 1,
				}
				return dao, expectedCalls
			},
		},
		{
			Name: "after successful deprovision if there is an error cleaning up the job state should be set to failed",
			JobMsg: []broker.JobMsg{{
				State: apb.JobState{
					State:  apb.StateSucceeded,
					Method: apb.JobMethodDeprovision,
				},
			}},
			DAO: func() (*mock.SubscriberDAO, map[string]int) {
				dao := mock.NewSubscriberDAO()
				dao.Errs["DeleteServiceInstance"] = errors.New("failed")
				calls := 0
				dao.AssertOn["SetState"] = func(args ...interface{}) error {
					calls++
					state := args[1].(apb.JobState)
					if state.Method != apb.JobMethodDeprovision {
						return fmt.Errorf("expected to have a provision job state")
					}
					if calls == 1 && state.State != apb.StateSucceeded {
						return fmt.Errorf("expected the job state to be %v but got %v", apb.StateSucceeded, state.State)
					} else if calls == 2 && state.State != apb.StateFailed {
						return fmt.Errorf("expected the job state to be %v but got %v", apb.StateFailed, state.State)
					}

					return nil
				}
				expectedCalls := map[string]int{
					"SetState":              2,
					"DeleteServiceInstance": 1,
				}
				return dao, expectedCalls
			},
		},
		{
			Name: "after successful unbind if there is an error cleaning up the job state should be set to failed",
			JobMsg: []broker.JobMsg{{
				State: apb.JobState{
					State:  apb.StateSucceeded,
					Method: apb.JobMethodUnbind,
				},
			}},
			DAO: func() (*mock.SubscriberDAO, map[string]int) {
				dao := mock.NewSubscriberDAO()
				dao.Object["GetServiceInstance"] = &apb.ServiceInstance{}
				dao.Object["GetBindInstance"] = &apb.BindInstance{ID: uID}
				dao.Errs["DeleteBinding"] = errors.New("failed")
				calls := 0
				dao.AssertOn["SetState"] = func(args ...interface{}) error {
					calls++
					state := args[1].(apb.JobState)
					if state.Method != apb.JobMethodUnbind {
						return fmt.Errorf("expected to have a unbind job state")
					}
					if calls == 1 && state.State != apb.StateSucceeded {
						return fmt.Errorf("expected the job state to be %v but got %v", apb.StateSucceeded, state.State)
					} else if calls == 2 && state.State != apb.StateFailed {
						return fmt.Errorf("expected the job state to be %v but got %v", apb.StateFailed, state.State)
					}

					return nil
				}
				expectedCalls := map[string]int{
					"SetState":      2,
					"DeleteBinding": 1,
				}
				return dao, expectedCalls
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.Name, func(t *testing.T) {
			for _, jmsg := range tc.JobMsg {
				dao, calls := tc.DAO()
				sub := broker.NewJobStateSubscriber(dao)
				sub.Notify(jmsg)
				if len(dao.AssertErrors()) != 0 {
					t.Fatal("unexpected error during data assertions ", dao.AssertErrors())
				}
				if err := dao.CheckCalls(calls); err != nil {
					t.Fatal("unexpected error checking calls ", err)
				}
			}
		})
	}
}
