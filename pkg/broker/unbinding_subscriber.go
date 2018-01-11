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
	"github.com/openshift/ansible-service-broker/pkg/dao"
	"github.com/openshift/ansible-service-broker/pkg/metrics"
)

// UnbindingWorkSubscriber - Listen for binding messages
type UnbindingWorkSubscriber struct {
	dao       *dao.Dao
	msgBuffer <-chan JobMsg
}

// NewUnbindingWorkSubscriber - Creates a new work subscriber
func NewUnbindingWorkSubscriber(dao *dao.Dao) *UnbindingWorkSubscriber {
	return &UnbindingWorkSubscriber{dao: dao}
}

// Subscribe - will start a work subscriber listening for bind job messages
func (b *UnbindingWorkSubscriber) Subscribe(msgBuffer <-chan JobMsg) {
	b.msgBuffer = msgBuffer

	go func() {
		log.Info("Listening for binding messages")
		for {
			msg := <-msgBuffer
			metrics.UnbindingJobFinished()

			log.Debug("Processed binding message from buffer")

			if msg.Error != "" {
				log.Errorf("bindsub: Uninding job reporting error: %s", msg.Error)
				if err := b.dao.SetState(msg.InstanceUUID, apb.JobState{
					Token:   msg.JobToken,
					State:   apb.StateFailed,
					Podname: msg.PodName,
					Method:  apb.JobMethodUnbind,
				}); err != nil {
					log.Errorf("failed to set state after unbind %v", err)
				}
			} else if msg.Msg == "" {
				if err := b.dao.SetState(msg.InstanceUUID, apb.JobState{
					Token:   msg.JobToken,
					State:   apb.StateInProgress,
					Podname: msg.PodName,
					Method:  apb.JobMethodUnbind,
				}); err != nil {
					log.Errorf("failed to set state after unbind %v", err)
				}
			} else {
				if err := b.dao.SetState(msg.InstanceUUID, apb.JobState{
					Token:   msg.JobToken,
					State:   apb.StateSucceeded,
					Podname: msg.PodName,
					Method:  apb.JobMethodUnbind,
				}); err != nil {
					log.Errorf("failed to set state after unbind %v", err)
				}
			}
		}
	}()
}
