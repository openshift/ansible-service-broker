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

// KubernetesClient - Client to interact with Kubernetes API
type KubernetesClient struct {
	Client       *clientset.Clientset
	ClientConfig *rest.Config
	log          *logging.Logger
}

// Kubernetes - Create a new kubernetes client if needed, returns reference
func Kubernetes() (*KubernetesClient, error) {
	once.Kubernetes.Do(func() {
		var log *logging.Logger
		client, err := newKubernetes(log)
		if err != nil {
			log.Error(err.Error())
			panic(err.Error())
		}
		instances.Kubernetes = client
	})
	if instances.Kubernetes == nil {
		return nil, errors.New("Kubernetes client instance is nil")
	}
	return instances.Kubernetes, nil
}

// GetSecretData - Returns the data inside of a given secret
func (k KubernetesClient) GetSecretData(secretName, namespace string) (map[string][]byte, error) {
	secretData, err := k.Client.CoreV1().Secrets(namespace).Get(secretName, meta_v1.GetOptions{})
	if err != nil {
		k.log.Errorf("Unable to load secret '%s' from namespace '%s'", secretName, namespace)
		return make(map[string][]byte), nil
	}
	k.log.Debugf("Found secret with name %v\n", secretName)

	return secretData.Data, nil
}

func createOnce(log *logging.Logger) {
	errMsg := "Something went wrong while initializing kubernetes client!\n"
	k8s, err := newKubernetes(log)
	if err != nil {
		log.Error(errMsg)
		// NOTE: Looking to leverage panic recovery to gracefully handle this
		// with things like retries or better intelligence, but the environment
		// is probably in a unrecoverable state as far as the broker is concerned,
		// and demands the attention of an operator.
		panic(err.Error())
	}

	instances.Kubernetes = k8s
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

func newKubernetes(log *logging.Logger) (*KubernetesClient, error) {
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
			return nil, err
		}
	}

	clientset, err := clientset.NewForConfig(clientConfig)
	if err != nil {
		log.Error("Failed to create LocalClientSet")
		return nil, err
	}

	k := &KubernetesClient{
		Client:       clientset,
		ClientConfig: clientConfig,
		log:          log,
	}
	return k, err
}
