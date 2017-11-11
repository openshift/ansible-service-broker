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

	logging "github.com/op/go-logging"
	"github.com/openshift/ansible-service-broker/pkg/apb"
	"github.com/openshift/ansible-service-broker/pkg/metrics"
)

// UpdateJob - Job to update
type UpdateJob struct {
	serviceInstance *apb.ServiceInstance
	clusterConfig   apb.ClusterConfig
	log             *logging.Logger
}

// UpdateMsg - Message to be returned from the update job
type UpdateMsg struct {
	InstanceUUID string `json:"instance_uuid"`
	JobToken     string `json:"job_token"`
	SpecID       string `json:"spec_id"`
	PodName      string `json:"podname"`
	Msg          string `json:"msg"`
	Error        string `json:"error"`
}

// Render - Display the update message.
func (m UpdateMsg) Render() string {
	render, _ := json.Marshal(m)
	return string(render)
}

// NewUpdateJob - Create a new update job.
func NewUpdateJob(serviceInstance *apb.ServiceInstance, clusterConfig apb.ClusterConfig,
	log *logging.Logger,
) *UpdateJob {
	return &UpdateJob{
		serviceInstance: serviceInstance,
		clusterConfig:   clusterConfig,
		log:             log,
	}
}

// Run - run the update job.
func (u *UpdateJob) Run(token string, msgBuffer chan<- WorkMsg) {
	metrics.UpdateJobStarted()
	podName, extCreds, err := apb.Update(u.serviceInstance, u.clusterConfig, u.log)

	if err != nil {
		u.log.Error("broker::Update error occurred.")
		u.log.Errorf("%s", err.Error())

		// send error message
		// can't have an error type in a struct you want marshalled
		// https://github.com/golang/go/issues/5161
		msgBuffer <- UpdateMsg{InstanceUUID: u.serviceInstance.ID.String(),
			JobToken: token, SpecID: u.serviceInstance.Spec.ID, PodName: "", Msg: "", Error: err.Error()}
		return
	}

	// send creds
	jsonmsg, err := json.Marshal(extCreds)
	if err != nil {
		msgBuffer <- UpdateMsg{InstanceUUID: u.serviceInstance.ID.String(),
			JobToken: token, SpecID: u.serviceInstance.Spec.ID, PodName: "", Msg: "", Error: err.Error()}
		return
	}

	msgBuffer <- UpdateMsg{InstanceUUID: u.serviceInstance.ID.String(),
		JobToken: token, SpecID: u.serviceInstance.Spec.ID, PodName: podName, Msg: string(jsonmsg), Error: ""}
}
