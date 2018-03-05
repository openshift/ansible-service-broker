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
	"github.com/automationbroker/bundle-lib/apb"
	"github.com/automationbroker/config"
	crd "github.com/openshift/ansible-service-broker/pkg/dao/crd"
	etcd "github.com/openshift/ansible-service-broker/pkg/dao/etcd"
	logutil "github.com/openshift/ansible-service-broker/pkg/util/logging"
)

var log = logutil.NewLog()

// NewDao - Create a new Dao object
func NewDao(c *config.Config) (Dao, error) {
	if c.GetString("dao.type") == "crd" {
		return crd.NewDao(c.GetString("openshift.namespace"))
	}
	return etcd.NewDao()

}

// Dao - object to interface with the data store.
type Dao interface {
	// GetSpec - Retrieve the spec for the kvp API.
	GetSpec(string) (*apb.Spec, error)

	// SetSpec - set spec for an id in the kvp API.
	SetSpec(string, *apb.Spec) error

	// DeleteSpec - Delete the spec for a given spec id.
	DeleteSpec(string) error

	// BatchSetSpecs - set specs based on SpecManifest in the kvp API.
	BatchSetSpecs(apb.SpecManifest) error

	// BatchGetSpecs - Retrieve all the specs for dir.
	BatchGetSpecs(string) ([]*apb.Spec, error)

	// BatchDeleteSpecs - set specs based on SpecManifest in the kvp API.
	BatchDeleteSpecs([]*apb.Spec) error

	// FindJobStateByState - Retrieve all the jobs that match the specified state
	FindJobStateByState(apb.State) ([]apb.RecoverStatus, error)

	// GetSvcInstJobsByState - Lookup all jobs of a given state for a specific instance
	GetSvcInstJobsByState(string, apb.State) ([]apb.JobState, error)

	// GetServiceInstance - Retrieve specific service instance from the kvp API.
	GetServiceInstance(string) (*apb.ServiceInstance, error)

	// SetServiceInstance - Set service instance for an id in the kvp API.
	SetServiceInstance(string, *apb.ServiceInstance) error

	// DeleteServiceInstance - Delete the service instance for an service instance id.
	DeleteServiceInstance(string) error

	// GetBindInstance - Retrieve a specific bind instance from the kvp API
	GetBindInstance(string) (*apb.BindInstance, error)

	// SetBindInstance - Set the bind instance for id in the kvp API.
	SetBindInstance(string, *apb.BindInstance) error

	// DeleteBindInstance - Delete the binding instance for an id in the kvp API.
	DeleteBindInstance(string) error

	// SetState - Set the Job State in the kvp API for id.
	SetState(string, apb.JobState) (string, error)

	// GetState - Retrieve a job state from the kvp API for an ID and Token.
	GetState(string, string) (apb.JobState, error)

	// GetStateByKey - Retrieve a job state from the kvp API for a job key
	GetStateByKey(key string) (apb.JobState, error)

	// IsNotFoundError - Will determine if the error is a not found error from the DAO implementation.
	IsNotFoundError(err error) bool
}
