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
	"fmt"

	"reflect"

	log "github.com/sirupsen/logrus"
	apiv1 "k8s.io/api/core/v1"
	meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/kubernetes/typed/core/v1"
)

var (
	// ErrorPodPullErr - Error indicating we could not pull the image.
	ErrorPodPullErr = fmt.Errorf("Unable to pull APB image from it's registry. Please contact your cluster admin")
	// ErrorActionNotFound - Error indicating pod does not have the action.
	ErrorActionNotFound = fmt.Errorf("action not found")
)

// UpdateDescriptionFn function that will should handle the LastDescription from the bundle.
type UpdateDescriptionFn func(string, string)

// ErrorCustomMsg - An error to propagate the custom error message to the callers
type ErrorCustomMsg struct {
	msg string
}

func (e ErrorCustomMsg) Error() string {
	// returns an Error with a custom message
	return e.msg
}

// IsErrorCustomMsg - true if it's a custom message error
func IsErrorCustomMsg(err error) bool {
	_, ok := err.(ErrorCustomMsg)
	return ok
}

// WatchPod - watches the pod until completion and will update the last
// description using the UpdateDescriptionFunction
func WatchPod(podName string, namespace string, podClient v1.PodInterface, updateFunc UpdateDescriptionFn) error {
	log.Debugf(
		"Watching pod [ %s ] in namespace [ %s ] for completion",
		podName,
		namespace,
	)

	w, err := podClient.Watch(meta_v1.ListOptions{})
	if err != nil {
		return fmt.Errorf("failed to watch pod %s in namespace %s error: %v", podName, namespace, err)
	}
	for podEvent := range w.ResultChan() {
		pod, ok := podEvent.Object.(*apiv1.Pod)
		if !ok {
			log.Errorf("watch did not return a apiv1.Pod instead returned %v", reflect.TypeOf(podEvent.Object))
			continue
		}
		if pod.Name != podName {
			log.Debugf("watching pods in namespace %s ignoring pod %s as it is not the pod we are looking for", namespace, pod.Name)
			continue
		}

		lastOp := pod.Annotations["apb_last_operation"]
		if lastOp != "" {
			updateFunc(lastOp, "")
		}
		podStatus := pod.Status
		log.Debugf("pod [%s] in phase %s", podName, podStatus.Phase)
		switch podStatus.Phase {
		case apiv1.PodFailed:
			w.Stop()
			if errorPullingImage(podStatus.ContainerStatuses) {
				return ErrorPodPullErr
			}
			return translateExitStatus(podName, podStatus)
		case apiv1.PodSucceeded:
			w.Stop()
			// Check for dashboard_url
			dashURL := pod.Annotations["apb_dashboard_url"]
			updateFunc("", dashURL)
			log.Debugf("Pod [ %s ] completed", podName)
			return nil
		default:
			log.Debugf("Pod [ %s ] %s", podName, podStatus.Phase)
		}
		if podEvent.Type == watch.Deleted {
			w.Stop()
			return fmt.Errorf("pod [ %s ] was unexpectedly deleted", podName)
		}
	}
	log.Debugf("finished watching pod %s in namespace %s ", podName, namespace)
	return nil
}

func errorPullingImage(conds []apiv1.ContainerStatus) bool {
	if len(conds) < 1 {
		log.Warningf("unable to get container status for APB pod")
		return false
	}
	// We should expect only a single container for our APB pod.
	// If this assumption changes then we may need to update this code.
	// Basis for the image strings is here:
	// https://github.com/kubernetes/kubernetes/blob/886e04f1fffbb04faf8a9f9ee141143b2684ae68/pkg/kubelet/images/types.go#L27
	status := conds[0].State.Waiting
	if status == nil {
		return false
	}

	if status.Reason == "ErrImagePull" {
		return true
	} else if status.Reason == "ImagePullBackOff" {
		return true
	}

	return false
}

func translateExitStatus(podName string, podStatus apiv1.PodStatus) error {
	conds := podStatus.ContainerStatuses
	if len(conds) < 1 {
		log.Warningf("unable to get container status for APB pod")
		return fmt.Errorf("Pod [ %s ] failed - Unable to determine exit code - %v", podName, podStatus.Message)
	}

	status := conds[0].State.Terminated
	if status == nil {
		return fmt.Errorf("Pod [ %s ] failed. Unable to determine status - %v", podName, podStatus.Message)
	}

	// return the termination message if it's not empty
	if status.Message != "" {
		return ErrorCustomMsg{msg: status.Message}
	}

	if status.ExitCode == 8 {
		log.Errorf("Pod [ %s ] failed - action's playbook not found.", podName)
		return ErrorActionNotFound
	} else if status.ExitCode != 0 {
		return fmt.Errorf("Pod [ %s ] failed with exit code [%d]", podName, status.ExitCode)
	}

	// exit code was 0 so not really an error
	log.Warning("Pod was marked as failed but exit code was 0 - %v", status.Message)
	return nil
}
