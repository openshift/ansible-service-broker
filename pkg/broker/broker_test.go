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

package broker

import (
	"fmt"
	"reflect"
	"testing"
	"time"

	"encoding/json"
	"github.com/automationbroker/bundle-lib/bundle"
	"github.com/automationbroker/bundle-lib/registries"
	"github.com/automationbroker/config"
	"github.com/openshift/ansible-service-broker/pkg/dao/mocks"
	ft "github.com/openshift/ansible-service-broker/pkg/fusortest"
	"github.com/pborman/uuid"
	"os"
)

func TestAddNameAndIDForSpecStripsTailingDash(t *testing.T) {
	spec1 := bundle.Spec{FQName: "1234567890123456789012345678901234567890-"}
	spec2 := bundle.Spec{FQName: "org/hello-world-apb"}
	spcs := []*bundle.Spec{&spec1, &spec2}
	addNameAndIDForSpec(spcs, "h")
	ft.AssertEqual(t, "h-1234567890123456789012345678901234567890", spcs[0].FQName)
	ft.AssertEqual(t, "h-org-hello-world-apb", spcs[1].FQName)
}

func TestAddIdForPlan(t *testing.T) {
	plan1 := bundle.Plan{Name: "default"}
	plans := []bundle.Plan{plan1}
	addIDForPlan(plans, "dh-sns-apb")
	ft.AssertNotEqual(t, plans[0].ID, "", "plan id not updated")
}

func TestNewAnsibleBroker(t *testing.T) {
	_, err := NewAnsibleBroker(&mocks.Dao{}, []registries.Registry{}, *NewWorkEngine(20, 2*time.Minute, &mocks.Dao{}), &config.Config{}, "new-space", NewWorkFactory())
	if err != nil {
		t.Fail()
	}
}

func TestGetServiceInstance(t *testing.T) {
	u := uuid.NewUUID()
	errIsNotFound := fmt.Errorf("is not found")
	unknownError := fmt.Errorf("is not found")
	testCases := []struct {
		name            string
		dao             *mocks.Dao
		config          Config
		addExpectations func(*mocks.Dao)
		shouldErr       error
		dashboardURL    string
	}{
		{
			name:   "get service instance without dashboard url",
			dao:    new(mocks.Dao),
			config: Config{},
			addExpectations: func(d *mocks.Dao) {
				d.On("GetServiceInstance", u.String()).Return(&bundle.ServiceInstance{
					ID:   u,
					Spec: &bundle.Spec{},
				}, nil)
			},
		},
		{
			name: "get service instance with dashboard url",
			dao:  new(mocks.Dao),
			config: Config{
				DashboardRedirector: "url.com",
			},
			addExpectations: func(d *mocks.Dao) {
				d.On("GetServiceInstance", u.String()).Return(&bundle.ServiceInstance{
					ID: u,
					Spec: &bundle.Spec{
						Alpha: map[string]interface{}{
							"dashboard_redirect": true,
						},
					},
				}, nil)
			},
			dashboardURL: fmt.Sprintf("http://%v/?id=%v", "url.com", u.String()),
		},
		{
			name: "get service instance with dashboard url and https",
			dao:  new(mocks.Dao),
			config: Config{
				DashboardRedirector: "https://url.com",
			},
			addExpectations: func(d *mocks.Dao) {
				d.On("GetServiceInstance", u.String()).Return(&bundle.ServiceInstance{
					ID: u,
					Spec: &bundle.Spec{
						Alpha: map[string]interface{}{
							"dashboard_redirect": true,
						},
					},
				}, nil)
			},
			dashboardURL: fmt.Sprintf("https://%v/?id=%v", "url.com", u.String()),
		},
		{
			name: "get service instance with spec alpha no dashboard redirect",
			dao:  new(mocks.Dao),
			config: Config{
				DashboardRedirector: "https://url.com",
			},
			addExpectations: func(d *mocks.Dao) {
				d.On("GetServiceInstance", u.String()).Return(&bundle.ServiceInstance{
					ID: u,
					Spec: &bundle.Spec{
						Alpha: map[string]interface{}{
							"testing": true,
						},
					},
				}, nil)
			},
		},
		{
			name:   "get service instance with spec alpha no dashboard redirect",
			dao:    new(mocks.Dao),
			config: Config{},
			addExpectations: func(d *mocks.Dao) {
				d.On("GetServiceInstance", u.String()).Return(&bundle.ServiceInstance{
					ID: u,
					Spec: &bundle.Spec{
						Alpha: map[string]interface{}{
							"testing": true,
						},
					},
				}, nil)
			},
		},
		{
			name:   "get service instance with spec alpha dashboard redirect invalid value",
			dao:    new(mocks.Dao),
			config: Config{},
			addExpectations: func(d *mocks.Dao) {
				d.On("GetServiceInstance", u.String()).Return(&bundle.ServiceInstance{
					ID: u,
					Spec: &bundle.Spec{
						Alpha: map[string]interface{}{
							"dashboard_redirect": "invalid_value",
						},
					},
				}, nil)
			},
		},
		{
			name:   "get service instance with spec alpha dashboard redirect false",
			dao:    new(mocks.Dao),
			config: Config{},
			addExpectations: func(d *mocks.Dao) {
				d.On("GetServiceInstance", u.String()).Return(&bundle.ServiceInstance{
					ID: u,
					Spec: &bundle.Spec{
						Alpha: map[string]interface{}{
							"dashboard_redirect": false,
						},
					},
				}, nil)
			},
		},
		{
			name:   "get service instance error is not found",
			dao:    new(mocks.Dao),
			config: Config{},
			addExpectations: func(d *mocks.Dao) {
				d.On("GetServiceInstance", u.String()).Return(nil, unknownError)
				d.On("IsNotFoundError", errIsNotFound).Return(false)
			},
			shouldErr: unknownError,
		},
		{
			name:   "get service instance error is not found",
			dao:    new(mocks.Dao),
			config: Config{},
			addExpectations: func(d *mocks.Dao) {
				d.On("GetServiceInstance", u.String()).Return(nil, errIsNotFound)
				d.On("IsNotFoundError", errIsNotFound).Return(true)
			},
			shouldErr: ErrorNotFound,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tc.addExpectations(tc.dao)
			a := AnsibleBroker{
				brokerConfig: tc.config,
				registry:     nil,
				namespace:    "test1",
				engine:       NewWorkEngine(20, 2*time.Minute, mockDao),
				dao:          tc.dao,
			}

			si, err := a.GetServiceInstance(u)
			if err != nil && err != tc.shouldErr {
				t.Fatalf("unknown error - %v", err)
			}
			if err != nil && err == tc.shouldErr {
				return
			}
			expected, err := tc.dao.GetServiceInstance(u.String())
			if !reflect.DeepEqual(si, *expected) {
				t.Fatalf("Invalid Service Instance\nexpected: %#v\nactual:%#v", *expected, si)
			}
			tc.dao.AssertExpectations(t)
			if si.DashboardURL != tc.dashboardURL {
				t.Fatalf("Invalid Service Instance dashboard URL \nexpected: %#v\nactual:%#v", tc.dashboardURL, si.DashboardURL)
			}
		})
	}
}

func TestGetMarkedSpecs(t *testing.T) {
	f, err := os.Open("./testdata/specs.json")
	if err != nil {
		t.Fail()
	}
	var specs []*bundle.Spec
	d := json.NewDecoder(f)
	d.Decode(&specs)
	defer f.Close()
	if err != nil {
		t.Fail()
	}
	m := getMarkedSpecs(specs)
	if _, ok := m["1dda1477cace09730bd8ed7a6505607e"]; !ok {
		t.Fail()
	}
	if _, ok := m["f86f8e54b99f9332e7610df228fc11d3"]; !ok {
		t.Fail()
	}
}

func TestGetSafeToDeleteSpecs(t *testing.T) {
	a := AnsibleBroker{dao: &mocks.Dao{}}
	f, err := os.Open("./testdata/specs.json")
	if err != nil {
		t.Fail()
	}
	d := json.NewDecoder(f)
	specs := make([]*bundle.Spec, 3)
	if err = d.Decode(&specs); err != nil {
		t.Fail()
	}
	f.Close()
	m := getMarkedSpecs(specs)
	s := getSafeToDeleteSpecs(a, m)
	if len(s) != 1 {
		t.Fail()
	}
	if s[0].ID != "f86f8e54b99f9332e7610df228fc11d3" {
		t.Fail()
	}
}

func TestConvertSpecListToMap(t *testing.T) {
	f, err := os.Open("./testdata/specs.json")
	if err != nil {
		t.Fail()
	}
	d := json.NewDecoder(f)
	var specs []*bundle.Spec
	if err = d.Decode(&specs); err != nil {
		t.Fail()
	}
	defer f.Close()
	sMap := convertSpecListToMap(specs)
	for _, spec := range specs {
		if _, ok := sMap[spec.ID]; !ok {
			t.Fail()
		}
	}
}

func TestGetNewAndUpdatedSpecs(t *testing.T) {
	f, err := os.Open("./testdata/specs.json")
	if err != nil {
		t.Fail()
	}
	var specs []*bundle.Spec
	var newSpecs []*bundle.Spec
	d := json.NewDecoder(f)
	if err = d.Decode(&specs); err != nil {
		t.Fail()
	}
	defer f.Close()
	daoSpecs := convertSpecListToMap(specs)
	newF, err := os.Open("./testdata/updatedSpecs.json")
	if err != nil {
		t.Fail()
	}
	newD := json.NewDecoder(newF)
	if err = newD.Decode(&newSpecs); err != nil {
		t.Fail()
	}
	defer newF.Close()
	n, u := getNewAndUpdatedSpecs(daoSpecs, newSpecs)
	if _, ok := n["0e991006d21029e47abe71acc255e807"]; !ok {
		t.Fail()
	}
	if _, ok := n["11bbd6c120e197ea6acacf7165749629"]; !ok {
		t.Fail()
	}
	if _, ok := u["1dda1477cace09730bd8ed7a6505607e"]; !ok {
		t.Fail()
	}
}

func TestMarkSpecsForDeletion(t *testing.T) {
	f, err := os.Open("./testdata/specs.json")
	if err != nil {
		t.Fail()
	}
	d := json.NewDecoder(f)
	var specs []*bundle.Spec
	var newSpecs []*bundle.Spec
	if err = d.Decode(&specs); err != nil {
		t.Fail()
	}
	defer f.Close()
	daoSpecs := convertSpecListToMap(specs)
	newF, err := os.Open("./testdata/updatedSpecs.json")
	if err != nil {
		t.Fail()
	}
	newD := json.NewDecoder(newF)
	if err = newD.Decode(&newSpecs); err != nil {
		t.Fail()
	}
	defer newF.Close()
	specManifest := convertSpecListToMap(newSpecs)
	markSpecsForDeletion(daoSpecs, specManifest)
	for id, spec := range daoSpecs {
		if id == "f6c4486b7fb0cdac4b58e193607f7011" || id == "1dda1477cace09730bd8ed7a6505607e" {
			if !spec.Delete {
				t.Fail()
			}
		}
	}
}
