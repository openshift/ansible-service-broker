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

// FakeServiceBindings implements ServiceBindingInterface
type FakeServiceBindings struct {
	Fake *FakeAutomationbrokerV1
	ns   string
}

var servicebindingsResource = schema.GroupVersionResource{Group: "automationbroker.io", Version: "v1", Resource: "servicebindings"}

var servicebindingsKind = schema.GroupVersionKind{Group: "automationbroker.io", Version: "v1", Kind: "ServiceBinding"}

// Get takes name of the serviceBinding, and returns the corresponding serviceBinding object, and an error if there is any.
func (c *FakeServiceBindings) Get(name string, options v1.GetOptions) (result *automationbroker_io_v1.ServiceBinding, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewGetAction(servicebindingsResource, c.ns, name), &automationbroker_io_v1.ServiceBinding{})

	if obj == nil {
		return nil, err
	}
	return obj.(*automationbroker_io_v1.ServiceBinding), err
}

// List takes label and field selectors, and returns the list of ServiceBindings that match those selectors.
func (c *FakeServiceBindings) List(opts v1.ListOptions) (result *automationbroker_io_v1.ServiceBindingList, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewListAction(servicebindingsResource, servicebindingsKind, c.ns, opts), &automationbroker_io_v1.ServiceBindingList{})

	if obj == nil {
		return nil, err
	}

	label, _, _ := testing.ExtractFromListOptions(opts)
	if label == nil {
		label = labels.Everything()
	}
	list := &automationbroker_io_v1.ServiceBindingList{}
	for _, item := range obj.(*automationbroker_io_v1.ServiceBindingList).Items {
		if label.Matches(labels.Set(item.Labels)) {
			list.Items = append(list.Items, item)
		}
	}
	return list, err
}

// Watch returns a watch.Interface that watches the requested serviceBindings.
func (c *FakeServiceBindings) Watch(opts v1.ListOptions) (watch.Interface, error) {
	return c.Fake.
		InvokesWatch(testing.NewWatchAction(servicebindingsResource, c.ns, opts))

}

// Create takes the representation of a serviceBinding and creates it.  Returns the server's representation of the serviceBinding, and an error, if there is any.
func (c *FakeServiceBindings) Create(serviceBinding *automationbroker_io_v1.ServiceBinding) (result *automationbroker_io_v1.ServiceBinding, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewCreateAction(servicebindingsResource, c.ns, serviceBinding), &automationbroker_io_v1.ServiceBinding{})

	if obj == nil {
		return nil, err
	}
	return obj.(*automationbroker_io_v1.ServiceBinding), err
}

// Update takes the representation of a serviceBinding and updates it. Returns the server's representation of the serviceBinding, and an error, if there is any.
func (c *FakeServiceBindings) Update(serviceBinding *automationbroker_io_v1.ServiceBinding) (result *automationbroker_io_v1.ServiceBinding, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewUpdateAction(servicebindingsResource, c.ns, serviceBinding), &automationbroker_io_v1.ServiceBinding{})

	if obj == nil {
		return nil, err
	}
	return obj.(*automationbroker_io_v1.ServiceBinding), err
}

// Delete takes name of the serviceBinding and deletes it. Returns an error if one occurs.
func (c *FakeServiceBindings) Delete(name string, options *v1.DeleteOptions) error {
	_, err := c.Fake.
		Invokes(testing.NewDeleteAction(servicebindingsResource, c.ns, name), &automationbroker_io_v1.ServiceBinding{})

	return err
}

// DeleteCollection deletes a collection of objects.
func (c *FakeServiceBindings) DeleteCollection(options *v1.DeleteOptions, listOptions v1.ListOptions) error {
	action := testing.NewDeleteCollectionAction(servicebindingsResource, c.ns, listOptions)

	_, err := c.Fake.Invokes(action, &automationbroker_io_v1.ServiceBindingList{})
	return err
}

// Patch applies the patch and returns the patched serviceBinding.
func (c *FakeServiceBindings) Patch(name string, pt types.PatchType, data []byte, subresources ...string) (result *automationbroker_io_v1.ServiceBinding, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewPatchSubresourceAction(servicebindingsResource, c.ns, name, data, subresources...), &automationbroker_io_v1.ServiceBinding{})

	if obj == nil {
		return nil, err
	}
	return obj.(*automationbroker_io_v1.ServiceBinding), err
}
