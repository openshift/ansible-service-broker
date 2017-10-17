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

	logging "github.com/op/go-logging"
	"github.com/openshift/ansible-service-broker/pkg/apb"
)

// ProvisionJob - Job to provision
type ProvisionJob struct {
	serviceInstance *apb.ServiceInstance
	clusterConfig   apb.ClusterConfig
	log             *logging.Logger
	task            string
}

// ProvisionMsg - Message to be returned from the provision job
type ProvisionMsg struct {
	InstanceUUID string `json:"instance_uuid"`
	JobToken     string `json:"job_token"`
	SpecID       string `json:"spec_id"`
	PodName      string `json:"podname"`
	Msg          string `json:"msg"`
	Error        string `json:"error"`
}

// Render - Display the provision message.
func (m ProvisionMsg) Render() string {
	render, _ := json.Marshal(m)
	return string(render)
}

// NewProvisionJob - Create a new provision job.
func NewProvisionJob(task string, serviceInstance *apb.ServiceInstance, clusterConfig apb.ClusterConfig,
	log *logging.Logger,
) *ProvisionJob {
	return &ProvisionJob{
		serviceInstance: serviceInstance,
		clusterConfig:   clusterConfig,
		log:             log,
		task:            task}
}

// Run - run the provision job.
func (p *ProvisionJob) Run(token string, msgBuffer chan<- WorkMsg) {
	podName, extCreds, err := apb.Provision(p.serviceInstance, p.clusterConfig, p.log)

	if err != nil {
		p.log.Error("broker::Provision error occurred.")
		p.log.Errorf("%s", err.Error())

		// send error message
		// can't have an error type in a struct you want marshalled
		// https://github.com/golang/go/issues/5161
		msgBuffer <- ProvisionMsg{InstanceUUID: p.serviceInstance.ID.String(),
			JobToken: token, SpecID: p.serviceInstance.Spec.ID, PodName: "", Msg: "", Error: err.Error()}
		return
	}

	// send creds
	jsonmsg, err := json.Marshal(extCreds)
	if err != nil {
		msgBuffer <- ProvisionMsg{InstanceUUID: p.serviceInstance.ID.String(),
			JobToken: token, SpecID: p.serviceInstance.Spec.ID, PodName: "", Msg: "", Error: err.Error()}
		return
	}

	msgBuffer <- ProvisionMsg{InstanceUUID: p.serviceInstance.ID.String(),
		JobToken: token, SpecID: p.serviceInstance.Spec.ID, PodName: podName, Msg: string(jsonmsg), Error: ""}
}
