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

// FakeJobStates implements JobStateInterface
type FakeJobStates struct {
	Fake *FakeAutomationbrokerV1
	ns   string
}

var jobstatesResource = schema.GroupVersionResource{Group: "automationbroker.io", Version: "v1", Resource: "jobstates"}

var jobstatesKind = schema.GroupVersionKind{Group: "automationbroker.io", Version: "v1", Kind: "JobState"}

// Get takes name of the jobState, and returns the corresponding jobState object, and an error if there is any.
func (c *FakeJobStates) Get(name string, options v1.GetOptions) (result *automationbroker_io_v1.JobState, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewGetAction(jobstatesResource, c.ns, name), &automationbroker_io_v1.JobState{})

	if obj == nil {
		return nil, err
	}
	return obj.(*automationbroker_io_v1.JobState), err
}

// List takes label and field selectors, and returns the list of JobStates that match those selectors.
func (c *FakeJobStates) List(opts v1.ListOptions) (result *automationbroker_io_v1.JobStateList, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewListAction(jobstatesResource, jobstatesKind, c.ns, opts), &automationbroker_io_v1.JobStateList{})

	if obj == nil {
		return nil, err
	}

	label, _, _ := testing.ExtractFromListOptions(opts)
	if label == nil {
		label = labels.Everything()
	}
	list := &automationbroker_io_v1.JobStateList{}
	for _, item := range obj.(*automationbroker_io_v1.JobStateList).Items {
		if label.Matches(labels.Set(item.Labels)) {
			list.Items = append(list.Items, item)
		}
	}
	return list, err
}

// Watch returns a watch.Interface that watches the requested jobStates.
func (c *FakeJobStates) Watch(opts v1.ListOptions) (watch.Interface, error) {
	return c.Fake.
		InvokesWatch(testing.NewWatchAction(jobstatesResource, c.ns, opts))

}

// Create takes the representation of a jobState and creates it.  Returns the server's representation of the jobState, and an error, if there is any.
func (c *FakeJobStates) Create(jobState *automationbroker_io_v1.JobState) (result *automationbroker_io_v1.JobState, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewCreateAction(jobstatesResource, c.ns, jobState), &automationbroker_io_v1.JobState{})

	if obj == nil {
		return nil, err
	}
	return obj.(*automationbroker_io_v1.JobState), err
}

// Update takes the representation of a jobState and updates it. Returns the server's representation of the jobState, and an error, if there is any.
func (c *FakeJobStates) Update(jobState *automationbroker_io_v1.JobState) (result *automationbroker_io_v1.JobState, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewUpdateAction(jobstatesResource, c.ns, jobState), &automationbroker_io_v1.JobState{})

	if obj == nil {
		return nil, err
	}
	return obj.(*automationbroker_io_v1.JobState), err
}

// Delete takes name of the jobState and deletes it. Returns an error if one occurs.
func (c *FakeJobStates) Delete(name string, options *v1.DeleteOptions) error {
	_, err := c.Fake.
		Invokes(testing.NewDeleteAction(jobstatesResource, c.ns, name), &automationbroker_io_v1.JobState{})

	return err
}

// DeleteCollection deletes a collection of objects.
func (c *FakeJobStates) DeleteCollection(options *v1.DeleteOptions, listOptions v1.ListOptions) error {
	action := testing.NewDeleteCollectionAction(jobstatesResource, c.ns, listOptions)

	_, err := c.Fake.Invokes(action, &automationbroker_io_v1.JobStateList{})
	return err
}

// Patch applies the patch and returns the patched jobState.
func (c *FakeJobStates) Patch(name string, pt types.PatchType, data []byte, subresources ...string) (result *automationbroker_io_v1.JobState, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewPatchSubresourceAction(jobstatesResource, c.ns, name, data, subresources...), &automationbroker_io_v1.JobState{})

	if obj == nil {
		return nil, err
	}
	return obj.(*automationbroker_io_v1.JobState), err
}
