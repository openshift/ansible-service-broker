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
// Red Hat trademarks are not licensed under Apache License, Version 2.
// No permission is granted to use or replicate Red Hat trademarks that
// are incorporated in this software or its documentation.
//

package broker

import (
	"github.com/automationbroker/bundle-lib/apb"
	"github.com/openshift/ansible-service-broker/pkg/dao"
)

// UnbindingWorkSubscriber - Listen for binding messages
type UnbindingWorkSubscriber struct {
	dao       dao.Dao
	msgBuffer <-chan JobMsg
}

// NewUnbindingWorkSubscriber - Creates a new work subscriber
func NewUnbindingWorkSubscriber(dao dao.Dao) *UnbindingWorkSubscriber {
	return &UnbindingWorkSubscriber{dao: dao}
}

// Subscribe - will start a work subscriber listening for bind job messages
func (b *UnbindingWorkSubscriber) Subscribe(msgBuffer <-chan JobMsg) {
	go func() {
		log.Info("Listening for binding messages")
		for msg := range msgBuffer {
			log.Debug("Processed binding message from buffer")
			if _, err := b.dao.SetState(msg.InstanceUUID, msg.State); err != nil {
				log.Errorf("failed to set state after unbind %v", err)
				continue
			}

			if msg.State.State == apb.StateSucceeded {
				svcInstance, err := b.dao.GetServiceInstance(msg.InstanceUUID)
				if err != nil {
					log.Errorf("Error getting service instance [ %s ] after unbind job",
						msg.InstanceUUID)
					setFailedUnbindJob(b.dao, msg)
					continue
				}

				bindInstance, err := b.dao.GetBindInstance(msg.BindingUUID)
				if err != nil {
					log.Errorf("Error getting bind instance [ %s ] after unbind job",
						msg.BindingUUID)
					setFailedUnbindJob(b.dao, msg)
					continue
				}

				if err := cleanupUnbind(bindInstance, svcInstance, &msg.ExtractedCredentials, b.dao); err != nil {
					log.Errorf("Failed cleaning up unbind after job, error: %v", err)
					setFailedUnbindJob(b.dao, msg)
					continue
				}
			}
		}
	}()
}

func setFailedUnbindJob(dao dao.Dao, dmsg JobMsg) {
	// have to set the state here manually as the logic that triggers this is in the subscriber
	dmsg.State.State = apb.StateFailed
	if _, err := dao.SetState(dmsg.InstanceUUID, dmsg.State); err != nil {
		log.Errorf("failed to set state after unbind %v", err)
	}
}

func cleanupUnbind(bindInstance *apb.BindInstance, serviceInstance *apb.ServiceInstance, bindExtCreds *apb.ExtractedCredentials, dao dao.Dao) error {
	var err error
	id := bindInstance.ID.String()

	if err = dao.DeleteBindInstance(id); err != nil {
		log.Errorf("failed to delete bind instance - %v", err)
		return err
	}

	serviceInstance.RemoveBinding(bindInstance.ID)
	if err = dao.SetServiceInstance(serviceInstance.ID.String(), serviceInstance); err != nil {
		log.Errorf("failed to set service instance [ %s ] during unbind of [ %s ]",
			serviceInstance.ID.String(), id)
		return err
	}
	log.Infof("Clean up of binding instance [ %s ] done. Unbinding successful", id)
	return nil
}
