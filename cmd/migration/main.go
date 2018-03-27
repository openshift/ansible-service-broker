package main

import (
	"flag"
	"fmt"
	"net/http"

	"github.com/coreos/etcd/pkg/transport"
)

var options struct {
	EtcdHost       string
	EtcdCAFile     string
	EtcdClientCert string
	EtcdClientKey  string
	EtcdPort       int
}

func init() {
	flag.IntVar(&options.EtcdPort, "port", 2379, "user '--port' option to specify the port to connect to etcd")
	flag.StringVar(&options.EtcdHost, "host", "", "host to connect to etcd server")
	flag.StringVar(&options.EtcdCAFile, "ca-file", "", "CA certificate file path to be used to connect over TLS with etcd server")
	flag.StringVar(&options.EtcdClientCert, "client-cert", "", "client cert file path to authenticate to etcd server. If used must also use client-key")
	flag.StringVar(&options.EtcdClientKey, "client-key", "", "client key file path to authenticate to etcd server. If used must also use client-cert")
	flag.Parse()
}

func main() {
	// First step is to connect to etcd, we will use etcd to retrieve all
	// of the bundles, job states, service instances and bindings
	if (options.EtcdClientCert != "" || options.EtcdClientKey != "") &&
		(options.EtcdClientCert == "" && options.EtcdClientKey == "") {
		fmt.Printf("To use Mutual TLS Authentication with etcd must have both client-cert and client-key")
		return
	}

	info := transport.TLSInfo{}
	if options.EtcdClientCert != "" && options.EtcdClientKey != "" {
		info.CertFile = options.EtcdClientCert
		info.KeyFile = options.EtcdClientKey
	}
	if options.EtcdCAFile != "" {
		info.CAFile = options.EtcdCAFile
	}

	cfg, err := info.ClientConfig()
	if err != nil {
		fmt.Printf("unable to create client configuration - %v", err)
		return
	}

	tr := &http.Transport{
                Proxy: http.ProxyFromEnvironment,
                Dial: (&net.Dialer{
                    Timeout: 30 * time.Second,
                    KeepAlive: 30 * time.Second,
                }).Dial,
                TLSHandshakeTimeout: 10 * time.Second,
                MaxIdleConnsPerHost: 500,
                TLSClientConfig: cfg,
        }
        endpoint := etcdEndPoint(options.EtcdCAFile, options.EtcdHost, options.EtcdPort)

        etcdClient, err := etcd.New(etcd.Config{
            Endpoints: endpoint,
            Transport: tr,
            HeaderTimeoutPerRequest: time.Second,
        })
        if err != nil {
            fmt.Printf("unable to create etcd client to connect to the server: %v", err)
        }

        specs, err := 


}

func etcdEndPoint(caFile, host string, port int) string {
        if caFile != "" {
            return fmt.Sprintf("https://%v:%v", host, port)
        }
            return fmt.Sprintf("http://%v:%v", host, port)
}
