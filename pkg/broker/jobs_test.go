package broker

import (
	"fmt"
	"testing"
	"time"

	"github.com/openshift/ansible-service-broker/pkg/apb"
)

func TestApbJobRun(t *testing.T) {
	timeoutSecs := 5
	serviceInstanceID := "16235516-9e5e-4c68-a541-33bda63413ee"
	specID := "a7fc708e-52cf-427e-88b2-2b750b607a27"
	//bindingID := "20c6ec16-c5bd-433a-815c-63cf0e2d2c9d"
	token := "4ac9529c-6a01-4daf-9e10-8b557e4885ae"

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
				executor:          apb.NewExecutor(),
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
				if first.State.State != apb.StateInProgress ||
					first.PodName != "" {
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
				executor:          apb.NewExecutor(),
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
