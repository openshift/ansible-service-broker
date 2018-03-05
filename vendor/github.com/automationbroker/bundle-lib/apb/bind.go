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

package apb

import (
	"github.com/automationbroker/bundle-lib/clients"
	"github.com/automationbroker/bundle-lib/runtime"

	log "github.com/sirupsen/logrus"
)

// Bind - Will run the APB with the bind action.
func (e *executor) Bind(
	instance *ServiceInstance, parameters *Parameters, bindingID string,
) <-chan StatusMessage {
	log.Info("============================================================")
	log.Info("                       BINDING                              ")
	log.Info("============================================================")
	log.Infof("ServiceInstance.ID: %s", instance.Spec.ID)
	log.Infof("ServiceInstance.Name: %v", instance.Spec.FQName)
	log.Infof("ServiceInstance.Image: %s", instance.Spec.Image)
	log.Infof("ServiceInstance.Description: %s", instance.Spec.Description)
	log.Infof("============================================================")

	go func() {
		e.actionStarted()
		executionContext, err := e.executeApb(
			"bind", instance.Spec, instance.Context, parameters)
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
			e.actionFinishedWithError(err)
			return
		}
		k8scli, err := clients.Kubernetes()
		if err != nil {
			log.Error("Something went wrong getting kubernetes client")
			e.actionFinishedWithError(err)
			return
		}

		if instance.Spec.Runtime >= 2 {
			err := runtime.WatchPod(executionContext.PodName, executionContext.Namespace,
				k8scli.Client.CoreV1().Pods(executionContext.Namespace), e.updateDescription)
			if err != nil {
				log.Errorf("Bind action failed - %v", err)
				e.actionFinishedWithError(err)
				return
			}
		}

		creds, err := ExtractCredentials(
			executionContext.PodName,
			executionContext.Namespace,
			instance.Spec.Runtime,
		)

		if err != nil {
			log.Errorf("apb::bind error occurred - %v", err)
			e.actionFinishedWithError(err)
			return
		}

		labels := map[string]string{"apbAction": "bind", "apbName": instance.Spec.FQName}
		err = runtime.Provider.CreateExtractedCredential(bindingID, clusterConfig.Namespace, creds.Credentials, labels)
		if err != nil {
			log.Errorf("apb::%v error occurred - %v", executionMethodProvision, err)
			e.actionFinishedWithError(err)
			return
		}
		e.extractedCredentials = creds
		e.actionFinishedWithSuccess()
	}()

	return e.statusChan
}
