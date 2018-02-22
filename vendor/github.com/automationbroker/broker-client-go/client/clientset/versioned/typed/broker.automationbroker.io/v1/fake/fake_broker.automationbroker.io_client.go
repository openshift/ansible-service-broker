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
package fake

import (
	rest "k8s.io/client-go/rest"
	testing "k8s.io/client-go/testing"
	v1 "github.com/automationbroker/broker-client-go/client/clientset/versioned/typed/broker.automationbroker.io/v1"
)


type FakeBroker.automationbroker.ioV1 struct {
	*testing.Fake
}

func (c *FakeBroker.automationbroker.ioV1) Bundles(namespace string) v1.BundleInterface {
	return &FakeBundles{c, namespace}
}

func (c *FakeBroker.automationbroker.ioV1) JobStates(namespace string) v1.JobStateInterface {
	return &FakeJobStates{c, namespace}
}

func (c *FakeBroker.automationbroker.ioV1) ServiceBindings(namespace string) v1.ServiceBindingInterface {
	return &FakeServiceBindings{c, namespace}
}

func (c *FakeBroker.automationbroker.ioV1) ServiceInstances(namespace string) v1.ServiceInstanceInterface {
	return &FakeServiceInstances{c, namespace}
}

// RESTClient returns a RESTClient that is used to communicate
// with API server by this client implementation.
func (c *FakeBroker.automationbroker.ioV1) RESTClient() rest.Interface {
	var ret *rest.RESTClient
	return ret
}
