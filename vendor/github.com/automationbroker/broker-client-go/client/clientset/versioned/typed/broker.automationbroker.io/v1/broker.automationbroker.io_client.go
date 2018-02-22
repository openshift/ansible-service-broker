/*
Copyright 2018 The Openshift Evangelists

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
package v1

import (
	rest "k8s.io/client-go/rest"
	"github.com/automationbroker/broker-client-go/client/clientset/versioned/scheme"
	v1 "github.com/automationbroker/broker-client-go/pkg/apis/broker.automationbroker.io/v1"
	serializer "k8s.io/apimachinery/pkg/runtime/serializer"
)


type Broker.automationbroker.ioV1Interface interface {
    RESTClient() rest.Interface
     BundlesGetter
     JobStatesGetter
     ServiceBindingsGetter
     ServiceInstancesGetter
    
}

// Broker.automationbroker.ioV1Client is used to interact with features provided by the broker.automationbroker.io group.
type Broker.automationbroker.ioV1Client struct {
	restClient rest.Interface
}

func (c *Broker.automationbroker.ioV1Client) Bundles(namespace string) BundleInterface {
	return newBundles(c, namespace)
}

func (c *Broker.automationbroker.ioV1Client) JobStates(namespace string) JobStateInterface {
	return newJobStates(c, namespace)
}

func (c *Broker.automationbroker.ioV1Client) ServiceBindings(namespace string) ServiceBindingInterface {
	return newServiceBindings(c, namespace)
}

func (c *Broker.automationbroker.ioV1Client) ServiceInstances(namespace string) ServiceInstanceInterface {
	return newServiceInstances(c, namespace)
}

// NewForConfig creates a new Broker.automationbroker.ioV1Client for the given config.
func NewForConfig(c *rest.Config) (*Broker.automationbroker.ioV1Client, error) {
	config := *c
	if err := setConfigDefaults(&config); err != nil {
		return nil, err
	}
	client, err := rest.RESTClientFor(&config)
	if err != nil {
		return nil, err
	}
	return &Broker.automationbroker.ioV1Client{client}, nil
}

// NewForConfigOrDie creates a new Broker.automationbroker.ioV1Client for the given config and
// panics if there is an error in the config.
func NewForConfigOrDie(c *rest.Config) *Broker.automationbroker.ioV1Client {
	client, err := NewForConfig(c)
	if err != nil {
		panic(err)
	}
	return client
}

// New creates a new Broker.automationbroker.ioV1Client for the given RESTClient.
func New(c rest.Interface) *Broker.automationbroker.ioV1Client {
	return &Broker.automationbroker.ioV1Client{c}
}

func setConfigDefaults(config *rest.Config) error {
	gv := v1.SchemeGroupVersion
	config.GroupVersion =  &gv
	config.APIPath = "/apis"
	config.NegotiatedSerializer = serializer.DirectCodecFactory{CodecFactory: scheme.Codecs}

	if config.UserAgent == "" {
		config.UserAgent = rest.DefaultKubernetesUserAgent()
	}

	return nil
}

// RESTClient returns a RESTClient that is used to communicate
// with API server by this client implementation.
func (c *Broker.automationbroker.ioV1Client) RESTClient() rest.Interface {
	if c == nil {
		return nil
	}
	return c.restClient
}
