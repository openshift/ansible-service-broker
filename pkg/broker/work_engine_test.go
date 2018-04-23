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

package broker

import (
	"sync"
	"testing"
	"time"

	"github.com/automationbroker/bundle-lib/bundle"
	"github.com/openshift/ansible-service-broker/pkg/dao"

	ft "github.com/openshift/ansible-service-broker/pkg/fusortest"
)

var engine *WorkEngine
var mockDao = &dao.MockDao{}
var testToken = engine.Token()

func init() {
	engine = NewWorkEngine(10, 1, mockDao)
}

type mockSubscriber struct {
	msg        JobMsg
	called     bool
	funcToCall func(msg JobMsg)
}

func (ms *mockSubscriber) Notify(msg JobMsg) {
	ms.msg = msg
	ms.called = true
	if ms.funcToCall != nil {
		ms.funcToCall(msg)
	}
}

func (ms *mockSubscriber) ID() string {
	return "mock"
}

type mockWorker struct {
	called bool
	wg     *sync.WaitGroup
}

func (mw *mockWorker) Run(token string, buffer chan<- JobMsg) {
	mw.called = true
	buffer <- JobMsg{Msg: "hello"}
	mw.wg.Done()
}

func TestAttachSubscriber(t *testing.T) {
	dasub := mockSubscriber{}
	subsBefore := len(engine.GetSubscribers(ProvisionTopic))
	err := engine.AttachSubscriber(&dasub, ProvisionTopic)
	if err != nil {
		t.Fatal(err)
	}
	subsAfter := len(engine.GetSubscribers(ProvisionTopic))
	if subsAfter != subsBefore+1 {
		t.Fatal("expected subscribers to increase by one")
	}
}

func TestInvalidWorkTopic(t *testing.T) {
	var faketopic WorkTopic
	faketopic = "fake"
	dasub := mockSubscriber{}
	err := engine.AttachSubscriber(&dasub, faketopic)
	if err == nil {
		t.Fail()
	}
	ft.AssertEqual(t, "invalid work topic", err.Error(), "invalid error")
}

type mockWork struct {
	funcToCall func(msg chan<- JobMsg)
}

func (mw *mockWork) Run(token string, msgBuffer chan<- JobMsg) {
	mw.funcToCall(msgBuffer)
}

func (mw *mockWork) ID() string {
	return "id"
}

func (mw *mockWork) Method() bundle.JobMethod {
	return bundle.JobMethodBind
}

func TestStartNewJob(t *testing.T) {

	cases := []struct {
		Name          string
		Work          func() Work
		Subscribers   func(wg *sync.WaitGroup) []WorkSubscriber
		ConfigureMock func()
		TestTimeout   time.Duration
	}{
		{
			Name: "test start new job sends a message and calls all subscribers",
			Work: func() Work {
				return &mockWork{
					funcToCall: func(msg chan<- JobMsg) {
						msg <- JobMsg{
							Msg: "test",
						}
					},
				}
			},
			ConfigureMock: func() {
				mockDao.On("SetState", "id", bundle.JobState{Token: testToken, State: "not yet started", Podname: "", Method: "bind", Error: "", Description: ""}).Return(testToken, nil)
			},
			Subscribers: func(wg *sync.WaitGroup) []WorkSubscriber {
				wg.Add(2)
				return []WorkSubscriber{&mockSubscriber{
					funcToCall: func(msg JobMsg) {
						wg.Done()
					},
				}, &mockSubscriber{
					funcToCall: func(msg JobMsg) {
						wg.Done()
					},
				}}
			},
			TestTimeout: 1,
		},
		{
			Name: "test start new job sends a message and calls all subscribers even if one timesout",
			Work: func() Work {
				return &mockWork{
					funcToCall: func(msg chan<- JobMsg) {
						msg <- JobMsg{
							Msg: "test",
						}
					},
				}
			},
			ConfigureMock: func() {
				mockDao.On("SetState", "id", bundle.JobState{Token: testToken, State: "not yet started", Podname: "", Method: "bind", Error: "", Description: ""}).Return(testToken, nil)
			},
			Subscribers: func(wg *sync.WaitGroup) []WorkSubscriber {
				wg.Add(2)
				return []WorkSubscriber{&mockSubscriber{
					funcToCall: func(msg JobMsg) {
						// force a timeout
						<-time.Tick(2 * time.Second)
						wg.Done()
					},
				}, &mockSubscriber{
					funcToCall: func(msg JobMsg) {
						wg.Done()
					},
				}}
			},
			TestTimeout: 3,
		},
	}

	done := func(wg *sync.WaitGroup) chan struct{} {
		d := make(chan struct{})
		go func() {
			wg.Wait()
			d <- struct{}{}
		}()
		return d
	}

	for _, tc := range cases {
		t.Run(tc.Name, func(t *testing.T) {
			tc.ConfigureMock()
			wg := &sync.WaitGroup{}
			engine := NewWorkEngine(10, 1, mockDao)
			for _, s := range tc.Subscribers(wg) {
				engine.AttachSubscriber(s, ProvisionTopic)
			}

			engine.StartNewAsyncJob(testToken, tc.Work(), ProvisionTopic)
			select {
			case <-time.Tick(tc.TestTimeout * time.Second):
				t.Fatal("test timed out !!")
			case <-done(wg):
				// check out channel is gone
				if _, ok := engine.jobChannels[testToken]; ok {
					t.Fatal("there should be no job channel present")
				}
			}
		})
	}
}
