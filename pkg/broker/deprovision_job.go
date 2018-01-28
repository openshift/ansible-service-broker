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
}

// NewDeprovisionJob - Create a deprovision job.
func NewDeprovisionJob(
	serviceInstance *apb.ServiceInstance, skipApbExecution bool,
) *DeprovisionJob {
	return &DeprovisionJob{
		serviceInstance:  serviceInstance,
		skipApbExecution: skipApbExecution,
	}
}

// Run - will run the deprovision job.
func (p *DeprovisionJob) Run(token string, msgBuffer chan<- JobMsg) {
	metrics.DeprovisionJobStarted()
	defer metrics.DeprovisionJobFinished()
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
	msgBuffer <- jobMsg

	if p.skipApbExecution {
		log.Debug("skipping deprovision and sending complete msg to channel")
		jobMsg.State.State = apb.StateSucceeded
		msgBuffer <- jobMsg
		return
	}

	podName, err := apb.Deprovision(p.serviceInstance)
	if err != nil {
		log.Errorf("broker::Deprovision error occurred. %v", err)
		errMsg := "Error occurred during deprovision. Please contact administrator if it persists."
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

	log.Debug("sending deprovision complete msg to channel")
	jobMsg.State.State = apb.StateSucceeded
	jobMsg.PodName = podName
	msgBuffer <- jobMsg
}
