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

	"github.com/automationbroker/bundle-lib/bundle"
	"github.com/automationbroker/bundle-lib/registries"
	"github.com/automationbroker/config"
	"github.com/openshift/ansible-service-broker/pkg/dao/mocks"
	ft "github.com/openshift/ansible-service-broker/pkg/fusortest"
	"github.com/pborman/uuid"
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
	_, err := NewAnsibleBroker(&mocks.Dao{}, []registries.Registry{}, *NewWorkEngine(20, 2*time.Minute), &config.Config{}, "new-space")
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
				engine:       NewWorkEngine(20, 2*time.Minute),
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
