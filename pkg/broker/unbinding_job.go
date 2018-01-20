//
// Copyright (c) 2017 Red Hat, Inc.
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
// Red Hat trademarks are not licensed under Apache License, Version 2.
// No permission is granted to use or replicate Red Hat trademarks that
// are incorporated in this software or its documentation.
//

package broker

import (
	"github.com/apex/log"
	"github.com/openshift/ansible-service-broker/pkg/apb"
	"github.com/openshift/ansible-service-broker/pkg/metrics"
)

// UnbindingJob - Job to provision
type UnbindingJob struct {
	serviceInstance  *apb.ServiceInstance
	bindInstance     *apb.BindInstance
	params           *apb.Parameters
	skipApbExecution bool
	unbind           apb.UnBinder
}

// NewUnbindingJob - Create a new binding job.
func NewUnbindingJob(serviceInstance *apb.ServiceInstance, bindInstance *apb.BindInstance, params *apb.Parameters, unbind apb.UnBinder, skipApbExecution bool) *UnbindingJob {
	return &UnbindingJob{
		serviceInstance:  serviceInstance,
		bindInstance:     bindInstance,
		params:           params,
		skipApbExecution: skipApbExecution,
		unbind:           unbind,
	}
}

// Run - run the binding job.
func (p *UnbindingJob) Run(token string, msgBuffer chan<- JobMsg) {
	metrics.UnbindingJobStarted()
	defer metrics.UnbindingJobFinished()
	jobMsg := JobMsg{
		InstanceUUID: p.serviceInstance.ID.String(),
		JobToken:     token,
		SpecID:       p.serviceInstance.Spec.ID,
		BindingUUID:  p.bindInstance.ID.String(),
		State: apb.JobState{
			State:  apb.StateInProgress,
			Method: apb.JobMethodUnbind,
			Token:  token,
		},
	}
	msgBuffer <- jobMsg
	log.Debugf("unbindjob: unbinding job (%v) started, calling apb.Unbind", token)

	if p.skipApbExecution {
		log.Info("unbinding job (%v) skipping apb execution", token)
		jobMsg.State.State = apb.StateSucceeded
		jobMsg.Msg = "unbind finished, execution skipped"
		msgBuffer <- jobMsg
		return
	}

	err := p.unbind(p.serviceInstance, p.params)

	log.Debug("unbindjob: returned from apb.Unbind")

	if err != nil {
		errMsg := "Error occurred during binding. Please contact administrator if it persists."
		log.Errorf("unbindjob::Unbinding error occurred.\n%s", err.Error())

		// send error message
		// can't have an error type in a struct you want marshalled
		// https://github.com/golang/go/issues/5161
		// Because we know the error we should return that error.
		if err == apb.ErrorPodPullErr {
			errMsg = err.Error()
		}
		jobMsg.State.State = apb.StateFailed
		jobMsg.State.Error = errMsg
		msgBuffer <- jobMsg
		return
	}

	log.Debug("unbindjob: Looks like we're done")
	jobMsg.Msg = "unbind finished"
	jobMsg.State.State = apb.StateSucceeded
	msgBuffer <- jobMsg
}
