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
	"github.com/openshift/ansible-service-broker/pkg/apb"
	"github.com/openshift/ansible-service-broker/pkg/dao"
	"github.com/openshift/ansible-service-broker/pkg/metrics"
)

// DeprovisionWorkSubscriber - Lissten for provision messages
type DeprovisionWorkSubscriber struct {
	dao       *dao.Dao
	msgBuffer <-chan JobMsg
}

// NewDeprovisionWorkSubscriber - Create a new work subscriber.
func NewDeprovisionWorkSubscriber(dao *dao.Dao) *DeprovisionWorkSubscriber {
	return &DeprovisionWorkSubscriber{dao: dao}
}

// Subscribe - will start the work subscriber listenning on the message buffer for deprovision messages.
func (d *DeprovisionWorkSubscriber) Subscribe(msgBuffer <-chan JobMsg) {
	d.msgBuffer = msgBuffer

	go func() {
		log.Info("Listening for deprovision messages")
		for {
			msg := <-msgBuffer
			metrics.DeprovisionJobFinished()
			log.Debug("Processed deprovision message from buffer")

			if msg.Error != "" {
				// Job failed, mark failure
				log.Errorf("Deprovision job reporting error: %s", msg.Error)
				setFailedDeprovisionJob(d.dao, msg)
				continue
			}

			instance, err := d.dao.GetServiceInstance(msg.InstanceUUID)
			if err != nil {
				log.Errorf(
					"Error occurred getting service instance [ %s ] after deprovision job:",
					msg.InstanceUUID,
				)
				setFailedDeprovisionJob(d.dao, msg)
				continue
			}

			// Job is not reporting error, cleanup after deprovision
			err = cleanupDeprovision(instance, d.dao)
			if err != nil {
				log.Error("Failed cleaning up deprovision after job, error: %s", err.Error())
				// Cleanup is reporting something has gone wrong. Deprovision overall
				// has not completed. Mark the job as failed.
				setFailedDeprovisionJob(d.dao, msg)
				continue
			}

			// No errors reported, deprovision action successfully performed and
			// broker has successfully cleaned up. Mark depro success
			if err := d.dao.SetState(msg.InstanceUUID, apb.JobState{
				Token:   msg.JobToken,
				State:   apb.StateSucceeded,
				Podname: msg.PodName,
				Method:  apb.JobMethodDeprovision,
			}); err != nil {
				log.Errorf("failed to set state after deprovision %#v", err)
			}
		}
	}()
}

func setFailedDeprovisionJob(dao *dao.Dao, dmsg JobMsg) {
	if err := dao.SetState(dmsg.InstanceUUID, apb.JobState{
		Token:   dmsg.JobToken,
		State:   apb.StateFailed,
		Podname: dmsg.PodName,
		Method:  apb.JobMethodDeprovision,
		Error:   dmsg.Error,
	}); err != nil {
		log.Errorf("failed to set state after deprovision %#v", err)
	}
}

func cleanupDeprovision(instance *apb.ServiceInstance, dao *dao.Dao) error {
	var err error
	id := instance.ID.String()

	if err = dao.DeleteExtractedCredentials(id); err != nil {
		log.Error("failed to delete extracted credentials - %#v", err)
		return err
	}

	if err = dao.DeleteServiceInstance(id); err != nil {
		log.Error("failed to delete service instance - %#v", err)
		return err
	}

	return nil
}
