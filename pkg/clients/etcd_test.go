package clients

import (
	"fmt"
	"net/http"
	"os"
	"strings"
	"sync"
	"testing"

	logging "github.com/op/go-logging"
)

func TestEtcd(t *testing.T) {
	gopath := os.Getenv("GOPATH")
	if gopath == "" {
		panic("GOPATH not set!")
	}

	filePath := strings.Join([]string{gopath, "src", "github.com", "openshift", "ansible-service-broker", "pkg", "clients", "testdata"}, "/")

	testCases := []struct {
		Config               EtcdConfig
		Name                 string
		ResetRun             bool
		ShouldError          bool
		NilOutExistingClient bool
	}{
		{
			Name: "only CA Cert",
			Config: EtcdConfig{
				EtcdHost:   "testing-etcd.svc",
				EtcdPort:   "2379",
				EtcdCaFile: fmt.Sprintf("%s/%s", filePath, "ca.crt"),
			},
			ResetRun:             true,
			NilOutExistingClient: true,
		},
		{
			Name: "only CA Cert",
			Config: EtcdConfig{
				EtcdHost:   "testing-etcd.svc",
				EtcdPort:   "2379",
				EtcdCaFile: fmt.Sprintf("%s/%s", filePath, "ca.crt"),
			},
			ResetRun:             true,
			NilOutExistingClient: true,
		},
		{
			Name: "Invalid state",
			Config: EtcdConfig{
				EtcdHost: "testing-etcd.svc",
				EtcdPort: "2379",
			},
			NilOutExistingClient: true,
			ShouldError:          true,
		},
		{
			Name: "Invalid configuration",
			Config: EtcdConfig{
				EtcdHost: "aklsjdfalskdfj   alskdfjaslkdfj",
				EtcdPort: "2379",
			},
			NilOutExistingClient: true,
			ResetRun:             true,
			ShouldError:          true,
		},
	}

	for _, tc := range testCases {
		t.Run(fmt.Sprintf("%s", tc.Name), func(t *testing.T) {
			defer func() {
				if r := recover(); r != nil {
					if !tc.ShouldError {
						panic(fmt.Sprintf("failed in test run - %v", r))
					}
				}
			}()

			log := logging.MustGetLogger("test")
			if tc.ResetRun {
				once.Etcd = sync.Once{}
			}
			if tc.NilOutExistingClient {
				instances.Etcd = nil
			}
			cl, err := Etcd(tc.Config, log)
			if !tc.ShouldError && err != nil {
				t.Fatalf("failed to get etcd client - %v client - %v", err, cl)
			} else if tc.ShouldError && err != nil {
				t.Logf("Should error and retrieved error - %v", err)
				return
			} else if tc.ShouldError && err == nil {
				t.Fatalf("failed to get error - %v", err)
			}
			t.Logf("client - %v", cl.Endpoints())
			if tc.Config.EtcdCaFile != "" {
				for _, ep := range cl.Endpoints() {
					if !strings.Contains(ep, "https") {
						t.Fatalf("If EtcdCaFile is set, we expect to connect over SSL")
					}
				}
			} else {
				for _, ep := range cl.Endpoints() {
					if strings.Contains(ep, "https") {
						t.Fatalf("If EtcdCaFile is not set, we expect to connect over without SSL")
					}
				}
			}
		})
	}
}

func TestNewTransport(t *testing.T) {
	gopath := os.Getenv("GOPATH")
	if gopath == "" {
		panic("GOPATH not set!")
	}

	filePath := strings.Join([]string{gopath, "src", "github.com", "openshift", "ansible-service-broker", "pkg", "clients", "testdata"}, "/")

	testCases := []struct {
		Config EtcdConfig
		Name   string
	}{
		{
			Name: "only CA Cert",
			Config: EtcdConfig{
				EtcdCaFile: fmt.Sprintf("%s/%s", filePath, "ca.crt"),
			},
		},
		{
			Name: "only authentication",
			Config: EtcdConfig{
				EtcdClientCert: fmt.Sprintf("%s/%s", filePath, "client.crt"),
				EtcdClientKey:  fmt.Sprintf("%s/%s", filePath, "client.key"),
			},
		},
		{
			Name:   "nothing",
			Config: EtcdConfig{},
		},
		{
			Name: "All options",
			Config: EtcdConfig{
				EtcdCaFile:     fmt.Sprintf("%s/%s", filePath, "ca.crt"),
				EtcdClientCert: fmt.Sprintf("%s/%s", filePath, "client.crt"),
				EtcdClientKey:  fmt.Sprintf("%s/%s", filePath, "client.key"),
			},
		},
	}

	for _, tc := range testCases {
		t.Run(fmt.Sprintf("%s", tc.Name), func(t *testing.T) {
			defer func() {
				if r := recover(); r != nil {
					t.Fatalf("panic if test - %v", tc.Name)
				}
			}()
			tr, err := newTransport(tc.Config)
			if err != nil {
				t.Fatalf("Test failed to get a new transport - %v", err)
			}
			transport, ok := tr.(*http.Transport)
			if !ok || transport == nil {
				t.Fatalf("unable to get transport %#v", tr)
			}
			//Default transport does not have a TLSClientConfig
			if tc.Config.EtcdCaFile == "" && tc.Config.EtcdClientCert == "" && tc.Config.EtcdClientKey == "" && transport.TLSClientConfig == nil {
				return
			}
			if tc.Config.EtcdCaFile != "" {
				if transport.TLSClientConfig.ClientCAs != nil {
					t.Fatal("Should have set the client CA for the transport")
				}
			} else {
				if transport.TLSClientConfig.ClientCAs != nil {
					t.Fatal("Should not have set the client CA for the transport")
				}
			}
			if tc.Config.EtcdClientKey != "" && tc.Config.EtcdClientCert != "" {
				if len(transport.TLSClientConfig.Certificates) == 0 {
					t.Fatal("Should have created a certificate")
				}
			} else {
				if len(transport.TLSClientConfig.Certificates) != 0 {
					t.Fatal("Should not have created a certificate")
				}
			}
		})
	}
}
