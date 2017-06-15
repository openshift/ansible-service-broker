package broker

import (
	"encoding/json"

	logging "github.com/op/go-logging"
	"github.com/openshift/ansible-service-broker/pkg/apb"
)

type ProvisionJob struct {
	serviceInstance *apb.ServiceInstance
	clusterConfig   apb.ClusterConfig
	log             *logging.Logger
}

type ProvisionMsg struct {
	InstanceUUID string `json:"instance_uuid"`
	JobToken     string `json:"job_token"`
	SpecId       string `json:"spec_id"`
	PodName      string `json:"podname"`
	Msg          string `json:"msg"`
	Error        string `json:"error"`
}

func (m ProvisionMsg) Render() string {
	render, _ := json.Marshal(m)
	return string(render)
}

func NewProvisionJob(
	serviceInstance *apb.ServiceInstance,
	clusterConfig apb.ClusterConfig,
	log *logging.Logger,
) *ProvisionJob {
	return &ProvisionJob{
		serviceInstance: serviceInstance,
		clusterConfig:   clusterConfig,
		log:             log}
}

func (p *ProvisionJob) Run(token string, msgBuffer chan<- WorkMsg) {
	podName, extCreds, err := apb.Provision(p.serviceInstance, p.clusterConfig, p.log)
	sm := apb.NewServiceAccountManager(p.log)

	if err != nil {
		p.log.Error("broker::Provision error occurred.")
		p.log.Error("%s", err.Error())

		p.log.Error("Attempting to destroy APB sandbox if it has been created")
		sm.DestroyApbSandbox(podName, p.serviceInstance.Context.Namespace)
		// send error message
		// can't have an error type in a struct you want marshalled
		// https://github.com/golang/go/issues/5161
		msgBuffer <- ProvisionMsg{InstanceUUID: p.instanceuuid.String(),
			JobToken: token, SpecId: p.spec.Id, PodName: "", Msg: "", Error: err.Error()}
		return
	}

	msgBuffer <- ProvisionMsg{InstanceUUID: p.instanceuuid.String(),
		JobToken: token, SpecId: p.spec.Id, PodName: podName, Msg: "", Error: ""}

	// need to get the pod name for the job state
	extCreds, extErr := apb.ExtractCredentials(podName, p.context.Namespace, p.log)
	if extErr != nil {
		p.log.Error("broker::Provision extError occurred.")
		p.log.Error("%s", extErr.Error())
		// send extError message
		// can't have an extError type in a struct you want marshalled
		// https://github.com/golang/go/issues/5161
		msgBuffer <- ProvisionMsg{InstanceUUID: p.instanceuuid.String(),
			JobToken: token, SpecId: p.spec.Id, PodName: podName, Msg: "", Error: extErr.Error()}
		return
	}

	p.log.Info("Destroying APB sandbox...")
	sm.DestroyApbSandbox(podName, p.serviceInstance.Context.Namespace)

	// send creds
	jsonmsg, _ := json.Marshal(extCreds)
	p.log.Debug("sending message to channel")
	msgBuffer <- ProvisionMsg{InstanceUUID: p.instanceuuid.String(),
		JobToken: token, SpecId: p.spec.Id, PodName: podName, Msg: string(jsonmsg), Error: ""}
}
