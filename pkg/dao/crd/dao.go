//
// Copyright (c) 2018 Red Hat, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//    http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//

package dao

import (
	"fmt"

	automationbrokerv1 "github.com/automationbroker/broker-client-go/client/clientset/versioned/typed/automationbroker.io/v1"
	v1 "github.com/automationbroker/broker-client-go/pkg/apis/automationbroker.io/v1"
	"github.com/openshift/ansible-service-broker/pkg/apb"
	"github.com/openshift/ansible-service-broker/pkg/clients"
	logutil "github.com/openshift/ansible-service-broker/pkg/util/logging"
	"github.com/pborman/uuid"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var log = logutil.NewLog()

const (
	// instanceLabel for the job state to track which instance created it.
	instanceLabel string = "instanceId"
)

// Dao - object to interface with the data store.
type Dao struct {
	client    automationbrokerv1.AutomationbrokerV1Interface
	namespace string
}

// NewDao - Create a new Dao object
func NewDao(namespace string) (*Dao, error) {
	dao := Dao{}

	crdClient, err := clients.CRDClient()
	if err != nil {
		return nil, err
	}
	dao.client = crdClient.AutomationbrokerV1()
	return &dao, nil
}

// GetSpec - Retrieve the spec from the k8s API.
func (d *Dao) GetSpec(id string) (*apb.Spec, error) {
	log.Debugf("get spec: %v", id)
	s, err := d.client.Bundles(d.namespace).Get(id, metav1.GetOptions{})
	if err != nil {
		log.Errorf("unable to get bundle from k8s api - %v", err)
		return nil, err
	}
	return bundleToSpec(s.Spec, s.GetName()), nil
}

// SetSpec - set spec for an id in the kvp API.
func (d *Dao) SetSpec(id string, spec *apb.Spec) error {
	log.Debugf("set spec: %v", id)
	bundleSpec := specToBundle(spec)
	b := v1.Bundle{
		ObjectMeta: metav1.ObjectMeta{
			Name:      id,
			Namespace: d.namespace,
		},
		Spec: bundleSpec,
	}
	_, err := d.client.Bundles(d.namespace).Create(&b)
	return err
}

// DeleteSpec - Delete the spec for a given spec id.
func (d *Dao) DeleteSpec(specID string) error {
	log.Debug("Dao::DeleteSpec-> [ %s ]", specID)
	return d.client.Bundles(d.namespace).Delete(specID, &metav1.DeleteOptions{})
}

// BatchSetSpecs - set specs based on SpecManifest in the kvp API.
func (d *Dao) BatchSetSpecs(specs apb.SpecManifest) error {
	for id, spec := range specs {
		err := d.SetSpec(id, spec)
		if err != nil {
			return err
		}
	}

	return nil
}

// BatchGetSpecs - Retrieve all the specs for dir.
func (d *Dao) BatchGetSpecs(dir string) ([]*apb.Spec, error) {
	log.Debugf("Dao::BatchGetSpecs")
	l, err := d.client.Bundles(d.namespace).List(metav1.ListOptions{})
	if err != nil {
		log.Errorf("unable to get batch specs - %v", err)
		return nil, err
	}
	specs := []*apb.Spec{}
	for _, b := range l.Items {
		specs = append(specs, bundleToSpec(b.Spec, b.GetName()))
	}
	return specs, nil
}

// BatchDeleteSpecs - set specs based on SpecManifest in the kvp API.
func (d *Dao) BatchDeleteSpecs(specs []*apb.Spec) error {
	for _, spec := range specs {
		err := d.DeleteSpec(spec.ID)
		if err != nil {
			return err
		}
	}
	return nil
}

// GetServiceInstance - Retrieve specific service instance from the kvp API.
func (d *Dao) GetServiceInstance(id string) (*apb.ServiceInstance, error) {
	log.Debugf("get service instance: %v", id)
	servInstance, err := d.client.ServiceInstances(d.namespace).Get(id, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}
	spec, err := d.GetSpec(servInstance.Spec.BundleID)
	if err != nil {
		return nil, err
	}
	return convertServiceInstanceToAPB(servInstance.Spec, spec, servInstance.GetName()), nil
}

// SetServiceInstance - Set service instance for an id in the kvp API.
func (d *Dao) SetServiceInstance(id string, serviceInstance *apb.ServiceInstance) error {
	log.Debugf("set service instance: %v", id)
	spec := convertServiceInstanceToCRD(serviceInstance)
	s := v1.ServiceInstance{
		ObjectMeta: metav1.ObjectMeta{
			Name:      id,
			Namespace: d.namespace,
		},
		Spec: spec,
	}
	_, err := d.client.ServiceInstances(d.namespace).Create(&s)
	if err != nil {
		log.Errorf("unable to save service instance - %v", err)
		return err
	}
	return nil
}

// DeleteServiceInstance - Delete the service instance for an service instance id.
func (d *Dao) DeleteServiceInstance(id string) error {
	log.Debugf("Dao::DeleteServiceInstance -> [ %s ]", id)
	return d.client.ServiceInstances(d.namespace).Delete(id, &metav1.DeleteOptions{})
}

// GetBindInstance - Retrieve a specific bind instance from the kvp API
func (d *Dao) GetBindInstance(id string) (*apb.BindInstance, error) {
	log.Debugf("get binidng instance: %v", id)
	bi, err := d.client.ServiceBindings(d.namespace).Get(id, metav1.GetOptions{})
	if err != nil {
		log.Errorf("unable to get bind instance.")
		return nil, err
	}
	return convertServiceBindingToAPB(bi.Spec, bi.GetName()), nil
}

// SetBindInstance - Set the bind instance for id in the kvp API.
func (d *Dao) SetBindInstance(id string, bindInstance *apb.BindInstance) error {
	log.Debugf("set binding instance: %v", id)
	b := convertServiceBindingToCRD(bindInstance)
	bi := v1.ServiceBinding{
		ObjectMeta: metav1.ObjectMeta{
			Name:      id,
			Namespace: d.namespace,
		},
		Spec: b,
	}
	_, err := d.client.ServiceBindings(d.namespace).Create(&bi)
	if err != nil {
		log.Errorf("unable to save service binding - %v", err)
		return err
	}
	return nil
}

// DeleteBindInstance - Delete the binding instance for an id in the kvp API.
func (d *Dao) DeleteBindInstance(id string) error {
	log.Debugf("Dao::DeleteBindInstance -> [ %s ]", id)
	err := d.client.ServiceBindings(d.namespace).Delete(id, &metav1.DeleteOptions{})
	return err
}

// SetState - Set the Job State in the kvp API for id.
func (d *Dao) SetState(instanceID string, state apb.JobState) (string, error) {
	log.Debugf("set job state for instance: %v token: %v", instanceID, state.Token)
	j := convertJobStateToCRD(&state)
	js := v1.JobState{
		ObjectMeta: metav1.ObjectMeta{
			Name:      state.Token,
			Namespace: d.namespace,
			Labels:    map[string]string{instanceLabel: instanceID},
		},
		Spec: j,
	}

	_, err := d.client.JobStates(d.namespace).Create(&js)
	if err != nil {
		log.Errorf("unable to create the job state - %v", err)
		return "", err
	}
	return state.Token, nil
}

// GetState - Retrieve a job state from the kvp API for an ID and Token.
func (d *Dao) GetState(id string, token string) (apb.JobState, error) {
	js, err := d.client.JobStates(d.namespace).Get(token, metav1.GetOptions{})
	if err != nil {
		log.Errorf("unable to get state for token: %v", err)
		return apb.JobState{}, err
	}
	j := convertJobStateToAPB(js.Spec, js.GetName())
	return *j, nil
}

// GetStateByKey - Retrieve a job state from the kvp API for a job key
func (d *Dao) GetStateByKey(key string) (apb.JobState, error) {
	return d.GetState("", key)
}

// FindJobStateByState - Retrieve all the jobs that match the specified state
func (d *Dao) FindJobStateByState(state apb.State) ([]apb.RecoverStatus, error) {
	log.Debugf("Dao::FindJobStateByState -> [%v]", state)
	jobStates, err := d.client.JobStates(d.namespace).List(metav1.ListOptions{
		FieldSelector: fmt.Sprintf("state=%v", state),
	})
	if err != nil {
		log.Errorf("unable to get job states for the state: %v - %v", state, err)
		return nil, err
	}

	rs := []apb.RecoverStatus{}
	for _, js := range jobStates.Items {
		rs = append(rs, apb.RecoverStatus{
			InstanceID: uuid.Parse(js.GetLabels()[instanceLabel]),
			State:      *convertJobStateToAPB(js.Spec, js.GetName()),
		})
	}
	return rs, nil
}

// GetSvcInstJobsByState - Lookup all jobs of a given state for a specific instance
func (d *Dao) GetSvcInstJobsByState(ID string, state apb.State) ([]apb.JobState, error) {
	log.Debugf("Dao::FindJobStateByState -> [%v]", state)
	jobStates, err := d.client.JobStates(d.namespace).List(metav1.ListOptions{
		FieldSelector: fmt.Sprintf("state=%v", state),
		LabelSelector: fmt.Sprintf("%v=%v", instanceLabel, ID),
	})
	if err != nil {
		log.Errorf("unable to get job states for the state: %v - %v", state, err)
		return nil, err
	}

	jss := []apb.JobState{}
	for _, js := range jobStates.Items {
		jss = append(jss, *convertJobStateToAPB(js.Spec, js.GetName()))
	}
	return jss, nil
}
