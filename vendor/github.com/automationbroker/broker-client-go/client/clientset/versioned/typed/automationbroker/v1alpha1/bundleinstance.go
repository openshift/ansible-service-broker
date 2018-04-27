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

// BundleInstancesGetter has a method to return a BundleInstanceInterface.
// A group's client should implement this interface.
type BundleInstancesGetter interface {
	BundleInstances(namespace string) BundleInstanceInterface
}

// BundleInstanceInterface has methods to work with BundleInstance resources.
type BundleInstanceInterface interface {
	Create(*v1alpha1.BundleInstance) (*v1alpha1.BundleInstance, error)
	Update(*v1alpha1.BundleInstance) (*v1alpha1.BundleInstance, error)
	Delete(name string, options *v1.DeleteOptions) error
	DeleteCollection(options *v1.DeleteOptions, listOptions v1.ListOptions) error
	Get(name string, options v1.GetOptions) (*v1alpha1.BundleInstance, error)
	List(opts v1.ListOptions) (*v1alpha1.BundleInstanceList, error)
	Watch(opts v1.ListOptions) (watch.Interface, error)
	Patch(name string, pt types.PatchType, data []byte, subresources ...string) (result *v1alpha1.BundleInstance, err error)
	BundleInstanceExpansion
}

// bundleInstances implements BundleInstanceInterface
type bundleInstances struct {
	client rest.Interface
	ns     string
}

// newBundleInstances returns a BundleInstances
func newBundleInstances(c *AutomationbrokerV1alpha1Client, namespace string) *bundleInstances {
	return &bundleInstances{
		client: c.RESTClient(),
		ns:     namespace,
	}
}

// Get takes name of the bundleInstance, and returns the corresponding bundleInstance object, and an error if there is any.
func (c *bundleInstances) Get(name string, options v1.GetOptions) (result *v1alpha1.BundleInstance, err error) {
	result = &v1alpha1.BundleInstance{}
	err = c.client.Get().
		Namespace(c.ns).
		Resource("bundleinstances").
		Name(name).
		VersionedParams(&options, scheme.ParameterCodec).
		Do().
		Into(result)
	return
}

// List takes label and field selectors, and returns the list of BundleInstances that match those selectors.
func (c *bundleInstances) List(opts v1.ListOptions) (result *v1alpha1.BundleInstanceList, err error) {
	result = &v1alpha1.BundleInstanceList{}
	err = c.client.Get().
		Namespace(c.ns).
		Resource("bundleinstances").
		VersionedParams(&opts, scheme.ParameterCodec).
		Do().
		Into(result)
	return
}

// Watch returns a watch.Interface that watches the requested bundleInstances.
func (c *bundleInstances) Watch(opts v1.ListOptions) (watch.Interface, error) {
	opts.Watch = true
	return c.client.Get().
		Namespace(c.ns).
		Resource("bundleinstances").
		VersionedParams(&opts, scheme.ParameterCodec).
		Watch()
}

// Create takes the representation of a bundleInstance and creates it.  Returns the server's representation of the bundleInstance, and an error, if there is any.
func (c *bundleInstances) Create(bundleInstance *v1alpha1.BundleInstance) (result *v1alpha1.BundleInstance, err error) {
	result = &v1alpha1.BundleInstance{}
	err = c.client.Post().
		Namespace(c.ns).
		Resource("bundleinstances").
		Body(bundleInstance).
		Do().
		Into(result)
	return
}

// Update takes the representation of a bundleInstance and updates it. Returns the server's representation of the bundleInstance, and an error, if there is any.
func (c *bundleInstances) Update(bundleInstance *v1alpha1.BundleInstance) (result *v1alpha1.BundleInstance, err error) {
	result = &v1alpha1.BundleInstance{}
	err = c.client.Put().
		Namespace(c.ns).
		Resource("bundleinstances").
		Name(bundleInstance.Name).
		Body(bundleInstance).
		Do().
		Into(result)
	return
}

// Delete takes name of the bundleInstance and deletes it. Returns an error if one occurs.
func (c *bundleInstances) Delete(name string, options *v1.DeleteOptions) error {
	return c.client.Delete().
		Namespace(c.ns).
		Resource("bundleinstances").
		Name(name).
		Body(options).
		Do().
		Error()
}

// DeleteCollection deletes a collection of objects.
func (c *bundleInstances) DeleteCollection(options *v1.DeleteOptions, listOptions v1.ListOptions) error {
	return c.client.Delete().
		Namespace(c.ns).
		Resource("bundleinstances").
		VersionedParams(&listOptions, scheme.ParameterCodec).
		Body(options).
		Do().
		Error()
}

// Patch applies the patch and returns the patched bundleInstance.
func (c *bundleInstances) Patch(name string, pt types.PatchType, data []byte, subresources ...string) (result *v1alpha1.BundleInstance, err error) {
	result = &v1alpha1.BundleInstance{}
	err = c.client.Patch(pt).
		Namespace(c.ns).
		Resource("bundleinstances").
		SubResource(subresources...).
		Name(name).
		Body(data).
		Do().
		Into(result)
	return
}
