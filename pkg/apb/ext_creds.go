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

package apb

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/openshift/ansible-service-broker/pkg/runtime"

	logging "github.com/op/go-logging"
)

// ExtractCredentials - Extract credentials from pod in a certain namespace.
func ExtractCredentials(
	podname string, namespace string, log *logging.Logger,
) (*ExtractedCredentials, error) {
	log.Debug("Calling monitorOutput on " + podname)
	bindOutput, err := monitorOutput(namespace, podname, log)
	if err != nil {
		return nil, err
	}

	if bindOutput == nil {
		return nil, nil
	}

	return buildExtractedCredentials(bindOutput)
}

func monitorOutput(namespace string, podname string, log *logging.Logger) ([]byte, error) {
	// TODO: Error handling here
	// It would also be nice to gather the script output that exec runs
	// instead of only getting the credentials

	for r := 1; r <= apbWatchRetries; r++ {
		// err will be the return code from the exec command
		// Use the error code to determine the state
		failedToExec := errors.New("exit status 1")
		credsNotAvailable := errors.New("exit status 2")

		output, err := runtime.RunCommand("kubectl", "exec", podname, gatherCredentialsCMD, "--namespace="+namespace)

		// cannot exec container, pod is done
		podFailed := strings.Contains(string(output), "current phase is Failed")
		podCompleted := strings.Contains(string(output), "current phase is Succeeded") ||
			strings.Contains(string(output), "cannot exec into a container in a completed pod")

		if err == nil {
			log.Notice("[%s] Bind credentials found", podname)
			return output, nil
		} else if podFailed {
			// pod has completed but is in failed state
			log.Notice("[%s] APB failed", podname)
			return nil, errors.New("APB failed")
		} else if podCompleted && err.Error() == failedToExec.Error() {
			log.Error("[%s] APB completed", podname)
			return nil, nil
		} else if err.Error() == failedToExec.Error() {
			log.Info(string(output))
			log.Warning("[%s] Retry attempt %d: Failed to exec into the container", podname, r)
		} else if err.Error() == credsNotAvailable.Error() {
			log.Info(string(output))
			log.Warning("[%s] Retry attempt %d: Bind credentials not available yet", podname, r)
		} else {
			log.Info(string(output))
			log.Warning("[%s] Retry attempt %d: Failed to exec into the container", podname, r)
		}

		log.Warning("[%s] Retry attempt %d: exec into %s failed", podname, r, podname)
		time.Sleep(time.Duration(apbWatchInterval) * time.Second)
	}

	timeout := fmt.Sprintf("[%s] ExecTimeout: Failed to gather bind credentials after %d retries", podname, apbWatchRetries)
	return nil, errors.New(timeout)
}

func buildExtractedCredentials(output []byte) (*ExtractedCredentials, error) {
	result, err := decodeOutput(output)
	if err != nil {
		return nil, err
	}

	creds := make(map[string]interface{})
	for k, v := range result {
		creds[k] = v
	}

	return &ExtractedCredentials{Credentials: creds}, nil
}

func decodeOutput(output []byte) (map[string]interface{}, error) {
	str := string(output)

	decodedjson, err := base64.StdEncoding.DecodeString(str)
	if err != nil {
		return nil, err
	}

	decoded := make(map[string]interface{})
	json.Unmarshal(decodedjson, &decoded)
	return decoded, nil
}
