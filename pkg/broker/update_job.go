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

package broker

import (
	"encoding/json"

	"github.com/openshift/ansible-service-broker/pkg/apb"
	"github.com/openshift/ansible-service-broker/pkg/metrics"
)

// UpdateJob - Job to update
type UpdateJob struct {
	serviceInstance *apb.ServiceInstance
}

// NewUpdateJob - Create a new update job.
func NewUpdateJob(serviceInstance *apb.ServiceInstance) *UpdateJob {
	return &UpdateJob{
		serviceInstance: serviceInstance,
	}
}

// Run - run the update job.
func (u *UpdateJob) Run(token string, msgBuffer chan<- JobMsg) {
	metrics.UpdateJobStarted()
	podName, extCreds, err := apb.Update(u.serviceInstance)

	if err != nil {
		log.Error("broker::Update error occurred.")
		log.Errorf("%s", err.Error())

		// Because we know the error we should return that error.
		if err == apb.ErrorPodPullErr {
			// send error message, can't have
			// an error type in a struct you want marshalled
			// https://github.com/golang/go/issues/5161
			msgBuffer <- JobMsg{InstanceUUID: u.serviceInstance.ID.String(),
				JobToken: token,
				SpecID:   u.serviceInstance.Spec.ID,
				PodName:  "",
				Msg:      "",
				Error:    err.Error()}
			return
		}
		//Unkown error defaulting to generic message.
		msgBuffer <- JobMsg{InstanceUUID: u.serviceInstance.ID.String(),
			JobToken: token,
			SpecID:   u.serviceInstance.Spec.ID,
			PodName:  "",
			Msg:      "",
			Error:    "Error occured during update. Please contact administrator if it presists."}
		return
	}

	// send creds
	jsonmsg, err := json.Marshal(extCreds)
	if err != nil {
		msgBuffer <- JobMsg{InstanceUUID: u.serviceInstance.ID.String(),
			JobToken: token, SpecID: u.serviceInstance.Spec.ID, PodName: "", Msg: "", Error: err.Error()}
		return
	}

	msgBuffer <- JobMsg{InstanceUUID: u.serviceInstance.ID.String(),
		JobToken: token, SpecID: u.serviceInstance.Spec.ID, PodName: podName, Msg: string(jsonmsg), Error: ""}
}
