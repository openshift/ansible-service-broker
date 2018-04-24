/*
Copyright (c) 2018 Red Hat, Inc.

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
	v1alpha1 "github.com/automationbroker/broker-client-go/client/clientset/versioned/typed/automationbroker/v1alpha1"
	rest "k8s.io/client-go/rest"
	testing "k8s.io/client-go/testing"
)

type FakeAutomationbrokerV1alpha1 struct {
	*testing.Fake
}

func (c *FakeAutomationbrokerV1alpha1) Bundles(namespace string) v1alpha1.BundleInterface {
	return &FakeBundles{c, namespace}
}

func (c *FakeAutomationbrokerV1alpha1) BundleBindings(namespace string) v1alpha1.BundleBindingInterface {
	return &FakeBundleBindings{c, namespace}
}

func (c *FakeAutomationbrokerV1alpha1) BundleInstances(namespace string) v1alpha1.BundleInstanceInterface {
	return &FakeBundleInstances{c, namespace}
}

// RESTClient returns a RESTClient that is used to communicate
// with API server by this client implementation.
func (c *FakeAutomationbrokerV1alpha1) RESTClient() rest.Interface {
	var ret *rest.RESTClient
	return ret
}
