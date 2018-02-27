package dao

import (
	"encoding/json"
	"reflect"

	"github.com/automationbroker/broker-client-go/pkg/apis/automationbroker.io/v1"
	"github.com/openshift/ansible-service-broker/pkg/apb"
	"github.com/pborman/uuid"
)

func specToBundle(spec *apb.Spec) v1.BundleSpec {
	// encode the metadata as string
	b, err := json.Marshal(spec.Metadata)
	if err != nil {
		log.Errorf("unable to marshal the metadata for spec to a json byte array - %v", err)
		return v1.BundleSpec{}
	}
	plans := []v1.Plan{}
	for _, specPlan := range spec.Plans {
		plans = append(plans, convertPlanToCRD(specPlan))
	}

	return v1.BundleSpec{
		Runtime:     spec.Runtime,
		Version:     spec.Version,
		FQName:      spec.FQName,
		Image:       spec.Image,
		Tags:        spec.Tags,
		Bindable:    spec.Bindable,
		Description: spec.Description,
		Async:       convertToAsyncType(spec.Async),
		Metadata:    string(b),
		Plans:       plans,
	}
}

func convertToAsyncType(s string) v1.AsyncType {
	switch s {
	case "optional":
		return v1.OptionalAsync
	case "required":
		return v1.RequiredAsync
	case "unsupported":
		return v1.Unsupported
	default:
		// Defaulting should never happen but defaulting to
		// required because Bundles by default should be run in async
		// because they will take time to spin up the new pod.
		return v1.RequiredAsync

	}
}

func convertPlanToCRD(plan apb.Plan) v1.Plan {
	b, err := json.Marshal(plan.Metadata)
	if err != nil {
		log.Errorf("unable to marshal the metadata for plan to a json byte array - %v", err)
		return v1.Plan{}
	}

	bindParams := []v1.Parameters{}
	params := []v1.Parameters{}
	for _, p := range plan.Parameters {
		params = append(params, convertParametersToCRD(p))
	}

	for _, p := range plan.BindParameters {
		bindParams = append(bindParams, convertParametersToCRD(p))
	}
	return v1.Plan{
		ID:             plan.ID,
		Name:           plan.Name,
		Description:    plan.Description,
		Metadata:       string(b),
		Free:           plan.Free,
		Bindable:       plan.Bindable,
		UpdatesTo:      plan.UpdatesTo,
		Parameters:     params,
		BindParameters: bindParams,
	}
}

func convertParametersToCRD(param apb.ParameterDescriptor) v1.Parameters {
	b, err := json.Marshal(map[string]interface{}{"default": param.Default})
	if err != nil {
		log.Errorf("unable to marshal the default for parameter to a json byte array - %v", err)
		return v1.Parameters{}
	}

	var v1Max *v1.NilableNumber
	if param.Maximum != nil {
		n := v1.NilableNumber(reflect.ValueOf(param.Maximum).Float())
		v1Max = &n
	}
	var v1exMax *v1.NilableNumber
	if param.ExclusiveMaximum != nil {
		n := v1.NilableNumber(reflect.ValueOf(param.ExclusiveMaximum).Float())
		v1exMax = &n
	}
	var v1Min *v1.NilableNumber
	if param.Minimum != nil {
		n := v1.NilableNumber(reflect.ValueOf(param.Minimum).Float())
		v1Min = &n
	}
	var v1exMin *v1.NilableNumber
	if param.ExclusiveMinimum != nil {
		n := v1.NilableNumber(reflect.ValueOf(param.ExclusiveMinimum).Float())
		v1exMin = &n
	}

	return v1.Parameters{
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
	}
}

func convertServiceInstanceToCRD(si *apb.ServiceInstance) v1.ServiceInstanceSpec {
	var b []byte
	if si.Parameters != nil {
		by, err := json.Marshal(si.Parameters)
		if err != nil {
			log.Errorf("unable to convert parameters to encoded json byte array -%v", err)
			return v1.ServiceInstanceSpec{}
		}
		b = by
	}

	bindingIDs := []string{}
	for key := range si.BindingIDs {
		bindingIDs = append(bindingIDs, key)
	}

	return v1.ServiceInstanceSpec{
		BundleID: si.Spec.ID,
		Context: v1.Context{
			Namespace: si.Context.Namespace,
			Plateform: si.Context.Platform,
		},
		Parameters: string(b),
		BindingIDs: bindingIDs,
	}
}

func convertServiceBindingToCRD(bi *apb.BindInstance) v1.ServiceBindingSpec {
	var b []byte
	if bi.Parameters != nil {
		by, err := json.Marshal(bi.Parameters)
		if err != nil {
			log.Errorf("Unable to marshal parameters to json byte array - %v", err)
			return v1.ServiceBindingSpec{}
		}
		b = by
	}
	return v1.ServiceBindingSpec{
		ServiceInstanceID: bi.ServiceID.String(),
		Parameters:        string(b),
		JobToken:          bi.CreateJobKey,
	}
}

func convertJobStateToCRD(js *apb.JobState) v1.JobStateSpec {
	return v1.JobStateSpec{
		State:       convertStateToCRD(js.State),
		Method:      convertJobMethodToCRD(js.Method),
		PodName:     js.Podname,
		Error:       js.Error,
		Description: js.Description,
	}
}

func convertJobMethodToCRD(j apb.JobMethod) v1.JobMethod {
	switch j {
	case apb.JobMethodProvision:
		return v1.JobMethodProvision
	case apb.JobMethodDeprovision:
		return v1.JobMethodDeprovision
	case apb.JobMethodBind:
		return v1.JobMethodBind
	case apb.JobMethodUnbind:
		return v1.JobMethodUnbind
	case apb.JobMethodUpdate:
		return v1.JobMethodUpdate
	}
	log.Errorf("unable to find the job method - %v", j)
	// This should never be called as all cases should already be covered.
	return v1.JobMethodProvision
}

func convertStateToCRD(s apb.State) v1.State {
	switch s {
	case apb.StateNotYetStarted:
		return v1.StateNotYetStarted
	case apb.StateInProgress:
		return v1.StateInProgress
	case apb.StateSucceeded:
		return v1.StateSucceeded
	case apb.StateFailed:
		return v1.StateFailed
	}
	// all cases should be coverd. we should never hit this code path.
	log.Errorf("Job state not found: %v", s)
	return v1.StateFailed
}

func bundleToSpec(spec v1.BundleSpec, id string) *apb.Spec {
	// encode the metadata as string
	m := map[string]interface{}{}
	err := json.Unmarshal([]byte(spec.Metadata), &m)
	if err != nil {
		log.Errorf("unable to unmarshal the metadata for spec - %v", err)
		return &apb.Spec{}
	}
	plans := []apb.Plan{}
	for _, specPlan := range spec.Plans {
		plans = append(plans, convertPlanToAPB(specPlan))
	}

	return &apb.Spec{
		ID:          id,
		Runtime:     spec.Runtime,
		Version:     spec.Version,
		FQName:      spec.FQName,
		Image:       spec.Image,
		Tags:        spec.Tags,
		Bindable:    spec.Bindable,
		Description: spec.Description,
		Async:       convertAsyncTypeToString(spec.Async),
		Metadata:    m,
		Plans:       plans,
	}
}

func convertAsyncTypeToString(a v1.AsyncType) string {
	switch a {
	case v1.OptionalAsync:
		return "optional"
	case v1.RequiredAsync:
		return "required"
	case v1.Unsupported:
		return "unsupported"
	}
	log.Errorf("unable to find the async type - %v", a)
	return "required"
}

func convertPlanToAPB(plan v1.Plan) apb.Plan {
	m := map[string]interface{}{}
	err := json.Unmarshal([]byte(plan.Metadata), &m)
	if err != nil {
		log.Errorf("unable to unmarshal the metadata for plan - %v", err)
		return apb.Plan{}
	}

	bindParams := []apb.ParameterDescriptor{}
	params := []apb.ParameterDescriptor{}
	for _, p := range plan.Parameters {
		params = append(params, convertParametersToAPB(p))
	}

	for _, p := range plan.BindParameters {
		bindParams = append(bindParams, convertParametersToAPB(p))
	}
	return apb.Plan{
		ID:             plan.ID,
		Name:           plan.Name,
		Description:    plan.Description,
		Metadata:       m,
		Free:           plan.Free,
		Bindable:       plan.Bindable,
		UpdatesTo:      plan.UpdatesTo,
		Parameters:     params,
		BindParameters: bindParams,
	}
}

func convertParametersToAPB(param v1.Parameters) apb.ParameterDescriptor {
	m := map[string]interface{}{}
	err := json.Unmarshal([]byte(param.Default), &m)
	if err != nil {
		log.Errorf("unable to unmarshal the default for parameter - %v", err)
		return apb.ParameterDescriptor{}
	}

	b := m["default"]

	var v1Max *apb.NilableNumber
	if param.Maximum != nil {
		n := apb.NilableNumber(reflect.ValueOf(param.Maximum).Float())
		v1Max = &n
	}
	var v1exMax *apb.NilableNumber
	if param.ExclusiveMaximum != nil {
		n := apb.NilableNumber(reflect.ValueOf(param.ExclusiveMaximum).Float())
		v1exMax = &n
	}
	var v1Min *apb.NilableNumber
	if param.Minimum != nil {
		n := apb.NilableNumber(reflect.ValueOf(param.Minimum).Float())
		v1Min = &n
	}
	var v1exMin *apb.NilableNumber
	if param.ExclusiveMinimum != nil {
		n := apb.NilableNumber(reflect.ValueOf(param.ExclusiveMinimum).Float())
		v1exMin = &n
	}

	return apb.ParameterDescriptor{
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
	}
}

func convertServiceInstanceToAPB(si v1.ServiceInstanceSpec, spec *apb.Spec, id string) *apb.ServiceInstance {
	parameters := &apb.Parameters{}
	if si.Parameters != "" {
		err := json.Unmarshal([]byte(si.Parameters), parameters)
		if err != nil {
			log.Errorf("unable to convert parameters to unmarshaled apb parameters -%v", err)
			return &apb.ServiceInstance{}
		}
	}

	bindingIDs := map[string]bool{}
	for _, val := range si.BindingIDs {
		bindingIDs[val] = true
	}

	return &apb.ServiceInstance{
		Spec: spec,
		Context: &apb.Context{
			Namespace: si.Context.Namespace,
			Platform:  si.Context.Plateform,
		},
		Parameters: parameters,
		BindingIDs: bindingIDs,
	}
}

func convertServiceBindingToAPB(bi v1.ServiceBindingSpec, id string) *apb.BindInstance {
	parameters := &apb.Parameters{}
	if bi.Parameters != "" {
		err := json.Unmarshal([]byte(bi.Parameters), parameters)
		if err != nil {
			log.Errorf("Unable to unmarshal parameters to apb parameters- %v", err)
			return &apb.BindInstance{}
		}
	}
	return &apb.BindInstance{
		ID:           uuid.Parse(id),
		ServiceID:    uuid.Parse(bi.ServiceInstanceID),
		Parameters:   parameters,
		CreateJobKey: bi.JobToken,
	}
}

func convertJobStateToAPB(js v1.JobStateSpec, id string) *apb.JobState {
	return &apb.JobState{
		Token:       id,
		State:       convertStateToAPB(js.State),
		Method:      convertJobMethodToAPB(js.Method),
		Podname:     js.PodName,
		Error:       js.Error,
		Description: js.Description,
	}
}

func convertJobMethodToAPB(j v1.JobMethod) apb.JobMethod {
	switch j {
	case v1.JobMethodProvision:
		return apb.JobMethodProvision
	case v1.JobMethodDeprovision:
		return apb.JobMethodDeprovision
	case v1.JobMethodBind:
		return apb.JobMethodBind
	case v1.JobMethodUnbind:
		return apb.JobMethodUnbind
	case v1.JobMethodUpdate:
		return apb.JobMethodUpdate
	}
	// We should have already covered all the cases above
	log.Errorf("Unable to find job method from - %v", j)
	return apb.JobMethodProvision
}

func convertStateToAPB(s v1.State) apb.State {
	switch s {
	case v1.StateNotYetStarted:
		return apb.StateNotYetStarted
	case v1.StateInProgress:
		return apb.StateInProgress
	case v1.StateSucceeded:
		return apb.StateSucceeded
	case v1.StateFailed:
		return apb.StateFailed
	}
	// We should have already covered all the cases above
	log.Errorf("Unable to find job state from - %v", s)
	return apb.StateFailed
}
