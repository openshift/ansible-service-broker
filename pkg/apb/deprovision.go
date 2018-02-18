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
	"github.com/openshift/ansible-service-broker/pkg/clients"
	"github.com/openshift/ansible-service-broker/pkg/metrics"
	"github.com/openshift/ansible-service-broker/pkg/runtime"
	"github.com/pkg/errors"
)

// Deprovision - runs the abp with the deprovision action.
func (e *Executor) Deprovision(instance *ServiceInstance) <-chan StatusMessage {
	log.Notice("============================================================")
	log.Notice("                      DEPROVISIONING                        ")
	log.Notice("============================================================")
	log.Noticef("ServiceInstance.Id: %s", instance.Spec.ID)
	log.Noticef("ServiceInstance.Name: %v", instance.Spec.FQName)
	log.Noticef("ServiceInstance.Image: %s", instance.Spec.Image)
	log.Noticef("ServiceInstance.Description: %s", instance.Spec.Description)
	log.Notice("============================================================")

	go func() {
		e.start()
		if instance.Spec.Image == "" {
			log.Error("No image field found on the apb instance.Spec (apb.yaml)")
			log.Error("apb instance.Spec requires [name] and [image] fields to be separate")
			log.Error("Are you trying to run a legacy ansibleapp without an image field?")
			e.finishWithError(errors.New("No image field found on instance.Spec"))
			return
		}

		// Might need to change up this interface to feed in instance ids
		metrics.ActionStarted("deprovision")
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
			e.finishWithError(err)
			return
		}
		k8scli, err := clients.Kubernetes()
		if err != nil {
			log.Error("Something went wrong getting kubernetes client")
			e.finishWithError(err)
			return
		}
		err = watchPod(executionContext.PodName, executionContext.Namespace,
			k8scli.Client.CoreV1().Pods(executionContext.Namespace), e.updateDescription)
		if err != nil {
			log.Errorf("Deprovision action failed - %v", err)
			e.finishWithError(err)
			return
		}

		e.finishWithSuccess()
	}()

	return e.statusChan
}
