package broker

import (
	"fmt"
	"testing"
	"time"

	"github.com/automationbroker/bundle-lib/bundle"
	"github.com/automationbroker/bundle-lib/runtime"
	"github.com/stretchr/testify/assert"
)

func TestApbJobRun(t *testing.T) {
	timeoutSecs := 5
	serviceInstanceID := "16235516-9e5e-4c68-a541-33bda63413ee"
	specID := "a7fc708e-52cf-427e-88b2-2b750b607a27"
	bindingID := "20c6ec16-c5bd-433a-815c-63cf0e2d2c9d"
	token := "4ac9529c-6a01-4daf-9e10-8b557e4885ae"
	podName := "apb-8f9268c2-1aaa-48f1-918d-eae920986c9f"
	extCreds := &bundle.ExtractedCredentials{
		Credentials: map[string]interface{}{"foo": "bar", "baz": "duder"},
	}

	cases := []struct {
		name                            string
		testJob                         *apbJob
		metricsJobStartHookTriggered    bool
		metricsJobFinishedHookTriggered bool
		validate                        func([]JobMsg) error
		expectedMsgCount                int
	}{
		{
			name: "should stream expected jobMsgs when descriptions are reported",
			testJob: &apbJob{
				serviceInstanceID: serviceInstanceID,
				specID:            specID,
				method:            bundle.JobMethodProvision,
				skipExecution:     false,
				executor: func() bundle.Executor {
					e := &bundle.MockExecutor{}
					e.On("PodName").Return(podName)
					e.On("LastStatus").Return(bundle.StatusMessage{
						State:       bundle.StateSucceeded,
						Description: "action finished with success",
					})
					e.On("ExtractedCredentials").Return(nil)
					e.On("DashboardURL").Return("http://foo.example.com")
					return e
				}(),
				run: func(exec bundle.Executor) <-chan bundle.StatusMessage {
					statusChan := make(chan bundle.StatusMessage)
					go func() {
						// Initial message sent from executor.actionStarted
						statusChan <- bundle.StatusMessage{
							State:       bundle.StateInProgress,
							Description: "action started",
						}
						// Two updateDescription calls
						statusChan <- bundle.StatusMessage{
							State:       bundle.StateInProgress,
							Description: "lastOp0",
						}
						statusChan <- bundle.StatusMessage{
							State:       bundle.StateInProgress,
							Description: "lastOp1",
						}
						// Final status sent by executor.actionFinishedWithSuccess
						statusChan <- bundle.StatusMessage{
							State:       bundle.StateSucceeded,
							Description: "action finished with success",
						}
						close(statusChan)
					}()
					return statusChan
				},
			},
			expectedMsgCount: 4,
			validate: func(messages []JobMsg) error {
				if len(messages) != 4 {
					return fmt.Errorf("expected 4 job messages")
				}
				first := messages[0]
				if first.State.State != bundle.StateInProgress {
					return fmt.Errorf("unexpected first message contents")
				}

				second := messages[1]
				if second.State.State != bundle.StateInProgress ||
					second.State.Description != "lastOp0" {
					return fmt.Errorf("unexpected second message contents")
				}

				third := messages[2]
				if third.State.State != bundle.StateInProgress ||
					third.State.Description != "lastOp1" {
					return fmt.Errorf("unexpected third message contents")
				}

				fourth := messages[3]
				if fourth.State.State != bundle.StateSucceeded {
					return fmt.Errorf("unexpected fourth message contents")
				}

				return nil
			},
		},
		{
			name: "should skip execution when requested",
			testJob: &apbJob{
				serviceInstanceID: serviceInstanceID,
				specID:            specID,
				method:            bundle.JobMethodProvision,
				skipExecution:     true,
				executor:          &bundle.MockExecutor{},
				run: func(exec bundle.Executor) <-chan bundle.StatusMessage {
					statusChan := make(chan bundle.StatusMessage)
					go func() {
						// Initial message sent from executor.actionStarted
						statusChan <- bundle.StatusMessage{
							State:       bundle.StateInProgress,
							Description: "action started",
						}
						// Two updateDescription calls
						statusChan <- bundle.StatusMessage{
							State:       bundle.StateInProgress,
							Description: "lastOp0",
						}
						statusChan <- bundle.StatusMessage{
							State:       bundle.StateInProgress,
							Description: "lastOp1",
						}
						// Final status sent by executor.actionFinishedWithSuccess
						statusChan <- bundle.StatusMessage{
							State:       bundle.StateSucceeded,
							Description: "action finished with success",
						}
						close(statusChan)
					}()
					return statusChan
				},
			},
			expectedMsgCount: 1,
			validate: func(messages []JobMsg) error {
				if len(messages) != 1 {
					return fmt.Errorf("expected 1 job messages")
				}
				// Since the apb is never executed, we're just expecting a single
				// success JobMsg
				if messages[0].State.State != bundle.StateSucceeded {
					return fmt.Errorf("unexpected second message contents")
				}
				return nil
			},
		},
		{
			name: "should send failed jobMsg when error reported on executor",
			testJob: &apbJob{
				serviceInstanceID: serviceInstanceID,
				specID:            specID,
				method:            bundle.JobMethodProvision,
				skipExecution:     false,
				executor: func() bundle.Executor {
					e := &bundle.MockExecutor{}
					e.On("PodName").Return(podName)
					e.On("LastStatus").Return(bundle.StatusMessage{
						State:       bundle.StateFailed,
						Error:       fmt.Errorf("Everything is on fire"),
						Description: "action finished with error",
					})
					e.On("ExtractedCredentials").Return(nil)
					return e
				}(),
				run: func(exec bundle.Executor) <-chan bundle.StatusMessage {
					statusChan := make(chan bundle.StatusMessage)
					go func() {
						// Initial message sent from executor.actionStarted
						statusChan <- bundle.StatusMessage{
							State:       bundle.StateInProgress,
							Description: "action started",
						}
						// Final status sent by executor.actionFinishedWithSuccess
						statusChan <- bundle.StatusMessage{
							State:       bundle.StateFailed,
							Error:       fmt.Errorf("Everything is on fire"),
							Description: "action finished with error",
						}
						close(statusChan)
					}()
					return statusChan
				},
			},
			expectedMsgCount: 2,
			validate: func(messages []JobMsg) error {
				if len(messages) != 2 {
					return fmt.Errorf("expected 2 job messages")
				}
				first := messages[0]
				if first.State.State != bundle.StateInProgress {
					return fmt.Errorf("unexpected first message contents")
				}

				second := messages[1]
				if second.State.State != bundle.StateFailed ||
					second.State.Error != "Everything is on fire" {
					return fmt.Errorf("unexpected second message contents")
				}

				return nil
			},
		},
		{
			name: "should pass extCreds on jobMsg if found on executor",
			testJob: &apbJob{
				serviceInstanceID: serviceInstanceID,
				specID:            specID,
				method:            bundle.JobMethodProvision,
				skipExecution:     false,
				executor: func() bundle.Executor {
					e := &bundle.MockExecutor{}
					e.On("PodName").Return(podName)
					e.On("LastStatus").Return(bundle.StatusMessage{
						State:       bundle.StateSucceeded,
						Description: "action finished with success",
					})
					e.On("ExtractedCredentials").Return(extCreds)
					e.On("DashboardURL").Return("http://foo.example.com")
					return e
				}(),
				run: func(exec bundle.Executor) <-chan bundle.StatusMessage {
					statusChan := make(chan bundle.StatusMessage)
					go func() {
						// Initial message sent from executor.actionStarted
						statusChan <- bundle.StatusMessage{
							State:       bundle.StateInProgress,
							Description: "action started",
						}
						// Final status sent by executor.actionFinishedWithSuccess
						statusChan <- bundle.StatusMessage{
							State:       bundle.StateSucceeded,
							Description: "action finished with success",
						}
						close(statusChan)
					}()
					return statusChan
				},
			},
			expectedMsgCount: 2,
			validate: func(messages []JobMsg) error {
				if len(messages) != 2 {
					return fmt.Errorf("expected 2 job messages")
				}

				first := messages[0]
				if first.State.State != bundle.StateInProgress {
					return fmt.Errorf("unexpected first message contents")
				}

				second := messages[1]
				if second.State.State != bundle.StateSucceeded {
					return fmt.Errorf("unexpected fourth message contents")
				}

				for expectedKey, expectedVal := range extCreds.Credentials {
					testVal, ok := second.ExtractedCredentials.Credentials[expectedKey]

					if !ok {
						return fmt.Errorf("expected credential key missing from final jobMsg")
					}
					if testVal != expectedVal {
						return fmt.Errorf("expected credential val not the same as the one from jobMsg")
					}
				}

				return nil
			},
		},
		{
			name: "job message should use binding uuid",
			testJob: &apbJob{
				serviceInstanceID: serviceInstanceID,
				bindingID:         &bindingID,
				specID:            specID,
				method:            bundle.JobMethodBind,
				skipExecution:     false,
				executor: func() bundle.Executor {
					e := &bundle.MockExecutor{}
					e.On("PodName").Return(podName)
					e.On("LastStatus").Return(bundle.StatusMessage{
						State:       bundle.StateSucceeded,
						Description: "action finished with success",
					})
					e.On("ExtractedCredentials").Return(extCreds)
					e.On("DashboardURL").Return("http://foo.example.com")
					return e
				}(),
				run: func(exec bundle.Executor) <-chan bundle.StatusMessage {
					statusChan := make(chan bundle.StatusMessage)
					go func() {
						// Initial message sent from executor.actionStarted
						statusChan <- bundle.StatusMessage{
							State:       bundle.StateInProgress,
							Description: "action started",
						}
						// Final status sent by executor.actionFinishedWithSuccess
						statusChan <- bundle.StatusMessage{
							State:       bundle.StateSucceeded,
							Description: "action finished with success",
						}
						close(statusChan)
					}()
					return statusChan
				},
			},
			expectedMsgCount: 2,
			validate: func(messages []JobMsg) error {
				if len(messages) != 2 {
					return fmt.Errorf("expected 2 job messages")
				}

				first := messages[0]
				if first.State.State != bundle.StateInProgress {
					return fmt.Errorf("unexpected first message contents")
				}

				second := messages[1]
				if second.State.State != bundle.StateSucceeded {
					return fmt.Errorf("unexpected fourth message contents")
				}
				if first.BindingUUID != bindingID && second.BindingUUID != bindingID {
					return fmt.Errorf("Binding id is not valid expected: %v - got: %v", bindingID, first.BindingUUID)
				}

				for expectedKey, expectedVal := range extCreds.Credentials {
					testVal, ok := second.ExtractedCredentials.Credentials[expectedKey]

					if !ok {
						return fmt.Errorf("expected credential key missing from final jobMsg")
					}
					if testVal != expectedVal {
						return fmt.Errorf("expected credential val not the same as the one from jobMsg")
					}
				}

				return nil
			},
		},
		{
			name: "should send failed jobMsg when error reported on executor pod pull error",
			testJob: &apbJob{
				serviceInstanceID: serviceInstanceID,
				specID:            specID,
				method:            bundle.JobMethodProvision,
				skipExecution:     false,
				executor: func() bundle.Executor {
					e := &bundle.MockExecutor{}
					e.On("PodName").Return(podName)
					e.On("LastStatus").Return(bundle.StatusMessage{
						State:       bundle.StateFailed,
						Error:       runtime.ErrorPodPullErr,
						Description: "action finished with error",
					})
					e.On("ExtractedCredentials").Return(nil)
					return e
				}(),
				run: func(exec bundle.Executor) <-chan bundle.StatusMessage {
					statusChan := make(chan bundle.StatusMessage)
					go func() {
						// Initial message sent from executor.actionStarted
						statusChan <- bundle.StatusMessage{
							State:       bundle.StateInProgress,
							Description: "action started",
						}
						// Final status sent by executor.actionFinishedWithSuccess
						statusChan <- bundle.StatusMessage{
							State:       bundle.StateFailed,
							Error:       runtime.ErrorPodPullErr,
							Description: "action finished with error",
						}
						close(statusChan)
					}()
					return statusChan
				},
			},
			expectedMsgCount: 2,
			validate: func(messages []JobMsg) error {
				if len(messages) != 2 {
					return fmt.Errorf("expected 2 job messages")
				}
				first := messages[0]
				if first.State.State != bundle.StateInProgress {
					return fmt.Errorf("unexpected first message contents")
				}

				second := messages[1]
				if second.State.State != bundle.StateFailed ||
					second.State.Error != runtime.ErrorPodPullErr.Error() {
					return fmt.Errorf("unexpected second message contents")
				}

				return nil
			},
		},
		{
			name: "should send failed jobMsg when error reported on executor pod pull error",
			testJob: &apbJob{
				serviceInstanceID: serviceInstanceID,
				specID:            specID,
				method:            bundle.JobMethodProvision,
				skipExecution:     false,
				executor: func() bundle.Executor {
					e := &bundle.MockExecutor{}
					e.On("PodName").Return(podName)
					e.On("LastStatus").Return(bundle.StatusMessage{
						State:       bundle.StateFailed,
						Error:       runtime.ErrorCustomMsg{},
						Description: "action finished with error",
					})
					e.On("ExtractedCredentials").Return(nil)
					return e
				}(),
				run: func(exec bundle.Executor) <-chan bundle.StatusMessage {
					statusChan := make(chan bundle.StatusMessage)
					go func() {
						// Initial message sent from executor.actionStarted
						statusChan <- bundle.StatusMessage{
							State:       bundle.StateInProgress,
							Description: "action started",
						}
						// Final status sent by executor.actionFinishedWithSuccess
						statusChan <- bundle.StatusMessage{
							State:       bundle.StateFailed,
							Error:       runtime.ErrorCustomMsg{},
							Description: "action finished with error",
						}
						close(statusChan)
					}()
					return statusChan
				},
			},
			expectedMsgCount: 2,
			validate: func(messages []JobMsg) error {
				if len(messages) != 2 {
					return fmt.Errorf("expected 2 job messages")
				}
				first := messages[0]
				if first.State.State != bundle.StateInProgress {
					return fmt.Errorf("unexpected first message contents")
				}

				second := messages[1]
				if second.State.State != bundle.StateFailed ||
					second.State.Error != "" {
					return fmt.Errorf("unexpected second message contents")
				}

				return nil
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			msgBuffer := make(chan JobMsg)
			messages := []JobMsg{}
			tc.testJob.metricsJobStartHook = func() {
				tc.metricsJobStartHookTriggered = true
			}
			tc.testJob.metricsJobFinishedHook = func() {
				tc.metricsJobFinishedHookTriggered = true
			}

			go func(m []JobMsg) {
				time.Sleep(time.Duration(timeoutSecs) * time.Second)
				if len(m) != tc.expectedMsgCount {
					panic(fmt.Sprintf("%s timed out", tc.name))
				}
			}(messages)

			go func() {
				fmt.Printf("Running test %s", tc.name)
				tc.testJob.Run(token, msgBuffer)
			}()

			for {
				msg := <-msgBuffer
				messages = append(messages, msg)
				if len(messages) == tc.expectedMsgCount {
					break
				}
			}

			if tc.validate != nil {
				if err := tc.validate(messages); err != nil {
					t.Fatal("unexpected errror validating job state", err)
				}
			}
		})
	}
}

func TestWork(t *testing.T) {
	cases := []struct {
		Name     string
		Work     func() []Work
		Validate func(t *testing.T, w []Work)
	}{
		{
			Name: "test work returns binding id when unbind or bind work",
			Work: func() []Work {
				id := "bindingID"
				jobs := []Work{&unbindJob{apbJob: apbJob{method: bundle.JobMethodUnbind, bindingID: &id}}, &bindJob{apbJob: apbJob{method: bundle.JobMethodBind, bindingID: &id}}}
				return jobs
			},
			Validate: func(t *testing.T, w []Work) {
				for _, work := range w {
					if work.ID() != "bindingID" {
						t.Fatalf("expected id to tbe bindingID but got %s ", work.ID())
					}
				}
			},
		},
		{
			Name: "test work returns service instance id when provision update or deprovision work",
			Work: func() []Work {
				id := "serviceInstanceID"
				j := []Work{&provisionJob{apbJob: apbJob{method: bundle.JobMethodProvision, serviceInstanceID: id}},
					&updateJob{apbJob: apbJob{method: bundle.JobMethodUpdate, serviceInstanceID: id}}, &deprovisionJob{apbJob: apbJob{method: bundle.JobMethodDeprovision, serviceInstanceID: id}}}
				return j
			},
			Validate: func(t *testing.T, work []Work) {
				for _, w := range work {
					if w.Method() == bundle.JobMethodUnbind || w.Method() == bundle.JobMethodBind {
						t.Fatalf("did not expect an unbind or bind method ")
					}
					if w.ID() != "serviceInstanceID" {
						t.Fatalf("expected id to be serviceInstanceID but got %s ", w.ID())
					}
				}
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.Name, func(t *testing.T) {
			tc.Validate(t, tc.Work())
		})
	}
}

func TestNewUnbindJob(t *testing.T) {
	cases := []struct {
		name      string
		bindingID string
		params    *bundle.Parameters
		si        *bundle.ServiceInstance
		skip      bool
		validate  func(t *testing.T, w Work)
	}{
		{
			name:      "ensure bindingID is passed to the job",
			bindingID: "test-binding-id-abcd123",
			params:    &bundle.Parameters{},
			si: &bundle.ServiceInstance{
				Spec: &bundle.Spec{
					ID: "test-spec",
				},
			},
			skip: false,
			validate: func(t *testing.T, work Work) {
				assert.Equal(t, "test-binding-id-abcd123", work.ID())
				assert.Equal(t, bundle.JobMethodUnbind, work.Method())
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			wf := NewWorkFactory()
			unbindjob := wf.NewUnbindJob(tc.bindingID, tc.params, tc.si, tc.skip)
			tc.validate(t, unbindjob)
		})
	}

}
