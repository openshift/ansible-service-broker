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

// FakeBundleInstances implements BundleInstanceInterface
type FakeBundleInstances struct {
	Fake *FakeAutomationbrokerV1alpha1
	ns   string
}

var bundleinstancesResource = schema.GroupVersionResource{Group: "automationbroker", Version: "v1alpha1", Resource: "bundleinstances"}

var bundleinstancesKind = schema.GroupVersionKind{Group: "automationbroker", Version: "v1alpha1", Kind: "BundleInstance"}

// Get takes name of the bundleInstance, and returns the corresponding bundleInstance object, and an error if there is any.
func (c *FakeBundleInstances) Get(name string, options v1.GetOptions) (result *v1alpha1.BundleInstance, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewGetAction(bundleinstancesResource, c.ns, name), &v1alpha1.BundleInstance{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.BundleInstance), err
}

// List takes label and field selectors, and returns the list of BundleInstances that match those selectors.
func (c *FakeBundleInstances) List(opts v1.ListOptions) (result *v1alpha1.BundleInstanceList, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewListAction(bundleinstancesResource, bundleinstancesKind, c.ns, opts), &v1alpha1.BundleInstanceList{})

	if obj == nil {
		return nil, err
	}

	label, _, _ := testing.ExtractFromListOptions(opts)
	if label == nil {
		label = labels.Everything()
	}
	list := &v1alpha1.BundleInstanceList{}
	for _, item := range obj.(*v1alpha1.BundleInstanceList).Items {
		if label.Matches(labels.Set(item.Labels)) {
			list.Items = append(list.Items, item)
		}
	}
	return list, err
}

// Watch returns a watch.Interface that watches the requested bundleInstances.
func (c *FakeBundleInstances) Watch(opts v1.ListOptions) (watch.Interface, error) {
	return c.Fake.
		InvokesWatch(testing.NewWatchAction(bundleinstancesResource, c.ns, opts))

}

// Create takes the representation of a bundleInstance and creates it.  Returns the server's representation of the bundleInstance, and an error, if there is any.
func (c *FakeBundleInstances) Create(bundleInstance *v1alpha1.BundleInstance) (result *v1alpha1.BundleInstance, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewCreateAction(bundleinstancesResource, c.ns, bundleInstance), &v1alpha1.BundleInstance{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.BundleInstance), err
}

// Update takes the representation of a bundleInstance and updates it. Returns the server's representation of the bundleInstance, and an error, if there is any.
func (c *FakeBundleInstances) Update(bundleInstance *v1alpha1.BundleInstance) (result *v1alpha1.BundleInstance, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewUpdateAction(bundleinstancesResource, c.ns, bundleInstance), &v1alpha1.BundleInstance{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.BundleInstance), err
}

// Delete takes name of the bundleInstance and deletes it. Returns an error if one occurs.
func (c *FakeBundleInstances) Delete(name string, options *v1.DeleteOptions) error {
	_, err := c.Fake.
		Invokes(testing.NewDeleteAction(bundleinstancesResource, c.ns, name), &v1alpha1.BundleInstance{})

	return err
}

// DeleteCollection deletes a collection of objects.
func (c *FakeBundleInstances) DeleteCollection(options *v1.DeleteOptions, listOptions v1.ListOptions) error {
	action := testing.NewDeleteCollectionAction(bundleinstancesResource, c.ns, listOptions)

	_, err := c.Fake.Invokes(action, &v1alpha1.BundleInstanceList{})
	return err
}

// Patch applies the patch and returns the patched bundleInstance.
func (c *FakeBundleInstances) Patch(name string, pt types.PatchType, data []byte, subresources ...string) (result *v1alpha1.BundleInstance, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewPatchSubresourceAction(bundleinstancesResource, c.ns, name, data, subresources...), &v1alpha1.BundleInstance{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.BundleInstance), err
}
