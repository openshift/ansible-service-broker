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
	"github.com/openshift/ansible-service-broker/pkg/dao"
	"github.com/openshift/ansible-service-broker/pkg/metrics"
)

// ProvisionWorkSubscriber - Lissten for provision messages
type ProvisionWorkSubscriber struct {
	dao       *dao.Dao
	msgBuffer <-chan JobMsg
}

// NewProvisionWorkSubscriber - Create a new work subscriber.
func NewProvisionWorkSubscriber(dao *dao.Dao) *ProvisionWorkSubscriber {
	return &ProvisionWorkSubscriber{dao: dao}
}

// Subscribe - will start the work subscriber listenning on the message buffer for provision messages.
func (p *ProvisionWorkSubscriber) Subscribe(msgBuffer <-chan JobMsg) {
	p.msgBuffer = msgBuffer

	go func() {
		log.Info("Listening for provision messages")
		for {
			msg := <-msgBuffer
			var extCreds *apb.ExtractedCredentials
			metrics.ProvisionJobFinished()

			log.Debug("Processed provision message from buffer")

			if msg.Error != "" {
				log.Errorf("Provision job reporting error: %s", msg.Error)
				if err := p.dao.SetState(msg.InstanceUUID, apb.JobState{
					Token:   msg.JobToken,
					State:   apb.StateFailed,
					Podname: msg.PodName,
					Method:  apb.JobMethodProvision,
				}); err != nil {
					log.Errorf("failed to set state after provision %v", err)
				}
			} else if msg.Msg == "" {
				// HACK: OMG this is horrible. We should probably pass in a
				// state. Since we'll also be using this to get more granular
				// updates one day.
				if err := p.dao.SetState(msg.InstanceUUID, apb.JobState{
					Token:   msg.JobToken,
					State:   apb.StateInProgress,
					Podname: msg.PodName,
					Method:  apb.JobMethodProvision,
				}); err != nil {
					log.Errorf("failed to set state after provision %v", err)
				}
			} else {
				if err := json.Unmarshal([]byte(msg.Msg), &extCreds); err != nil {
					log.Errorf("failed to unmarshal extracted credentials after provision %v", err)
				}
				if err := p.dao.SetState(msg.InstanceUUID, apb.JobState{
					Token:   msg.JobToken,
					State:   apb.StateSucceeded,
					Podname: msg.PodName,
					Method:  apb.JobMethodProvision,
				}); err != nil {
					log.Errorf("failed to set state after provision %v", err)
				}
				if err := p.dao.SetExtractedCredentials(msg.InstanceUUID, extCreds); err != nil {
					log.Errorf("failed to set extracted credentials after provision %v", err)
				}
			}
		}
	}()
}
