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
	automationbroker_io_v1 "github.com/automationbroker/broker-client-go/pkg/apis/automationbroker.io/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	labels "k8s.io/apimachinery/pkg/labels"
	schema "k8s.io/apimachinery/pkg/runtime/schema"
	types "k8s.io/apimachinery/pkg/types"
	watch "k8s.io/apimachinery/pkg/watch"
	testing "k8s.io/client-go/testing"
)

// FakeServiceInstances implements ServiceInstanceInterface
type FakeServiceInstances struct {
	Fake *FakeAutomationbrokerV1
	ns   string
}

var serviceinstancesResource = schema.GroupVersionResource{Group: "automationbroker.io", Version: "v1", Resource: "serviceinstances"}

var serviceinstancesKind = schema.GroupVersionKind{Group: "automationbroker.io", Version: "v1", Kind: "ServiceInstance"}

// Get takes name of the serviceInstance, and returns the corresponding serviceInstance object, and an error if there is any.
func (c *FakeServiceInstances) Get(name string, options v1.GetOptions) (result *automationbroker_io_v1.ServiceInstance, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewGetAction(serviceinstancesResource, c.ns, name), &automationbroker_io_v1.ServiceInstance{})

	if obj == nil {
		return nil, err
	}
	return obj.(*automationbroker_io_v1.ServiceInstance), err
}

// List takes label and field selectors, and returns the list of ServiceInstances that match those selectors.
func (c *FakeServiceInstances) List(opts v1.ListOptions) (result *automationbroker_io_v1.ServiceInstanceList, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewListAction(serviceinstancesResource, serviceinstancesKind, c.ns, opts), &automationbroker_io_v1.ServiceInstanceList{})

	if obj == nil {
		return nil, err
	}

	label, _, _ := testing.ExtractFromListOptions(opts)
	if label == nil {
		label = labels.Everything()
	}
	list := &automationbroker_io_v1.ServiceInstanceList{}
	for _, item := range obj.(*automationbroker_io_v1.ServiceInstanceList).Items {
		if label.Matches(labels.Set(item.Labels)) {
			list.Items = append(list.Items, item)
		}
	}
	return list, err
}

// Watch returns a watch.Interface that watches the requested serviceInstances.
func (c *FakeServiceInstances) Watch(opts v1.ListOptions) (watch.Interface, error) {
	return c.Fake.
		InvokesWatch(testing.NewWatchAction(serviceinstancesResource, c.ns, opts))

}

// Create takes the representation of a serviceInstance and creates it.  Returns the server's representation of the serviceInstance, and an error, if there is any.
func (c *FakeServiceInstances) Create(serviceInstance *automationbroker_io_v1.ServiceInstance) (result *automationbroker_io_v1.ServiceInstance, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewCreateAction(serviceinstancesResource, c.ns, serviceInstance), &automationbroker_io_v1.ServiceInstance{})

	if obj == nil {
		return nil, err
	}
	return obj.(*automationbroker_io_v1.ServiceInstance), err
}

// Update takes the representation of a serviceInstance and updates it. Returns the server's representation of the serviceInstance, and an error, if there is any.
func (c *FakeServiceInstances) Update(serviceInstance *automationbroker_io_v1.ServiceInstance) (result *automationbroker_io_v1.ServiceInstance, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewUpdateAction(serviceinstancesResource, c.ns, serviceInstance), &automationbroker_io_v1.ServiceInstance{})

	if obj == nil {
		return nil, err
	}
	return obj.(*automationbroker_io_v1.ServiceInstance), err
}

// Delete takes name of the serviceInstance and deletes it. Returns an error if one occurs.
func (c *FakeServiceInstances) Delete(name string, options *v1.DeleteOptions) error {
	_, err := c.Fake.
		Invokes(testing.NewDeleteAction(serviceinstancesResource, c.ns, name), &automationbroker_io_v1.ServiceInstance{})

	return err
}

// DeleteCollection deletes a collection of objects.
func (c *FakeServiceInstances) DeleteCollection(options *v1.DeleteOptions, listOptions v1.ListOptions) error {
	action := testing.NewDeleteCollectionAction(serviceinstancesResource, c.ns, listOptions)

	_, err := c.Fake.Invokes(action, &automationbroker_io_v1.ServiceInstanceList{})
	return err
}

// Patch applies the patch and returns the patched serviceInstance.
func (c *FakeServiceInstances) Patch(name string, pt types.PatchType, data []byte, subresources ...string) (result *automationbroker_io_v1.ServiceInstance, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewPatchSubresourceAction(serviceinstancesResource, c.ns, name, data, subresources...), &automationbroker_io_v1.ServiceInstance{})

	if obj == nil {
		return nil, err
	}
	return obj.(*automationbroker_io_v1.ServiceInstance), err
}
