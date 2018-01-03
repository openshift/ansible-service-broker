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

// UpdateJob - Job to update
type UpdateJob struct {
	serviceInstance *apb.ServiceInstance
	update          apb.Updater
}

// NewUpdateJob - Create a new update job.
func NewUpdateJob(serviceInstance *apb.ServiceInstance, update apb.Updater) *UpdateJob {
	return &UpdateJob{
		serviceInstance: serviceInstance,
		update:          update,
	}
}

// Run - run the update job.
func (u *UpdateJob) Run(token string, msgBuffer chan<- JobMsg) {
	metrics.UpdateJobStarted()
	var (
		stateUpdates = make(chan apb.JobState)
		jobMsg       = JobMsg{
			InstanceUUID: u.serviceInstance.ID.String(),
			JobToken:     token,
			SpecID:       u.serviceInstance.Spec.ID,
			State: apb.JobState{
				State:  apb.StateInProgress,
				Method: apb.JobMethodUpdate,
				Token:  token,
			},
		}
		podName  string
		err      error
		errMsg   = "Error occurred during update. Please contact administrator if it persists."
		extCreds *apb.ExtractedCredentials
	)
	go func() {
		defer func() {
			close(stateUpdates)
			metrics.UpdateJobFinished()
		}()
		msgBuffer <- jobMsg
		podName, extCreds, err = u.update(u.serviceInstance, stateUpdates)
	}()
	for su := range stateUpdates {
		su.Token = token
		su.Method = apb.JobMethodUpdate
		msgBuffer <- JobMsg{InstanceUUID: u.serviceInstance.ID.String(), JobToken: token, State: su, PodName: su.Podname}
	}
	if err != nil {
		log.Errorf(" broker::Update error occurred. %v", err)
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
	if extCreds != nil {
		jobMsg.ExtractedCredentials = *extCreds
	}
	jobMsg.State.Description = "update job completed"
	jobMsg.PodName = podName
	msgBuffer <- jobMsg
}
