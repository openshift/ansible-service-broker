package broker

import (
	"encoding/json"

	logging "github.com/op/go-logging"
	"github.com/openshift/ansible-service-broker/pkg/apb"
	"github.com/openshift/ansible-service-broker/pkg/dao"
)

type ProvisionWorkSubscriber struct {
	dao       *dao.Dao
	log       *logging.Logger
	msgBuffer <-chan WorkMsg
}

func NewProvisionWorkSubscriber(dao *dao.Dao, log *logging.Logger) *ProvisionWorkSubscriber {
	return &ProvisionWorkSubscriber{dao: dao, log: log}
}

func (p *ProvisionWorkSubscriber) Subscribe(msgBuffer <-chan WorkMsg) {
	p.msgBuffer = msgBuffer

	var pmsg *ProvisionMsg
	var extCreds *apb.ExtractedCredentials
	go func() {
		p.log.Info("Listening for provision messages")
		for {
			msg := <-msgBuffer

			p.log.Debug("Processed message from buffer")
			// HACK: this seems like a hack, there's probably a better way to
			// get the data sent through instead of a string
			json.Unmarshal([]byte(msg.Render()), &pmsg)

			if pmsg.Error != "" {
				p.dao.SetState(pmsg.InstanceUUID, apb.JobState{Token: pmsg.JobToken, State: apb.StateFailed, Podname: pmsg.PodName})
			} else if pmsg.Msg == "" {
				// HACK: OMG this is horrible. We should probably pass in a
				// state. Since we'll also be using this to get more granular
				// updates one day.
				p.dao.SetState(pmsg.InstanceUUID, apb.JobState{Token: pmsg.JobToken, State: apb.StateInProgress, Podname: pmsg.PodName})
			} else {
				json.Unmarshal([]byte(pmsg.Msg), &extCreds)
				p.dao.SetState(pmsg.InstanceUUID, apb.JobState{Token: pmsg.JobToken, State: apb.StateSucceeded, Podname: pmsg.PodName})
				p.dao.SetExtractedCredentials(pmsg.InstanceUUID, extCreds)
			}
		}
	}()
}
