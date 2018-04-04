//
// Copyright (c) 2018 Red Hat, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//    http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//

package clients

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"time"

	"github.com/automationbroker/config"
	"github.com/coreos/etcd/pkg/transport"
	"github.com/coreos/etcd/version"
	log "github.com/sirupsen/logrus"

	etcd "github.com/coreos/etcd/client"
)

// EtcdConfig - Etcd configuration
type EtcdConfig struct {
	EtcdHost       string `yaml:"etcd_host"`
	EtcdCaFile     string `yaml:"etcd_ca_file"`
	EtcdClientCert string `yaml:"etcd_client_cert"`
	EtcdClientKey  string `yaml:"etcd_client_key"`
	EtcdPort       int    `yaml:"etcd_port"`
}

var etcdConfig EtcdConfig

// InitEtcdConfig - Initialize the configuration for etcd.
func InitEtcdConfig(config *config.Config) {
	etcdConfig = EtcdConfig{
		EtcdHost:       config.GetString("dao.etcd_host"),
		EtcdPort:       config.GetInt("dao.etcd_port"),
		EtcdCaFile:     config.GetString("dao.etcd_ca_file"),
		EtcdClientKey:  config.GetString("dao.etcd_client_key"),
		EtcdClientCert: config.GetString("dao.etcd_client_cert"),
	}
}

// GetEtcdVersion - Connects to ETCD cluster and retrieves server/version info
func GetEtcdVersion(ec EtcdConfig) (string, string, error) {
	// The next etcd release (1.4) will have client.GetVersion()
	// We'll use this to test our etcd connection for now
	etcdURL := fmt.Sprintf("http://%s:%v/version", ec.EtcdHost, ec.EtcdPort)
	resp, err := http.Get(etcdURL)
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

// Etcd - Create a new etcd client if needed, returns reference
func Etcd() (etcd.Client, error) {
	errMsg := "Something went wrong intializing etcd client!"
	once.Etcd.Do(func() {
		client, err := newEtcd()
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

	if instances.Etcd == nil {
		return nil, errors.New("Etcd client instance is nil")
	}

	return instances.Etcd, nil
}

func newEtcd() (etcd.Client, error) {
	// TODO: Config validation
	endpoints := []string{etcdEndpoint()}
	transport, err := newTransport()
	log.Info("== ETCD CX ==")
	log.Infof("EtcdHost: %s", etcdConfig.EtcdHost)
	log.Infof("EtcdPort: %v", etcdConfig.EtcdPort)
	log.Infof("Endpoints: %v", endpoints)

	etcdClient, err := etcd.New(etcd.Config{
		Endpoints:               endpoints,
		Transport:               transport,
		HeaderTimeoutPerRequest: time.Second,
	})
	if err != nil {
		return nil, err
	}

	return etcdClient, err
}

func newTransport() (etcd.CancelableTransport, error) {
	if etcdConfig.EtcdClientCert == "" && etcdConfig.EtcdClientKey == "" && etcdConfig.EtcdCaFile == "" {
		return etcd.DefaultTransport, nil
	}
	info := transport.TLSInfo{}
	if etcdConfig.EtcdClientCert != "" && etcdConfig.EtcdClientKey != "" {
		info.CertFile = etcdConfig.EtcdClientCert
		info.KeyFile = etcdConfig.EtcdClientKey
	}

	if etcdConfig.EtcdCaFile != "" {
		info.CAFile = etcdConfig.EtcdCaFile
	}

	cfg, err := info.ClientConfig()
	if err != nil {
		return nil, err
	}
	// Copied from etcd.DefaultTransport declaration.
	// TODO: Determine if transport needs optimization
	tr := &http.Transport{
		Proxy: http.ProxyFromEnvironment,
		Dial: (&net.Dialer{
			Timeout:   30 * time.Second,
			KeepAlive: 30 * time.Second,
		}).Dial,
		TLSHandshakeTimeout: 10 * time.Second,
		MaxIdleConnsPerHost: 500,
		TLSClientConfig:     cfg,
	}
	return tr, nil
}

func etcdEndpoint() string {
	if etcdConfig.EtcdCaFile != "" {
		return fmt.Sprintf("https://%s:%v", etcdConfig.EtcdHost, etcdConfig.EtcdPort)
	}
	return fmt.Sprintf("http://%s:%v", etcdConfig.EtcdHost, etcdConfig.EtcdPort)
}
