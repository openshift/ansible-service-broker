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

	"github.com/automationbroker/bundle-lib/bundle"
	"github.com/automationbroker/bundle-lib/runtime"
	"github.com/openshift/ansible-service-broker/pkg/metrics"
	log "github.com/sirupsen/logrus"
)

type metricsHookFn func()
type runFn func(bundle.Executor) <-chan bundle.StatusMessage

type apbJob struct {
	serviceInstanceID      string
	specID                 string
	bindingID              *string
	method                 bundle.JobMethod
	metricsJobStartHook    metricsHookFn
	metricsJobFinishedHook metricsHookFn
	executor               bundle.Executor
	run                    runFn

	// NOTE: skipExecution is an artifact of an older time when we did not have
	// spec level support for some async actions (like bind). In time, this should
	// be entirely removed.
	skipExecution bool
}

func (j *apbJob) ID() string {
	if (j.method == bundle.JobMethodUnbind || j.method == bundle.JobMethodBind) && j.bindingID != nil {
		return *j.bindingID
	}
	return j.serviceInstanceID
}

func (j *apbJob) Method() bundle.JobMethod {
	return j.method
}

func (j *apbJob) Run(token string, msgBuffer chan<- JobMsg) {
	var (
		err     error
		podName string
		jobMsg  JobMsg
		exec    = j.executor
		errMsg  = fmt.Sprintf(
			"Error occurred during %s. Please contact administrator if the issue persists.", j.method)
	)

	j.metricsJobStartHook()
	defer j.metricsJobFinishedHook()

	if j.skipExecution {
		log.Debugf("skipExecution: True for %s, sending complete msg to channel", j.method)
		msgBuffer <- j.createJobMsg(
			"", token, bundle.StateSucceeded, fmt.Sprintf("%s job completed", j.method))
		return
	}

	for status := range j.run(exec) {
		podName = exec.PodName()
		jobMsg = j.createJobMsg(podName, token, status.State, status.Description)
		if status.State == bundle.StateInProgress {
			// Only send intermediate messages since the final ones are processed
			// and messaged separately (otherwise we'll double up).
			msgBuffer <- jobMsg
		}
	}

	err = exec.LastStatus().Error

	if err != nil {
		log.Errorf("broker::%s error occurred. %s", j.method, err.Error())

		if err == runtime.ErrorPodPullErr {
			errMsg = err.Error()
		} else if runtime.IsErrorCustomMsg(err) {
			errMsg = err.Error()
		}

		jobMsg.State.State = bundle.StateFailed
		// send error message, can't have
		// an error type in a struct you want marshalled
		// https://github.com/golang/go/issues/5161
		jobMsg.State.Error = err.Error()
		jobMsg.State.Description = errMsg
		msgBuffer <- jobMsg
		return
	}

	extCreds := exec.ExtractedCredentials()
	if extCreds != nil {
		jobMsg.ExtractedCredentials = *extCreds
	}

	// pull out dashboard url from exec.
	if exec.DashboardURL() != "" {
		jobMsg.DashboardURL = exec.DashboardURL()
	}

	jobMsg.State.State = bundle.StateSucceeded
	jobMsg.State.Description = fmt.Sprintf("%s job completed", j.method)
	msgBuffer <- jobMsg
}

func (j *apbJob) createJobMsg(
	podName string, token string,
	state bundle.State, description string,
) JobMsg {
	jobMsg := JobMsg{
		PodName:      podName,
		InstanceUUID: j.serviceInstanceID,
		JobToken:     token,
		SpecID:       j.specID,
		State: bundle.JobState{
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

type workFactory struct {
}

// NewWorkFactory will return a work factory capable of creating different kinds of work
func NewWorkFactory() WorkFactory {
	return &workFactory{}
}

// NewProvisionJob will setup a Work implementation that will perform the provision work
func (wf *workFactory) NewProvisionJob(si *bundle.ServiceInstance) Work {
	return &provisionJob{
		apbJob: apbJob{
			executor:               bundle.NewExecutor(bundle.ExecutorConfig{}),
			serviceInstanceID:      si.ID.String(),
			specID:                 si.Spec.ID,
			method:                 bundle.JobMethodProvision,
			metricsJobStartHook:    metrics.ProvisionJobStarted,
			metricsJobFinishedHook: metrics.ProvisionJobFinished,
			skipExecution:          false,
			run: func(exec bundle.Executor) <-chan bundle.StatusMessage {
				return exec.Provision(si)
			},
		},
		serviceInstance: si,
	}
}

// NewDeprovisionJob will setup a Work implementation that will perform the deprovision work
func (wf *workFactory) NewDeprovisionJob(si *bundle.ServiceInstance, skipExecution bool) Work {
	return &deprovisionJob{
		apbJob: apbJob{
			executor:               bundle.NewExecutor(bundle.ExecutorConfig{}),
			serviceInstanceID:      si.ID.String(),
			specID:                 si.Spec.ID,
			method:                 bundle.JobMethodDeprovision,
			metricsJobStartHook:    metrics.DeprovisionJobStarted,
			metricsJobFinishedHook: metrics.DeprovisionJobFinished,
			skipExecution:          skipExecution,
			run: func(e bundle.Executor) <-chan bundle.StatusMessage {
				return e.Deprovision(si)
			},
		},
		serviceInstance: si,
	}
}

// NewUnbindJob will setup a Work implementation that will perform the unbind work
func (wf *workFactory) NewUnbindJob(bindingID string, params *bundle.Parameters, si *bundle.ServiceInstance, skipExecution bool) Work {
	return &unbindJob{
		apbJob: apbJob{
			executor:               bundle.NewExecutor(bundle.ExecutorConfig{}),
			serviceInstanceID:      si.ID.String(),
			specID:                 si.Spec.ID,
			method:                 bundle.JobMethodUnbind,
			metricsJobStartHook:    metrics.UnbindJobStarted,
			metricsJobFinishedHook: metrics.UnbindJobFinished,
			skipExecution:          skipExecution,
			run: func(e bundle.Executor) <-chan bundle.StatusMessage {
				return e.Unbind(si, params, bindingID)
			},
		},
	}
}

// NewBindJob will setup a Work implementation that will perform the bind work
func (wf *workFactory) NewBindJob(bindingID string, bindingParams *bundle.Parameters, si *bundle.ServiceInstance) Work {
	return &bindJob{
		apbJob: apbJob{
			executor:               bundle.NewExecutor(bundle.ExecutorConfig{}),
			serviceInstanceID:      si.ID.String(),
			specID:                 si.Spec.ID,
			bindingID:              &bindingID,
			method:                 bundle.JobMethodBind,
			metricsJobStartHook:    metrics.BindJobStarted,
			metricsJobFinishedHook: metrics.BindJobFinished,
			skipExecution:          false,
			run: func(e bundle.Executor) <-chan bundle.StatusMessage {
				return e.Bind(si, bindingParams, bindingID)
			},
		},
	}
}

// NewUpdateJob will setup a Work implementation that will perform the update work
func (wf *workFactory) NewUpdateJob(si *bundle.ServiceInstance) Work {
	return &updateJob{
		apbJob: apbJob{
			executor:               bundle.NewExecutor(bundle.ExecutorConfig{}),
			serviceInstanceID:      si.ID.String(),
			specID:                 si.Spec.ID,
			method:                 bundle.JobMethodUpdate,
			metricsJobStartHook:    metrics.UpdateJobStarted,
			metricsJobFinishedHook: metrics.UpdateJobFinished,
			skipExecution:          false,
			run: func(exec bundle.Executor) <-chan bundle.StatusMessage {
				return exec.Update(si)
			},
		},
	}
}

type provisionJob struct {
	apbJob
	serviceInstance *bundle.ServiceInstance
}

type deprovisionJob struct {
	apbJob
	serviceInstance *bundle.ServiceInstance
	skipExecution   bool
}

type bindJob struct {
	apbJob
	serviceInstance *bundle.ServiceInstance
	bindingID       string
	params          *bundle.Parameters
}

type unbindJob struct {
	apbJob
	serviceInstance *bundle.ServiceInstance
	bindingID       string
	params          *bundle.Parameters
	skipExecution   bool
}

type updateJob struct {
	apbJob
	serviceInstance *bundle.ServiceInstance
}
