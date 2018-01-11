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
	"github.com/openshift/ansible-service-broker/pkg/dao"
	"github.com/openshift/ansible-service-broker/pkg/metrics"
)

// BindingWorkSubscriber - Listen for binding messages
type BindingWorkSubscriber struct {
	dao       *dao.Dao
	msgBuffer <-chan JobMsg
}

// NewBindingWorkSubscriber - Creates a new work subscriber
func NewBindingWorkSubscriber(dao *dao.Dao) *BindingWorkSubscriber {
	return &BindingWorkSubscriber{dao: dao}
}

// Subscribe - will start a work subscriber listening for bind job messages
func (b *BindingWorkSubscriber) Subscribe(msgBuffer <-chan JobMsg) {
	b.msgBuffer = msgBuffer

	go func() {
		log.Info("Listening for binding messages")
		for {
			msg := <-msgBuffer
			var extCreds *apb.ExtractedCredentials
			metrics.BindingJobFinished()

			log.Debug("Processed binding message from buffer")

			if msg.Error != "" {
				log.Errorf("bindsub: Binding job reporting error: %s", msg.Error)
				if err := b.dao.SetState(msg.InstanceUUID, apb.JobState{
					Token:   msg.JobToken,
					State:   apb.StateFailed,
					Podname: msg.PodName,
					Method:  apb.JobMethodBind,
				}); err != nil {
					log.Errorf("failed to set state after bind %v", err)
				}
			} else if msg.Msg == "" {
				if err := b.dao.SetState(msg.InstanceUUID, apb.JobState{
					Token:   msg.JobToken,
					State:   apb.StateInProgress,
					Podname: msg.PodName,
					Method:  apb.JobMethodBind,
				}); err != nil {
					log.Errorf("failed to set state after bind %v", err)
				}
			} else {
				log.Debug("bindsub: getting creds")
				if err := json.Unmarshal([]byte(msg.Msg), &extCreds); err != nil {
					log.Errorf("failed to unmarshal extracted credentials after bind %v", err)
				}
				if err := b.dao.SetState(msg.InstanceUUID, apb.JobState{
					Token:   msg.JobToken,
					State:   apb.StateSucceeded,
					Podname: msg.PodName,
					Method:  apb.JobMethodBind,
				}); err != nil {
					log.Errorf("failed to set state after bind %v", err)
				}

				log.Debug("CALL SetExtractedCredentials $v - %v", msg.BindingUUID, extCreds)
				if err := b.dao.SetExtractedCredentials(msg.BindingUUID, extCreds); err != nil {
					log.Errorf("failed to set extracted credentials after bind %v", err)
				}
			}
		}
	}()
}
