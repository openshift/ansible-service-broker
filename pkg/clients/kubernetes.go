package clients

import (
	"errors"
	logging "github.com/op/go-logging"
	restclient "k8s.io/client-go/rest"

	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
	"k8s.io/kubernetes/pkg/client/clientset_generated/clientset"
)

// Kubernetes - Create a new kubernetes client if needed, returns reference
func Kubernetes(log *logging.Logger) (*clientset.Clientset, error) {
	errMsg := "Something went wrong while initializing kubernetes client!"
	once.Kubernetes.Do(func() {
		client, err := newKubernetes(log)
		if err != nil {
			log.Error(errMsg)
			log.Error(err.Error())
			instances.Kubernetes = clientResult{nil, err}
		}
		instances.Kubernetes = clientResult{client, nil}
	})

	err := instances.Kubernetes.err
	if err != nil {
		log.Error(errMsg)
		log.Error(err.Error())
		return nil, err
	}

	if client, ok := instances.Kubernetes.client.(*clientset.Clientset); ok {
		return client, nil
	} else {
		return nil, errors.New(errMsg)
	}
}

func createClientConfigFromFile(configPath string) (*restclient.Config, error) {
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

func newKubernetes(log *logging.Logger) (*clientset.Clientset, error) {
	// NOTE: Both the external and internal client object are using the same
	// clientset library. Internal clientset normally uses a different
	// library
	clientConfig, err := restclient.InClusterConfig()
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

	return clientset, err
}
