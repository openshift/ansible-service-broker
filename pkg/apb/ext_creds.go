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
	"bufio"
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"runtime/debug"

	"github.com/openshift/ansible-service-broker/pkg/clients"
	"github.com/openshift/ansible-service-broker/pkg/version"
	"k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/remotecommand"
)

const (
	// GatherCredentialsCommand - Command used when execing for bind credentials
	// moving this constant here because eventuall Extrating creds will
	// need to be moved to runtime. Therefore keeping all of this together
	// makes sense
	GatherCredentialsCommand = "broker-bind-creds"
)

type extractCreds func(string, string) (*ExtractedCredentials, error)

// ExtractCredentials - Extract credentials from pod in a certain namespace.
// needs the podname, namespace and the runtime version.
func ExtractCredentials(podname string, ns string, runtime int) (*ExtractedCredentials, error) {
	extractCredsFunc, err := getExtractCreds(runtime)
	if err != nil {
		return nil, err
	}
	return extractCredsFunc(podname, ns)
}

// ExtractCredentialsAsFile - Extract credentials from running APB using exec
func ExtractCredentialsAsFile(podname string, namespace string) (*ExtractedCredentials, error) {
	// TODO: Error handling here
	// It would also be nice to gather the script output that exec runs
	// instead of only getting the credentials

	/*	for r := 1; r <= apbWatchRetries; r++ {
			// err will be the return code from the exec command
			// Use the error code to determine the state
			failedToExec := errors.New("exit status 1")
			credsNotAvailable := errors.New("exit status 2")

			output, err := runtime.RunCommand(
				"kubectl",
				"exec",
				podname,
				GatherCredentialsCommand,
				"--namespace="+namespace,
			)

			// cannot exec container, pod is done
			podFailed := strings.Contains(string(output), "current phase is Failed")
			podCompleted := strings.Contains(string(output), "current phase is Succeeded") ||
				strings.Contains(string(output), "cannot exec into a container in a completed pod")

			if err == nil {
				log.Notice("[%s] Bind credentials found", podname)
				decodedOutput, err := decodeOutput(output)
				if err != nil {
					return nil, err
				}
				return buildExtractedCredentials(decodedOutput)
			} else if podFailed {
				// pod has completed but is in failed state
				return nil, fmt.Errorf("[%s] APB failed", podname)
			} else if podCompleted && err.Error() == failedToExec.Error() {
				log.Notice("[%s] APB completed", podname)
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
	*/
	k8sClient, err := clients.Kubernetes(log)
	if err != nil {
		fmt.Printf("error creating k8s client:")
		fmt.Printf("%v", err.Error())
		return
	}

	clientConfig := k8sClient.ClientConfig
	clientConfig.GroupVersion = &v1.SchemeGroupVersion
	clientConfig.NegotiatedSerializer =
		serializer.DirectCodecFactory{CodecFactory: scheme.Codecs}

	// NOTE: kubectl exec simply sets the API path to /api when where is no
	// Group, which is the case for pod exec.
	clientConfig.APIPath = "/api"

	log.Infof("%v\n", clientConfig)
	log.Infof("%s", string(debug.Stack()))

	restClient, err := rest.RESTClientFor(clientConfig)
	if err != nil {
		fmt.Printf("error creating rest client:")
		fmt.Printf("%v", err.Error())
		return
	}

	req := restClient.Post().
		Resource("pods").
		Name(podname).
		Namespace(namespace).
		SubResource("exec")

	req.VersionedParams(&v1.PodExecOptions{
		Command: GatherCredentialsCommand,
		Stdin:   false,
		Stdout:  true,
		Stderr:  true,
		TTY:     false,
	}, v1.ParameterCodec)

	exec, err := remotecommand.NewSPDYExecutor(clientConfig, "POST", req.URL())
	if err != nil {
		fmt.Printf("error getting new remotecommand executor")
		fmt.Printf("%v", err.Error())
	}

	for r := 1; r <= apbWatchRetries; r++ {
		var stdoutBuffer, stderrBuffer bytes.Buffer
		stdoutWriter := bufio.NewWriter(&stdoutBuffer)
		stderrWriter := bufio.NewWriter(&stderrBuffer)

		err = exec.Stream(remotecommand.StreamOptions{
			Stdout: stdoutWriter,
			Stderr: stderrWriter,
		})
		if err == nil {
			log.Notice("[%s] Bind credentials found", podname)
			decodedOutput, err := decodeOutput(output)
			if err != nil {
				return nil, err
			}
			return buildExtractedCredentials(decodedOutput)
		}
		//Get Pods to determine if the pod is still alive.

	}

	return nil, fmt.Errorf("[%s] ExecTimeout: Failed to gather bind credentials after %d retries", podname, apbWatchRetries)
}

// ExtractCredentialsAsSecret - Extract credentials from APB as secret in namespace.
func ExtractCredentialsAsSecret(podname string, namespace string) (*ExtractedCredentials, error) {
	k8s, err := clients.Kubernetes()
	if err != nil {
		return nil, fmt.Errorf("Unable to retrive kubernetes client - %v", err)
	}

	secret, err := k8s.GetSecretData(podname, namespace)
	if err != nil {
		return nil, fmt.Errorf("Unable to retrieve secret [ %v ] - %v", podname, err)
	}

	return buildExtractedCredentials(secret["fields"])
}

func getExtractCreds(runtimeVersion int) (extractCreds, error) {
	if runtimeVersion == 1 {
		return ExtractCredentialsAsFile, nil
	} else if runtimeVersion >= 2 {
		return ExtractCredentialsAsSecret, nil
	} else {
		return nil, fmt.Errorf(
			"Unexpected runtime version [%v], support %v <= runtimeVersion <= %v",
			runtimeVersion,
			version.MinRuntimeVersion,
			version.MaxRuntimeVersion,
		)
	}
}

func buildExtractedCredentials(output []byte) (*ExtractedCredentials, error) {

	creds := make(map[string]interface{})
	json.Unmarshal(output, &creds)

	return &ExtractedCredentials{Credentials: creds}, nil
}

func decodeOutput(output []byte) ([]byte, error) {
	str := string(output)

	decodedjson, err := base64.StdEncoding.DecodeString(str)
	if err != nil {
		return nil, err
	}

	return decodedjson, nil
}
