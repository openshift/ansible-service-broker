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
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
)

// Deprovision - runs the abp with the deprovision action.
func (e *executor) Deprovision(instance *ServiceInstance) <-chan StatusMessage {
	log.Infof("============================================================")
	log.Infof("                      DEPROVISIONING                        ")
	log.Infof("============================================================")
	log.Infof("ServiceInstance.Id: %s", instance.Spec.ID)
	log.Infof("ServiceInstance.Name: %v", instance.Spec.FQName)
	log.Infof("ServiceInstance.Image: %s", instance.Spec.Image)
	log.Infof("ServiceInstance.Description: %s", instance.Spec.Description)
	log.Infof("============================================================")

	go func() {
		e.actionStarted()
		if instance.Spec.Image == "" {
			log.Error("No image field found on the apb instance.Spec (apb.yaml)")
			log.Error("apb instance.Spec requires [name] and [image] fields to be separate")
			log.Error("Are you trying to run a legacy ansibleapp without an image field?")
			e.actionFinishedWithError(errors.New("No image field found on instance.Spec"))
			return
		}

		// Might need to change up this interface to feed in instance ids
		executionContext, err := e.executeApb("deprovision", instance.Spec,
			instance.Context, instance.Parameters)
		defer runtime.Provider.DestroySandbox(
			executionContext.PodName,
			executionContext.Namespace,
			executionContext.Targets,
			clusterConfig.Namespace,
			clusterConfig.KeepNamespace,
			clusterConfig.KeepNamespaceOnError,
		)
		if err != nil {
			log.Errorf("Problem executing apb [%s] deprovision", executionContext.PodName)
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
			log.Errorf("Deprovision action failed - %v", err)
			e.actionFinishedWithError(err)
			return
		}
		err = runtime.Provider.DeleteExtractedCredential(instance.ID.String(), clusterConfig.Namespace)
		if err != nil {
			log.Errorf("unable to delete the extracted credentials - %v", err)
			e.actionFinishedWithError(err)
			return
		}

		e.actionFinishedWithSuccess()
	}()

	return e.statusChan
}
