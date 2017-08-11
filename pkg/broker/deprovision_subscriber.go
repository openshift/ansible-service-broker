package broker

import (
	"encoding/json"

	logging "github.com/op/go-logging"
	"github.com/openshift/ansible-service-broker/pkg/apb"
	"github.com/openshift/ansible-service-broker/pkg/dao"
)

// DeprovisionWorkSubscriber - Lissten for provision messages
type DeprovisionWorkSubscriber struct {
	dao       *dao.Dao
	log       *logging.Logger
	msgBuffer <-chan WorkMsg
}

// NewDeprovisionWorkSubscriber - Create a new work subscriber.
func NewDeprovisionWorkSubscriber(dao *dao.Dao, log *logging.Logger) *DeprovisionWorkSubscriber {
	return &DeprovisionWorkSubscriber{dao: dao, log: log}
}

// Subscribe - will start the work subscriber listenning on the message buffer for deprovision messages.
func (d *DeprovisionWorkSubscriber) Subscribe(msgBuffer <-chan WorkMsg) {
	d.msgBuffer = msgBuffer
	var dmsg *DeprovisionMsg

	go func() {
		d.log.Info("Listening for deprovision messages")
		for {
			msg := <-msgBuffer

			d.log.Debug("Processed deprovision message from buffer")
			json.Unmarshal([]byte(msg.Render()), &dmsg)

			if dmsg.Error != "" {
				// Job failed, mark failure
				setFailedDeprovisionJob(d.dao, dmsg)
				return
			}

			instance, err := d.dao.GetServiceInstance(dmsg.InstanceUUID)
			if err != nil {
				d.log.Errorf(
					"Error occurred getting service instance [ %s ] after deprovision job:",
					dmsg.InstanceUUID,
				)
				d.log.Errorf("%s", err.Error())
				setFailedDeprovisionJob(d.dao, dmsg)
				return
			}

			// Job is not reporting error, cleanup after deprovision
			err = cleanupDeprovision(dmsg.PodName, instance, d.dao, d.log)
			if err != nil {
				d.log.Error("Failed cleaning up deprovision after job, error: %s", err.Error())
				// Cleanup is reporting something has gone wrong. Deprovision overall
				// has not completed. Mark the job as failed.
				setFailedDeprovisionJob(d.dao, dmsg)
				return
			}

			// No errors reported, deprovision action successfully performed and
			// broker has successfully cleaned up. Mark depro success
			d.dao.SetState(dmsg.InstanceUUID, apb.JobState{Token: dmsg.JobToken,
				State: apb.StateSucceeded, Podname: dmsg.PodName})
		}
	}()
}

func setFailedDeprovisionJob(dao *dao.Dao, dmsg *DeprovisionMsg) {
	dao.SetState(dmsg.InstanceUUID, apb.JobState{
		Token:   dmsg.JobToken,
		State:   apb.StateFailed,
		Podname: dmsg.PodName,
	})
}

func cleanupDeprovision(
	podName string, instance *apb.ServiceInstance, dao *dao.Dao, log *logging.Logger,
) error {
	var err error
	id := instance.ID.String()
	sm := apb.NewServiceAccountManager(log)
	log.Info("Destroying APB sandbox...")
	sm.DestroyApbSandbox(podName, instance.Context.Namespace)

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
