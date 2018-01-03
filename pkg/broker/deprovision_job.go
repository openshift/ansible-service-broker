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

// DeprovisionJob - Job to deprovision.
type DeprovisionJob struct {
	serviceInstance  *apb.ServiceInstance
	skipApbExecution bool
	deprovision      apb.Deprovisioner
}

// NewDeprovisionJob - Create a deprovision job.
func NewDeprovisionJob(serviceInstance *apb.ServiceInstance,
	skipApbExecution bool, deprovision apb.Deprovisioner,
) *DeprovisionJob {
	return &DeprovisionJob{
		serviceInstance:  serviceInstance,
		skipApbExecution: skipApbExecution,
		deprovision:      deprovision,
	}
}

// Run - will run the deprovision job.
func (p *DeprovisionJob) Run(token string, msgBuffer chan<- JobMsg) {
	metrics.DeprovisionJobStarted()
	jobMsg := JobMsg{
		InstanceUUID: p.serviceInstance.ID.String(),
		JobToken:     token,
		SpecID:       p.serviceInstance.Spec.ID,
		State: apb.JobState{
			State:  apb.StateInProgress,
			Method: apb.JobMethodDeprovision,
			Token:  token,
		},
	}
	stateUpdates := make(chan apb.JobState)

	var (
		errMsg  = "Error occurred during deprovision. Please contact administrator if it persists."
		podName string
		jobErr  error
	)

	go func() {
		defer func() {
			close(stateUpdates)
			metrics.DeprovisionJobFinished()
		}()
		msgBuffer <- jobMsg
		if p.skipApbExecution {
			log.Debug("skipping deprovision and sending complete msg to channel")
			jobMsg.State.State = apb.StateSucceeded
			msgBuffer <- jobMsg
			return
		}
		podName, jobErr = p.deprovision(p.serviceInstance, stateUpdates)
	}()
	//read our status updates and send on updated JobMsgs for the subscriber to persist
	for su := range stateUpdates {
		su.Token = token
		su.Method = apb.JobMethodDeprovision
		msgBuffer <- JobMsg{InstanceUUID: p.serviceInstance.ID.String(), JobToken: token, State: su, PodName: su.Podname}
	}
	//Once our job is complete and the status updates channel closed evaluate jobErr to see was it successful or not
	if jobErr != nil {
		// send error message, can't have
		// an error type in a struct you want marshalled
		// https://github.com/golang/go/issues/5161
		// set the full error and log it
		jobMsg.State.Error = jobErr.Error()
		log.Errorf("broker::Deprovision error occurred. %v", jobErr)
		if jobErr == apb.ErrorPodPullErr {
			// Because we know the error we should send it back.
			errMsg = jobErr.Error()
		}
		jobMsg.State.State = apb.StateFailed
		// set the description as a displayable error
		jobMsg.State.Description = errMsg
		msgBuffer <- jobMsg
		return
	}
	// no error so success
	log.Debug("sending deprovision complete msg to channel")
	jobMsg.State.State = apb.StateSucceeded
	jobMsg.PodName = podName
	jobMsg.State.Description = "completed deprovision job"
	msgBuffer <- jobMsg
}
