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

package apb

import (
	"fmt"
	"time"

	"github.com/openshift/ansible-service-broker/pkg/clients"

	apiv1 "k8s.io/api/core/v1"
)

var (
	// ErrorPodPullErr - Error indicating we could not pull the image.
	ErrorPodPullErr = fmt.Errorf("Unable to pull APB image from it's registry. Please contact your cluster admin")
	// ErrorActionNotFound - Error indicating pod does not have the action.
	ErrorActionNotFound = fmt.Errorf("action not found")
)

func watchPod(podName string, namespace string) error {
	log.Debugf(
		"Watching pod [ %s ] in namespace [ %s ] for completion",
		podName,
		namespace,
	)

	k8scli, err := clients.Kubernetes()
	if err != nil {
		return fmt.Errorf("Unable to retrive kubernetes client - %v", err)
	}

	for r := 1; r <= apbWatchRetries; r++ {
		log.Info("Watch pod [ %s ] tick %d", podName, r)

		podStatus, err := k8scli.GetPodStatus(podName, namespace)
		if err != nil {
			return err
		}

		switch podStatus.Phase {
		case apiv1.PodFailed:
			if errorPullingImage(podStatus.ContainerStatuses) {
				return ErrorPodPullErr
			}

			// handle the return code from the pod
			exitCode := getExitCode(podStatus.ContainerStatuses)
			if exitCode == 8 {
				log.Errorf("Pod [ %s ] failed - action not found.", podName)
				return ErrorActionNotFound
			} else if exitCode > 0 {
				return fmt.Errorf("Pod [ %s ] failed with exit code [%d]", podName, exitCode)
			} else if exitCode < 0 {
				return fmt.Errorf("Pod [ %s ] failed. Unable to determine exit code [%d]", podName, exitCode)
			}

			return fmt.Errorf("Pod [ %s ] failed - %v", podName, podStatus.Message)
		case apiv1.PodSucceeded:
			log.Debugf("Pod [ %s ] completed", podName)
			return nil
		default:
			log.Debugf("Pod [ %s ] %s", podName, podStatus.Phase)
		}

		time.Sleep(time.Duration(apbWatchInterval) * time.Second)
	}

	return fmt.Errorf("Timed out while watching pod %s for completion", podName)
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

func getExitCode(conds []apiv1.ContainerStatus) int32 {
	if len(conds) < 1 {
		log.Warningf("unable to get container status for APB pod")
		return -120 // picked some weird number
	}
	status := conds[0].State.Terminated
	if status == nil {
		return -127 // picked some weird number
	}

	return status.ExitCode
}
