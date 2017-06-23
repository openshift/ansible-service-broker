package clients

import (
	"fmt"
	"time"

	logging "github.com/op/go-logging"

	"github.com/coreos/etcd/client"
)

type EtcdConfig struct {
	EtcdHost string `yaml:"etcd_host"`
	EtcdPort string `yaml:"etcd_port"`
}

func NewEtcd(config EtcdConfig, log *logging.Logger) error {
	// TODO: Config validation
	endpoints := []string{etcdEndpoint(config.EtcdHost, config.EtcdPort)}

	log.Info("== ETCD CX ==")
	log.Infof("EtcdHost: %s", config.EtcdHost)
	log.Infof("EtcdPort: %s", config.EtcdPort)
	log.Infof("Endpoints: %v", endpoints)

	etcdClient, err := client.New(client.Config{
		Endpoints:               endpoints,
		Transport:               client.DefaultTransport,
		HeaderTimeoutPerRequest: time.Second,
	})
	if err != nil {
		return err
	}

	Clients.EtcdClient = etcdClient
	return nil
}

func etcdEndpoint(host string, port string) string {
	return fmt.Sprintf("http://%s:%s", host, port)
}
