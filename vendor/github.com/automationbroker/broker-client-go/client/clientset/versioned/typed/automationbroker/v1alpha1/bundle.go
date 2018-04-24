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

// BundlesGetter has a method to return a BundleInterface.
// A group's client should implement this interface.
type BundlesGetter interface {
	Bundles(namespace string) BundleInterface
}

// BundleInterface has methods to work with Bundle resources.
type BundleInterface interface {
	Create(*v1alpha1.Bundle) (*v1alpha1.Bundle, error)
	Update(*v1alpha1.Bundle) (*v1alpha1.Bundle, error)
	Delete(name string, options *v1.DeleteOptions) error
	DeleteCollection(options *v1.DeleteOptions, listOptions v1.ListOptions) error
	Get(name string, options v1.GetOptions) (*v1alpha1.Bundle, error)
	List(opts v1.ListOptions) (*v1alpha1.BundleList, error)
	Watch(opts v1.ListOptions) (watch.Interface, error)
	Patch(name string, pt types.PatchType, data []byte, subresources ...string) (result *v1alpha1.Bundle, err error)
	BundleExpansion
}

// bundles implements BundleInterface
type bundles struct {
	client rest.Interface
	ns     string
}

// newBundles returns a Bundles
func newBundles(c *AutomationbrokerV1alpha1Client, namespace string) *bundles {
	return &bundles{
		client: c.RESTClient(),
		ns:     namespace,
	}
}

// Get takes name of the bundle, and returns the corresponding bundle object, and an error if there is any.
func (c *bundles) Get(name string, options v1.GetOptions) (result *v1alpha1.Bundle, err error) {
	result = &v1alpha1.Bundle{}
	err = c.client.Get().
		Namespace(c.ns).
		Resource("bundles").
		Name(name).
		VersionedParams(&options, scheme.ParameterCodec).
		Do().
		Into(result)
	return
}

// List takes label and field selectors, and returns the list of Bundles that match those selectors.
func (c *bundles) List(opts v1.ListOptions) (result *v1alpha1.BundleList, err error) {
	result = &v1alpha1.BundleList{}
	err = c.client.Get().
		Namespace(c.ns).
		Resource("bundles").
		VersionedParams(&opts, scheme.ParameterCodec).
		Do().
		Into(result)
	return
}

// Watch returns a watch.Interface that watches the requested bundles.
func (c *bundles) Watch(opts v1.ListOptions) (watch.Interface, error) {
	opts.Watch = true
	return c.client.Get().
		Namespace(c.ns).
		Resource("bundles").
		VersionedParams(&opts, scheme.ParameterCodec).
		Watch()
}

// Create takes the representation of a bundle and creates it.  Returns the server's representation of the bundle, and an error, if there is any.
func (c *bundles) Create(bundle *v1alpha1.Bundle) (result *v1alpha1.Bundle, err error) {
	result = &v1alpha1.Bundle{}
	err = c.client.Post().
		Namespace(c.ns).
		Resource("bundles").
		Body(bundle).
		Do().
		Into(result)
	return
}

// Update takes the representation of a bundle and updates it. Returns the server's representation of the bundle, and an error, if there is any.
func (c *bundles) Update(bundle *v1alpha1.Bundle) (result *v1alpha1.Bundle, err error) {
	result = &v1alpha1.Bundle{}
	err = c.client.Put().
		Namespace(c.ns).
		Resource("bundles").
		Name(bundle.Name).
		Body(bundle).
		Do().
		Into(result)
	return
}

// Delete takes name of the bundle and deletes it. Returns an error if one occurs.
func (c *bundles) Delete(name string, options *v1.DeleteOptions) error {
	return c.client.Delete().
		Namespace(c.ns).
		Resource("bundles").
		Name(name).
		Body(options).
		Do().
		Error()
}

// DeleteCollection deletes a collection of objects.
func (c *bundles) DeleteCollection(options *v1.DeleteOptions, listOptions v1.ListOptions) error {
	return c.client.Delete().
		Namespace(c.ns).
		Resource("bundles").
		VersionedParams(&listOptions, scheme.ParameterCodec).
		Body(options).
		Do().
		Error()
}

// Patch applies the patch and returns the patched bundle.
func (c *bundles) Patch(name string, pt types.PatchType, data []byte, subresources ...string) (result *v1alpha1.Bundle, err error) {
	result = &v1alpha1.Bundle{}
	err = c.client.Patch(pt).
		Namespace(c.ns).
		Resource("bundles").
		SubResource(subresources...).
		Name(name).
		Body(data).
		Do().
		Into(result)
	return
}
