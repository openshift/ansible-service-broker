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
	"testing"

	"k8s.io/apimachinery/pkg/api/errors"

	apiv1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	clientset "k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/fake"
)

func TestKubernetesCreateServiceAccount(t *testing.T) {
	k, err := Kubernetes()
	if err != nil {
		t.Fail()
	}

	testCases := []struct {
		name      string
		client    clientset.Interface
		podName   string
		namespace string
		isErr     bool
	}{
		{
			name: "failed",
			client: fake.NewSimpleClientset(&apiv1.ServiceAccount{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "new-pod-1",
					Namespace: "namespace",
				},
			}),
			podName:   "new-pod-1",
			namespace: "namespace",
			isErr:     true,
		},
		{
			name:      "ok",
			client:    fake.NewSimpleClientset(),
			podName:   "new-pod-1",
			namespace: "namespace",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			k.Client = tc.client
			err := k.CreateServiceAccount(tc.podName, tc.namespace)
			if err != nil {
				if tc.isErr && !errors.IsAlreadyExists(err) {
					t.Fatalf("error occurend but not already exists")
					return
				}
				return
			}
			if tc.isErr {
				t.Fatalf("Should have errored and did not")
				return
			}
			_, err = k.Client.CoreV1().ServiceAccounts(tc.namespace).Get(tc.podName, metav1.GetOptions{})
			if err != nil {
				t.Fatalf("Unable to get created svc account")
				return
			}
		})
	}
}
