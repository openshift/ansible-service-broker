package broker

import (
	"encoding/json"

	"github.com/fusor/ansible-service-broker/pkg/apb"
	logging "github.com/op/go-logging"
)

type ProvisionJob struct {
	spec          *apb.Spec
	parameters    *apb.Parameters
	clusterConfig apb.ClusterConfig
	log           *logging.Logger
}

type ProvisionMsg struct {
	JobToken string `json:"job_token"`
	SpecId   string `json:"spec_id"`
	Msg      string `json:"msg"`
	Error    string `json:"error"`
}

func (m ProvisionMsg) Render() string {
	render, _ := json.Marshal(m)
	return string(render)
}

func NewProvisionJob(
	spec *apb.Spec, parameters *apb.Parameters,
	clusterConfig apb.ClusterConfig, log *logging.Logger,
) *ProvisionJob {
	return &ProvisionJob{spec: spec, parameters: parameters,
		clusterConfig: clusterConfig, log: log}
}

func (p *ProvisionJob) Run(token string, msgBuffer chan<- WorkMsg) {
	/*
		// DEMO
		p.log.Notice("Sleep for a bit then return fake credentials")
		time.Sleep(20 * time.Second)
		creds := make(map[string]interface{})
		creds["username"] = "foobar"
		creds["password"] = "b@rb@z"
		//extCreds := apb.ExtractedCredentials{Credentials: creds}
		p.log.Notice(fmt.Sprintf("Dumping creds: %v", creds))
	*/

	extCreds, err := apb.Provision(p.spec, p.parameters, p.clusterConfig, p.log)
	if err != nil {
		p.log.Error("broker::Provision error occurred.")
		p.log.Error("%s", err.Error())
		// send error message
		// can't have an error type in a struct you want marshalled
		// https://github.com/golang/go/issues/5161
		msgBuffer <- ProvisionMsg{JobToken: token, SpecId: p.spec.Id, Msg: "", Error: err.Error()}
	}

	// send creds
	jsonmsg, _ := json.Marshal(extCreds)
	p.log.Notice("sending message to channel")
	msgBuffer <- ProvisionMsg{JobToken: token, SpecId: p.spec.Id, Msg: string(jsonmsg), Error: ""}
}
