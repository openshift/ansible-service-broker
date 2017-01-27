package dao

import (
	"fmt"
	"github.com/coreos/etcd/client"
	"github.com/fusor/ansible-service-broker/pkg/ansibleapp"
	"github.com/op/go-logging"
	"golang.org/x/net/context"
	"time"
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
	return spec.LoadJSON(raw), nil
}

func (d *Dao) SetSpec(id string, spec *ansibleapp.Spec) error {
	payload := spec.DumpJSON()
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
		spec.LoadJSON(payload)
		specs[i] = spec
		d.log.Debug("Batch idx [ %d ] -> [ %s ]", i, spec.Id)
	}

	return specs, nil
}

////////////////////////////////////////////////////////////
// Key generators
////////////////////////////////////////////////////////////
func specKey(id string) string {
	return fmt.Sprintf("/spec/%s", id)
}
