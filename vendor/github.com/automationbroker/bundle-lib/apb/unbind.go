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

// Unbind - runs the abp with the unbind action.
func (e *executor) Unbind(
	instance *ServiceInstance, parameters *Parameters, bindingID string,
) <-chan StatusMessage {
	log.Infof("============================================================")
	log.Infof("                       UNBINDING                            ")
	log.Infof("============================================================")
	log.Infof("ServiceInstance.ID: %s", instance.Spec.ID)
	log.Infof("ServiceInstance.Name: %v", instance.Spec.FQName)
	log.Infof("ServiceInstance.Image: %s", instance.Spec.Image)
	log.Infof("ServiceInstance.Description: %s", instance.Spec.Description)
	log.Infof("============================================================")

	go func() {
		e.actionStarted()
		executionContext, err := e.executeApb("unbind", instance.Spec,
			instance.Context, parameters)
		defer runtime.Provider.DestroySandbox(
			executionContext.PodName,
			executionContext.Namespace,
			executionContext.Targets,
			clusterConfig.Namespace,
			clusterConfig.KeepNamespace,
			clusterConfig.KeepNamespaceOnError,
		)
		if err != nil {
			log.Errorf("Problem executing apb [%s] unbind", executionContext.PodName)
			e.actionFinishedWithError(err)
			return
		}

		k8scli, err := clients.Kubernetes()
		if err != nil {
			log.Error("Something went wrong getting kubernetes client")
			e.actionFinishedWithError(err)
			return
		}

		err = runtime.WatchPod(executionContext.PodName, executionContext.Namespace,
			k8scli.Client.CoreV1().Pods(executionContext.Namespace), e.updateDescription)
		if err != nil {
			log.Errorf("Unbind action failed - %v", err)
			e.actionFinishedWithError(err)
			return
		}
		// Delete the binding extracted credential here.
		err = runtime.Provider.DeleteExtractedCredential(bindingID, clusterConfig.Namespace)
		if err != nil {
			log.Infof("Unbind failed to delete extracted credential m- %v", err)
		}

		e.actionFinishedWithSuccess()
	}()

	return e.statusChan
}
