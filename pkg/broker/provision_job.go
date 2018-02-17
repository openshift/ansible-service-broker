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

// NewProvisionJob - Create a new provision job.
func NewProvisionJob(serviceInstance *apb.ServiceInstance, provision apb.Provisioner) *ProvisionJob {
	return &ProvisionJob{
		serviceInstance: serviceInstance,
		provision:       provision,
	}
}

// Run - run the provision job.
func (p *ProvisionJob) Run(token string, msgBuffer chan<- JobMsg) {
	// receives state updates during the provision action
	stateUpdates := make(chan apb.JobState, 1)
	metrics.ProvisionJobStarted()
	jobMsg := JobMsg{
		InstanceUUID: p.serviceInstance.ID.String(),
		JobToken:     token,
		SpecID:       p.serviceInstance.Spec.ID,
		State: apb.JobState{
			State:       apb.StateInProgress,
			Method:      apb.JobMethodProvision,
			Token:       token,
			Description: "provision job started",
		},
	}

	var (
		err      error
		errMsg   = "Error occurred during provision. Please contact administrator if it persists."
		podName  string
		extCreds *apb.ExtractedCredentials
	)

	go func() {
		defer func() {
			close(stateUpdates)
			metrics.ProvisionJobFinished()
		}()
		msgBuffer <- jobMsg
		podName, extCreds, err = p.provision(p.serviceInstance, stateUpdates)
	}()
	for stateUpdate := range stateUpdates {
		stateUpdate.Token = token
		stateUpdate.Method = apb.JobMethodProvision
		msgBuffer <- JobMsg{InstanceUUID: p.serviceInstance.ID.String(), JobToken: token, State: stateUpdate, PodName: stateUpdate.Podname}
	}
	// Once our job is finished evaluate if there was an err and update the state accordingly
	if err != nil {
		log.Errorf("broker::Provision error occurred. %s", err.Error())

		// Because we know the error we should return that error.
		if err == apb.ErrorPodPullErr {
			errMsg = err.Error()
		}

		jobMsg.State.State = apb.StateFailed
		// send error message, can't have
		// an error type in a struct you want marshalled
		// https://github.com/golang/go/issues/5161
		jobMsg.State.Error = err.Error()
		jobMsg.State.Description = errMsg
		msgBuffer <- jobMsg
		return
	}
	jobMsg.State.State = apb.StateSucceeded
	jobMsg.State.Podname = podName
	// send creds
	if nil != extCreds {
		jobMsg.ExtractedCredentials = *extCreds
	}
	jobMsg.PodName = podName
	jobMsg.State.Description = "provision job completed"
	msgBuffer <- jobMsg
}
