package clients

import (
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
func Etcd(config EtcdConfig, log *logging.Logger) *etcd.Client {
	errMsg := "Something went wrong intializing etcd client!"
	once.Etcd.Do(func() {
		client, err := newEtcd(config, log)
		if err != nil {
			log.Error(errMsg)
			// NOTE: Looking to leverage panic recovery to gracefully handle this
			// with things like retries or better intelligence, but the environment
			// is probably in a unrecoverable state as far as the broker is concerned,
			// and demands the attention of an operator.
			panic(err.Error())
		}
		instances.Etcd = client
	})

	return instances.Etcd
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
