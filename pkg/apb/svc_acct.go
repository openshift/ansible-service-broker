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
// Red Hat trademarks are not licensed under Apache License, Version 2.
// No permission is granted to use or replicate Red Hat trademarks that
// are incorporated in this software or its documentation.
//

package apb

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/openshift/ansible-service-broker/pkg/clients"
	"github.com/openshift/ansible-service-broker/pkg/runtime"
	apicorev1 "k8s.io/kubernetes/pkg/api/v1"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	logging "github.com/op/go-logging"
	yaml "gopkg.in/yaml.v2"
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

// CreateApbSandbox - Sets up ServiceAccount based apb sandbox
// Returns service account name to be used as a handle for destroying
// the sandbox at the conclusion of running the apb
func (s *ServiceAccountManager) CreateApbSandbox(
	executionContext ExecutionContext,
	apbRole string,
) (string, error) {
	apbID := executionContext.PodName
	svcAccountName := executionContext.PodName
	roleBindingName := executionContext.PodName

	svcAcctM := map[string]interface{}{
		"apiVersion": "v1",
		"kind":       "ServiceAccount",
		"metadata": map[string]string{
			"name":      svcAccountName,
			"namespace": executionContext.Namespace,
		},
	}

	roleBindingM := map[string]interface{}{
		"apiVersion": "v1",
		"kind":       "RoleBinding",
		"metadata": map[string]string{
			"name":      roleBindingName,
			"namespace": executionContext.Namespace,
		},
		"subjects": []map[string]string{
			map[string]string{
				"kind":      "ServiceAccount",
				"name":      svcAccountName,
				"namespace": executionContext.Namespace,
			},
		},
		"roleRef": map[string]string{
			"name": strings.ToLower(apbRole),
		},
	}
	targetRoleBindingsM := []map[string]interface{}{}
	for _, target := range executionContext.Targets {
		targetRoleBindingsM = append(targetRoleBindingsM,
			map[string]interface{}{
				"apiVersion": "v1",
				"kind":       "RoleBinding",
				"metadata": map[string]string{
					"name":      roleBindingName,
					"namespace": target,
				},
				"subjects": []map[string]string{
					map[string]string{
						"kind":      "ServiceAccount",
						"name":      svcAccountName,
						"namespace": executionContext.Namespace,
					},
				},
				"roleRef": map[string]string{
					"name": apbRole,
				},
			},
		)
	}

	s.createResourceDir()
	rFilePath, err := s.writeResourceFile(apbID, &svcAcctM, &roleBindingM, &targetRoleBindingsM)
	if err != nil {
		return "", err
	}

	s.log.Debug("Trying to create apb sandbox: [ %s ], with  %s permissions in namespace %s", apbID, apbRole, executionContext.Namespace)
	// Create resources in cluster
	s.createResources(rFilePath, executionContext.Namespace)

	s.log.Info("Successfully created apb sandbox: [ %s ], with %s permissions in namespace %s", apbID, apbRole, executionContext.Namespace)

	return apbID, nil
}

func (s *ServiceAccountManager) createResources(rFilePath string, namespace string) error {
	s.log.Debug("Creating resources from file at path: %s", rFilePath)
	output, err := runtime.RunCommand("oc", "create", "-f", rFilePath)
	// TODO: Parse output somehow to validate things got created?
	if err != nil {
		s.log.Error("Something went wrong trying to create resources in cluster")
		s.log.Error("Returned error:")
		s.log.Error(err.Error())
		s.log.Error("oc create -f output:")
		s.log.Error(string(output))
		return err
	}
	s.log.Debug("Successfully created resources, oc create -f output:")
	s.log.Debug("\n" + string(output))
	return nil
}

func (s *ServiceAccountManager) writeResourceFile(handle string,
	svcAcctM *map[string]interface{}, roleBindingM *map[string]interface{}, targetRoleBindingsM *[]map[string]interface{},
) (string, error) {
	// Create file if doesn't already exist
	filePath, err := s.createFile(handle)
	if err != nil {
		return "", err // Bubble, error logged in createFile
	}

	file, err := os.OpenFile(filePath, os.O_RDWR, 0644)
	defer file.Close()

	if err != nil {
		s.log.Error("Something went wrong writing resources to file!")
		s.log.Error(err.Error())
		return "", err
	}

	file.WriteString("---\n")
	svcAcctY, err := yaml.Marshal(svcAcctM)
	if err != nil {
		s.log.Error("Something went wrong marshalling svc acct to yaml")
		s.log.Error(err.Error())
		return "", err
	}
	file.WriteString(string(svcAcctY))

	file.WriteString("---\n")
	roleBindingY, err := yaml.Marshal(roleBindingM)
	if err != nil {
		s.log.Error("Something went wrong marshalling role binding to yaml")
		s.log.Error(err.Error())
		return "", err
	}
	file.WriteString(string(roleBindingY))

	for _, bindingM := range *targetRoleBindingsM {
		targetRoleBindingY, err := yaml.Marshal(bindingM)
		if err != nil {
			s.log.Error("Something went wrong marshalling role binding to yaml")
			s.log.Error(err.Error())
			return "", err
		}
		file.WriteString("---\n")
		file.WriteString(string(targetRoleBindingY))
	}

	s.log.Info("Successfully wrote resources to %s", filePath)
	return filePath, nil
}

func (s *ServiceAccountManager) createResourceDir() {
	dir := resourceDir()
	s.log.Debug("Creating resource file dir: %s", dir)
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		os.Mkdir(dir, os.ModePerm)
	}
}

func (s *ServiceAccountManager) createFile(handle string) (string, error) {
	rFilePath := filePathFromHandle(handle)
	s.log.Debug("Creating resource file %s", rFilePath)

	if _, err := os.Stat(rFilePath); os.IsNotExist(err) {
		// Valid behavior if the file does not exist, create
		file, err := os.Create(rFilePath)
		// Handle file creation error
		if err != nil {
			s.log.Error("Something went wrong touching new resource file!")
			s.log.Error(err.Error())
			return "", err
		}
		defer file.Close()
	} else if err != nil {
		// Bubble any non-expected errors
		return "", err
	}

	return rFilePath, nil
}

// DestroyApbSandbox - Destroys the apb sandbox
func (s *ServiceAccountManager) DestroyApbSandbox(executionContext ExecutionContext, clusterConfig ClusterConfig) error {
	s.log.Info("Destroying APB sandbox...")
	if executionContext.PodName == "" {
		s.log.Info("Requested destruction of APB sandbox with empty handle, skipping.")
		return nil
	}
	k8scli, err := clients.Kubernetes(s.log)
	if err != nil {
		return err
	}
	pod, err := k8scli.CoreV1().Pods(executionContext.Namespace).Get(executionContext.PodName, metav1.GetOptions{})
	if err != nil {
		s.log.Errorf("Unable to retrieve pod - %v", err)
	}
	if shouldDeleteNamspace(clusterConfig, pod, err) {
		if clusterConfig.Namespace != executionContext.Namespace {
			s.log.Debug("Deleting namespace %s", executionContext.Namespace)
			k8scli.CoreV1().Namespaces().Delete(executionContext.Namespace, &metav1.DeleteOptions{})
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
		s.log.Error(err.Error())
		s.log.Error("oc delete output:")
		s.log.Error(string(output))
		return err
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
			s.log.Error(err.Error())
			s.log.Error("oc delete output:")
			s.log.Error(string(output))
			return err
		}
		s.log.Debug("Successfully deleted rolebinding %s, namespace %s", executionContext.PodName, target)
		s.log.Debug("oc delete output:")
		s.log.Debug(string(output))

	}

	// If file doesn't exist, ignore
	// "If there is an error, it will be of type *PathError"
	// We don't care, because it's gone
	os.Remove(filePathFromHandle(executionContext.PodName))

	return nil
}

func shouldDeleteNamspace(clusterConfig ClusterConfig, pod *apicorev1.Pod, getPodErr error) bool {
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
