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
	v1alpha1 "github.com/automationbroker/broker-client-go/pkg/apis/automationbroker/v1alpha1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	labels "k8s.io/apimachinery/pkg/labels"
	schema "k8s.io/apimachinery/pkg/runtime/schema"
	types "k8s.io/apimachinery/pkg/types"
	watch "k8s.io/apimachinery/pkg/watch"
	testing "k8s.io/client-go/testing"
)

// FakeBundleBindings implements BundleBindingInterface
type FakeBundleBindings struct {
	Fake *FakeAutomationbrokerV1alpha1
	ns   string
}

var bundlebindingsResource = schema.GroupVersionResource{Group: "automationbroker", Version: "v1alpha1", Resource: "bundlebindings"}

var bundlebindingsKind = schema.GroupVersionKind{Group: "automationbroker", Version: "v1alpha1", Kind: "BundleBinding"}

// Get takes name of the bundleBinding, and returns the corresponding bundleBinding object, and an error if there is any.
func (c *FakeBundleBindings) Get(name string, options v1.GetOptions) (result *v1alpha1.BundleBinding, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewGetAction(bundlebindingsResource, c.ns, name), &v1alpha1.BundleBinding{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.BundleBinding), err
}

// List takes label and field selectors, and returns the list of BundleBindings that match those selectors.
func (c *FakeBundleBindings) List(opts v1.ListOptions) (result *v1alpha1.BundleBindingList, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewListAction(bundlebindingsResource, bundlebindingsKind, c.ns, opts), &v1alpha1.BundleBindingList{})

	if obj == nil {
		return nil, err
	}

	label, _, _ := testing.ExtractFromListOptions(opts)
	if label == nil {
		label = labels.Everything()
	}
	list := &v1alpha1.BundleBindingList{}
	for _, item := range obj.(*v1alpha1.BundleBindingList).Items {
		if label.Matches(labels.Set(item.Labels)) {
			list.Items = append(list.Items, item)
		}
	}
	return list, err
}

// Watch returns a watch.Interface that watches the requested bundleBindings.
func (c *FakeBundleBindings) Watch(opts v1.ListOptions) (watch.Interface, error) {
	return c.Fake.
		InvokesWatch(testing.NewWatchAction(bundlebindingsResource, c.ns, opts))

}

// Create takes the representation of a bundleBinding and creates it.  Returns the server's representation of the bundleBinding, and an error, if there is any.
func (c *FakeBundleBindings) Create(bundleBinding *v1alpha1.BundleBinding) (result *v1alpha1.BundleBinding, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewCreateAction(bundlebindingsResource, c.ns, bundleBinding), &v1alpha1.BundleBinding{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.BundleBinding), err
}

// Update takes the representation of a bundleBinding and updates it. Returns the server's representation of the bundleBinding, and an error, if there is any.
func (c *FakeBundleBindings) Update(bundleBinding *v1alpha1.BundleBinding) (result *v1alpha1.BundleBinding, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewUpdateAction(bundlebindingsResource, c.ns, bundleBinding), &v1alpha1.BundleBinding{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.BundleBinding), err
}

// Delete takes name of the bundleBinding and deletes it. Returns an error if one occurs.
func (c *FakeBundleBindings) Delete(name string, options *v1.DeleteOptions) error {
	_, err := c.Fake.
		Invokes(testing.NewDeleteAction(bundlebindingsResource, c.ns, name), &v1alpha1.BundleBinding{})

	return err
}

// DeleteCollection deletes a collection of objects.
func (c *FakeBundleBindings) DeleteCollection(options *v1.DeleteOptions, listOptions v1.ListOptions) error {
	action := testing.NewDeleteCollectionAction(bundlebindingsResource, c.ns, listOptions)

	_, err := c.Fake.Invokes(action, &v1alpha1.BundleBindingList{})
	return err
}

// Patch applies the patch and returns the patched bundleBinding.
func (c *FakeBundleBindings) Patch(name string, pt types.PatchType, data []byte, subresources ...string) (result *v1alpha1.BundleBinding, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewPatchSubresourceAction(bundlebindingsResource, c.ns, name, data, subresources...), &v1alpha1.BundleBinding{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.BundleBinding), err
}
