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

	logging "github.com/op/go-logging"
	"github.com/openshift/ansible-service-broker/pkg/runtime"
)

// Bind - Will run the APB with the bind action.
func Bind(
	instance *ServiceInstance,
	parameters *Parameters,
	log *logging.Logger,
) (string, *ExtractedCredentials, error) {
	log.Notice("============================================================")
	log.Notice("                       BINDING                              ")
	log.Notice("============================================================")
	log.Notice(fmt.Sprintf("ServiceInstance.ID: %s", instance.Spec.ID))
	log.Notice(fmt.Sprintf("ServiceInstance.Name: %v", instance.Spec.FQName))
	log.Notice(fmt.Sprintf("ServiceInstance.Image: %s", instance.Spec.Image))
	log.Notice(fmt.Sprintf("ServiceInstance.Description: %s", instance.Spec.Description))
	log.Notice("============================================================")

	executionContext, err := ExecuteApb("bind", instance.Spec, instance.Context, parameters, log)
	defer runtime.Provider.DestroySandbox(
		executionContext.PodName,
		executionContext.Namespace,
		executionContext.Targets,
		clusterConfig.Namespace,
		clusterConfig.KeepNamespace,
		clusterConfig.KeepNamespaceOnError,
	)
	if err != nil {
		log.Errorf("Problem executing apb [%s] bind", executionContext.PodName)
		return executionContext.PodName, nil, err
	}

	if instance.Spec.Runtime >= 2 {
		err := watchPod(executionContext.PodName, executionContext.Namespace, log)
		if err != nil {
			log.Errorf("APB Execution failed - %v", err)
			return executionContext.PodName, nil, err
		}
	}

	creds, err := ExtractCredentials(
		executionContext.PodName,
		executionContext.Namespace,
		instance.Spec.Runtime,
		log,
	)
	if err != nil {
		log.Errorf("apb::bind error occurred - %v", err)
		return executionContext.PodName, creds, err
	}
	return executionContext.PodName, creds, err
}
