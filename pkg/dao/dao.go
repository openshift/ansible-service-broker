package dao

import (
	"fmt"
	"github.com/coreos/etcd/client"
	"github.com/fusor/ansible-service-broker/pkg/ansibleapp"
	"github.com/op/go-logging"
	"golang.org/x/net/context"
	"time"
)

type DaoConfig struct {
	EtcdHost string `yaml:"etcd_host"`
	EtcdPort string `yaml:"etcd_port"`
}

type Dao struct {
	config    DaoConfig
	log       *logging.Logger
	endpoints []string
	client    client.Client
	kapi      client.KeysAPI // Used to interact with kvp API over HTTP
}

func NewDao(config DaoConfig, log *logging.Logger) (*Dao, error) {
	var err error
	dao := Dao{
		config: config,
		log:    log,
	}

	// TODO: Config validation

	dao.endpoints = []string{etcdEndpoint(config.EtcdHost, config.EtcdPort)}

	log.Debug("Instantiating new Dao")
	log.Debug(fmt.Sprintf("EtcdHost: %s", config.EtcdHost))
	log.Debug(fmt.Sprintf("EtcdPort: %s", config.EtcdPort))
	log.Debug(fmt.Sprintf("Endpoint: %v", dao.endpoints))

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

func (d *Dao) BatchSetSpec(specs map[string]*ansibleapp.Spec) error {
	// TODO: Is there no batch insert in the etcd api?
	for id, spec := range specs {
		err := d.SetSpec(id, spec)
		if err != nil {
			return err
		}
	}

	return nil
}

////////////////////////////////////////////////////////////
// Key generators
////////////////////////////////////////////////////////////
func specKey(id string) string {
	return fmt.Sprintf("/spec/%s", id)
}
