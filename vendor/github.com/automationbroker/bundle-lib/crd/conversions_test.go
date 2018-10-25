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
	"encoding/base64"
	"testing"

	yaml "gopkg.in/yaml.v1"

	"github.com/automationbroker/broker-client-go/pkg/apis/automationbroker/v1alpha1"
	"github.com/automationbroker/bundle-lib/apb"
	"github.com/pborman/uuid"
	"github.com/stretchr/testify/assert"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestConvertJobMethodToCRD(t *testing.T) {
	testCases := []struct {
		name     string
		input    apb.JobMethod
		expected v1alpha1.JobMethod
	}{
		{
			name:     "bundle provision job method",
			input:    apb.JobMethodProvision,
			expected: v1alpha1.JobMethodProvision,
		},
		{
			name:     "bundle deprovision job method",
			input:    apb.JobMethodDeprovision,
			expected: v1alpha1.JobMethodDeprovision,
		},
		{
			name:     "bundle bind job method",
			input:    apb.JobMethodBind,
			expected: v1alpha1.JobMethodBind,
		},
		{
			name:     "bundle unbind job method",
			input:    apb.JobMethodUnbind,
			expected: v1alpha1.JobMethodUnbind,
		},
		{
			name:     "bundle update job method",
			input:    apb.JobMethodUpdate,
			expected: v1alpha1.JobMethodUpdate,
		},
		{
			name:     "bundle empty job method",
			input:    "",
			expected: v1alpha1.JobMethodProvision,
		},
		{
			name:     "bundle unknown job method",
			input:    "unknown",
			expected: v1alpha1.JobMethodProvision,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.expected, ConvertJobMethodToCRD(tc.input))
		})
	}
}

func TestConvertJobMethodToAPB(t *testing.T) {
	testCases := []struct {
		name     string
		input    v1alpha1.JobMethod
		expected apb.JobMethod
	}{
		{
			name:     "crd provision job method",
			input:    v1alpha1.JobMethodProvision,
			expected: apb.JobMethodProvision,
		},
		{
			name:     "crd deprovision job method",
			input:    v1alpha1.JobMethodDeprovision,
			expected: apb.JobMethodDeprovision,
		},
		{
			name:     "crd bind job method",
			input:    v1alpha1.JobMethodBind,
			expected: apb.JobMethodBind,
		},
		{
			name:     "crd unbind job method",
			input:    v1alpha1.JobMethodUnbind,
			expected: apb.JobMethodUnbind,
		},
		{
			name:     "crd update job method",
			input:    v1alpha1.JobMethodUpdate,
			expected: apb.JobMethodUpdate,
		},
		{
			name:     "empty job method",
			input:    "",
			expected: apb.JobMethodProvision,
		},
		{
			name:     "unknown job method",
			input:    "unknown",
			expected: apb.JobMethodProvision,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.expected, ConvertJobMethodToAPB(tc.input))
		})
	}
}

func TestConvertStateToAPB(t *testing.T) {
	testCases := []struct {
		name     string
		input    v1alpha1.State
		expected apb.State
	}{
		{
			name:     "crd not yet started state",
			input:    v1alpha1.StateNotYetStarted,
			expected: apb.StateNotYetStarted,
		},
		{
			name:     "crd in progress state",
			input:    v1alpha1.StateInProgress,
			expected: apb.StateInProgress,
		},
		{
			name:     "crd succeeded state",
			input:    v1alpha1.StateSucceeded,
			expected: apb.StateSucceeded,
		},
		{
			name:     "crd failed state",
			input:    v1alpha1.StateFailed,
			expected: apb.StateFailed,
		},
		{
			name:     "empty state",
			input:    "",
			expected: apb.StateFailed,
		},
		{
			name:     "unknown state",
			input:    "unknown",
			expected: apb.StateFailed,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.expected, ConvertStateToAPB(tc.input))
		})
	}
}

func TestConvertStateToCRD(t *testing.T) {
	testCases := []struct {
		name     string
		input    apb.State
		expected v1alpha1.State
	}{
		{
			name:     "bundle not yet started state",
			input:    apb.StateNotYetStarted,
			expected: v1alpha1.StateNotYetStarted,
		},
		{
			name:     "bundle in progress state",
			input:    apb.StateInProgress,
			expected: v1alpha1.StateInProgress,
		},
		{
			name:     "bundle succeeded state",
			input:    apb.StateSucceeded,
			expected: v1alpha1.StateSucceeded,
		},
		{
			name:     "bundle failed state",
			input:    apb.StateFailed,
			expected: v1alpha1.StateFailed,
		},
		{
			name:     "empty state",
			input:    "",
			expected: v1alpha1.StateFailed,
		},
		{
			name:     "unknown state",
			input:    "unknown",
			expected: v1alpha1.StateFailed,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.expected, ConvertStateToCRD(tc.input))
		})
	}
}

func TestConvertServiceBindingToAPB(t *testing.T) {
	uid := uuid.New()

	testCases := []struct {
		name        string
		input       v1alpha1.BundleBinding
		expected    *apb.BindInstance
		expectederr bool
	}{
		{
			name:  "BundleBinding zero value",
			input: v1alpha1.BundleBinding{},
			expected: &apb.BindInstance{
				Parameters: &apb.Parameters{},
			},
		},
		{
			name: "invalid json string should return error",
			input: v1alpha1.BundleBinding{
				Spec: v1alpha1.BundleBindingSpec{
					BundleInstance: v1alpha1.LocalObjectReference{
						Name: "mynameis",
					},
					// removed final curly to make it invalid json
					Parameters: `{"_apb_creds":"letmein","foo":"bar"`,
				},
			},
			expected:    &apb.BindInstance{},
			expectederr: true,
		},
		{
			name: "parameters should get copied",
			input: v1alpha1.BundleBinding{
				ObjectMeta: metav1.ObjectMeta{
					Name:      uid,
					Namespace: "testing",
				},
				Spec: v1alpha1.BundleBindingSpec{
					BundleInstance: v1alpha1.LocalObjectReference{
						Name: uid,
					},
					Parameters: `{"_apb_creds":"letmein","foo":"bar"}`,
				},
			},
			expected: &apb.BindInstance{
				ID:        uuid.Parse(uid),
				ServiceID: uuid.Parse(uid),
				Parameters: &apb.Parameters{
					"foo":        "bar",
					"_apb_creds": "letmein",
				},
				CreateJobKey: uid,
			},
			expectederr: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			output, err := ConvertServiceBindingToAPB(tc.input, tc.input.GetName())
			if tc.expectederr {
				assert.Error(t, err)
			} else if err != nil {
				t.Fatalf("unexpected error during test: %v\n", err)
			}
			assert.Equal(t, tc.expected, output)
		})
	}
}

func TestConvertServiceBindingToCRD(t *testing.T) {
	uid := uuid.New()

	testCases := []struct {
		name        string
		input       *apb.BindInstance
		expected    v1alpha1.BundleBinding
		expectederr bool
	}{
		{
			name:     "BindInstance zero value",
			input:    &apb.BindInstance{},
			expected: v1alpha1.BundleBinding{},
		},
		{
			name: "parameters should get copied",
			input: &apb.BindInstance{
				ID:        uuid.Parse(uid),
				ServiceID: uuid.Parse(uid),
				Parameters: &apb.Parameters{
					"foo":        "bar",
					"_apb_creds": "letmein",
				},
			},
			expected: v1alpha1.BundleBinding{
				Spec: v1alpha1.BundleBindingSpec{
					BundleInstance: v1alpha1.LocalObjectReference{
						Name: uid,
					},
					Parameters: `{"_apb_creds":"letmein","foo":"bar"}`,
				},
			},
			expectederr: false,
		},
		{
			name: "invalid parameters should return error",
			input: &apb.BindInstance{
				ID:        uuid.Parse(uid),
				ServiceID: uuid.Parse(uid),
				Parameters: &apb.Parameters{
					// force json marshal to fail
					"foo": make(chan int),
				},
			},
			expected:    v1alpha1.BundleBinding{},
			expectederr: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			output, err := ConvertServiceBindingToCRD(tc.input)
			if tc.expectederr {
				assert.Error(t, err)
			} else if err != nil {
				t.Fatalf("unexpected error during test: %v\n", err)
			}
			assert.Equal(t, tc.expected, output)
		})
	}
}

func TestConvertServiceInstanceToAPB(t *testing.T) {
	uid := uuid.New()

	testCases := []struct {
		name        string
		input       v1alpha1.BundleInstance
		spec        *apb.Spec
		expected    *apb.ServiceInstance
		expectederr bool
	}{
		{
			name:  "BindInstance zero value",
			input: v1alpha1.BundleInstance{},
			spec:  &apb.Spec{},
			expected: &apb.ServiceInstance{
				ID:         uuid.Parse(uid),
				Spec:       &apb.Spec{},
				Context:    &apb.Context{},
				Parameters: &apb.Parameters{},
				BindingIDs: map[string]bool{},
			},
		},
		{
			name: "invalid parameter specs",
			input: v1alpha1.BundleInstance{
				Spec: v1alpha1.BundleInstanceSpec{
					Bundle: v1alpha1.LocalObjectReference{Name: uid},
					Context: v1alpha1.Context{
						Namespace: "testnamespace",
						Platform:  "kubernetes",
					},
					Parameters: `"_apb_creds":"letmein","foo":"bar"}`,
				},
				Status: v1alpha1.BundleInstanceStatus{
					Bindings: []v1alpha1.LocalObjectReference{
						{
							Name: "a binding",
						},
					},
				},
			},
			spec:        &apb.Spec{},
			expected:    &apb.ServiceInstance{},
			expectederr: true,
		},
		{
			name: "parameters should get copied",
			input: v1alpha1.BundleInstance{
				Spec: v1alpha1.BundleInstanceSpec{
					Bundle: v1alpha1.LocalObjectReference{Name: uid},
					Context: v1alpha1.Context{
						Namespace: "testnamespace",
						Platform:  "kubernetes",
					},
					Parameters:   `{"_apb_creds":"letmein","foo":"bar"}`,
					DashboardURL: "http://example.com/dashboard",
				},
				Status: v1alpha1.BundleInstanceStatus{
					Bindings: []v1alpha1.LocalObjectReference{
						{
							Name: "a binding",
						},
					},
				},
			},
			spec: &apb.Spec{},
			expected: &apb.ServiceInstance{
				ID:   uuid.Parse(uid),
				Spec: &apb.Spec{},
				Context: &apb.Context{
					Namespace: "testnamespace",
					Platform:  "kubernetes",
				},
				Parameters: &apb.Parameters{
					"foo":        "bar",
					"_apb_creds": "letmein",
				},
				BindingIDs: map[string]bool{
					"a binding": true,
				},
				DashboardURL: "http://example.com/dashboard",
			},

			expectederr: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			output, err := ConvertServiceInstanceToAPB(tc.input, tc.spec, uid)
			if tc.expectederr {
				assert.Error(t, err)
			} else if err != nil {
				t.Fatalf("unexpected error during test: %v\n", err)
			}
			assert.Equal(t, tc.expected, output)
		})
	}
}

func TestConvertSpecToBundle(t *testing.T) {
	uid := uuid.New()

	testCases := []struct {
		name        string
		input       *apb.Spec
		expected    v1alpha1.BundleSpec
		expectederr bool
	}{
		{
			name:  "apb.Spec zero value",
			input: &apb.Spec{},
			expected: v1alpha1.BundleSpec{
				Async:    convertToAsyncType("required"),
				Metadata: "null",
				Alpha:    "null",
				Plans:    []v1alpha1.Plan{},
			},
		},
		{
			name:        "invalid alpha",
			expectederr: true,
			input: &apb.Spec{
				Alpha: map[string]interface{}{
					"foo": make(chan int),
				},
			},
			expected: v1alpha1.BundleSpec{},
		},
		{
			name:        "invalid metadata",
			expectederr: true,
			input: &apb.Spec{
				Metadata: map[string]interface{}{
					"foo": make(chan int),
				},
			},
			expected: v1alpha1.BundleSpec{},
		},
		{
			name:        "invalid plan metadata",
			expectederr: true,
			input: &apb.Spec{
				Plans: []apb.Plan{
					{
						Name: "blowup",
						Metadata: map[string]interface{}{
							"foo": make(chan int),
						},
					},
				},
			},
			expected: v1alpha1.BundleSpec{},
		},
		{
			name:        "invalid plan parameters",
			expectederr: true,
			input: &apb.Spec{
				Plans: []apb.Plan{
					{
						Name: "blowup",
						Metadata: map[string]interface{}{
							"_apb_creds": "letmein",
						},
						Parameters: []apb.ParameterDescriptor{
							{
								Name:    "param1",
								Type:    "string",
								Default: make(chan int),
							},
						},
					},
				},
			},
			expected: v1alpha1.BundleSpec{},
		},
		{
			name:        "invalid plan bind parameters",
			expectederr: true,
			input: &apb.Spec{
				Plans: []apb.Plan{
					{
						Name: "blowup",
						Metadata: map[string]interface{}{
							"_apb_creds": "letmein",
						},
						BindParameters: []apb.ParameterDescriptor{
							{
								Name:    "param1",
								Type:    "string",
								Default: make(chan int),
							},
						},
					},
				},
			},
			expected: v1alpha1.BundleSpec{},
		},
		{
			name: "parameters should get copied",
			input: &apb.Spec{
				ID:          uid,
				Runtime:     2,
				Version:     "1.2.3",
				FQName:      "chevy/camaro-apb",
				Image:       "chevy/cavalier-apb",
				Tags:        []string{"cars", "chevy"},
				Bindable:    true,
				Description: "description",
				Async:       "optional",
				Metadata: map[string]interface{}{
					"_apb_creds": "letmein",
					"foo":        "bar",
				},
				Alpha: map[string]interface{}{
					"alpha_apb_creds": "letmein",
					"alphafoo":        "bar",
				},
				Plans: []apb.Plan{
					{
						Name:     "dev",
						Bindable: true,
						Metadata: map[string]interface{}{
							"plan_param1": "letmein",
							"plan_param2": "bar",
						},
						Parameters: []apb.ParameterDescriptor{
							{
								Name:        "param1",
								Type:        "string",
								Description: "parameter one",
							},
							{
								Name:             "param2",
								Type:             "int",
								Description:      "parameter two",
								Default:          10,
								Maximum:          bundleNilableNumber(float64(20)),
								ExclusiveMaximum: bundleNilableNumber(float64(40)),
								Minimum:          bundleNilableNumber(float64(5)),
								ExclusiveMinimum: bundleNilableNumber(float64(5)),
							},
						},
						BindParameters: []apb.ParameterDescriptor{
							{
								Name:        "bindparam1",
								Type:        "string",
								Description: "bind parameter one",
							},
							{
								Name:        "bindparam2",
								Type:        "int",
								Description: "bind parameter two",
								Default:     10,
							},
						},
					},
				},
			},
			expected: v1alpha1.BundleSpec{
				Runtime:     2,
				Version:     "1.2.3",
				FQName:      "chevy/camaro-apb",
				Image:       "chevy/cavalier-apb",
				Tags:        []string{"cars", "chevy"},
				Bindable:    true,
				Description: "description",
				Async:       convertToAsyncType("optional"),
				Metadata:    `{"_apb_creds":"letmein","foo":"bar"}`,
				Alpha:       `{"alpha_apb_creds":"letmein","alphafoo":"bar"}`,
				Plans: []v1alpha1.Plan{
					{
						Name:     "dev",
						Bindable: true,
						Metadata: `{"plan_param1":"letmein","plan_param2":"bar"}`,
						Parameters: []v1alpha1.Parameter{
							{
								Name:        "param1",
								Type:        "string",
								Description: "parameter one",
								Default:     "{\"default\":null}",
							},
							{
								Name:             "param2",
								Type:             "int",
								Description:      "parameter two",
								Default:          "{\"default\":10}",
								Maximum:          v1alpha1NilableNumber(float64(20)),
								ExclusiveMaximum: v1alpha1NilableNumber(float64(40)),
								Minimum:          v1alpha1NilableNumber(float64(5)),
								ExclusiveMinimum: v1alpha1NilableNumber(float64(5)),
							},
						},
						BindParameters: []v1alpha1.Parameter{
							{
								Name:        "bindparam1",
								Type:        "string",
								Description: "bind parameter one",
								Default:     "{\"default\":null}",
							},
							{
								Name:        "bindparam2",
								Type:        "int",
								Description: "bind parameter two",
								Default:     "{\"default\":10}",
							},
						},
					},
				},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {

			output, err := ConvertSpecToBundle(tc.input)
			if tc.expectederr {
				assert.Error(t, err)
			} else if err != nil {
				t.Fatalf("unexpected error during test: %v\n", err)
			}

			assert.Equal(t, tc.expected, output)
		})
	}
}

func TestConvertBundleToSpec(t *testing.T) {
	uid := uuid.New()

	testCases := []struct {
		name        string
		input       v1alpha1.BundleSpec
		expected    *apb.Spec
		expectederr bool
	}{
		{
			name:        "BundleSpec zero value",
			input:       v1alpha1.BundleSpec{},
			expected:    &apb.Spec{},
			expectederr: true,
		},
		{
			name:        "invalid metadata",
			expectederr: true,
			input: v1alpha1.BundleSpec{
				Metadata: `{"_apb_creds":"letmein","foo":"bar"`,
			},
			expected: &apb.Spec{},
		},
		{
			name:        "invalid alpha",
			expectederr: true,
			input: v1alpha1.BundleSpec{
				Metadata: `{}`,
				Alpha:    `"alpha_apb_creds":"letmein","alphafoo":"bar"}`,
			},
			expected: &apb.Spec{},
		},
		{
			name:        "invalid plan metadata",
			expectederr: true,
			input: v1alpha1.BundleSpec{
				Metadata: `{}`,
				Alpha:    `{}`,
				Plans: []v1alpha1.Plan{
					{
						Name:     "dev",
						Metadata: `"plan_param1":"letmein","plan_param2":"bar"`,
					},
				},
			},
			expected: &apb.Spec{},
		},
		{
			name:        "invalid plan parameters",
			expectederr: true,
			input: v1alpha1.BundleSpec{
				Alpha:    `{"alpha_apb_creds":"letmein","alphafoo":"bar"}`,
				Metadata: `{"_apb_creds":"letmein","foo":"bar"}`,
				Plans: []v1alpha1.Plan{
					{
						Name:     "dev",
						Metadata: `{"plan_param1":"letmein","plan_param2":"bar"}`,
						Parameters: []v1alpha1.Parameter{
							{
								Name:        "param1",
								Type:        "string",
								Description: "parameter one",
								Default:     `"default":null}`,
							},
						},
					},
				},
			},
			expected: &apb.Spec{},
		},
		{
			name:        "invalid plan bind parameters",
			expectederr: true,
			input: v1alpha1.BundleSpec{
				Alpha:    `{"alpha_apb_creds":"letmein","alphafoo":"bar"}`,
				Metadata: `{"_apb_creds":"letmein","foo":"bar"}`,
				Plans: []v1alpha1.Plan{
					{
						Name:     "dev",
						Metadata: `{"plan_param1":"letmein","plan_param2":"bar"}`,
						BindParameters: []v1alpha1.Parameter{
							{
								Name:        "param1",
								Type:        "string",
								Description: "parameter one",
								Default:     `"default":null}`,
							},
						},
					},
				},
			},
			expected: &apb.Spec{},
		},
		{
			name: "parameters should get copied",
			input: v1alpha1.BundleSpec{
				Runtime:     2,
				Version:     "1.2.3",
				FQName:      "chevy/camaro-apb",
				Image:       "chevy/cavalier-apb",
				Tags:        []string{"cars", "chevy"},
				Bindable:    true,
				Description: "description",
				Async:       convertToAsyncType("optional"),
				Metadata:    `{"_apb_creds":"letmein","foo":"bar"}`,
				Alpha:       `{"alpha_apb_creds":"letmein","alphafoo":"bar"}`,
				Plans: []v1alpha1.Plan{
					{
						Name:     "dev",
						Bindable: true,
						Metadata: `{"plan_param1":"letmein","plan_param2":"bar"}`,
						Parameters: []v1alpha1.Parameter{
							{
								Name:        "param1",
								Type:        "string",
								Description: "parameter one",
								Default:     "{\"default\":null}",
							},
							{
								Name:             "param2",
								Type:             "int",
								Description:      "parameter two",
								Default:          "{\"default\":10}",
								Maximum:          v1alpha1NilableNumber(float64(20)),
								ExclusiveMaximum: v1alpha1NilableNumber(float64(40)),
								Minimum:          v1alpha1NilableNumber(float64(5)),
								ExclusiveMinimum: v1alpha1NilableNumber(float64(5)),
							},
						},
						BindParameters: []v1alpha1.Parameter{
							{
								Name:        "bindparam1",
								Type:        "string",
								Description: "bind parameter one",
								Default:     "{\"default\":null}",
							},
							{
								Name:        "bindparam2",
								Type:        "int",
								Description: "bind parameter two",
								Default:     "{\"default\":10}",
							},
						},
					},
				},
			},

			expected: &apb.Spec{
				ID:          uid,
				Runtime:     2,
				Version:     "1.2.3",
				FQName:      "chevy/camaro-apb",
				Image:       "chevy/cavalier-apb",
				Tags:        []string{"cars", "chevy"},
				Bindable:    true,
				Description: "description",
				Async:       "optional",
				Metadata: map[string]interface{}{
					"_apb_creds": "letmein",
					"foo":        "bar",
				},
				Alpha: map[string]interface{}{
					"alpha_apb_creds": "letmein",
					"alphafoo":        "bar",
				},
				Plans: []apb.Plan{
					{
						Name:     "dev",
						Bindable: true,
						Metadata: map[string]interface{}{
							"plan_param1": "letmein",
							"plan_param2": "bar",
						},
						Parameters: []apb.ParameterDescriptor{
							{
								Name:        "param1",
								Type:        "string",
								Description: "parameter one",
							},
							{
								Name:             "param2",
								Type:             "int",
								Description:      "parameter two",
								Default:          float64(10),
								Maximum:          bundleNilableNumber(float64(20)),
								ExclusiveMaximum: bundleNilableNumber(float64(40)),
								Minimum:          bundleNilableNumber(float64(5)),
								ExclusiveMinimum: bundleNilableNumber(float64(5)),
							},
						},
						BindParameters: []apb.ParameterDescriptor{
							{
								Name:        "bindparam1",
								Type:        "string",
								Description: "bind parameter one",
							},
							{
								Name:        "bindparam2",
								Type:        "int",
								Description: "bind parameter two",
								Default:     float64(10),
							},
						},
					},
				},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {

			output, err := ConvertBundleToSpec(tc.input, tc.expected.ID)
			if tc.expectederr {
				assert.Error(t, err)
			} else if err != nil {
				t.Fatalf("unexpected error during test: %v\n", err)
			}

			assert.Equal(t, tc.expected, output)
		})
	}
}

func TestConvertServiceInstanceToCRD(t *testing.T) {
	uid := uuid.New()

	testCases := []struct {
		name        string
		input       *apb.ServiceInstance
		expected    v1alpha1.BundleInstance
		expectederr bool
		panics      bool
	}{
		{
			name:     "nil spec should cause error",
			input:    &apb.ServiceInstance{},
			expected: v1alpha1.BundleInstance{},
			panics:   true,
		},
		{
			name:     "nil ServiceInstance should cause error",
			input:    nil,
			expected: v1alpha1.BundleInstance{},
			panics:   true,
		},
		{
			name: "BindInstance zero value",
			input: &apb.ServiceInstance{
				Spec:    &apb.Spec{},
				Context: &apb.Context{},
			},
			expected: v1alpha1.BundleInstance{
				Status: v1alpha1.BundleInstanceStatus{
					Bindings: []v1alpha1.LocalObjectReference{},
				},
			},
		},
		{
			name: "invalid parameters should return error",
			input: &apb.ServiceInstance{
				Parameters: &apb.Parameters{
					// force json marshal to fail
					"foo": make(chan int),
				},
			},
			expected:    v1alpha1.BundleInstance{},
			expectederr: true,
		},
		{
			name: "parameters should get copied",
			input: &apb.ServiceInstance{
				ID: uuid.Parse(uid),
				Spec: &apb.Spec{
					ID: uid,
				},
				Context: &apb.Context{
					Namespace: "testnamespace",
					Platform:  "kubernetes",
				},
				Parameters: &apb.Parameters{
					"foo":        "bar",
					"_apb_creds": "letmein",
				},
				BindingIDs: map[string]bool{
					"a binding": true,
				},
				DashboardURL: "http://example.com/dashboard",
			},
			expected: v1alpha1.BundleInstance{
				Spec: v1alpha1.BundleInstanceSpec{
					Bundle: v1alpha1.LocalObjectReference{Name: uid},
					Context: v1alpha1.Context{
						Namespace: "testnamespace",
						Platform:  "kubernetes",
					},
					Parameters:   `{"_apb_creds":"letmein","foo":"bar"}`,
					DashboardURL: "http://example.com/dashboard",
				},
				Status: v1alpha1.BundleInstanceStatus{
					Bindings: []v1alpha1.LocalObjectReference{
						{
							Name: "a binding",
						},
					},
				},
			},
			expectederr: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {

			if tc.panics {
				assert.Panics(t, func() { ConvertServiceInstanceToCRD(tc.input) })
				return
			}

			output, err := ConvertServiceInstanceToCRD(tc.input)
			if tc.expectederr {
				assert.Error(t, err)
			} else if err != nil {
				t.Fatalf("unexpected error during test: %v\n", err)
			}

			assert.Equal(t, tc.expected, output)
		})
	}
}

func bundleNilableNumber(i float64) *apb.NilableNumber {
	n := apb.NilableNumber(i)
	return &n
}

func v1alpha1NilableNumber(i float64) *v1alpha1.NilableNumber {
	n := v1alpha1.NilableNumber(i)
	return &n
}

// TestConvertSpecToBundleUsingEncodedSpec uses a base64 encoded apb.yml spec
// to verify that the conversion code works with what the broker would normally
// see.
//
func TestConvertSpecToBundleUsingEncodedSpec(t *testing.T) {
	// Here is the yaml we encoded
	//
	// version: 1.0
	// name: testapp
	// description: your description
	// bindable: False
	// async: optional
	// metadata:
	//   displayName: testapp
	// plans:
	//   - name: default
	//     description: This default plan deploys testapp
	//     free: True
	//     metadata: {}
	//     parameters:
	//     - name: countwithrange
	//       title: Count Chocula
	//       type: int
	//       required: true
	//       updatable: true
	//       display_type: text
	//       maximum: 10
	//       minimum: 2
	//     - name: exclusiveberries
	//       title: Franken Berry
	//       type: int
	//       required: true
	//       updatable: true
	//       display_type: text
	//       maximum: 10
	//       exclusive_maximum: 10
	//       minimum: 2
	//       exclusive_minimum: 2
	//

	encodedapb := `dmVyc2lvbjogMS4wDQpuYW1lOiB0ZXN0YXBwDQpkZXNjcmlwdGlvbjogeW91ciBkZXNjcmlwdGlvbg0KYmluZGFibGU6IEZhbHNlDQphc3luYzogb3B0aW9uYWwNCm1ldGFkYXRhOg0KICBkaXNwbGF5TmFtZTogdGVzdGFwcA0KcGxhbnM6DQogIC0gbmFtZTogZGVmYXVsdA0KICAgIGRlc2NyaXB0aW9uOiBUaGlzIGRlZmF1bHQgcGxhbiBkZXBsb3lzIHRlc3RhcHANCiAgICBmcmVlOiBUcnVlDQogICAgbWV0YWRhdGE6IHt9DQogICAgcGFyYW1ldGVyczoNCiAgICAtIG5hbWU6IGNvdW50d2l0aHJhbmdlDQogICAgICB0aXRsZTogQ291bnQgQ2hvY3VsYQ0KICAgICAgdHlwZTogaW50DQogICAgICByZXF1aXJlZDogdHJ1ZQ0KICAgICAgdXBkYXRhYmxlOiB0cnVlDQogICAgICBkaXNwbGF5X3R5cGU6IHRleHQNCiAgICAgIG1heGltdW06IDEwDQogICAgICBtaW5pbXVtOiAyDQogICAgLSBuYW1lOiBleGNsdXNpdmViZXJyaWVzDQogICAgICB0aXRsZTogRnJhbmtlbiBCZXJyeQ0KICAgICAgdHlwZTogaW50DQogICAgICByZXF1aXJlZDogdHJ1ZQ0KICAgICAgdXBkYXRhYmxlOiB0cnVlDQogICAgICBkaXNwbGF5X3R5cGU6IHRleHQNCiAgICAgIG1heGltdW06IDEwDQogICAgICBleGNsdXNpdmVfbWF4aW11bTogMTANCiAgICAgIG1pbmltdW06IDINCiAgICAgIGV4Y2x1c2l2ZV9taW5pbXVtOiAy`

	decodedyaml, err := base64.StdEncoding.DecodeString(encodedapb)
	if err != nil {
		t.Fatal(err)
	}

	// This is a spec created from an encoded apb.yml.
	spec := &apb.Spec{}
	if err = yaml.Unmarshal(decodedyaml, spec); err != nil {
		t.Fatal(err)
	}
	expected := v1alpha1.BundleSpec{
		Version:     "1.0",
		FQName:      "testapp",
		Bindable:    false,
		Description: "your description",
		Async:       convertToAsyncType("optional"),
		Metadata:    `{"displayName":"testapp"}`,
		Alpha:       "null",
		Plans: []v1alpha1.Plan{
			{
				Name:        "default",
				Bindable:    false,
				Free:        true,
				Metadata:    `{}`,
				Description: "This default plan deploys testapp",
				Parameters: []v1alpha1.Parameter{
					{
						Name:        "countwithrange",
						Type:        "int",
						Title:       "Count Chocula",
						Required:    true,
						Updatable:   true,
						Default:     "{\"default\":null}",
						Maximum:     v1alpha1NilableNumber(float64(10)),
						Minimum:     v1alpha1NilableNumber(float64(2)),
						DisplayType: "text",
					},
					{
						Name:             "exclusiveberries",
						Type:             "int",
						Title:            "Franken Berry",
						Required:         true,
						Updatable:        true,
						Default:          "{\"default\":null}",
						Maximum:          v1alpha1NilableNumber(float64(10)),
						ExclusiveMaximum: v1alpha1NilableNumber(float64(10)),
						Minimum:          v1alpha1NilableNumber(float64(2)),
						ExclusiveMinimum: v1alpha1NilableNumber(float64(2)),
						DisplayType:      "text",
					},
				},
				BindParameters: []v1alpha1.Parameter{},
			},
		},
	}
	output, err := ConvertSpecToBundle(spec)
	assert.Equal(t, expected, output)
}

func TestConvertToAsyncTypeToString(t *testing.T) {
	testCases := []struct {
		name     string
		input    v1alpha1.AsyncType
		expected string
	}{
		{
			name:     "optional",
			input:    v1alpha1.OptionalAsync,
			expected: "optional",
		},
		{
			name:     "required",
			input:    v1alpha1.RequiredAsync,
			expected: "required",
		},
		{
			name:     "unsupported",
			input:    v1alpha1.Unsupported,
			expected: "unsupported",
		},
		{
			name:     "unknown",
			input:    v1alpha1.AsyncType("unknown"),
			expected: "required",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.expected, convertAsyncTypeToString(tc.input))
		})
	}
}

func TestConvertToAsyncType(t *testing.T) {
	testCases := []struct {
		name     string
		input    string
		expected v1alpha1.AsyncType
	}{
		{
			name:     "optional",
			input:    "optional",
			expected: v1alpha1.OptionalAsync,
		},
		{
			name:     "required",
			input:    "required",
			expected: v1alpha1.RequiredAsync,
		},
		{
			name:     "unsupported",
			input:    "unsupported",
			expected: v1alpha1.Unsupported,
		},
		{
			name:     "unknown",
			input:    "unknown",
			expected: v1alpha1.RequiredAsync,
		},
		{
			name:     "mismatched case",
			input:    "Optional",
			expected: v1alpha1.RequiredAsync,
		},
		{
			name:     "empty string",
			input:    "",
			expected: v1alpha1.RequiredAsync,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.expected, convertToAsyncType(tc.input))
		})
	}
}
