package clients

import (
	"errors"

	clientset "github.com/automationbroker/broker-client-go/client/clientset/versioned"
	log "github.com/sirupsen/logrus"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/util/homedir"
)

// CRD - Client to interact with automationbroker crd API
type CRD struct {
	clientset.Interface
}

// CRDClient - Create a new kubernetes client if needed, returns reference
func CRDClient() (*CRD, error) {
	once.CRD.Do(func() {
		client, err := newCRDClient()
		if err != nil {
			log.Error(err.Error())
			panic(err.Error())
		}
		instances.CRD = client
	})
	if instances.CRD == nil {
		return nil, errors.New("CRDClient client instance is nil")
	}
	return instances.CRD, nil
}

func newCRDClient() (*CRD, error) {
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
	c := &CRD{clientset}
	return c, err
}
