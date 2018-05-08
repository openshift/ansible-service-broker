package broker

import (
	"fmt"
	"testing"
	"time"

	apb "github.com/automationbroker/bundle-lib/bundle"
)

func TestApbJobRun(t *testing.T) {
	timeoutSecs := 5
	serviceInstanceID := "16235516-9e5e-4c68-a541-33bda63413ee"
	specID := "a7fc708e-52cf-427e-88b2-2b750b607a27"
	//bindingID := "20c6ec16-c5bd-433a-815c-63cf0e2d2c9d"
	token := "4ac9529c-6a01-4daf-9e10-8b557e4885ae"
	podName := "apb-8f9268c2-1aaa-48f1-918d-eae920986c9f"
	extCreds := &apb.ExtractedCredentials{
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
				method:            apb.JobMethodProvision,
				skipExecution:     false,
				executor: func() apb.Executor {
					e := &apb.MockExecutor{}
					e.On("PodName").Return(podName)
					e.On("LastStatus").Return(apb.StatusMessage{
						State:       apb.StateSucceeded,
						Description: "action finished with success",
					})
					e.On("ExtractedCredentials").Return(nil)
					e.On("DashboardURL").Return("http://foo.example.com")
					return e
				}(),
				run: func(exec apb.Executor) <-chan apb.StatusMessage {
					statusChan := make(chan apb.StatusMessage)
					go func() {
						// Initial message sent from executor.actionStarted
						statusChan <- apb.StatusMessage{
							State:       apb.StateInProgress,
							Description: "action started",
						}
						// Two updateDescription calls
						statusChan <- apb.StatusMessage{
							State:       apb.StateInProgress,
							Description: "lastOp0",
						}
						statusChan <- apb.StatusMessage{
							State:       apb.StateInProgress,
							Description: "lastOp1",
						}
						// Final status sent by executor.actionFinishedWithSuccess
						statusChan <- apb.StatusMessage{
							State:       apb.StateSucceeded,
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
				if first.State.State != apb.StateInProgress {
					return fmt.Errorf("unexpected first message contents")
				}

				second := messages[1]
				if second.State.State != apb.StateInProgress ||
					second.State.Description != "lastOp0" {
					return fmt.Errorf("unexpected second message contents")
				}

				third := messages[2]
				if third.State.State != apb.StateInProgress ||
					third.State.Description != "lastOp1" {
					return fmt.Errorf("unexpected third message contents")
				}

				fourth := messages[3]
				if fourth.State.State != apb.StateSucceeded {
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
				method:            apb.JobMethodProvision,
				skipExecution:     true,
				executor:          &apb.MockExecutor{},
				run: func(exec apb.Executor) <-chan apb.StatusMessage {
					statusChan := make(chan apb.StatusMessage)
					go func() {
						// Initial message sent from executor.actionStarted
						statusChan <- apb.StatusMessage{
							State:       apb.StateInProgress,
							Description: "action started",
						}
						// Two updateDescription calls
						statusChan <- apb.StatusMessage{
							State:       apb.StateInProgress,
							Description: "lastOp0",
						}
						statusChan <- apb.StatusMessage{
							State:       apb.StateInProgress,
							Description: "lastOp1",
						}
						// Final status sent by executor.actionFinishedWithSuccess
						statusChan <- apb.StatusMessage{
							State:       apb.StateSucceeded,
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
				if messages[0].State.State != apb.StateSucceeded {
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
				method:            apb.JobMethodProvision,
				skipExecution:     false,
				executor: func() apb.Executor {
					e := &apb.MockExecutor{}
					e.On("PodName").Return(podName)
					e.On("LastStatus").Return(apb.StatusMessage{
						State:       apb.StateFailed,
						Error:       fmt.Errorf("Everything is on fire"),
						Description: "action finished with error",
					})
					e.On("ExtractedCredentials").Return(nil)
					return e
				}(),
				run: func(exec apb.Executor) <-chan apb.StatusMessage {
					statusChan := make(chan apb.StatusMessage)
					go func() {
						// Initial message sent from executor.actionStarted
						statusChan <- apb.StatusMessage{
							State:       apb.StateInProgress,
							Description: "action started",
						}
						// Final status sent by executor.actionFinishedWithSuccess
						statusChan <- apb.StatusMessage{
							State:       apb.StateFailed,
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
				if first.State.State != apb.StateInProgress {
					return fmt.Errorf("unexpected first message contents")
				}

				second := messages[1]
				if second.State.State != apb.StateFailed ||
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
				method:            apb.JobMethodProvision,
				skipExecution:     false,
				executor: func() apb.Executor {
					e := &apb.MockExecutor{}
					e.On("PodName").Return(podName)
					e.On("LastStatus").Return(apb.StatusMessage{
						State:       apb.StateSucceeded,
						Description: "action finished with success",
					})
					e.On("ExtractedCredentials").Return(extCreds)
					e.On("DashboardURL").Return("http://foo.example.com")
					return e
				}(),
				run: func(exec apb.Executor) <-chan apb.StatusMessage {
					statusChan := make(chan apb.StatusMessage)
					go func() {
						// Initial message sent from executor.actionStarted
						statusChan <- apb.StatusMessage{
							State:       apb.StateInProgress,
							Description: "action started",
						}
						// Final status sent by executor.actionFinishedWithSuccess
						statusChan <- apb.StatusMessage{
							State:       apb.StateSucceeded,
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
				if first.State.State != apb.StateInProgress {
					return fmt.Errorf("unexpected first message contents")
				}

				second := messages[1]
				if second.State.State != apb.StateSucceeded {
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
