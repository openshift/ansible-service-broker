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

package runtime

import (
	"testing"

	"fmt"

	"github.com/automationbroker/bundle-lib/clients"
	core1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/kubernetes/fake"
	ktesting "k8s.io/client-go/testing"
)

func TestWatchPod(t *testing.T) {
	k8scli, err := clients.Kubernetes()
	if err != nil {
		t.Fatal()
	}

	podStateUpdater := func(watcher *watch.FakeWatcher, podUpdates []*core1.Pod) {
		for _, podUpdate := range podUpdates {
			watcher.Modify(podUpdate)
		}
	}

	cases := []struct {
		Name            string
		PodClient       func() (*fake.Clientset, *watch.FakeWatcher)
		UpdatePodStates func(watcher *watch.FakeWatcher)
		ExpectError     bool
		Validate        func(status []string) error
	}{
		{
			Name: "should get error and state update when pod fails",
			PodClient: func() (*fake.Clientset, *watch.FakeWatcher) {
				kfake := &fake.Clientset{}
				podWatch := watch.NewFake()
				kfake.AddWatchReactor("pods", ktesting.DefaultWatchReactor(podWatch, nil))
				return kfake, podWatch
			},
			ExpectError: true,
			UpdatePodStates: func(watcher *watch.FakeWatcher) {
				podStates := []*core1.Pod{{
					ObjectMeta: metav1.ObjectMeta{
						Name: "test",
						Annotations: map[string]string{
							"apb_last_operation": "lastop0",
						},
					},
					Status: core1.PodStatus{
						Phase: core1.PodRunning,
					},
				}, {
					ObjectMeta: metav1.ObjectMeta{
						Name: "test",
						Annotations: map[string]string{
							"apb_last_operation": "lastop1",
						},
					},
					Status: core1.PodStatus{
						Phase: core1.PodFailed,
					},
				}}
				podStateUpdater(watcher, podStates)
			},
			Validate: func(status []string) error {
				if len(status) != 2 {
					return fmt.Errorf("expected 2 status updates")
				}
				for i, s := range status {
					if s != fmt.Sprintf("lastop%v", i) {
						return fmt.Errorf("expected description to be lastop%v but got %v", i, s)
					}
				}
				return nil
			},
		},
		{
			Name: "should get state updates when pod succeeds and no error",
			PodClient: func() (*fake.Clientset, *watch.FakeWatcher) {
				kfake := &fake.Clientset{}
				podWatch := watch.NewFake()
				kfake.AddWatchReactor("pods", ktesting.DefaultWatchReactor(podWatch, nil))
				return kfake, podWatch
			},
			UpdatePodStates: func(watcher *watch.FakeWatcher) {
				podStates := []*core1.Pod{{
					ObjectMeta: metav1.ObjectMeta{
						Name: "test",
						Annotations: map[string]string{
							"apb_last_operation": "lastop0",
						},
					},
					Status: core1.PodStatus{
						Phase: core1.PodRunning,
					},
				}, {
					ObjectMeta: metav1.ObjectMeta{
						Name: "test",
						Annotations: map[string]string{
							"apb_last_operation": "lastop1",
						},
					},
					Status: core1.PodStatus{
						Phase: core1.PodSucceeded,
					},
				}}
				podStateUpdater(watcher, podStates)

			},
			Validate: func(status []string) error {
				if len(status) != 2 {
					return fmt.Errorf("expected 2 status updates")
				}
				for i, s := range status {
					if s != fmt.Sprintf("lastop%v", i) {
						return fmt.Errorf("expected description to be lastop%v but got %v", i, s)
					}
				}
				return nil
			},
		},
		{
			Name: "should get state updates error if pod unexpectedly deleted",
			PodClient: func() (*fake.Clientset, *watch.FakeWatcher) {
				kfake := &fake.Clientset{}
				podWatch := watch.NewFake()
				kfake.AddWatchReactor("pods", ktesting.DefaultWatchReactor(podWatch, nil))
				return kfake, podWatch
			},
			UpdatePodStates: func(watcher *watch.FakeWatcher) {
				podStates := []*core1.Pod{{
					ObjectMeta: metav1.ObjectMeta{
						Name: "test",
						Annotations: map[string]string{
							"apb_last_operation": "lastop0",
						},
					},
					Status: core1.PodStatus{
						Phase: core1.PodRunning,
					},
				}}
				podStateUpdater(watcher, podStates)
				watcher.Delete(podStates[0])
			},
			ExpectError: true,
			Validate: func(status []string) error {
				if len(status) != 2 {
					return fmt.Errorf("expected 2 status updates")
				}
				for _, s := range status {
					if s != "lastop0" {
						return fmt.Errorf("expected description to be lastop0 but got %v", s)
					}
				}
				return nil
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.Name, func(t *testing.T) {
			var watchErr error
			var dashURL string
			podClient, podWatch := tc.PodClient()
			descriptions := []string{}
			done := make(chan bool)
			k8scli.Client = podClient

			go func() {
				watchErr = defaultWatchRunningBundle("test", "test", func(d, newDashURL string) {
					fmt.Printf("got newDescription -> %v\n", d)
					fmt.Printf("got newDashURL-> %v\n", newDashURL)
					if d != "" {
						descriptions = append(descriptions, d)
					}

					if dashURL != "" {
						dashURL = newDashURL
					}
				})
				done <- true
			}()
			go tc.UpdatePodStates(podWatch)

			<-done

			if nil != tc.Validate {
				fmt.Printf("NSK: Now trying to validate the descriptions: %v", descriptions)
				if err := tc.Validate(descriptions); err != nil {
					t.Fatal("unexpected errror validating job state", err)
				}
			}
			if tc.ExpectError && watchErr == nil {
				t.Fatal("expected a watch err but got none")
			}

			if !tc.ExpectError && watchErr != nil {
				t.Fatal("did not expect a watch err but got one ", watchErr)
			}

		})
	}
}
