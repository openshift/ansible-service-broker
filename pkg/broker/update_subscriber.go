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

// UpdateWorkSubscriber - Lissten for provision messages
type UpdateWorkSubscriber struct {
	dao SubscriberDAO
}

// NewUpdateWorkSubscriber - Create a new work subscriber.
func NewUpdateWorkSubscriber(dao SubscriberDAO) *UpdateWorkSubscriber {
	return &UpdateWorkSubscriber{dao: dao}
}

// Subscribe - will start the work subscriber listenning on the message buffer for provision messages.
func (u *UpdateWorkSubscriber) Subscribe(msgBuffer <-chan JobMsg) {
	go func() {
		log.Info("Listening for update messages")
		for msg := range msgBuffer {
			log.Debug("received update message from buffer")
			if _, err := u.dao.SetState(msg.InstanceUUID, msg.State); err != nil {
				log.Errorf("failed to set state after update %v", err)
			}
		}
	}()
}
