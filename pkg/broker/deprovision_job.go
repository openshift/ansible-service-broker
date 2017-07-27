package broker

import (
	"encoding/json"

	logging "github.com/op/go-logging"
	"github.com/openshift/ansible-service-broker/pkg/apb"
	"github.com/openshift/ansible-service-broker/pkg/dao"
)

type DeprovisionJob struct {
	serviceInstance *apb.ServiceInstance
	clusterConfig   apb.ClusterConfig
	dao             *dao.Dao
	log             *logging.Logger
}

type DeprovisionMsg struct {
	InstanceUUID string `json:"instance_uuid"`
	JobToken     string `json:"job_token"`
	SpecId       string `json:"spec_id"`
	Error        string `json:"error"`
	Msg          string `json:"msg"`
}

func (m DeprovisionMsg) Render() string {
	render, _ := json.Marshal(m)
	return string(render)
}

func NewDeprovisionJob(
	serviceInstance *apb.ServiceInstance,
	clusterConfig apb.ClusterConfig,
	dao *dao.Dao,
	log *logging.Logger,
) *DeprovisionJob {
	return &DeprovisionJob{
		serviceInstance: serviceInstance,
		clusterConfig:   clusterConfig,
		dao:             dao,
		log:             log}
}

func (p *DeprovisionJob) Run(token string, msgBuffer chan<- WorkMsg) {
	podName, err := apb.Deprovision(p.serviceInstance, p.clusterConfig, p.log)
	err = cleanupDeprovision(err, podName, p.serviceInstance, p.dao, p.log)
	if err != nil {
		p.log.Error("broker::Deprovision error occurred.")
		p.log.Error("%s", err.Error())
		// send error message
		// can't have an error type in a struct you want marshalled
		// https://github.com/golang/go/issues/5161
		msgBuffer <- DeprovisionMsg{InstanceUUID: p.serviceInstance.Id.String(),
			JobToken: token, SpecId: p.serviceInstance.Spec.Id, Error: err.Error()}
		return
	}

	// send creds
	p.log.Debug("sending message to channel")
	msgBuffer <- DeprovisionMsg{InstanceUUID: p.serviceInstance.Id.String(),
		JobToken: token, SpecId: p.serviceInstance.Spec.Id, Error: "", Msg: "_deprovision"}
}
