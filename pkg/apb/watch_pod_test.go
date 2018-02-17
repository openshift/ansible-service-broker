package apb

import (
	"testing"

	"time"

	"fmt"

	core1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/kubernetes/fake"
	"k8s.io/client-go/kubernetes/typed/core/v1"
	ktesting "k8s.io/client-go/testing"
)

func TestWatchPod(t *testing.T) {

	podStateUpdater := func(watcher *watch.FakeWatcher, podUpdates []*core1.Pod) {
		for _, podUpdate := range podUpdates {
			watcher.Modify(podUpdate)
		}
	}

	cases := []struct {
		Name            string
		PodClient       func() (v1.PodInterface, *watch.FakeWatcher)
		UpdatePodStates func(watcher *watch.FakeWatcher)
		ExpectError     bool
		Validate        func(status []JobState) error
	}{
		{
			Name: "should get error and state update when pod fails",
			PodClient: func() (v1.PodInterface, *watch.FakeWatcher) {
				kfake := &fake.Clientset{}
				podWatch := watch.NewFake()
				kfake.AddWatchReactor("pods", ktesting.DefaultWatchReactor(podWatch, nil))
				return kfake.CoreV1().Pods("test"), podWatch
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
			Validate: func(status []JobState) error {
				if len(status) != 2 {
					return fmt.Errorf("expected 2 status updates")
				}
				for i, s := range status {
					if s.Description != fmt.Sprintf("lastop%v", i) {
						return fmt.Errorf("expected description to be lastop%v but got %v", i, s.Description)
					}
					if s.State != StateInProgress {
						return fmt.Errorf("expected state to be %v but was %v", StateInProgress, s.State)
					}
				}
				return nil
			},
		},
		{
			Name: "should get state updates when pod succeeds and no error",
			PodClient: func() (v1.PodInterface, *watch.FakeWatcher) {
				kfake := &fake.Clientset{}
				podWatch := watch.NewFake()
				kfake.AddWatchReactor("pods", ktesting.DefaultWatchReactor(podWatch, nil))
				return kfake.CoreV1().Pods("test"), podWatch
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
			Validate: func(status []JobState) error {
				if len(status) != 2 {
					return fmt.Errorf("expected 2 status updates")
				}
				for i, s := range status {
					if s.Description != fmt.Sprintf("lastop%v", i) {
						return fmt.Errorf("expected description to be lastop%v but got %v", i, s.Description)
					}
					if s.State != StateInProgress {
						return fmt.Errorf("expected state to be %v but was %v", StateInProgress, s.State)
					}
				}
				return nil
			},
		},
		{
			Name: "should get state updates error if pod unexpectedly deleted",
			PodClient: func() (v1.PodInterface, *watch.FakeWatcher) {
				kfake := &fake.Clientset{}
				podWatch := watch.NewFake()
				kfake.AddWatchReactor("pods", ktesting.DefaultWatchReactor(podWatch, nil))
				return kfake.CoreV1().Pods("test"), podWatch
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
			Validate: func(status []JobState) error {
				if len(status) != 2 {
					return fmt.Errorf("expected 2 status updates")
				}
				for _, s := range status {
					if s.Description != "lastop0" {
						return fmt.Errorf("expected description to be lastop0 but got %v", s.Description)
					}
					if s.State != StateInProgress {
						return fmt.Errorf("expected state to be %v but was %v", StateInProgress, s.State)
					}
				}
				return nil
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.Name, func(t *testing.T) {
			statusReceiver := make(chan JobState)
			podClient, podWatch := tc.PodClient()
			time.AfterFunc(100*time.Millisecond, func() {
				close(statusReceiver)
			})
			var watchErr error
			go func() {
				watchErr = watchPod("test", "test", podClient, statusReceiver)
			}()
			go tc.UpdatePodStates(podWatch)
			var state []JobState
			for s := range statusReceiver {
				state = append(state, s)
			}
			if nil != tc.Validate {
				if err := tc.Validate(state); err != nil {
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
