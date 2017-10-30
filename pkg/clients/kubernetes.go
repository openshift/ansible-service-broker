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

package clients

import (
	"errors"

	logging "github.com/op/go-logging"
	meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
	"k8s.io/kubernetes/pkg/client/clientset_generated/clientset"
)

// Kubernetes - Create a new kubernetes client if needed, returns reference
func Kubernetes(log *logging.Logger) (*clientset.Clientset, error) {
	once.Kubernetes.Do(func() { createOnce(log) })
	if instances.Kubernetes == nil {
		return nil, errors.New("Kubernetes client instance is nil")
	}
	return instances.Kubernetes, nil
}

// KubernetesConfig - Retrieve or create a new kubernetes configuration.
func KubernetesConfig(log *logging.Logger) (*rest.Config, error) {
	once.Kubernetes.Do(func() { createOnce(log) })
	if instances.KubernetesConfig == nil {
		return nil, errors.New("Kubernetes client config instance is nil")
	}
	return instances.KubernetesConfig, nil
}

// GetSecretData - Returns the data inside of a given secret
func GetSecretData(secretName, namespace string) (map[string][]byte, error) {
	var log logging.Logger
	k8scli, err := Kubernetes(&log)
	if err != nil {
		return nil, err
	}

	secretData, err := k8scli.CoreV1().Secrets(namespace).Get(secretName, meta_v1.GetOptions{})
	if err != nil {
		log.Errorf("Unable to load secret '%s' from namespace '%s'", secretName, namespace)
		return make(map[string][]byte), nil
	}
	log.Debugf("Found secret with name %v\n", secretName)

	return secretData.Data, nil
}

func createOnce(log *logging.Logger) {
	errMsg := "Something went wrong while initializing kubernetes client!\n"
	client, clientConfig, err := newKubernetes(log)
	if err != nil {
		log.Error(errMsg)
		// NOTE: Looking to leverage panic recovery to gracefully handle this
		// with things like retries or better intelligence, but the environment
		// is probably in a unrecoverable state as far as the broker is concerned,
		// and demands the attention of an operator.
		panic(err.Error())
	}
	instances.Kubernetes = client
	instances.KubernetesConfig = clientConfig
}

func createClientConfigFromFile(configPath string) (*rest.Config, error) {
	clientConfig, err := clientcmd.LoadFromFile(configPath)
	if err != nil {
		return nil, err
	}

	config, err := clientcmd.NewDefaultClientConfig(*clientConfig, &clientcmd.ConfigOverrides{}).ClientConfig()
	if err != nil {
		return nil, err
	}
	return config, nil
}

func newKubernetes(log *logging.Logger) (*clientset.Clientset, *rest.Config, error) {
	// NOTE: Both the external and internal client object are using the same
	// clientset library. Internal clientset normally uses a different
	// library
	clientConfig, err := rest.InClusterConfig()
	if err != nil {
		log.Warning("Failed to create a InternalClientSet: %v.", err)

		log.Debug("Checking for a local Cluster Config")
		clientConfig, err = createClientConfigFromFile(homedir.HomeDir() + "/.kube/config")
		if err != nil {
			log.Error("Failed to create LocalClientSet")
			return nil, nil, err
		}
	}

	clientset, err := clientset.NewForConfig(clientConfig)
	if err != nil {
		log.Error("Failed to create LocalClientSet")
		return nil, nil, err
	}

	return clientset, clientConfig, err
}
