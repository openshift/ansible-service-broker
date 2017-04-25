package broker

import (
	"encoding/json"

	"github.com/fusor/ansible-service-broker/pkg/apb"
	"github.com/fusor/ansible-service-broker/pkg/dao"
)

type ProvisionWorkSubscriber struct {
	dao       *dao.Dao
	msgBuffer <-chan WorkMsg
}

func NewProvisionWorkSubscriber(dao *dao.Dao) *ProvisionWorkSubscriber {
	return &ProvisionWorkSubscriber{dao: dao}
}

func (p *ProvisionWorkSubscriber) Subscribe(msgBuffer <-chan WorkMsg) {
	p.msgBuffer = msgBuffer

	var pmsg *ProvisionMsg
	var extCreds *apb.ExtractedCredentials
	go func() {
		for {
			msg := <-msgBuffer

			// HACK: this seems like a hack, there's probably a better way to
			// get the data sent through instead of a string
			json.Unmarshal([]byte(msg.Render()), &pmsg)

			if pmsg.Error != "" {
				p.dao.SetState(pmsg.InstanceUUID, apb.JobState{Token: pmsg.JobToken, State: apb.StateFailed})
			} else {
				json.Unmarshal([]byte(pmsg.Msg), &extCreds)
				p.dao.SetState(pmsg.InstanceUUID, apb.JobState{Token: pmsg.JobToken, State: apb.StateSucceeded})
				p.dao.SetExtractedCredentials(pmsg.InstanceUUID, extCreds)
			}
		}
	}()
}
