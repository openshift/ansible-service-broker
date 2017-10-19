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
	"github.com/openshift/ansible-service-broker/pkg/dao"
	"github.com/openshift/ansible-service-broker/pkg/metrics"
)

// UpdateWorkSubscriber - Lissten for provision messages
type UpdateWorkSubscriber struct {
	dao       *dao.Dao
	log       *logging.Logger
	msgBuffer <-chan WorkMsg
}

// NewUpdateWorkSubscriber - Create a new work subscriber.
func NewUpdateWorkSubscriber(dao *dao.Dao, log *logging.Logger) *UpdateWorkSubscriber {
	return &UpdateWorkSubscriber{dao: dao, log: log}
}

// Subscribe - will start the work subscriber listenning on the message buffer for provision messages.
func (u *UpdateWorkSubscriber) Subscribe(msgBuffer <-chan WorkMsg) {
	u.msgBuffer = msgBuffer

	go func() {
		u.log.Info("Listening for provision messages")
		for {
			msg := <-msgBuffer
			var umsg *UpdateMsg
			var extCreds *apb.ExtractedCredentials
			metrics.UpdateJobFinished()

			u.log.Debug("Processed provision message from buffer")
			// HACK: this seems like a hack, there's probably a better way to
			// get the data sent through instead of a string
			json.Unmarshal([]byte(msg.Render()), &umsg)

			if umsg.Error != "" {
				u.log.Errorf("Update job reporting error: %s", umsg.Error)
				u.dao.SetState(umsg.InstanceUUID, apb.JobState{
					Token:   umsg.JobToken,
					State:   apb.StateFailed,
					Podname: umsg.PodName,
					Method:  apb.JobMethodUpdate,
				})
			} else if umsg.Msg == "" {
				// HACK: OMG this is horrible. We should probably pass in a
				// state. Since we'll also be using this to get more granular
				// updates one day.
				u.dao.SetState(umsg.InstanceUUID, apb.JobState{
					Token:   umsg.JobToken,
					State:   apb.StateInProgress,
					Podname: umsg.PodName,
					Method:  apb.JobMethodUpdate,
				})
			} else {
				json.Unmarshal([]byte(umsg.Msg), &extCreds)
				u.dao.SetState(umsg.InstanceUUID, apb.JobState{
					Token:   umsg.JobToken,
					State:   apb.StateSucceeded,
					Podname: umsg.PodName,
					Method:  apb.JobMethodUpdate,
				})
				u.dao.SetExtractedCredentials(umsg.InstanceUUID, extCreds)
			}
		}
	}()
}
