//
// Copyright (c) 2017 Red Hat, Inc.
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
// Red Hat trademarks are not licensed under Apache License, Version 2.
// No permission is granted to use or replicate Red Hat trademarks that
// are incorporated in this software or its documentation.
//

package apb

import (
	"fmt"
	"strings"
	"time"

	"github.com/openshift/ansible-service-broker/pkg/runtime"

	logging "github.com/op/go-logging"
)

const (
	podStatusRunning   = "Running"
	podStatusCompleted = "Completed"
	podStatusError     = "Error"
)

func watchPod(podName string, namespace string, log *logging.Logger) (string, error) {
	log.Debugf(
		"Watching pod [ %s ] in namespace [ %s ] for completion", podName, namespace)

	for r := 1; r <= apbWatchRetries; r++ {
		log.Info("Watch pod [ %s ] tick %d", podName, r)
		output, err := runtime.RunCommand(
			"kubectl", "get", "pod", "--no-headers=true", "--namespace="+namespace, podName)

		outStr := string(output)

		isPodRunning := strings.Contains(outStr, podStatusRunning)
		didPodComplete := strings.Contains(outStr, podStatusCompleted)
		didPodError := strings.Contains(outStr, podStatusError)

		if err != nil {
			log.Infof("Got error from watch pod cmd: %s\n error: %s\n output: %s",
				podName, string(err.Error()), outStr)
		} else if didPodError {
			return outStr, fmt.Errorf("Pod %s is reporting error", podName)
		} else if didPodComplete {
			return outStr, nil
		} else if isPodRunning {
			log.Info("Pod %s still running, continuing to watch", podName)
		} else {
			log.Info("Pod completion not found, continuing to watch")
			log.Infof("%s", outStr)
		}

		time.Sleep(time.Duration(apbWatchInterval) * time.Second)
	}

	err := fmt.Errorf(
		"Timed out while watching pod %s for completion", podName)
	return "", err
}
