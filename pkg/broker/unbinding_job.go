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
	"github.com/openshift/ansible-service-broker/pkg/apb"
	"github.com/openshift/ansible-service-broker/pkg/metrics"
	"github.com/pborman/uuid"
)

// UnbindingJob - Job to provision
type UnbindingJob struct {
	serviceInstance  *apb.ServiceInstance
	bindingUUID      uuid.UUID
	params           *apb.Parameters
	skipApbExecution bool
}

// NewUnbindingJob - Create a new binding job.
func NewUnbindingJob(serviceInstance *apb.ServiceInstance, bindingUUID uuid.UUID, params *apb.Parameters, skipApbExecution bool) *UnbindingJob {
	return &UnbindingJob{
		serviceInstance:  serviceInstance,
		bindingUUID:      bindingUUID,
		params:           params,
		skipApbExecution: skipApbExecution,
	}
}

// Run - run the binding job.
func (p *UnbindingJob) Run(token string, msgBuffer chan<- JobMsg) {
	metrics.UnbindingJobStarted()

	log.Debugf("unbindjob: unbinding job (%v) started, calling apb.Unbind", token)

	if p.skipApbExecution {
		log.Info("unbinding job (%v) skipping apb execution", token)
		msgBuffer <- JobMsg{InstanceUUID: p.serviceInstance.ID.String(),
			BindingUUID: p.bindingUUID.String(), JobToken: token,
			SpecID: p.serviceInstance.Spec.ID, PodName: "",
			Msg: "unbind finished, execution skipped", Error: ""}
		return
	}

	err := apb.Unbind(p.serviceInstance, p.params)

	log.Debug("unbindjob: returned from apb.Unbind")

	if err != nil {
		log.Errorf("unbindjob::Unbinding error occurred.\n%s", err.Error())

		// send error message
		// can't have an error type in a struct you want marshalled
		// https://github.com/golang/go/issues/5161
		msgBuffer <- JobMsg{InstanceUUID: p.serviceInstance.ID.String(),
			BindingUUID: p.bindingUUID.String(), JobToken: token,
			SpecID: p.serviceInstance.Spec.ID, PodName: "", Msg: "", Error: err.Error()}
		return
	}

	log.Debug("unbindjob: Looks like we're done")
	msgBuffer <- JobMsg{InstanceUUID: p.serviceInstance.ID.String(),
		BindingUUID: p.bindingUUID.String(), JobToken: token,
		SpecID: p.serviceInstance.Spec.ID, PodName: "", Msg: "unbind finished", Error: ""}
}
