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
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"strconv"
	"strings"
	"sync"
	"testing"

	"github.com/coreos/etcd/version"
)

func TestEtcd(t *testing.T) {
	gopath := os.Getenv("GOPATH")
	if gopath == "" {
		panic("GOPATH not set!")
	}

	filePath := strings.Join([]string{gopath, "src", "github.com", "automationbroker", "bundle-lib", "clients", "testdata"}, "/")

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
				EtcdPort:   2379,
				EtcdCaFile: fmt.Sprintf("%s/%s", filePath, "ca.crt"),
			},
			ResetRun:             true,
			NilOutExistingClient: true,
		},
		{
			Name: "only CA Cert",
			Config: EtcdConfig{
				EtcdHost:   "testing-etcd.svc",
				EtcdPort:   2379,
				EtcdCaFile: fmt.Sprintf("%s/%s", filePath, "ca.crt"),
			},
			ResetRun:             true,
			NilOutExistingClient: true,
		},
		{
			Name: "Invalid state",
			Config: EtcdConfig{
				EtcdHost: "testing-etcd.svc",
				EtcdPort: 2379,
			},
			NilOutExistingClient: true,
			ShouldError:          true,
		},
		{
			Name: "Invalid configuration",
			Config: EtcdConfig{
				EtcdHost: "aklsjdfalskdfj   alskdfjaslkdfj",
				EtcdPort: 2379,
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

			if tc.ResetRun {
				once.Etcd = sync.Once{}
			}
			if tc.NilOutExistingClient {
				instances.Etcd = nil
			}
			etcdConfig = tc.Config
			cl, err := Etcd()
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

	filePath := strings.Join([]string{gopath, "src", "github.com", "automationbroker", "bundle-lib", "clients", "testdata"}, "/")

	testCases := []struct {
		Config      EtcdConfig
		Name        string
		ShouldError bool
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
		{
			Name: "invalid options",
			Config: EtcdConfig{
				EtcdCaFile:     fmt.Sprintf("%s/%s", filePath, "ca.crt"),
				EtcdClientCert: fmt.Sprintf("%s/%s", filePath, "client-unknown.crt"),
				EtcdClientKey:  fmt.Sprintf("%s/%s", filePath, "client.key"),
			},
			ShouldError: true,
		},
	}

	for _, tc := range testCases {
		t.Run(fmt.Sprintf("%s", tc.Name), func(t *testing.T) {
			defer func() {
				if r := recover(); r != nil {
					t.Fatalf("panic if test - %v", tc.Name)
				}
			}()
			etcdConfig = tc.Config
			tr, err := newTransport()
			if err != nil && !tc.ShouldError {
				t.Fatalf("Test failed to get a new transport - %v", err)
			} else if err != nil && tc.ShouldError {
				return
			} else if err == nil && tc.ShouldError {
				t.Fatalf("Should have errored - %v", tc.Name)
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

func TestEtcdVersion(t *testing.T) {
	testCases := []struct {
		Name           string
		ServerVersion  string
		ClusterVersion string
		ShouldError    bool
		ServerFunc     func(http.ResponseWriter, *http.Request)
	}{
		{
			ServerVersion:  "3.2.1",
			ClusterVersion: "3.2.1",
			Name:           "Happy Path Status OK and body is valid",
			ServerFunc: func(w http.ResponseWriter, r *http.Request) {
				b, _ := json.Marshal(version.Versions{Server: "3.2.1", Cluster: "4.1.2"})
				w.WriteHeader(http.StatusOK)
				w.Write(b)
			},
		},
		{
			Name: "Error Path Status OK and body is not valid",
			ServerFunc: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
			},
			ShouldError: true,
		},
		{
			Name: "Error Path Status Not OK",
			ServerFunc: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusInternalServerError)
			},
			ShouldError: true,
		},
	}

	for _, tc := range testCases {
		t.Run(fmt.Sprintf("%s", tc.Name), func(t *testing.T) {
			ts := httptest.NewServer(http.HandlerFunc(tc.ServerFunc))
			defer ts.Close()
			u, _ := url.Parse(ts.URL)
			port, _ := strconv.Atoi(u.Port())
			con := EtcdConfig{
				EtcdHost: u.Hostname(),
				EtcdPort: port,
			}

			serverVer, clusterVersion, err := GetEtcdVersion(con)
			if err != nil && !tc.ShouldError {
				t.Fatalf("unable to get version - %v", err)
			} else if err != nil && tc.ShouldError {
				return
			} else if tc.ShouldError && err == nil {
				t.Fatalf("Getting etcd version should return error - %v", tc.Name)
			}
			if serverVer != tc.ServerVersion && clusterVersion != tc.ClusterVersion {
				t.Fatalf("server version expected: %v but got %v \n cluster version expected: %v but got %v",
					tc.ServerVersion,
					serverVer,
					tc.ClusterVersion,
					clusterVersion)
			}
		})
	}
}
