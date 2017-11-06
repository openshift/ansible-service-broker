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
	"github.com/openshift/ansible-service-broker/pkg/dao"
	"github.com/openshift/ansible-service-broker/pkg/metrics"
)

// BindingWorkSubscriber - Listen for binding messages
type BindingWorkSubscriber struct {
	dao       *dao.Dao
	log       *logging.Logger
	msgBuffer <-chan WorkMsg
}

func NewBindingWorkSubscriber(dao *dao.Dao, log *logging.Logger) *BindingWorkSubscriber {
	return &BindingWorkSubscriber{dao: dao, log: log}
}

func (b *BindingWorkSubscriber) Subscribe(msgBuffer <-chan WorkMsg) {
	b.msgBuffer = msgBuffer

	go func() {
		b.log.Info("Listening for binding messages")
		for {
			msg := <-msgBuffer
			var bmsg *BindingMsg
			//var extCreds *apb.ExtractedCredentials
			metrics.BindingJobFinished()

			b.log.Debug("Processed binding message from buffer")
			// HACK: this seems like a hack, there's probably a better way to
			// get the data sent through instead of a string
			json.Unmarshal([]byte(msg.Render()), &bmsg)

			/*
				if pmsg.Error != "" {
					b.log.Errorf("Provision job reporting error: %s", pmsg.Error)
					b.dao.SetState(pmsg.InstanceUUID, apb.JobState{
						Token:   pmsg.JobToken,
						State:   apb.StateFailed,
						Podname: pmsg.PodName,
						Method:  apb.JobMethodProvision,
					})
				} else if pmsg.Msg == "" {
					// HACK: OMG this is horrible. We should probably pass in a
					// state. Since we'll also be using this to get more granular
					// updates one day.
					b.dao.SetState(pmsg.InstanceUUID, apb.JobState{
						Token:   pmsg.JobToken,
						State:   apb.StateInProgress,
						Podname: pmsg.PodName,
						Method:  apb.JobMethodProvision,
					})
				} else {
					json.Unmarshal([]byte(pmsg.Msg), &extCreds)
					b.dao.SetState(pmsg.InstanceUUID, apb.JobState{
						Token:   pmsg.JobToken,
						State:   apb.StateSucceeded,
						Podname: pmsg.PodName,
						Method:  apb.JobMethodProvision,
					})
					b.dao.SetExtractedCredentials(pmsg.InstanceUUID, extCreds)
				}
			*/
		}
	}()
}
