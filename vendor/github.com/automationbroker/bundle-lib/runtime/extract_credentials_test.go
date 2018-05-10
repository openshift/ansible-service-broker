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

package runtime

import (
	"encoding/json"
	"fmt"
	"reflect"
	"testing"

	"github.com/automationbroker/bundle-lib/clients"
	ft "github.com/stretchr/testify/assert"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/fake"
)

func TestExitGracefully(t *testing.T) {
	output := []byte("eyJkYiI6ICJmdXNvcl9ndWVzdGJvb2tfZGIiLCAidXNlciI6ICJkdWRlcl90d28iLCAicGFzcyI6ICJkb2c4dHdvIn0=")

	_, err := decodeOutput(output)
	ft.Equal(t, err, nil)
}

func TestInt(t *testing.T) {
	output := []byte("eyJEQl9OQU1FIjogImZvb2JhciIsICJEQl9QQVNTV09SRCI6ICJzdXBlcnNlY3JldCIsICJEQl9UWVBFIjogIm15c3FsIiwgIkRCX1BPUlQiOiAzMzA2LCAiREJfVVNFUiI6ICJkdWRlciIsICJEQl9IT1NUIjogIm15aW5zdGFuY2UuMTIzNDU2Nzg5MDEyLnVzLWVhc3QtMS5yZHMuYW1hem9uYXdzLmNvbSJ9")

	decoded, err := decodeOutput(output)
	if err != nil {
		t.Log(err.Error())
	}

	do := make(map[string]interface{})
	json.Unmarshal(decoded, &do)
	ft.Equal(t, do["DB_NAME"], "foobar", "name does not match")
	ft.Equal(t, do["DB_PASSWORD"], "supersecret", "password does not match")
	ft.Equal(t, do["DB_TYPE"], "mysql", "type does not match")
	ft.Equal(t, do["DB_PORT"], float64(3306), "port does not match")
	ft.Equal(t, do["DB_USER"], "duder", "user does not match")
	ft.Equal(t, do["DB_HOST"], "myinstance.123456789012.us-east-1.rds.amazonaws.com", "invalid hostname")
}

func TestExtractCredentials(t *testing.T) {
	k, err := clients.Kubernetes()
	if err != nil {
		t.Fail()
	}
	testCases := []struct {
		name      string
		expected  []byte
		shouldErr bool
		runtime   int
		client    *fake.Clientset
		podname   string
		namespace string
	}{
		/*
			                This test is very hard to write as there is no fake for remotecommand.
			                As this is being deprecated the time to get the is test working seems like too
			                much work.
			                {
						name:                "runtime version 1",
						expected:            extractCredentialsAsFile,
						shouldErr:           false,
						runtime:             1,
						setFakeClientConfig: true,
						pod: &api.Pod{
							ObjectMeta: metav1.ObjectMeta{Name: "foo", Namespace: "test", ResourceVersion: "10"},
							Spec: api.PodSpec{
								RestartPolicy: api.RestartPolicyAlways,
								DNSPolicy:     api.DNSClusterFirst,
								Containers: []api.Container{
									{
										Name: "bar",
									},
								},
								InitContainers: []api.Container{
									{
										Name: "initfoo",
									},
								},
							},
							Status: api.PodStatus{
								Phase: api.PodRunning,
							},
						},
					},
		*/
		{
			name:      "runtime greater than equal 2 no secret found",
			expected:  []byte(`{"db": "name"}`),
			shouldErr: true,
			runtime:   2,
			client:    fake.NewSimpleClientset(),
			podname:   "foo",
			namespace: "bar",
		},
		{
			name:     "runtime greater than equal 2",
			expected: []byte(`{"db": "name"}`),
			runtime:  2,
			client: fake.NewSimpleClientset(&v1.Secret{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "foo",
					Namespace: "bar",
				},
				Data: map[string][]byte{"fields": []byte(`{"db": "name"}`)},
			}),
			podname:   "foo",
			namespace: "bar",
		},
		{
			name:      "invalid runtime",
			expected:  []byte{},
			shouldErr: true,
			runtime:   0,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			k.Client = tc.client
			p := provider{}
			ec, err := p.ExtractCredentials(tc.podname, tc.namespace, tc.runtime)
			fmt.Printf("%v", tc.name)
			if err != nil && tc.shouldErr {
				return
			} else if err != nil {
				t.Fatalf("unexpected err: %v", err)
			} else if err == nil && tc.shouldErr {
				t.Fatalf("should error")
			}
			fmt.Printf("%v", ec)
			if !reflect.DeepEqual(ec, tc.expected) {
				t.Fatalf("extracted credentials do not match expected: %q got: %q", tc.expected, ec)
			}

		})
	}
}
