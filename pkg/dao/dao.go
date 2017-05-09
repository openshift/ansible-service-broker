package dao

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/coreos/etcd/client"
	"github.com/coreos/etcd/version"
	"github.com/fusor/ansible-service-broker/pkg/apb"
	logging "github.com/op/go-logging"
)

type Config struct {
	EtcdHost string `yaml:"etcd_host"`
	EtcdPort string `yaml:"etcd_port"`
}

type Dao struct {
	config    Config
	log       *logging.Logger
	endpoints []string
	client    client.Client
	kapi      client.KeysAPI // Used to interact with kvp API over HTTP
}

func NewDao(config Config, log *logging.Logger) (*Dao, error) {
	var err error
	dao := Dao{
		config: config,
		log:    log,
	}

	// TODO: Config validation

	dao.endpoints = []string{etcdEndpoint(config.EtcdHost, config.EtcdPort)}

	log.Info("== ETCD CX ==")
	log.Info(fmt.Sprintf("EtcdHost: %s", config.EtcdHost))
	log.Info(fmt.Sprintf("EtcdPort: %s", config.EtcdPort))
	log.Info(fmt.Sprintf("Endpoints: %v", dao.endpoints))

	dao.client, err = client.New(client.Config{
		Endpoints:               dao.endpoints,
		Transport:               client.DefaultTransport,
		HeaderTimeoutPerRequest: time.Second,
	})
	if err != nil {
		return nil, err
	}

	dao.kapi = client.NewKeysAPI(dao.client)

	return &dao, nil
}

func etcdEndpoint(host string, port string) string {
	return fmt.Sprintf("http://%s:%s", host, port)
}

func (d *Dao) SetRaw(key string, val string) error {
	d.log.Debug(fmt.Sprintf("Dao::SetRaw [ %s ] -> [ %s ]", key, val))
	_, err := d.kapi.Set(context.Background(), key, val /*opts*/, nil)
	return err
}

func (d *Dao) GetEtcdVersion(config Config) (string, string, error) {

	// The next etcd release (1.4) will have client.GetVersion()
	// We'll use this to test our etcd connection for now
	resp, err := http.Get("http://" + config.EtcdHost + ":" + config.EtcdPort + "/version")
	if err != nil {
		return "", "", err
	}

	defer resp.Body.Close()
	body, _ := ioutil.ReadAll(resp.Body)

	switch resp.StatusCode {
	case http.StatusOK:
		var vresp version.Versions
		if err := json.Unmarshal(body, &vresp); err != nil {
			return "", "", err
		}
		return vresp.Server, vresp.Cluster, nil
	default:
		var connectErr error
		if err := json.Unmarshal(body, &connectErr); err != nil {
			return "", "", err
		}
		return "", "", connectErr
	}
}

func (d *Dao) GetRaw(key string) (string, error) {
	res, err := d.kapi.Get(context.Background(), key /*opts*/, nil)
	if err != nil {
		return "", err
	}

	val := res.Node.Value
	d.log.Debug(fmt.Sprintf("Dao::GetRaw [ %s ] -> [ %s ]", key, val))
	return val, nil
}

// TODO: Streaming interface? Going to need to optimize all this for
// a full-load catalog response of 10k
// This is more likely to be paged given current proposal
// In which case, we need paged Batch gets
// 2 steps?
// GET /spec/manifest [/*ordered ids*/]
// BatchGet(offset, count)?
func (d *Dao) BatchGetRaw(dir string) (*[]string, error) {
	d.log.Debug("Dao::BatchGetRaw")

	var res *client.Response
	var err error

	opts := &client.GetOptions{Recursive: true}
	if res, err = d.kapi.Get(context.Background(), dir, opts); err != nil {
		return nil, err
	}

	specNodes := res.Node.Nodes
	specCount := len(specNodes)

	d.log.Debug("Successfully loaded [ %d ] objects from etcd dir [ %s ]", specCount, dir)

	payloads := make([]string, specCount)
	for i, node := range specNodes {
		payloads[i] = node.Value
	}

	return &payloads, nil
}

func (d *Dao) GetSpec(id string) (*apb.Spec, error) {
	raw, err := d.GetRaw(specKey(id))
	if err != nil {
		return nil, err
	}

	spec := &apb.Spec{}
	apb.LoadJSON(raw, spec)
	return spec, nil
}

func (d *Dao) SetSpec(id string, spec *apb.Spec) error {
	payload, err := apb.DumpJSON(spec)
	if err != nil {
		return err
	}

	return d.SetRaw(specKey(id), payload)
}

func (d *Dao) BatchSetSpecs(specs apb.SpecManifest) error {
	// TODO: Is there no batch insert in the etcd api?
	for id, spec := range specs {
		err := d.SetSpec(id, spec)
		if err != nil {
			return err
		}
	}

	return nil
}

func (d *Dao) BatchGetSpecs(dir string) ([]*apb.Spec, error) {
	var payloads *[]string
	var err error
	if payloads, err = d.BatchGetRaw(dir); err != nil {
		return []*apb.Spec{}, err
	}

	specs := make([]*apb.Spec, len(*payloads))
	for i, payload := range *payloads {
		spec := &apb.Spec{}
		apb.LoadJSON(payload, spec)
		specs[i] = spec
		d.log.Debug("Batch idx [ %d ] -> [ %s ]", i, spec.Id)
	}

	return specs, nil
}

func (d *Dao) GetServiceInstance(id string) (*apb.ServiceInstance, error) {
	var raw string
	var err error
	if raw, err = d.GetRaw(serviceInstanceKey(id)); err != nil {
		return nil, err
	}

	spec := &apb.ServiceInstance{}
	err = apb.LoadJSON(raw, spec)
	if err != nil {
		return nil, err
	}

	return spec, nil
}

func (d *Dao) SetServiceInstance(
	id string, serviceInstance *apb.ServiceInstance,
) error {
	payload, err := apb.DumpJSON(serviceInstance)
	if err != nil {
		return err
	}

	return d.SetRaw(serviceInstanceKey(id), payload)
}

func (d *Dao) DeleteServiceInstance(id string) error {
	d.log.Debug(fmt.Sprintf("Dao::DeleteServiceInstance -> [ %s ]", id))
	_, err := d.kapi.Delete(context.Background(), serviceInstanceKey(id), nil)
	return err
}

func (d *Dao) GetBindInstance(id string) (*apb.BindInstance, error) {
	var raw string
	var err error
	if raw, err = d.GetRaw(bindInstanceKey(id)); err != nil {
		return nil, err
	}

	spec := &apb.BindInstance{}
	err = apb.LoadJSON(raw, spec)
	if err != nil {
		return nil, err
	}

	return spec, nil
}

func (d *Dao) SetBindInstance(
	id string, bindInstance *apb.BindInstance,
) error {
	payload, err := apb.DumpJSON(bindInstance)
	if err != nil {
		return err
	}

	return d.SetRaw(bindInstanceKey(id), payload)
}

func (d *Dao) GetExtractedCredentials(id string) (*apb.ExtractedCredentials, error) {
	raw, err := d.GetRaw(extractedCredentialsKey(id))
	if err != nil {
		return nil, err
	}

	extractedCredentials := &apb.ExtractedCredentials{}
	apb.LoadJSON(raw, extractedCredentials)
	return extractedCredentials, nil
}

func (d *Dao) SetExtractedCredentials(
	id string, extractedCredentials *apb.ExtractedCredentials,
) error {
	payload, err := apb.DumpJSON(extractedCredentials)
	if err != nil {
		return err
	}

	return d.SetRaw(extractedCredentialsKey(id), payload)
}

func (d *Dao) SetState(id string, state apb.JobState) error {
	payload, err := apb.DumpJSON(state)
	if err != nil {
		return err
	}

	return d.SetRaw(stateKey(id, state.Token), payload)
}

func (d *Dao) GetState(id string, token string) (apb.JobState, error) {
	raw, err := d.GetRaw(stateKey(id, token))
	if err != nil {
		return apb.JobState{State: apb.StateFailed}, err
	}

	state := apb.JobState{}
	apb.LoadJSON(raw, &state)
	return state, nil
}

func (d *Dao) DeleteBindInstance(id string) error {
	d.log.Debug(fmt.Sprintf("Dao::DeleteBindInstance -> [ %s ]", id))
	_, err := d.kapi.Delete(context.Background(), bindInstanceKey(id), nil)
	return err
}

////////////////////////////////////////////////////////////
// Key generators
////////////////////////////////////////////////////////////

func stateKey(id string, jobid string) string {
	//func stateKey(id string) string {
	return fmt.Sprintf("/state/%s/job/%s", id, jobid)
	//return fmt.Sprintf("/state/%s", id)
}

func extractedCredentialsKey(id string) string {
	return fmt.Sprintf("/extracted_credentials/%s", id)
}

func specKey(id string) string {
	return fmt.Sprintf("/spec/%s", id)
}

func serviceInstanceKey(id string) string {
	return fmt.Sprintf("/service_instance/%s", id)
}

func bindInstanceKey(id string) string {
	return fmt.Sprintf("/bind_instance/%s", id)
}
