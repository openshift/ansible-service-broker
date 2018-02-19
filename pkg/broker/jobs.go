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
// Red Hat trademarks are not licensed under Apache License, Version 2.
// No permission is granted to use or replicate Red Hat trademarks that
// are incorporated in this software or its documentation.
//

package broker

import (
	"fmt"
	"github.com/openshift/ansible-service-broker/pkg/apb"
	"github.com/openshift/ansible-service-broker/pkg/metrics"
)

type metricsHookFn func()
type runFn func(*apb.Executor) <-chan apb.StatusMessage

type apbJob struct {
	serviceInstanceID      string
	specID                 string
	bindingID              *string
	method                 apb.JobMethod
	metricsJobStartHook    metricsHookFn
	metricsJobFinishedHook metricsHookFn
	run                    runFn

	// NOTE: skipExecution is an artifact of an older time when we did not have
	// spec level support for some async actions (like bind). In time, this should
	// be entirely removed.
	skipExecution bool
}

func (j *apbJob) Run(token string, msgBuffer chan<- JobMsg) {
	var (
		err     error
		podName string
	)
	errMsg := fmt.Sprintf(
		"Error occurred during %s. Please contact administrator if the issue persists.", j.method)

	j.metricsJobStartHook()
	defer j.metricsJobFinishedHook()

	// Initial jobMsg
	jobMsg := j.createJobMsg(
		"", token, apb.StateInProgress, fmt.Sprintf("%s job started", j.method))
	msgBuffer <- jobMsg

	if j.skipExecution {
		log.Debugf("skipExecution: True for %s, sending complete msg to channel", j.method)
		jobMsg.State.State = apb.StateSucceeded
		jobMsg.State.Description = fmt.Sprintf("%s job completed", j.method)
		msgBuffer <- jobMsg
		return
	}

	exec := apb.NewExecutor()
	for status := range j.run(exec) {
		podName = exec.GetPodName()
		msgBuffer <- j.createJobMsg(podName, token, status.State, status.Description)
	}

	err = exec.GetLastStatus().Error

	if err != nil {
		log.Errorf("broker::%s error occurred. %s", j.method, err.Error())

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

	extCreds := exec.GetExtractedCredentials()
	if extCreds != nil {
		jobMsg.ExtractedCredentials = *extCreds
	}

	jobMsg.State.State = apb.StateSucceeded
	jobMsg.PodName = podName
	jobMsg.State.Description = fmt.Sprintf("%s job completed", j.method)
	msgBuffer <- jobMsg
}

func (j *apbJob) createJobMsg(
	podName string, token string,
	state apb.State, description string,
) JobMsg {
	jobMsg := JobMsg{
		PodName:      podName,
		InstanceUUID: j.serviceInstanceID,
		JobToken:     token,
		SpecID:       j.specID,
		State: apb.JobState{
			State:       state,
			Method:      j.method,
			Token:       token,
			Description: description,
		},
	}

	if j.bindingID != nil {
		jobMsg.BindingUUID = *j.bindingID
	}

	return jobMsg
}

// ProvisionJob - Job to provision.
type ProvisionJob struct {
	serviceInstance *apb.ServiceInstance
}

// Run - Run the provision job.
func (j *ProvisionJob) Run(token string, msgBuffer chan<- JobMsg) {
	job := apbJob{
		serviceInstanceID:      j.serviceInstance.ID.String(),
		specID:                 j.serviceInstance.Spec.ID,
		method:                 apb.JobMethodProvision,
		metricsJobStartHook:    metrics.ProvisionJobStarted,
		metricsJobFinishedHook: metrics.ProvisionJobFinished,
		skipExecution:          false,
		run: func(exec *apb.Executor) <-chan apb.StatusMessage {
			return exec.Provision(j.serviceInstance)
		},
	}
	job.Run(token, msgBuffer)
}

// DeprovisionJob - Job to deprovision.
type DeprovisionJob struct {
	serviceInstance *apb.ServiceInstance
	skipExecution   bool
}

// Run - Run the deprovision job.
func (j *DeprovisionJob) Run(token string, msgBuffer chan<- JobMsg) {
	job := apbJob{
		serviceInstanceID:      j.serviceInstance.ID.String(),
		specID:                 j.serviceInstance.Spec.ID,
		method:                 apb.JobMethodDeprovision,
		metricsJobStartHook:    metrics.DeprovisionJobStarted,
		metricsJobFinishedHook: metrics.DeprovisionJobFinished,
		skipExecution:          j.skipExecution,
		run: func(e *apb.Executor) <-chan apb.StatusMessage {
			return e.Deprovision(j.serviceInstance)
		},
	}
	job.Run(token, msgBuffer)
}

// BindJob - Job to bind.
type BindJob struct {
	serviceInstance *apb.ServiceInstance
	bindingID       string
	params          *apb.Parameters
}

// Run - Run the bind job.
func (j *BindJob) Run(token string, msgBuffer chan<- JobMsg) {
	job := apbJob{
		serviceInstanceID:      j.serviceInstance.ID.String(),
		specID:                 j.serviceInstance.Spec.ID,
		bindingID:              &j.bindingID,
		method:                 apb.JobMethodBind,
		metricsJobStartHook:    metrics.BindJobStarted,
		metricsJobFinishedHook: metrics.BindJobFinished,
		skipExecution:          false,
		run: func(e *apb.Executor) <-chan apb.StatusMessage {
			return e.Bind(j.serviceInstance, j.params)
		},
	}
	job.Run(token, msgBuffer)
}

// UnbindJob - Job to unbind.
type UnbindJob struct {
	serviceInstance *apb.ServiceInstance
	bindingID       string
	params          *apb.Parameters
	skipExecution   bool
}

// Run - Run the unbind job.
func (j *UnbindJob) Run(token string, msgBuffer chan<- JobMsg) {
	job := apbJob{
		serviceInstanceID:      j.serviceInstance.ID.String(),
		specID:                 j.serviceInstance.Spec.ID,
		bindingID:              &j.bindingID,
		method:                 apb.JobMethodUnbind,
		metricsJobStartHook:    metrics.UnbindJobStarted,
		metricsJobFinishedHook: metrics.UnbindJobFinished,
		skipExecution:          j.skipExecution,
		run: func(e *apb.Executor) <-chan apb.StatusMessage {
			return e.Unbind(j.serviceInstance, j.params)
		},
	}
	job.Run(token, msgBuffer)
}

// UpdateJob - Job to update.
type UpdateJob struct {
	serviceInstance *apb.ServiceInstance
}

// Run - Run the update job.
func (j *UpdateJob) Run(token string, msgBuffer chan<- JobMsg) {
	job := apbJob{
		serviceInstanceID:      j.serviceInstance.ID.String(),
		specID:                 j.serviceInstance.Spec.ID,
		method:                 apb.JobMethodUpdate,
		metricsJobStartHook:    metrics.UpdateJobStarted,
		metricsJobFinishedHook: metrics.UpdateJobFinished,
		skipExecution:          false,
		run: func(e *apb.Executor) <-chan apb.StatusMessage {
			return e.Update(j.serviceInstance)
		},
	}
	job.Run(token, msgBuffer)
}
