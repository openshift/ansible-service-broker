package broker

import (
	"encoding/json"

	logging "github.com/op/go-logging"
	"github.com/openshift/ansible-service-broker/pkg/apb"
	"github.com/openshift/ansible-service-broker/pkg/dao"
)

// DeprovisionJob - Job to deprovision.
type DeprovisionJob struct {
	serviceInstance *apb.ServiceInstance
	clusterConfig   apb.ClusterConfig
	dao             *dao.Dao
	log             *logging.Logger
}

// DeprovisionMsg - Message returned for a deprovison job.
type DeprovisionMsg struct {
	InstanceUUID string `json:"instance_uuid"`
	PodName      string `json:"podname"`
	JobToken     string `json:"job_token"`
	SpecID       string `json:"spec_id"`
	Error        string `json:"error"`
}

// Render - render the message
func (m DeprovisionMsg) Render() string {
	render, _ := json.Marshal(m)
	return string(render)
}

// NewDeprovisionJob - Create a deprovision job.
func NewDeprovisionJob(serviceInstance *apb.ServiceInstance, clusterConfig apb.ClusterConfig,
	dao *dao.Dao, log *logging.Logger,
) *DeprovisionJob {
	return &DeprovisionJob{
		serviceInstance: serviceInstance,
		clusterConfig:   clusterConfig,
		dao:             dao,
		log:             log}
}

// Run - will run the deprovision job.
func (p *DeprovisionJob) Run(token string, msgBuffer chan<- WorkMsg) {
	podName, err := apb.Deprovision(p.serviceInstance, p.clusterConfig, p.log)
	if err != nil {
		p.log.Error("broker::Deprovision error occurred.")
		p.log.Errorf("%s", err.Error())
		msgBuffer <- DeprovisionMsg{InstanceUUID: p.serviceInstance.ID.String(), PodName: podName,
			JobToken: token, SpecID: p.serviceInstance.Spec.ID, Error: err.Error()}
		return
	}

	p.log.Debug("sending deprovision complete msg to channel")
	msgBuffer <- DeprovisionMsg{InstanceUUID: p.serviceInstance.ID.String(), PodName: podName,
		JobToken: token, SpecID: p.serviceInstance.Spec.ID, Error: ""}
}
