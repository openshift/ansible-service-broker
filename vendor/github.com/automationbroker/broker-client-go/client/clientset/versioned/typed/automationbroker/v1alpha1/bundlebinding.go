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
package v1alpha1

import (
	scheme "github.com/automationbroker/broker-client-go/client/clientset/versioned/scheme"
	v1alpha1 "github.com/automationbroker/broker-client-go/pkg/apis/automationbroker/v1alpha1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	types "k8s.io/apimachinery/pkg/types"
	watch "k8s.io/apimachinery/pkg/watch"
	rest "k8s.io/client-go/rest"
)

// BundleBindingsGetter has a method to return a BundleBindingInterface.
// A group's client should implement this interface.
type BundleBindingsGetter interface {
	BundleBindings(namespace string) BundleBindingInterface
}

// BundleBindingInterface has methods to work with BundleBinding resources.
type BundleBindingInterface interface {
	Create(*v1alpha1.BundleBinding) (*v1alpha1.BundleBinding, error)
	Update(*v1alpha1.BundleBinding) (*v1alpha1.BundleBinding, error)
	Delete(name string, options *v1.DeleteOptions) error
	DeleteCollection(options *v1.DeleteOptions, listOptions v1.ListOptions) error
	Get(name string, options v1.GetOptions) (*v1alpha1.BundleBinding, error)
	List(opts v1.ListOptions) (*v1alpha1.BundleBindingList, error)
	Watch(opts v1.ListOptions) (watch.Interface, error)
	Patch(name string, pt types.PatchType, data []byte, subresources ...string) (result *v1alpha1.BundleBinding, err error)
	BundleBindingExpansion
}

// bundleBindings implements BundleBindingInterface
type bundleBindings struct {
	client rest.Interface
	ns     string
}

// newBundleBindings returns a BundleBindings
func newBundleBindings(c *AutomationbrokerV1alpha1Client, namespace string) *bundleBindings {
	return &bundleBindings{
		client: c.RESTClient(),
		ns:     namespace,
	}
}

// Get takes name of the bundleBinding, and returns the corresponding bundleBinding object, and an error if there is any.
func (c *bundleBindings) Get(name string, options v1.GetOptions) (result *v1alpha1.BundleBinding, err error) {
	result = &v1alpha1.BundleBinding{}
	err = c.client.Get().
		Namespace(c.ns).
		Resource("bundlebindings").
		Name(name).
		VersionedParams(&options, scheme.ParameterCodec).
		Do().
		Into(result)
	return
}

// List takes label and field selectors, and returns the list of BundleBindings that match those selectors.
func (c *bundleBindings) List(opts v1.ListOptions) (result *v1alpha1.BundleBindingList, err error) {
	result = &v1alpha1.BundleBindingList{}
	err = c.client.Get().
		Namespace(c.ns).
		Resource("bundlebindings").
		VersionedParams(&opts, scheme.ParameterCodec).
		Do().
		Into(result)
	return
}

// Watch returns a watch.Interface that watches the requested bundleBindings.
func (c *bundleBindings) Watch(opts v1.ListOptions) (watch.Interface, error) {
	opts.Watch = true
	return c.client.Get().
		Namespace(c.ns).
		Resource("bundlebindings").
		VersionedParams(&opts, scheme.ParameterCodec).
		Watch()
}

// Create takes the representation of a bundleBinding and creates it.  Returns the server's representation of the bundleBinding, and an error, if there is any.
func (c *bundleBindings) Create(bundleBinding *v1alpha1.BundleBinding) (result *v1alpha1.BundleBinding, err error) {
	result = &v1alpha1.BundleBinding{}
	err = c.client.Post().
		Namespace(c.ns).
		Resource("bundlebindings").
		Body(bundleBinding).
		Do().
		Into(result)
	return
}

// Update takes the representation of a bundleBinding and updates it. Returns the server's representation of the bundleBinding, and an error, if there is any.
func (c *bundleBindings) Update(bundleBinding *v1alpha1.BundleBinding) (result *v1alpha1.BundleBinding, err error) {
	result = &v1alpha1.BundleBinding{}
	err = c.client.Put().
		Namespace(c.ns).
		Resource("bundlebindings").
		Name(bundleBinding.Name).
		Body(bundleBinding).
		Do().
		Into(result)
	return
}

// Delete takes name of the bundleBinding and deletes it. Returns an error if one occurs.
func (c *bundleBindings) Delete(name string, options *v1.DeleteOptions) error {
	return c.client.Delete().
		Namespace(c.ns).
		Resource("bundlebindings").
		Name(name).
		Body(options).
		Do().
		Error()
}

// DeleteCollection deletes a collection of objects.
func (c *bundleBindings) DeleteCollection(options *v1.DeleteOptions, listOptions v1.ListOptions) error {
	return c.client.Delete().
		Namespace(c.ns).
		Resource("bundlebindings").
		VersionedParams(&listOptions, scheme.ParameterCodec).
		Body(options).
		Do().
		Error()
}

// Patch applies the patch and returns the patched bundleBinding.
func (c *bundleBindings) Patch(name string, pt types.PatchType, data []byte, subresources ...string) (result *v1alpha1.BundleBinding, err error) {
	result = &v1alpha1.BundleBinding{}
	err = c.client.Patch(pt).
		Namespace(c.ns).
		Resource("bundlebindings").
		SubResource(subresources...).
		Name(name).
		Body(data).
		Do().
		Into(result)
	return
}
