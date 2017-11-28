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

package apb

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/openshift/ansible-service-broker/pkg/clients"
	"github.com/openshift/ansible-service-broker/pkg/metrics"
	"github.com/openshift/ansible-service-broker/pkg/runtime"
	apicorev1 "k8s.io/kubernetes/pkg/api/v1"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	logging "github.com/op/go-logging"
)

// ServiceAccountManager - managers the service account methods
type ServiceAccountManager struct {
	log *logging.Logger
}

// NewServiceAccountManager - Creates a new service account manager
func NewServiceAccountManager(log *logging.Logger) ServiceAccountManager {
	return ServiceAccountManager{
		log: log,
	}
}

// DestroyApbSandbox - Destroys the apb sandbox
func (s *ServiceAccountManager) DestroyApbSandbox(executionContext ExecutionContext, clusterConfig ClusterConfig) {
	s.log.Info("Destroying APB sandbox...")
	if executionContext.PodName == "" {
		s.log.Info("Requested destruction of APB sandbox with empty handle, skipping.")
		return
	}
	k8scli, err := clients.Kubernetes(s.log)
	if err != nil {
		s.log.Error("Something went wrong getting kubernetes client")
		s.log.Errorf("%s", err.Error())
		return
	}
	pod, err := k8scli.Client.CoreV1().Pods(executionContext.Namespace).Get(executionContext.PodName, metav1.GetOptions{})
	if err != nil {
		s.log.Errorf("Unable to retrieve pod - %v", err)
	}
	if shouldDeleteNamespace(clusterConfig, pod, err) {
		if clusterConfig.Namespace != executionContext.Namespace {
			s.log.Debug("Deleting namespace %s", executionContext.Namespace)
			k8scli.Client.CoreV1().Namespaces().Delete(executionContext.Namespace, &metav1.DeleteOptions{})
			// This is keeping track of namespaces.
			metrics.SandboxDeleted()
		} else {
			// We should not be attempting to run pods in the ASB namespace, if we are, something is seriously wrong.
			panic(fmt.Errorf("Broker is attempting to delete its own namespace"))
		}

	} else {
		s.log.Debugf("Keeping namespace alive due to configuration")
	}
	s.log.Debugf("Deleting rolebinding %s, namespace %s", executionContext.PodName, executionContext.Namespace)
	output, err := runtime.RunCommand(
		"oc", "delete", "rolebinding", executionContext.PodName, "--namespace="+executionContext.Namespace,
	)
	if err != nil {
		s.log.Error("Something went wrong trying to destroy the rolebinding!")
		s.log.Errorf("%s", err.Error())
		s.log.Error("oc delete output:")
		s.log.Errorf("%s", string(output))
		return
	}
	s.log.Debug("Successfully deleted rolebinding %s, namespace %s", executionContext.PodName, executionContext.Namespace)
	s.log.Debug("oc delete output:")
	s.log.Debug(string(output))

	for _, target := range executionContext.Targets {
		s.log.Debugf("Deleting rolebinding %s, namespace %s", executionContext.PodName, target)
		output, err = runtime.RunCommand(
			"oc", "delete", "rolebinding", executionContext.PodName, "--namespace="+target,
		)
		if err != nil {
			s.log.Error("Something went wrong trying to destroy the rolebinding!")
			s.log.Errorf("%s", err.Error())
			s.log.Error("oc delete output:")
			s.log.Error(string(output))
			return
		}
		s.log.Debug("Successfully deleted rolebinding %s, namespace %s", executionContext.PodName, target)
		s.log.Debug("oc delete output:")
		s.log.Debugf("%s", string(output))
	}

	// If file doesn't exist, ignore
	// "If there is an error, it will be of type *PathError"
	// We don't care, because it's gone
	os.Remove(filePathFromHandle(executionContext.PodName))

	return
}

func shouldDeleteNamespace(clusterConfig ClusterConfig, pod *apicorev1.Pod, getPodErr error) bool {
	if clusterConfig.KeepNamespace {
		return false
	}

	if clusterConfig.KeepNamespaceOnError {
		if pod.Status.Phase == apicorev1.PodFailed || pod.Status.Phase == apicorev1.PodUnknown || getPodErr != nil {
			return false
		}
	}
	return true
}

func resourceDir() string {
	return filepath.FromSlash("/tmp/asb-resource-files")
}

func filePathFromHandle(handle string) string {
	return filepath.Join(resourceDir(), handle+".yaml")
}
