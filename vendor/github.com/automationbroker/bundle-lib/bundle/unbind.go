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

package bundle

import (
	"fmt"

	"github.com/automationbroker/bundle-lib/runtime"
	"github.com/pborman/uuid"
	log "github.com/sirupsen/logrus"
)

const (
	unbindAction = "unbind"
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
		// Create namespace name that will be used to generate a name.
		ns := fmt.Sprintf("%s-%.4s-", instance.Spec.FQName, unbindAction)
		// Create the podname
		pn := fmt.Sprintf("bundle-%s", uuid.New())
		targets := []string{instance.Context.Namespace}
		labels := map[string]string{
			"bundle-fqname":   instance.Spec.FQName,
			"bundle-action":   unbindAction,
			"bundle-pod-name": pn,
		}

		serviceAccount, namespace, err := runtime.Provider.CreateSandbox(pn, ns, targets, clusterConfig.SandboxRole, labels)
		if err != nil {
			log.Errorf("Problem executing bundle create sandbox [%s] unbind", pn)
			e.actionFinishedWithError(err)
			return
		}
		ec := runtime.ExecutionContext{
			BundleName: pn,
			Targets:    targets,
			Metadata:   labels,
			Action:     unbindAction,
			Image:      instance.Spec.Image,
			Account:    serviceAccount,
			Location:   namespace,
		}
		ec, err = e.executeApb(ec, instance, parameters)
		defer runtime.Provider.DestroySandbox(
			ec.BundleName,
			ec.Location,
			ec.Targets,
			clusterConfig.Namespace,
			clusterConfig.KeepNamespace,
			clusterConfig.KeepNamespaceOnError,
		)
		if err != nil {
			log.Errorf("Problem executing bundle [%s] unbind", ec.BundleName)
			e.actionFinishedWithError(err)
			return
		}

		err = runtime.Provider.WatchRunningBundle(ec.BundleName, ec.Location, e.updateDescription)
		if err != nil {
			log.Errorf("Unbind action failed - %v", err)
			e.actionFinishedWithError(err)
			return
		}
		// pod execution is complete so transfer state back
		err = e.stateManager.CopyState(
			ec.BundleName,
			e.stateManager.Name(instance.ID.String()),
			ec.Location,
			e.stateManager.MasterNamespace(),
		)
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
