package clients

import (
	"errors"
	"fmt"
	"time"

	logging "github.com/op/go-logging"

	etcd "github.com/coreos/etcd/client"
)

// EtcdConfig - Etcd configuration
type EtcdConfig struct {
	EtcdHost string `yaml:"etcd_host"`
	EtcdPort string `yaml:"etcd_port"`
}

// Etcd - Create a new etcd client if needed, returns reference
func Etcd(config EtcdConfig, log *logging.Logger) (*etcd.Client, error) {
	errMsg := "Something went wrong intializing etcd client!"
	once.Etcd.Do(func() {
		client, err := newEtcd(config, log)
		if err != nil {
			log.Error(errMsg)
			log.Error(err.Error())
			instances.Etcd = clientResult{nil, err}
		}
		instances.Etcd = clientResult{client, nil}
	})

	err := instances.Etcd.err
	if err != nil {
		log.Error(errMsg)
		log.Error(err.Error())
		return nil, err
	}

	if client, ok := instances.Etcd.client.(*etcd.Client); ok {
		return client, nil
	} else {
		return nil, errors.New(errMsg)
	}
}

func newEtcd(config EtcdConfig, log *logging.Logger) (*etcd.Client, error) {
	// TODO: Config validation
	endpoints := []string{etcdEndpoint(config.EtcdHost, config.EtcdPort)}

	log.Info("== ETCD CX ==")
	log.Infof("EtcdHost: %s", config.EtcdHost)
	log.Infof("EtcdPort: %s", config.EtcdPort)
	log.Infof("Endpoints: %v", endpoints)

	etcdClient, err := etcd.New(etcd.Config{
		Endpoints:               endpoints,
		Transport:               etcd.DefaultTransport,
		HeaderTimeoutPerRequest: time.Second,
	})
	if err != nil {
		return nil, err
	}

	return &etcdClient, err
}

func etcdEndpoint(host string, port string) string {
	return fmt.Sprintf("http://%s:%s", host, port)
}
