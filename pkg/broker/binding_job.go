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
	"encoding/json"

	"github.com/openshift/ansible-service-broker/pkg/apb"
	"github.com/openshift/ansible-service-broker/pkg/metrics"
	"github.com/pborman/uuid"
)

// BindingJob - Job to provision
type BindingJob struct {
	serviceInstance *apb.ServiceInstance
	bindingUUID     uuid.UUID
	params          *apb.Parameters
}

// NewBindingJob - Create a new binding job.
func NewBindingJob(serviceInstance *apb.ServiceInstance, bindingUUID uuid.UUID, params *apb.Parameters) *BindingJob {
	return &BindingJob{
		serviceInstance: serviceInstance,
		bindingUUID:     bindingUUID,
		params:          params,
	}
}

// Run - run the binding job.
func (p *BindingJob) Run(token string, msgBuffer chan<- JobMsg) {
	metrics.BindingJobStarted()

	log.Debug("bindjob: binding job started, calling apb.Bind")

	podName, extCreds, err := apb.Bind(p.serviceInstance, p.params)

	log.Debug("bindjob: returned from apb.Bind")

	if err != nil {
		log.Errorf("bindjob::Binding error occurred.\n%s", err.Error())

		// send error message
		// can't have an error type in a struct you want marshalled
		// https://github.com/golang/go/issues/5161
		msgBuffer <- JobMsg{InstanceUUID: p.serviceInstance.ID.String(),
			BindingUUID: p.bindingUUID.String(), JobToken: token,
			SpecID: p.serviceInstance.Spec.ID, PodName: "", Msg: "", Error: err.Error()}
		return
	}

	// send creds
	jsonmsg, err := json.Marshal(extCreds)
	if err != nil {
		msgBuffer <- JobMsg{InstanceUUID: p.serviceInstance.ID.String(),
			BindingUUID: p.bindingUUID.String(), JobToken: token,
			SpecID: p.serviceInstance.Spec.ID, PodName: "", Msg: "", Error: err.Error()}
		return
	}

	log.Debug("bindjob: looks like we're done, sending credentials")
	msgBuffer <- JobMsg{InstanceUUID: p.serviceInstance.ID.String(),
		BindingUUID: p.bindingUUID.String(), JobToken: token,
		SpecID: p.serviceInstance.Spec.ID, PodName: podName, Msg: string(jsonmsg), Error: ""}
}
