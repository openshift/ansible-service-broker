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
	"net/http"
	"sync"

	automationbrokerv1 "github.com/automationbroker/broker-client-go/client/clientset/versioned/typed/automationbroker/v1alpha1"
	v1 "github.com/automationbroker/broker-client-go/pkg/apis/automationbroker/v1alpha1"
	"github.com/automationbroker/bundle-lib/apb"
	"github.com/automationbroker/bundle-lib/bundle"
	"github.com/automationbroker/bundle-lib/clients"
	"github.com/automationbroker/bundle-lib/crd"
	"github.com/pborman/uuid"
	log "github.com/sirupsen/logrus"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type arrayErrors []error

func (a arrayErrors) Error() string {
	return fmt.Sprintf("%#v", a)
}

const (
	// instanceLabel for the job state to track which instance created it.
	jobStateInstanceLabel string = "instanceId"
	jobStateLabel         string = "state"
)

// Dao - object to interface with the data store.
type Dao struct {
	client       automationbrokerv1.AutomationbrokerV1alpha1Interface
	namespace    string
	bundleLock   sync.Mutex
	bindingLock  sync.Mutex
	instanceLock sync.Mutex
}

// NewDao - Create a new Dao object
func NewDao(namespace string) (*Dao, error) {
	dao := Dao{namespace: namespace,
		bundleLock:   sync.Mutex{},
		bindingLock:  sync.Mutex{},
		instanceLock: sync.Mutex{},
	}

	crdClient, err := clients.CRDClient()
	if err != nil {
		return nil, err
	}
	dao.client = crdClient.AutomationbrokerV1alpha1()
	return &dao, nil
}

// GetSpec - Retrieve the spec from the k8s API.
func (d *Dao) GetSpec(id string) (*bundle.Spec, error) {
	log.Debugf("get spec: %v", id)
	s, err := d.client.Bundles(d.namespace).Get(id, metav1.GetOptions{})
	if err != nil {
		log.Errorf("unable to get bundle from k8s api - %v", err)
		return nil, err
	}
	return crd.ConvertBundleToSpec(s.Spec, s.GetName())
}

// SetSpec - set spec for an id in the kvp API.
func (d *Dao) SetSpec(id string, spec *bundle.Spec) error {
	defer d.bundleLock.Unlock()
	d.bundleLock.Lock()
	log.Debugf("set spec: %v", id)
	bundleSpec, err := crd.ConvertSpecToBundle(spec)
	if err != nil {
		return err
	}
	b := v1.Bundle{
		ObjectMeta: metav1.ObjectMeta{
			Name:      id,
			Namespace: d.namespace,
		},
		Spec: bundleSpec,
	}
	_, err = d.client.Bundles(d.namespace).Create(&b)
	return err
}

// DeleteSpec - Delete the spec for a given spec id.
func (d *Dao) DeleteSpec(specID string) error {
	log.Debugf("Dao::DeleteSpec-> [ %s ]", specID)
	return d.client.Bundles(d.namespace).Delete(specID, &metav1.DeleteOptions{})
}

// BatchSetSpecs - set specs based on SpecManifest in the kvp API.
func (d *Dao) BatchSetSpecs(specs bundle.SpecManifest) error {
	for id, spec := range specs {
		err := d.SetSpec(id, spec)
		if err != nil {
			log.Warningf("Error loading SPEC '%v'", spec.FQName)
			log.Debugf("SPEC '%v' error: %v", spec.FQName, err)
		}
	}
	return nil
}

// BatchGetSpecs - Retrieve all the specs for dir.
func (d *Dao) BatchGetSpecs(dir string) ([]*bundle.Spec, error) {
	log.Debugf("Dao::BatchGetSpecs")
	l, err := d.client.Bundles(d.namespace).List(metav1.ListOptions{})
	if err != nil {
		log.Errorf("unable to get batch specs - %v", err)
		return nil, err
	}
	specs := []*bundle.Spec{}
	// capture all the errors and still try to save the correct bundles
	errs := arrayErrors{}
	for _, b := range l.Items {
		spec, err := crd.ConvertBundleToSpec(b.Spec, b.GetName())
		if err != nil {
			errs = append(errs, err)
			continue
		}
		specs = append(specs, spec)
	}
	if len(errs) > 0 {
		return specs, errs
	}
	return specs, nil
}

// BatchDeleteSpecs - set specs based on SpecManifest in the kvp API.
func (d *Dao) BatchDeleteSpecs(specs []*bundle.Spec) error {
	for _, spec := range specs {
		err := d.DeleteSpec(spec.ID)
		if err != nil {
			return err
		}
	}
	return nil
}

// GetServiceInstance - Retrieve specific service instance from the kvp API.
func (d *Dao) GetServiceInstance(id string) (*bundle.ServiceInstance, error) {
	log.Debugf("get service instance: %v", id)
	servInstance, err := d.client.BundleInstances(d.namespace).Get(id, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}
	spec, err := d.GetSpec(servInstance.Spec.Bundle.Name)
	if err != nil {
		return nil, err
	}
	return crd.ConvertServiceInstanceToAPB(*servInstance, spec, servInstance.GetName())
}

// SetServiceInstance - Set service instance for an id in the kvp API.
func (d *Dao) SetServiceInstance(id string, serviceInstance *bundle.ServiceInstance) error {
	defer d.instanceLock.Unlock()
	d.instanceLock.Lock()
	log.Debugf("set service instance: %v", id)
	spec, err := crd.ConvertServiceInstanceToCRD(serviceInstance)
	if err != nil {
		return err
	}
	if si, err := d.client.BundleInstances(d.namespace).Get(id, metav1.GetOptions{}); err == nil {
		log.Debugf("updating service instance: %v", id)
		si.Spec = spec.Spec
		si.Status.Bindings = intersectionOfBindings(serviceInstance.BindingIDs, si.Status.Bindings)
		_, err := d.client.BundleInstances(d.namespace).Update(si)
		if err != nil {
			log.Errorf("unable to get service instance - %v", err)
			return err
		}
		return nil
	}
	s := v1.BundleInstance{
		ObjectMeta: metav1.ObjectMeta{
			Name:      id,
			Namespace: d.namespace,
		},
		Spec:   spec.Spec,
		Status: spec.Status,
	}

	_, err = d.client.BundleInstances(d.namespace).Create(&s)
	if err != nil {
		log.Errorf("unable to save service instance - %v", err)
		return err
	}
	return nil
}

func intersectionOfBindings(bindings map[string]bool, bind []v1.LocalObjectReference) []v1.LocalObjectReference {
	newBindings := []v1.LocalObjectReference{}
	alreadyChecked := map[string]bool{}
	// If one was deleted then we are not adding a binding and do not need to update them.
	// Because of the sync we can reason that nothing else is updating this
	deleted := false
	log.Debugf("\n\nbindings: %#v\nbind: %#v", bindings, bind)
	for _, b := range bind {
		if add, ok := bindings[b.Name]; !ok || add {
			newBindings = append(newBindings, b)
		} else {
			deleted = true
		}
		alreadyChecked[b.Name] = true
	}
	log.Debugf("\n\nnewBindings: %#v\nalreadyChecked: %#v", newBindings, alreadyChecked)
	// If we did not have a deletion then we need to add the binding.
	if !deleted {
		for k, v := range bindings {
			if _, ok := alreadyChecked[k]; !ok {
				if v {
					newBindings = append(newBindings, v1.LocalObjectReference{Name: k})
				}
			}
		}
	}
	log.Debugf("\n\nnewBindings: %#v\nalreadyChecked: %#v", newBindings, alreadyChecked)
	return newBindings
}

// DeleteServiceInstance - Delete the service instance for an service instance id.
func (d *Dao) DeleteServiceInstance(id string) error {
	log.Debugf("Dao::DeleteServiceInstance -> [ %s ]", id)
	return d.client.BundleInstances(d.namespace).Delete(id, &metav1.DeleteOptions{})
}

// GetBindInstance - Retrieve a specific bind instance from the kvp API
func (d *Dao) GetBindInstance(id string) (*bundle.BindInstance, error) {
	log.Debugf("get binding instance: %v", id)
	bi, err := d.client.BundleBindings(d.namespace).Get(id, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}
	return crd.ConvertServiceBindingToAPB(*bi, bi.GetName())
}

// SetBindInstance - Set the bind instance for id in the kvp API.
func (d *Dao) SetBindInstance(id string, bindInstance *bundle.BindInstance) error {
	defer d.instanceLock.Unlock()
	d.instanceLock.Lock()
	log.Debugf("set binding instance: %v", id)
	b, err := crd.ConvertServiceBindingToCRD(bindInstance)
	if err != nil {
		return err
	}
	bi := v1.BundleBinding{
		ObjectMeta: metav1.ObjectMeta{
			Name:      id,
			Namespace: d.namespace,
		},
		Spec:   b.Spec,
		Status: b.Status,
	}
	_, err = d.client.BundleBindings(d.namespace).Create(&bi)
	if err != nil && apierrors.IsAlreadyExists(err) {
		// looks like we already have this state, probably created by
		// another goroutine. Let's try to update the existing one instead.
		if binding, err := d.client.BundleBindings(d.namespace).Get(id, metav1.GetOptions{}); err == nil {
			binding.Spec = b.Spec
			_, err := d.client.BundleBindings(d.namespace).Update(binding)

			if err != nil && apierrors.IsConflict(err) {
				log.Debugf("Binding %v already exists, skipping update because of conflict.", id)
			} else if err != nil {
				log.Errorf("Unable to update the binding %v, after a failed creation. Reason: %v - %v",
					id, apierrors.ReasonForError(err), err)

				return err
			}
		}
	} else if err != nil {
		log.Errorf("unable to save service binding. Reason: %v - %v", apierrors.ReasonForError(err), err)
		return err
	}
	return nil
}

// DeleteBindInstance - Delete the binding instance for an id in the kvp API.
func (d *Dao) DeleteBindInstance(id string) error {
	log.Debugf("Dao::DeleteBindInstance -> [ %s ]", id)
	err := d.client.BundleBindings(d.namespace).Delete(id, &metav1.DeleteOptions{})
	return err
}

// SetState - Set the Job State in the kvp API for id.
func (d *Dao) SetState(instanceID string, state bundle.JobState) (string, error) {
	log.Debugf("set job state for instance: %v token: %v", instanceID, state.Token)
	n := metav1.Now()
	switch state.Method {
	case apb.JobMethodBind, apb.JobMethodUnbind:
		defer d.bindingLock.Unlock()
		d.bindingLock.Lock()
		// get the binding based on instance ID //update the job based on the token.
		bi, err := d.client.BundleBindings(d.namespace).Get(instanceID, metav1.GetOptions{})
		if err != nil {
			log.Errorf("Could not find binding %v associated with job state %v - %v",
				instanceID, state.Token, err)

			return state.Token, err
		}
		if bi.Status.Jobs == nil {
			bi.Status.Jobs = map[string]v1.Job{}
		}
		bi.Status.Jobs[state.Token] = v1.Job{
			Description:      state.Description,
			LastModifiedTime: &n,
			Method:           crd.ConvertJobMethodToCRD(state.Method),
			Podname:          state.Podname,
			State:            crd.ConvertStateToCRD(state.State),
			Error:            state.Error,
		}
		bi.Status.LastDescription = state.Description
		bi.Status.State = crd.ConvertStateToCRD(state.State)
		_, err = d.client.BundleBindings(d.namespace).Update(bi)
		if err != nil {
			if apierrors.IsConflict(err) {
				// detect if the error was a conflict or not. Conflicts occur
				// when two things attempt to update the same resource
				// simultaneously
				log.Warningf("detected a conflicting update of job state %v on binding %v",
					state.Token, instanceID)
			}

			log.Errorf("Unable to update the job state %v on the binding %v. Reason: %v - %v",
				state.Token, instanceID, apierrors.ReasonForError(err), err)

			return state.Token, err
		}
	case apb.JobMethodUpdate, apb.JobMethodDeprovision, apb.JobMethodProvision:
		defer d.bindingLock.Unlock()
		d.bindingLock.Lock()
		// get the binding based on instance id //update the job based on the token
		si, err := d.client.BundleInstances(d.namespace).Get(instanceID, metav1.GetOptions{})
		if err != nil {
			log.Errorf("Could not find instance %v associated with job state %v - %v",
				instanceID, state.Token, err)

			return state.Token, err
		}
		if si.Status.Jobs == nil {
			si.Status.Jobs = map[string]v1.Job{}
		}
		si.Status.Jobs[state.Token] = v1.Job{
			Description:      state.Description,
			LastModifiedTime: &n,
			Method:           crd.ConvertJobMethodToCRD(state.Method),
			Podname:          state.Podname,
			State:            crd.ConvertStateToCRD(state.State),
			Error:            state.Error,
		}
		si.Status.LastDescription = state.Description
		si.Status.State = crd.ConvertStateToCRD(state.State)
		_, err = d.client.BundleInstances(d.namespace).Update(si)
		if err != nil {
			log.Errorf("Unable to update the job state %v on the instance %v: %v",
				state.Token, instanceID, err)

			return state.Token, err
		}
	}

	// looks like we're good
	return state.Token, nil
}

// GetState - Retrieve a job state from the kvp API for an ID and Token.
func (d *Dao) GetState(id string, token string) (bundle.JobState, error) {
	// get the binding based on instance ID //update the job based on the token.
	var job v1.Job
	bi, err := d.client.BundleBindings(d.namespace).Get(id, metav1.GetOptions{})
	if err != nil && !d.IsNotFoundError(err) {
		log.Errorf("Could not find binding %v associated with job state %v - %v", id, token, err)
		return bundle.JobState{}, fmt.Errorf("Could not find binding %v associated with job state %v",
			id, token)
	} else if d.IsNotFoundError(err) {
		si, err := d.client.BundleInstances(d.namespace).Get(id, metav1.GetOptions{})
		if err != nil || si.Status.Jobs == nil {
			log.Errorf("Could not find instance %v associated with job state %v - %v",
				id, token, err)

			return bundle.JobState{}, err
		}
		j, ok := si.Status.Jobs[token]
		if !ok {
			log.Errorf("Unable to get the job state: %v - %v", token, err)
			return bundle.JobState{}, fmt.Errorf("unable to find job state %v", token)
		}
		job = j
	} else {
		if bi.Status.Jobs == nil {
			log.Errorf("binding %v has no associated job states: %v - %v", id, token, err)
			return bundle.JobState{}, err
		}
		j, ok := bi.Status.Jobs[token]
		if !ok {
			log.Errorf("binding %v does not have job state: %v - %v", id, token, err)
			return bundle.JobState{}, fmt.Errorf("unable to find job state %v", token)
		}

		job = j
	}
	return bundle.JobState{
		Description: job.Description,
		Method:      crd.ConvertJobMethodToAPB(job.Method),
		Podname:     job.Podname,
		Token:       token,
		State:       crd.ConvertStateToAPB(job.State),
		Error:       job.Error,
	}, nil
}

// GetStateByKey - Retrieve a job state from the kvp API for a job key
func (d *Dao) GetStateByKey(key string) (bundle.JobState, error) {
	bi, err := d.client.BundleBindings(d.namespace).Get(key, metav1.GetOptions{})
	if err != nil {
		log.Errorf("Unable to get the job state: %v - %v", key, err)
		return bundle.JobState{}, err
	}
	for token, j := range bi.Status.Jobs {
		// Assuming a single bind job happens per binding instance.
		if j.Method == v1.JobMethodBind {
			return bundle.JobState{
				Description: j.Description,
				Method:      crd.ConvertJobMethodToAPB(j.Method),
				Podname:     j.Podname,
				Token:       token,
				State:       crd.ConvertStateToAPB(j.State),
				Error:       j.Error,
			}, nil
		}
	}
	return bundle.JobState{}, &apierrors.StatusError{ErrStatus: metav1.Status{
		Status: metav1.StatusFailure,
		Code:   http.StatusNotFound,
		Reason: metav1.StatusReasonNotFound,
	}}
}

// FindJobStateByState - Retrieve all the jobs that match the specified state
func (d *Dao) FindJobStateByState(state bundle.State) ([]bundle.RecoverStatus, error) {

	sis, err := d.client.BundleInstances(d.namespace).List(metav1.ListOptions{})
	if err != nil {
		log.Errorf("unable to get instance jobs for the state: %v - %v", state, err)
		return nil, err
	}

	bis, err := d.client.BundleBindings(d.namespace).List(metav1.ListOptions{})
	if err != nil {
		log.Errorf("unable to get binding jobs for the state: %v - %v", state, err)
		return nil, err
	}

	// build the status information for recovery purposes
	rss := []bundle.RecoverStatus{}

	for _, si := range sis.Items {
		for token, j := range si.Status.Jobs {
			if state == crd.ConvertStateToAPB(j.State) {
				rss = append(rss,
					bundle.RecoverStatus{InstanceID: uuid.Parse(si.GetName()), State: bundle.JobState{
						Description: j.Description,
						Method:      crd.ConvertJobMethodToAPB(j.Method),
						Podname:     j.Podname,
						Token:       token,
						State:       crd.ConvertStateToAPB(j.State),
						Error:       j.Error,
					}})
			}
		}
	}

	for _, bi := range bis.Items {
		for token, j := range bi.Status.Jobs {
			if state == crd.ConvertStateToAPB(j.State) {
				rss = append(rss,
					bundle.RecoverStatus{InstanceID: uuid.Parse(bi.GetName()), State: bundle.JobState{
						Description: j.Description,
						Method:      crd.ConvertJobMethodToAPB(j.Method),
						Podname:     j.Podname,
						Token:       token,
						State:       crd.ConvertStateToAPB(j.State),
						Error:       j.Error,
					}})
			}
		}
	}

	return rss, nil
}

// GetSvcInstJobsByState - Lookup all jobs of a given state for a specific instance
func (d *Dao) GetSvcInstJobsByState(ID string, state bundle.State) ([]bundle.JobState, error) {
	// get the binding based on instance ID //update the job based on the token.
	jobs := []bundle.JobState{}
	bi, err := d.client.BundleBindings(d.namespace).Get(ID, metav1.GetOptions{})
	if err != nil && !d.IsNotFoundError(err) {
		log.Errorf("Unable to get the job state: %v - %v", ID, err)
		return []bundle.JobState{}, fmt.Errorf("unable to find job state %v", ID)
	} else if d.IsNotFoundError(err) {
		si, err := d.client.BundleInstances(d.namespace).Get(ID, metav1.GetOptions{})
		if err != nil {
			log.Errorf("Unable to get the job state: %v - %v", ID, err)
			return []bundle.JobState{}, err
		}
		for token, job := range si.Status.Jobs {
			if job.State == crd.ConvertStateToCRD(state) {
				jobs = append(jobs, bundle.JobState{
					Description: job.Description,
					Method:      crd.ConvertJobMethodToAPB(job.Method),
					Podname:     job.Podname,
					Token:       token,
					State:       crd.ConvertStateToAPB(job.State),
					Error:       job.Error,
				})
			}
		}
	} else {
		for token, job := range bi.Status.Jobs {
			if job.State == crd.ConvertStateToCRD(state) {
				jobs = append(jobs, bundle.JobState{
					Description: job.Description,
					Method:      crd.ConvertJobMethodToAPB(job.Method),
					Podname:     job.Podname,
					Token:       token,
					State:       crd.ConvertStateToAPB(job.State),
					Error:       job.Error,
				})
			}
		}
	}
	return jobs, nil
}

// IsNotFoundError - Will determine if the error is an apimachinary IsNotFound error.
func (d *Dao) IsNotFoundError(err error) bool {
	return apierrors.IsNotFound(err)
}

// DeleteBinding - Delete the binding instance and remove the association with the service instance.
func (d *Dao) DeleteBinding(bindingInstance bundle.BindInstance, serviceInstance bundle.ServiceInstance) error {
	if err := d.DeleteBindInstance(bindingInstance.ID.String()); err != nil {
		return err
	}
	serviceInstance.RemoveBinding(bindingInstance.ID)
	return d.SetServiceInstance(serviceInstance.ID.String(), &serviceInstance)
}
