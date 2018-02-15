package broker_test

import (
	"testing"

	"fmt"

	"time"

	"sync"

	"github.com/openshift/ansible-service-broker/pkg/apb"
	"github.com/openshift/ansible-service-broker/pkg/broker"
)

func TestProvisionWorkSubscriber_Subscribe(t *testing.T) {
	cases := []struct {
		Name   string
		JobMsg broker.JobMsg
		DAO    func() (*mockProvisionSubscriberDAO, map[string]int)
	}{
		{
			Name: "should set state and credentials when job is successful",
			JobMsg: broker.JobMsg{
				State: apb.JobState{
					State:  apb.StateSucceeded,
					Method: apb.JobMethodProvision,
				},
				ExtractedCredentials: apb.ExtractedCredentials{
					Credentials: map[string]interface{}{"user": "test", "pass": "test"},
				},
			},
			DAO: func() (*mockProvisionSubscriberDAO, map[string]int) {
				dao := newProvisionSubscriberDAO()
				dao.assertOn["SetExtractedCredentials"] = func(args ...interface{}) error {
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
				dao.assertOn["SetState"] = func(args ...interface{}) error {
					state := args[1].(apb.JobState)
					if state.Method != apb.JobMethodProvision {
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
					Method: apb.JobMethodProvision,
				},
			},
			DAO: func() (*mockProvisionSubscriberDAO, map[string]int) {
				dao := newProvisionSubscriberDAO()
				dao.assertOn["SetState"] = func(args ...interface{}) error {
					state := args[1].(apb.JobState)
					if state.Method != apb.JobMethodProvision {
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
					Method: apb.JobMethodProvision,
				},
			},
			DAO: func() (*mockProvisionSubscriberDAO, map[string]int) {
				dao := newProvisionSubscriberDAO()
				dao.assertOn["SetState"] = func(args ...interface{}) error {
					state := args[1].(apb.JobState)
					if state.Method != apb.JobMethodProvision {
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
			sub := broker.NewProvisionWorkSubscriber(dao)
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

type mockProvisionSubscriberDAO struct {
	calls     map[string]int
	err       error
	assertErr []error
	assertOn  map[string]func(...interface{}) error
}

func (mp *mockProvisionSubscriberDAO) SetExtractedCredentials(id string, extCreds *apb.ExtractedCredentials) error {
	assert := mp.assertOn["SetExtractedCredentials"]
	if nil != assert {
		if err := assert(id, extCreds); err != nil {
			mp.assertErr = append(mp.assertErr, err)
			return err
		}
	}
	mp.calls["SetExtractedCredentials"]++
	return mp.err

}
func (mp *mockProvisionSubscriberDAO) SetState(id string, state apb.JobState) (string, error) {
	assert := mp.assertOn["SetState"]
	if nil != assert {
		if err := assert(id, state); err != nil {
			mp.assertErr = append(mp.assertErr, err)
			return "", err
		}
	}
	mp.calls["SetState"]++
	return "", mp.err

}

func (mp *mockProvisionSubscriberDAO) CheckCalls(calls map[string]int) error {
	for k, v := range calls {
		if mp.calls[k] != v {
			return fmt.Errorf("expected %d calls to %s but got %d ", v, k, mp.calls[k])
		}
	}
	return nil
}

func (mp *mockProvisionSubscriberDAO) AssertErrors() []error {
	return mp.assertErr
}

func newProvisionSubscriberDAO() *mockProvisionSubscriberDAO {
	return &mockProvisionSubscriberDAO{
		calls:    map[string]int{},
		assertOn: map[string]func(...interface{}) error{},
	}
}
