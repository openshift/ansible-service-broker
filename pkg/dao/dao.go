package dao

import (
	"fmt"
	"time"

	"github.com/coreos/etcd/client"
	"github.com/fusor/ansible-service-broker/pkg/ansibleapp"
	"golang.org/x/net/context"
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

func (d *Dao) GetSpec(id string) (*ansibleapp.Spec, error) {
	raw, err := d.GetRaw(specKey(id))
	if err != nil {
		return nil, err
	}

	spec := &ansibleapp.Spec{}
	ansibleapp.LoadJSON(raw, spec)
	return spec, nil
}

func (d *Dao) SetSpec(id string, spec *ansibleapp.Spec) error {
	payload, err := ansibleapp.DumpJSON(spec)
	if err != nil {
		return err
	}

	return d.SetRaw(specKey(id), payload)
}

func (d *Dao) BatchSetSpecs(specs ansibleapp.SpecManifest) error {
	// TODO: Is there no batch insert in the etcd api?
	for id, spec := range specs {
		err := d.SetSpec(id, spec)
		if err != nil {
			return err
		}
	}

	return nil
}

func (d *Dao) BatchGetSpecs(dir string) ([]*ansibleapp.Spec, error) {
	var payloads *[]string
	var err error
	if payloads, err = d.BatchGetRaw(dir); err != nil {
		return []*ansibleapp.Spec{}, err
	}

	specs := make([]*ansibleapp.Spec, len(*payloads))
	for i, payload := range *payloads {
		spec := &ansibleapp.Spec{}
		ansibleapp.LoadJSON(payload, spec)
		specs[i] = spec
		d.log.Debug("Batch idx [ %d ] -> [ %s ]", i, spec.Id)
	}

	return specs, nil
}

func (d *Dao) GetServiceInstance(id string) (*ansibleapp.ServiceInstance, error) {
	var raw string
	var err error
	if raw, err = d.GetRaw(serviceInstanceKey(id)); err != nil {
		return nil, err
	}

	spec := &ansibleapp.ServiceInstance{}
	err = ansibleapp.LoadJSON(raw, spec)
	if err != nil {
		return nil, err
	}

	return spec, nil
}

func (d *Dao) SetServiceInstance(
	id string, serviceInstance *ansibleapp.ServiceInstance,
) error {
	payload, err := ansibleapp.DumpJSON(serviceInstance)
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

func (d *Dao) GetBindInstance(id string) (*ansibleapp.BindInstance, error) {
	var raw string
	var err error
	if raw, err = d.GetRaw(bindInstanceKey(id)); err != nil {
		return nil, err
	}

	spec := &ansibleapp.BindInstance{}
	err = ansibleapp.LoadJSON(raw, spec)
	if err != nil {
		return nil, err
	}

	return spec, nil
}

func (d *Dao) SetBindInstance(
	id string, bindInstance *ansibleapp.BindInstance,
) error {
	payload, err := ansibleapp.DumpJSON(bindInstance)
	if err != nil {
		return err
	}

	return d.SetRaw(bindInstanceKey(id), payload)
}

func (d *Dao) DeleteBindInstance(id string) error {
	d.log.Debug(fmt.Sprintf("Dao::DeleteBindInstance -> [ %s ]", id))
	_, err := d.kapi.Delete(context.Background(), bindInstanceKey(id), nil)
	return err
}

////////////////////////////////////////////////////////////
// Key generators
////////////////////////////////////////////////////////////
func specKey(id string) string {
	return fmt.Sprintf("/spec/%s", id)
}

func serviceInstanceKey(id string) string {
	return fmt.Sprintf("/service_instance/%s", id)
}

func bindInstanceKey(id string) string {
	return fmt.Sprintf("/bind_instance/%s", id)
}
