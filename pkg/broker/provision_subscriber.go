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
)

// ProvisionWorkSubscriber - Lissten for provision messages
type ProvisionWorkSubscriber struct {
	dao       *dao.Dao
	log       *logging.Logger
	msgBuffer <-chan WorkMsg
}

// NewProvisionWorkSubscriber - Create a new work subscriber.
func NewProvisionWorkSubscriber(dao *dao.Dao, log *logging.Logger) *ProvisionWorkSubscriber {
	return &ProvisionWorkSubscriber{dao: dao, log: log}
}

// Subscribe - will start the work subscriber listenning on the message buffer for provision messages.
func (p *ProvisionWorkSubscriber) Subscribe(msgBuffer <-chan WorkMsg) {
	p.msgBuffer = msgBuffer

	go func() {
		p.log.Info("Listening for provision messages")
		for {
			msg := <-msgBuffer
			var pmsg *ProvisionMsg
			var extCreds *apb.ExtractedCredentials

			p.log.Debug("Processed provision message from buffer")
			// HACK: this seems like a hack, there's probably a better way to
			// get the data sent through instead of a string
			json.Unmarshal([]byte(msg.Render()), &pmsg)

			if pmsg.Error != "" {
				p.log.Errorf("Provision job reporting error: %s", pmsg.Error)
				p.dao.SetState(pmsg.InstanceUUID, apb.JobState{
					Token:         pmsg.JobToken,
					State:         apb.StateFailed,
					Podname:       pmsg.PodName,
					APBMethodType: apb.JobStateAPBMethodTypeProvision,
				})
			} else if pmsg.Msg == "" {
				// HACK: OMG this is horrible. We should probably pass in a
				// state. Since we'll also be using this to get more granular
				// updates one day.
				p.dao.SetState(pmsg.InstanceUUID, apb.JobState{
					Token:         pmsg.JobToken,
					State:         apb.StateInProgress,
					Podname:       pmsg.PodName,
					APBMethodType: apb.JobStateAPBMethodTypeProvision,
				})
			} else {
				json.Unmarshal([]byte(pmsg.Msg), &extCreds)
				p.dao.SetState(pmsg.InstanceUUID, apb.JobState{
					Token:         pmsg.JobToken,
					State:         apb.StateSucceeded,
					Podname:       pmsg.PodName,
					APBMethodType: apb.JobStateAPBMethodTypeProvision,
				})
				p.dao.SetExtractedCredentials(pmsg.InstanceUUID, extCreds)
			}
		}
	}()
}
