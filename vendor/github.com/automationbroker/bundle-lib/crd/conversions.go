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

package crd

import (
	"encoding/json"
	"fmt"
	"reflect"

	log "github.com/sirupsen/logrus"

	"github.com/automationbroker/broker-client-go/pkg/apis/automationbroker/v1alpha1"
	"github.com/automationbroker/bundle-lib/bundle"

	"github.com/pborman/uuid"
)

type arrayErrors []error

func (a arrayErrors) Error() string {
	return fmt.Sprintf("%#v", a)
}

// ConvertSpecToBundle will convert a bundle Spec to a Bundle CRD resource type.
func ConvertSpecToBundle(spec *bundle.Spec) (v1alpha1.BundleSpec, error) {
	// encode the metadata as string
	metadataBytes, err := json.Marshal(spec.Metadata)
	if err != nil {
		log.Errorf("unable to marshal the metadata for spec to a json byte array - %v", err)
		return v1alpha1.BundleSpec{}, err
	}
	plans := []v1alpha1.Plan{}
	// encode the alpha as string
	alphaBytes, err := json.Marshal(spec.Alpha)
	if err != nil {
		log.Errorf("unable to marshal the alpha for spec to a json byte array - %v", err)
		return v1alpha1.BundleSpec{}, err
	}
	errs := arrayErrors{}
	for _, specPlan := range spec.Plans {
		plan, err := convertPlanToCRD(specPlan)
		if err != nil {
			errs = append(errs, err)
		}
		plans = append(plans, plan)
	}
	if len(errs) > 0 {
		return v1alpha1.BundleSpec{}, errs
	}

	return v1alpha1.BundleSpec{
		Runtime:     spec.Runtime,
		Version:     spec.Version,
		FQName:      spec.FQName,
		Image:       spec.Image,
		Tags:        spec.Tags,
		Bindable:    spec.Bindable,
		Description: spec.Description,
		Async:       convertToAsyncType(spec.Async),
		Metadata:    string(metadataBytes),
		Alpha:       string(alphaBytes),
		Plans:       plans,
	}, nil
}

// ConvertBundleToSpec accepts a bundle-client-go BundleSpec along with its id
// (which is often the Bundle's name), and will convert these into a bundle
// Spec type.
func ConvertBundleToSpec(spec v1alpha1.BundleSpec, id string) (*bundle.Spec, error) {
	// TODO: Should this also just accept the bundle and automatically pull out
	// the name as the ID?

	// encode the metadata as string
	metadataMap := map[string]interface{}{}
	err := json.Unmarshal([]byte(spec.Metadata), &metadataMap)
	if err != nil {
		log.Errorf("unable to unmarshal the metadata for spec - %v", err)
		return &bundle.Spec{}, err
	}
	plans := []bundle.Plan{}
	// encode the alpha as string
	alphaMap := map[string]interface{}{}
	err = json.Unmarshal([]byte(spec.Alpha), &alphaMap)
	if err != nil {
		log.Errorf("unable to unmarshal the alpha for spec - %v", err)
		return &bundle.Spec{}, err
	}
	errs := arrayErrors{}
	for _, specPlan := range spec.Plans {
		plan, err := convertPlanToAPB(specPlan)
		if err != nil {
			errs = append(errs, err)
			continue
		}
		plans = append(plans, plan)
	}

	if len(errs) > 0 {
		return &bundle.Spec{}, errs
	}

	return &bundle.Spec{
		ID:          id,
		Runtime:     spec.Runtime,
		Version:     spec.Version,
		FQName:      spec.FQName,
		Image:       spec.Image,
		Tags:        spec.Tags,
		Bindable:    spec.Bindable,
		Description: spec.Description,
		Async:       convertAsyncTypeToString(spec.Async),
		Metadata:    metadataMap,
		Alpha:       alphaMap,
		Plans:       plans,
	}, nil
}

// ConvertServiceInstanceToCRD will take a bundle ServiceInstance and convert
// it to a ServiceInstanceSpec CRD type.
func ConvertServiceInstanceToCRD(si *bundle.ServiceInstance) (v1alpha1.BundleInstance, error) {
	var b []byte
	if si.Parameters != nil {
		by, err := json.Marshal(si.Parameters)
		if err != nil {
			log.Errorf("unable to convert parameters to encoded json byte array -%v", err)
			return v1alpha1.BundleInstance{}, err
		}
		b = by
	}

	bindings := []v1alpha1.LocalObjectReference{}
	for key := range si.BindingIDs {
		bindings = append(bindings, v1alpha1.LocalObjectReference{Name: key})
	}

	return v1alpha1.BundleInstance{
		Spec: v1alpha1.BundleInstanceSpec{
			Bundle: v1alpha1.LocalObjectReference{Name: si.Spec.ID},
			Context: v1alpha1.Context{
				Namespace: si.Context.Namespace,
				Platform:  si.Context.Platform,
			},
			Parameters:   string(b),
			DashboardURL: si.DashboardURL,
		},
		Status: v1alpha1.BundleInstanceStatus{
			Bindings: bindings,
		},
	}, nil
}

// ConvertServiceInstanceToAPB will take a ServiceInstanceSpec its associated
// bundle Spec, as well as an id (often the ServiceInstance's name), and will
// convert those to a bundle ServiceInstance.
func ConvertServiceInstanceToAPB(si v1alpha1.BundleInstance, spec *bundle.Spec, id string) (*bundle.ServiceInstance, error) {
	// TODO: Should this conversion just accept a ServiceInstance and automatically
	// dereference the bundle from the Service Instance?

	parameters := &bundle.Parameters{}
	if si.Spec.Parameters != "" {
		err := json.Unmarshal([]byte(si.Spec.Parameters), parameters)
		if err != nil {
			log.Errorf("unable to convert parameters to unmarshaled bundle parameters -%v", err)
			return &bundle.ServiceInstance{}, err
		}
	}

	bindingIDs := map[string]bool{}
	for _, val := range si.Status.Bindings {
		bindingIDs[val.Name] = true
	}

	return &bundle.ServiceInstance{
		ID:   uuid.Parse(id),
		Spec: spec,
		Context: &bundle.Context{
			Namespace: si.Spec.Context.Namespace,
			Platform:  si.Spec.Context.Platform,
		},
		Parameters:   parameters,
		BindingIDs:   bindingIDs,
		DashboardURL: si.Spec.DashboardURL,
	}, nil
}

// ConvertServiceBindingToCRD will take a bundle BindInstance and convert it
// to a ServiceBindingSpec CRD type.
func ConvertServiceBindingToCRD(bi *bundle.BindInstance) (v1alpha1.BundleBinding, error) {
	var b []byte
	if bi.Parameters != nil {
		by, err := json.Marshal(bi.Parameters)
		if err != nil {
			log.Errorf("Unable to marshal parameters to json byte array - %v", err)
			return v1alpha1.BundleBinding{}, err
		}
		b = by
	}
	return v1alpha1.BundleBinding{
		Spec: v1alpha1.BundleBindingSpec{
			BundleInstance: v1alpha1.LocalObjectReference{Name: bi.ServiceID.String()},
			Parameters:     string(b),
		},
	}, nil
}

// ConvertServiceBindingToAPB accepts a bundle-client-go ServiceBindingSpec
// along with its id (which is often the ServiceBinding's name), and will convert
// these into a bundle BindInstance.
func ConvertServiceBindingToAPB(bi v1alpha1.BundleBinding, id string) (*bundle.BindInstance, error) {
	// TODO: Same as above, accept the full ServiceBinding?
	parameters := &bundle.Parameters{}
	if bi.Spec.Parameters != "" {
		err := json.Unmarshal([]byte(bi.Spec.Parameters), parameters)
		if err != nil {
			log.Errorf("Unable to unmarshal parameters to bundle parameters- %v", err)
			return &bundle.BindInstance{}, err
		}
	}
	return &bundle.BindInstance{
		ID:         uuid.Parse(id),
		ServiceID:  uuid.Parse(bi.Spec.BundleInstance.Name),
		Parameters: parameters,
	}, nil
}

// ConvertStateToCRD will take an bundle State type and convert it to a
// broker-client-go State type.
func ConvertStateToCRD(s bundle.State) v1alpha1.State {
	switch s {
	case bundle.StateNotYetStarted:
		return v1alpha1.StateNotYetStarted
	case bundle.StateInProgress:
		return v1alpha1.StateInProgress
	case bundle.StateSucceeded:
		return v1alpha1.StateSucceeded
	case bundle.StateFailed:
		return v1alpha1.StateFailed
	}
	// all cases should be covered. we should never hit this code path.
	log.Errorf("Job state not found: %v", s)
	return v1alpha1.StateFailed
}

// ConvertStateToAPB will take a bundle-client-go State type and convert it to a
// bundle State type.
func ConvertStateToAPB(s v1alpha1.State) bundle.State {
	switch s {
	case v1alpha1.StateNotYetStarted:
		return bundle.StateNotYetStarted
	case v1alpha1.StateInProgress:
		return bundle.StateInProgress
	case v1alpha1.StateSucceeded:
		return bundle.StateSucceeded
	case v1alpha1.StateFailed:
		return bundle.StateFailed
	}
	// We should have already covered all the cases above
	log.Errorf("Unable to find job state from - %v", s)
	return bundle.StateFailed
}

// ConvertJobMethodToCRD will convert the bundle job method to the crd job method.
func ConvertJobMethodToCRD(j bundle.JobMethod) v1alpha1.JobMethod {
	switch j {
	case bundle.JobMethodProvision:
		return v1alpha1.JobMethodProvision
	case bundle.JobMethodDeprovision:
		return v1alpha1.JobMethodDeprovision
	case bundle.JobMethodBind:
		return v1alpha1.JobMethodBind
	case bundle.JobMethodUnbind:
		return v1alpha1.JobMethodUnbind
	case bundle.JobMethodUpdate:
		return v1alpha1.JobMethodUpdate
	}
	log.Errorf("unable to find the job method - %v", j)
	// This should never be called as all cases should already be covered.
	return v1alpha1.JobMethodProvision
}

// ConvertJobMethodToAPB will convert crd job method to bundle job method.
func ConvertJobMethodToAPB(j v1alpha1.JobMethod) bundle.JobMethod {
	switch j {
	case v1alpha1.JobMethodProvision:
		return bundle.JobMethodProvision
	case v1alpha1.JobMethodDeprovision:
		return bundle.JobMethodDeprovision
	case v1alpha1.JobMethodBind:
		return bundle.JobMethodBind
	case v1alpha1.JobMethodUnbind:
		return bundle.JobMethodUnbind
	case v1alpha1.JobMethodUpdate:
		return bundle.JobMethodUpdate
	}
	// We should have already covered all the cases above
	log.Errorf("Unable to find job method from - %v", j)
	return bundle.JobMethodProvision
}

////////////////////////////////////////////////////////////
// Internal
////////////////////////////////////////////////////////////

func convertToAsyncType(s string) v1alpha1.AsyncType {
	switch s {
	case "optional":
		return v1alpha1.OptionalAsync
	case "required":
		return v1alpha1.RequiredAsync
	case "unsupported":
		return v1alpha1.Unsupported
	default:
		// Defaulting should never happen but defaulting to
		// required because Bundles by default should be run in async
		// because they will take time to spin up the new pod.
		return v1alpha1.RequiredAsync

	}
}

func convertPlanToCRD(plan bundle.Plan) (v1alpha1.Plan, error) {
	b, err := json.Marshal(plan.Metadata)
	if err != nil {
		log.Errorf("unable to marshal the metadata for plan to a json byte array - %v", err)
		return v1alpha1.Plan{}, err
	}

	bindParams := []v1alpha1.Parameter{}
	params := []v1alpha1.Parameter{}
	errs := arrayErrors{}
	for _, p := range plan.Parameters {
		param, err := convertParametersToCRD(p)
		if err != nil {
			errs = append(errs, err)
			continue
		}
		params = append(params, param)
	}

	for _, p := range plan.BindParameters {
		param, err := convertParametersToCRD(p)
		if err != nil {
			errs = append(errs, err)
			continue
		}
		bindParams = append(bindParams, param)
	}
	if len(errs) > 0 {
		return v1alpha1.Plan{}, err
	}
	return v1alpha1.Plan{
		ID:             plan.ID,
		Name:           plan.Name,
		Description:    plan.Description,
		Metadata:       string(b),
		Free:           plan.Free,
		Bindable:       plan.Bindable,
		UpdatesTo:      plan.UpdatesTo,
		Parameters:     params,
		BindParameters: bindParams,
	}, nil
}

func convertParametersToCRD(param bundle.ParameterDescriptor) (v1alpha1.Parameter, error) {
	b, err := json.Marshal(map[string]interface{}{"default": param.Default})
	if err != nil {
		log.Errorf("unable to marshal the default for parameter to a json byte array - %v", err)
		return v1alpha1.Parameter{}, err
	}

	var v1Max *v1alpha1.NilableNumber
	if param.Maximum != nil {
		n := v1alpha1.NilableNumber(reflect.ValueOf(param.Maximum).Float())
		v1Max = &n
	}
	var v1exMax *v1alpha1.NilableNumber
	if param.ExclusiveMaximum != nil {
		n := v1alpha1.NilableNumber(reflect.ValueOf(param.ExclusiveMaximum).Float())
		v1exMax = &n
	}
	var v1Min *v1alpha1.NilableNumber
	if param.Minimum != nil {
		n := v1alpha1.NilableNumber(reflect.ValueOf(param.Minimum).Float())
		v1Min = &n
	}
	var v1exMin *v1alpha1.NilableNumber
	if param.ExclusiveMinimum != nil {
		n := v1alpha1.NilableNumber(reflect.ValueOf(param.ExclusiveMinimum).Float())
		v1exMin = &n
	}

	return v1alpha1.Parameter{
		Name:                param.Name,
		Title:               param.Title,
		Type:                param.Type,
		Description:         param.Description,
		Default:             string(b),
		DeprecatedMaxLength: param.DeprecatedMaxlength,
		MaxLength:           param.MaxLength,
		MinLength:           param.MinLength,
		Pattern:             param.Pattern,
		MultipleOf:          param.MultipleOf,
		Maximum:             v1Max,
		ExclusiveMaximum:    v1exMax,
		ExclusiveMinimum:    v1exMin,
		Minimum:             v1Min,
		Enum:                param.Enum,
		Required:            param.Required,
		Updatable:           param.Updatable,
		DisplayType:         param.DisplayType,
		DisplayGroup:        param.DisplayGroup,
	}, nil
}

func convertAsyncTypeToString(a v1alpha1.AsyncType) string {
	switch a {
	case v1alpha1.OptionalAsync:
		return "optional"
	case v1alpha1.RequiredAsync:
		return "required"
	case v1alpha1.Unsupported:
		return "unsupported"
	}
	log.Errorf("unable to find the async type - %v", a)
	return "required"
}

func convertPlanToAPB(plan v1alpha1.Plan) (bundle.Plan, error) {
	m := map[string]interface{}{}
	err := json.Unmarshal([]byte(plan.Metadata), &m)
	if err != nil {
		log.Errorf("unable to unmarshal the metadata for plan - %v", err)
		return bundle.Plan{}, err
	}

	bindParams := []bundle.ParameterDescriptor{}
	params := []bundle.ParameterDescriptor{}
	errs := arrayErrors{}
	for _, p := range plan.Parameters {
		param, err := convertParametersToAPB(p)
		if err != nil {
			errs = append(errs, err)
			continue
		}
		params = append(params, param)
	}

	for _, p := range plan.BindParameters {
		param, err := convertParametersToAPB(p)
		if err != nil {
			errs = append(errs, err)
			continue
		}
		bindParams = append(bindParams, param)
	}
	return bundle.Plan{
		ID:             plan.ID,
		Name:           plan.Name,
		Description:    plan.Description,
		Metadata:       m,
		Free:           plan.Free,
		Bindable:       plan.Bindable,
		UpdatesTo:      plan.UpdatesTo,
		Parameters:     params,
		BindParameters: bindParams,
	}, nil
}

func convertParametersToAPB(param v1alpha1.Parameter) (bundle.ParameterDescriptor, error) {
	m := map[string]interface{}{}
	err := json.Unmarshal([]byte(param.Default), &m)
	if err != nil {
		log.Errorf("unable to unmarshal the default for parameter - %v", err)
		return bundle.ParameterDescriptor{}, err
	}

	b := m["default"]

	var v1Max *bundle.NilableNumber
	if param.Maximum != nil {
		n := bundle.NilableNumber(reflect.ValueOf(param.Maximum).Float())
		v1Max = &n
	}
	var v1exMax *bundle.NilableNumber
	if param.ExclusiveMaximum != nil {
		n := bundle.NilableNumber(reflect.ValueOf(param.ExclusiveMaximum).Float())
		v1exMax = &n
	}
	var v1Min *bundle.NilableNumber
	if param.Minimum != nil {
		n := bundle.NilableNumber(reflect.ValueOf(param.Minimum).Float())
		v1Min = &n
	}
	var v1exMin *bundle.NilableNumber
	if param.ExclusiveMinimum != nil {
		n := bundle.NilableNumber(reflect.ValueOf(param.ExclusiveMinimum).Float())
		v1exMin = &n
	}

	return bundle.ParameterDescriptor{
		Name:                param.Name,
		Title:               param.Title,
		Type:                param.Type,
		Description:         param.Description,
		Default:             b,
		DeprecatedMaxlength: param.DeprecatedMaxLength,
		MaxLength:           param.MaxLength,
		MinLength:           param.MinLength,
		Pattern:             param.Pattern,
		MultipleOf:          param.MultipleOf,
		Maximum:             v1Max,
		ExclusiveMaximum:    v1exMax,
		ExclusiveMinimum:    v1exMin,
		Minimum:             v1Min,
		Enum:                param.Enum,
		Required:            param.Required,
		Updatable:           param.Updatable,
		DisplayType:         param.DisplayType,
		DisplayGroup:        param.DisplayGroup,
	}, nil
}
