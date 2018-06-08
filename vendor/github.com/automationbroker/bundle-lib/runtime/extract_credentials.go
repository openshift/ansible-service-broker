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
	"bufio"
	"bytes"
	"encoding/base64"
	"fmt"
	"time"

	"github.com/automationbroker/bundle-lib/clients"
	log "github.com/sirupsen/logrus"
	"k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/remotecommand"
)

const (
	// GatherCredentialsCommand - Command used when execing for bind credentials
	// moving this constant here because eventually Extracting creds will
	// need to be moved to runtime. Therefore keeping all of this together
	// makes sense
	GatherCredentialsCommand = "broker-bind-creds"
	bundleWatchInterval      = 5
	bundleWatchRetries       = 7200
)

// ExtractCredentialsFunc - the func that should be used to extract credentials
// Params:
// pod name - name of the container that the APB is running as
// namespace - name of the namespace where the container is running.
type extractCredentialsFunc func(string, string) ([]byte, error)

// ExtractCredentials - Extract credentials from pod in a certain namespace.
// needs the podname, namespace and the runtime version.
func (p provider) ExtractCredentials(podname string, ns string, runtime int) ([]byte, error) {
	extractCredsFunc, err := getExtractCreds(runtime)
	if err != nil {
		return nil, err
	}
	return extractCredsFunc(podname, ns)
}

// ExtractCredentialsAsFile - Extract credentials from running APB using exec
func extractCredentialsAsFile(podname string, namespace string) ([]byte, error) {
	k8scli, err := clients.Kubernetes()
	if err != nil {
		log.Errorf("error creating k8s client: %v", err)
		return nil, nil
	}

	clientConfig := k8scli.ClientConfig
	clientConfig.GroupVersion = &v1.SchemeGroupVersion
	clientConfig.NegotiatedSerializer =
		serializer.DirectCodecFactory{CodecFactory: scheme.Codecs}

	// NOTE: kubectl exec simply sets the API path to /api when there is no
	// Group, which is the case for pod exec.
	clientConfig.APIPath = "/api"

	restClient, err := rest.RESTClientFor(clientConfig)
	if err != nil {
		log.Errorf("error creating rest client: %v", err)
		return nil, err
	}

	req := restClient.Post().
		Resource("pods").
		Name(podname).
		Namespace(namespace).
		SubResource("exec")

	req.VersionedParams(&v1.PodExecOptions{
		Command: []string{GatherCredentialsCommand},
		Stdin:   false,
		Stdout:  true,
		Stderr:  true,
		TTY:     false,
	}, scheme.ParameterCodec)

	exec, err := remotecommand.NewSPDYExecutor(clientConfig, "POST", req.URL())
	if err != nil {
		log.Errorf("error getting new remotecommand executor - %v", err)
	}

	for r := 1; r <= bundleWatchRetries; r++ {
		var stdoutBuffer, stderrBuffer bytes.Buffer
		stdoutWriter := bufio.NewWriter(&stdoutBuffer)
		stderrWriter := bufio.NewWriter(&stderrBuffer)

		cmderr := exec.Stream(remotecommand.StreamOptions{
			Stdout: stdoutWriter,
			Stderr: stderrWriter,
		})
		if cmderr == nil {
			log.Infof("[%v] bind credentials found", podname)
			decodedOutput, err := decodeOutput(stdoutBuffer.Bytes())
			if err != nil {
				return nil, err
			}
			return decodedOutput, nil
		}
		//Get Pods to determine if the pod is still alive.
		status, err := k8scli.GetPodStatus(podname, namespace)
		if err != nil {
			//If pod can not be found then something is very wrong.
			log.Errorf("unable to find pod: %v in namespace: %v - err: %v", podname, namespace, err)
			return nil, err
		}
		switch status.Phase {
		case v1.PodFailed:
			// pod has completed but is in failed state
			log.Errorf("pod: %v in namespace: %v failed", podname, namespace)
			return nil, fmt.Errorf("[%v] APB failed", podname)
		case v1.PodSucceeded:
			log.Infof("pod: %v in namespace: %v has been completed", podname, namespace)
			return nil, nil
		default:
			log.Infof("command output: %v - err: %v", stdoutBuffer.String(), stderrBuffer.String())
			log.Infof("retry attempt: %v pod: %v in namespace: %v failed to exec into the container", r, podname, namespace)
		}
		time.Sleep(time.Duration(bundleWatchInterval) * time.Second)
	}

	return nil, fmt.Errorf("[%s] ExecTimeout: Failed to gather bind credentials after %d retries", podname, bundleWatchRetries)
}

// ExtractCredentialsAsSecret - Extract credentials from APB as secret in namespace.
func extractCredentialsAsSecret(podname string, namespace string) ([]byte, error) {
	k8s, err := clients.Kubernetes()
	if err != nil {
		return nil, fmt.Errorf("Unable to retrive kubernetes client - %v", err)
	}

	secret, err := k8s.GetSecretData(podname, namespace)
	if err != nil {
		return nil, fmt.Errorf("Unable to retrieve secret [ %v ] - %v", podname, err)
	}

	return secret["fields"], nil
}

func getExtractCreds(runtimeVersion int) (extractCredentialsFunc, error) {
	if runtimeVersion == 1 {
		log.Infof("Runtime version 1 is being deprecated.\nYou should move the Bundle to use the latest bundle base")
		return extractCredentialsAsFile, nil
	} else if runtimeVersion >= 2 {
		return extractCredentialsAsSecret, nil
	} else {
		return nil, fmt.Errorf(
			"Unexpected runtime version [%v], support %v <= runtimeVersion <= %v",
			runtimeVersion,
			1,
			2,
		)
	}
}

func decodeOutput(output []byte) ([]byte, error) {
	str := string(output)

	decodedjson, err := base64.StdEncoding.DecodeString(str)
	if err != nil {
		return nil, err
	}

	return decodedjson, nil
}
