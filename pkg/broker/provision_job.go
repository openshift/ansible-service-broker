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
	"github.com/openshift/ansible-service-broker/pkg/apb"
	"github.com/openshift/ansible-service-broker/pkg/metrics"
)

// ProvisionJob - Job to provision
type ProvisionJob struct {
	serviceInstance *apb.ServiceInstance
	provision       apb.Provisioner
}

// Provisioner defines a function that knows how to provision an apb
type Provisioner func(si *apb.ServiceInstance) (string, *apb.ExtractedCredentials, error)

// NewProvisionJob - Create a new provision job.
func NewProvisionJob(serviceInstance *apb.ServiceInstance, provision apb.Provisioner) *ProvisionJob {
	return &ProvisionJob{
		serviceInstance: serviceInstance,
		provision:       provision,
	}
}

// Run - run the provision job.
func (p *ProvisionJob) Run(token string, msgBuffer chan<- JobMsg) {
	metrics.ProvisionJobStarted()
	defer metrics.ProvisionJobFinished()
	jobMsg := JobMsg{
		InstanceUUID: p.serviceInstance.ID.String(),
		JobToken:     token,
		SpecID:       p.serviceInstance.Spec.ID,
		State: apb.JobState{
			State:  apb.StateInProgress,
			Method: apb.JobMethodProvision,
			Token:  token,
		},
	}
	msgBuffer <- jobMsg
	podName, extCreds, err := p.provision(p.serviceInstance)

	if err != nil {
		log.Errorf("broker::Provision error occurred. %s", err.Error())
		errMsg := "Error occurred during provision. Please contact administrator if it persists."
		// Because we know the error we should return that error.
		if err == apb.ErrorPodPullErr {
			errMsg = err.Error()
		}
		// send error message, can't have
		// an error type in a struct you want marshalled
		// https://github.com/golang/go/issues/5161
		jobMsg.State.State = apb.StateFailed
		jobMsg.State.Error = errMsg
		msgBuffer <- jobMsg
		return
	}

	// send creds
	jobMsg.State.State = apb.StateSucceeded
	jobMsg.State.Podname = podName
	if nil != extCreds {
		jobMsg.ExtractedCredentials = *extCreds
	}
	jobMsg.PodName = podName
	msgBuffer <- jobMsg
}
