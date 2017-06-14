package broker

import (
	"encoding/json"

	logging "github.com/op/go-logging"
	"github.com/openshift/ansible-service-broker/pkg/apb"
	"github.com/pborman/uuid"
)

type ProvisionJob struct {
	instanceuuid  uuid.UUID
	spec          *apb.Spec
	context       *apb.Context
	parameters    *apb.Parameters
	clusterConfig apb.ClusterConfig
	log           *logging.Logger
}

type ProvisionMsg struct {
	InstanceUUID string `json:"instance_uuid"`
	JobToken     string `json:"job_token"`
	SpecId       string `json:"spec_id"`
	Msg          string `json:"msg"`
	Error        string `json:"error"`
}

func (m ProvisionMsg) Render() string {
	render, _ := json.Marshal(m)
	return string(render)
}

func NewProvisionJob(
	instanceuuid uuid.UUID, spec *apb.Spec, context *apb.Context,
	parameters *apb.Parameters, clusterConfig apb.ClusterConfig,
	log *logging.Logger,
) *ProvisionJob {
	return &ProvisionJob{
		instanceuuid:  instanceuuid,
		spec:          spec,
		context:       context,
		parameters:    parameters,
		clusterConfig: clusterConfig,
		log:           log}
}

func (p *ProvisionJob) Run(token string, msgBuffer chan<- WorkMsg) {
	podName, extCreds, err := apb.Provision(p.spec, p.context, p.parameters, p.clusterConfig, p.log)
	sm := apb.NewServiceAccountManager(p.log)
	if err != nil {
		p.log.Error("broker::Provision error occurred.")
		p.log.Error("%s", err.Error())

		p.log.Error("Attempting to destroy APB sandbox if it has been created")
		sm.DestroyApbSandbox(podName, p.context.Namespace)
		// send error message
		// can't have an error type in a struct you want marshalled
		// https://github.com/golang/go/issues/5161
		msgBuffer <- ProvisionMsg{InstanceUUID: p.instanceuuid.String(),
			JobToken: token, SpecId: p.spec.Id, Msg: "", Error: err.Error()}
		return
	}

	p.log.Info("Destroying APB sandbox...")
	sm.DestroyApbSandbox(podName, p.context.Namespace)

	// send creds
	jsonmsg, _ := json.Marshal(extCreds)
	p.log.Debug("sending message to channel")
	msgBuffer <- ProvisionMsg{InstanceUUID: p.instanceuuid.String(),
		JobToken: token, SpecId: p.spec.Id, Msg: string(jsonmsg), Error: ""}
}
