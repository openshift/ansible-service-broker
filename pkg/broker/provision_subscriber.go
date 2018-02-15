//
// Copyright (c) 2018 Red Hat, Inc.
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

// ProvisionWorkSubscriber - Listen for provision messages
type ProvisionWorkSubscriber struct {
	dao SubscriberDAO
}

// NewProvisionWorkSubscriber - Create a new work subscriber.
func NewProvisionWorkSubscriber(dao SubscriberDAO) *ProvisionWorkSubscriber {
	return &ProvisionWorkSubscriber{dao: dao}
}

// Subscribe - will start the work subscriber listening on the message buffer for provision messages.
func (p *ProvisionWorkSubscriber) Subscribe(msgBuffer <-chan JobMsg) {
	go func() {
		log.Info("Listening for provision messages")
		for msg := range msgBuffer {
			log.Debug("received provision message from buffer")

			if _, err := p.dao.SetState(msg.InstanceUUID, msg.State); err != nil {
				log.Errorf("failed to set state after provision %v", err)
			}
		}
	}()
}
