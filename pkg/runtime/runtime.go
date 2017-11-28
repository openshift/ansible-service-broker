//
// Copyright (c) 2017 Red Hat, Inc.
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
	"fmt"

	"github.com/openshift/ansible-service-broker/pkg/clients"
	"github.com/openshift/ansible-service-broker/pkg/metrics"

	logging "github.com/op/go-logging"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	apicorev1 "k8s.io/kubernetes/pkg/api/v1"
	rbac "k8s.io/kubernetes/pkg/apis/rbac/v1beta1"
)

// Provider - Variable for accessing provider functions
var Provider *provider

// Runtime - Abstraction for broker actions
type Runtime interface {
	CreateSandbox(string, string, []string, string)
	DestroySandbox(string, string, []string, string, bool, bool)
}

// Variables for interacting with runtimes
type provider struct {
	log *logging.Logger
	coe
}

// Abstraction for actions that are different between runtimes
type coe interface{}

// Different runtimes
type openshift struct{}
type kubernetes struct{}

// NewRuntime - Initialize provider variable
func NewRuntime(log *logging.Logger) {
	Provider = &provider{log: log}
}

// CreateSandbox - Translate the broker CreateSandbox call into cluster resource calls
func (p provider) CreateSandbox(podName string, namespace string, targets []string, apbRole string) (string, error) {
	k8scli, err := clients.Kubernetes(p.log)
	if err != nil {
		return "", err
	}

	err = k8scli.CreateServiceAccount(podName, namespace)
	if err != nil {
		return "", err
	}

	p.log.Debug("Trying to create apb sandbox: [ %s ], with %s permissions in namespace %s", podName, apbRole, namespace)

	subjects := []rbac.Subject{
		rbac.Subject{
			Kind:      "ServiceAccount",
			Name:      podName,
			Namespace: namespace,
		},
	}

	roleRef := rbac.RoleRef{
		APIGroup: "rbac.authorization.k8s.io",
		Kind:     "ClusterRole",
		Name:     apbRole,
	}

	// targetNamespace and namespace are the same
	err = k8scli.CreateRoleBinding(podName, subjects, namespace, namespace, roleRef)
	if err != nil {
		return "", err
	}

	for _, target := range targets {
		err = k8scli.CreateRoleBinding(podName, subjects, namespace, target, roleRef)
		if err != nil {
			return "", err
		}
	}

	p.log.Info("Successfully created apb sandbox: [ %s ], with %s permissions in namespace %s", podName, apbRole, namespace)

	return podName, nil
}

// DestroySandbox - Translate the broker DestorySandbox call into cluster resource calls
func (p provider) DestroySandbox(podName string, namespace string, targets []string, configNamespace string, keepNamespace bool, keepNamespaceOnError bool) {
	p.log.Info("Destroying APB sandbox...")
	if podName == "" {
		p.log.Info("Requested destruction of APB sandbox with empty handle, skipping.")
		return
	}
	k8scli, err := clients.Kubernetes(p.log)
	if err != nil {
		p.log.Error("Something went wrong getting kubernetes client")
		p.log.Errorf("%s", err.Error())
		return
	}
	pod, err := k8scli.Client.CoreV1().Pods(namespace).Get(podName, metav1.GetOptions{})
	if err != nil {
		p.log.Errorf("Unable to retrieve pod - %v", err)
	}
	if shouldDeleteNamespace(keepNamespace, keepNamespaceOnError, pod, err) {
		if configNamespace != namespace {
			p.log.Debug("Deleting namespace %s", namespace)
			k8scli.Client.CoreV1().Namespaces().Delete(namespace, &metav1.DeleteOptions{})
			// This is keeping track of namespaces.
			metrics.SandboxDeleted()
		} else {
			// We should not be attempting to run pods in the ASB namespace, if we are, something is seriously wrong.
			panic(fmt.Errorf("Broker is attempting to delete its own namespace"))
		}

	} else {
		p.log.Debugf("Keeping namespace alive due to configuration")
	}
	p.log.Debugf("Deleting rolebinding %s, namespace %s", podName, namespace)

	err = k8scli.DeleteRoleBinding(podName, namespace)
	if err != nil {
		p.log.Error("Something went wrong trying to destroy the rolebinding! - %v", err)
		return
	}
	p.log.Notice("Successfully deleted rolebinding %s, namespace %s", podName, namespace)

	for _, target := range targets {
		p.log.Debugf("Deleting rolebinding %s, namespace %s", podName, target)
		err = k8scli.DeleteRoleBinding(podName, target)
		if err != nil {
			p.log.Error("Something went wrong trying to destroy the rolebinding!")
			return
		}
		p.log.Notice("Successfully deleted rolebinding %s, namespace %s", podName, target)
	}
	return
}

func shouldDeleteNamespace(keepNamespace bool, keepNamespaceOnError bool, pod *apicorev1.Pod, getPodErr error) bool {
	if keepNamespace {
		return false
	}

	if keepNamespaceOnError {
		if pod.Status.Phase == apicorev1.PodFailed || pod.Status.Phase == apicorev1.PodUnknown || getPodErr != nil {
			return false
		}
	}
	return true
}
