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
	"bufio"
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"time"

	"github.com/automationbroker/bundle-lib/clients"
	"github.com/automationbroker/bundle-lib/runtime"
	log "github.com/sirupsen/logrus"
	"k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/remotecommand"
)

const (
	// GatherCredentialsCommand - Command used when execing for bind credentials
	// moving this constant here because eventually Extrating creds will
	// need to be moved to runtime. Therefore keeping all of this together
	// makes sense
	GatherCredentialsCommand = "broker-bind-creds"
)

var (
	// ErrExtractedCredentialsNotFound - Extracted Credentials are not found.
	ErrExtractedCredentialsNotFound = fmt.Errorf("credentials not found")
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

	for r := 1; r <= apbWatchRetries; r++ {
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
			return buildExtractedCredentials(decodedOutput)
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
		time.Sleep(time.Duration(apbWatchInterval) * time.Second)
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
			1,
			2,
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

// GetExtractedCredentials - Will get the extracted credentials for a caller of the APB package.
func GetExtractedCredentials(id string) (*ExtractedCredentials, error) {
	creds, err := runtime.Provider.GetExtractedCredential(id, clusterConfig.Namespace)
	if err != nil {
		switch {
		case err == runtime.ErrCredentialsNotFound:
			log.Debugf("extracted credential secret not found - %v", id)
			return nil, ErrExtractedCredentialsNotFound
		default:
			log.Errorf("unable to get the extracted credential secret - %v", err)
			return nil, err
		}
	}
	return &ExtractedCredentials{Credentials: creds}, nil
}

// DeleteExtractedCredentials - Will delete the extracted credentials for a caller of the APB package.
// Please use this method with caution.
func DeleteExtractedCredentials(id string) error {
	return runtime.Provider.DeleteExtractedCredential(id, clusterConfig.Namespace)
}

// SetExtractedCredentials - Will set new credentials for an id.
// Please use this method with caution.
func SetExtractedCredentials(id string, creds *ExtractedCredentials) error {
	return runtime.Provider.CreateExtractedCredential(id, clusterConfig.Namespace, creds.Credentials, nil)
}
