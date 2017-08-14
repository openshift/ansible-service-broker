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

	"github.com/openshift/ansible-service-broker/pkg/runtime"

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
	namespace string,
	apbID string,
	apbRole string,
	roleScope string,
) (string, error) {
	svcAccountName := apbID
	roleBindingName := apbID
	roleKind := roleScope + "Binding"

	svcAcctM := map[string]interface{}{
		"apiVersion": "v1",
		"kind":       "ServiceAccount",
		"metadata": map[string]string{
			"name":      apbID,
			"namespace": namespace,
		},
	}

	roleBindingM := map[string]interface{}{
		"apiVersion": "v1",
		"kind":       roleKind,
		"metadata": map[string]string{
			"name":      roleBindingName,
			"namespace": namespace,
		},
		"subjects": []map[string]string{
			map[string]string{
				"kind":      "ServiceAccount",
				"name":      svcAccountName,
				"namespace": namespace,
			},
		},
		"roleRef": map[string]string{
			"name": strings.ToLower(apbRole),
		},
	}

	s.createResourceDir()
	rFilePath, err := s.writeResourceFile(apbID, &svcAcctM, &roleBindingM)
	if err != nil {
		return "", err
	}

	// Create resources in cluster
	s.createResources(rFilePath, namespace)

	s.log.Info("Successfully created apb sandbox: [ %s ]", apbID)

	return apbID, nil
}

func (s *ServiceAccountManager) createResources(rFilePath string, namespace string) error {
	s.log.Debug("Creating resources from file at path: %s", rFilePath)
	output, err := runtime.RunCommand("oc", "create", "-f", rFilePath, "--namespace="+namespace)
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
	svcAcctM *map[string]interface{}, roleBindingM *map[string]interface{},
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
func (s *ServiceAccountManager) DestroyApbSandbox(executionContext ApbExecutionContext) error {
	s.log.Info("Destroying APB sandbox...")
	if executionContext.PodName == "" {
		s.log.Info("Requested destruction of APB sandbox with empty handle, skipping.")
		return nil
	}

	s.log.Debug("Deleting serviceaccount %s, namespace %s", executionContext.PodName, executionContext.Namespace)
	output, err := runtime.RunCommand(
		"oc", "delete", "serviceaccount", executionContext.PodName, "--namespace="+executionContext.Namespace,
	)
	if err != nil {
		s.log.Error("Something went wrong trying to destroy the serviceaccount!")
		s.log.Error(err.Error())
		s.log.Error("oc delete output:")
		s.log.Error(string(output))
		return err
	}
	s.log.Debug("Successfully deleted serviceaccount %s, namespace %s", executionContext.PodName, executionContext.Namespace)
	s.log.Debug("oc delete output:")
	s.log.Debug(string(output))

	s.log.Debug("Deleting rolebinding %s, namespace %s", executionContext.PodName, executionContext.Namespace)
	output, err = runtime.RunCommand(
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

	// If file doesn't exist, ignore
	// "If there is an error, it will be of type *PathError"
	// We don't care, because it's gone
	os.Remove(filePathFromHandle(executionContext.PodName))

	return nil
}

func resourceDir() string {
	return filepath.FromSlash("/tmp/asb-resource-files")
}

func filePathFromHandle(handle string) string {
	return filepath.Join(resourceDir(), handle+".yaml")
}
