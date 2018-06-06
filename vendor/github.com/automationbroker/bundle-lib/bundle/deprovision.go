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
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
)

const (
	deprovisionAction = "deprovision"
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
		// Create namespace name that will be used to generate a name.
		ns := fmt.Sprintf("%s-%.4s-", instance.Spec.FQName, deprovisionAction)
		// Create the podname
		pn := fmt.Sprintf("bundle-%s", uuid.New())
		targets := []string{instance.Context.Namespace}
		labels := map[string]string{
			"bundle-fqname":   instance.Spec.FQName,
			"bundle-action":   deprovisionAction,
			"bundle-pod-name": pn,
		}
		serviceAccount, namespace, err := runtime.Provider.CreateSandbox(pn, ns, targets, clusterConfig.SandboxRole, labels)
		if err != nil {
			log.Errorf("Problem executing bundle create sandbox [%s] deprovision", pn)
			e.actionFinishedWithError(err)
			return
		}
		ec := runtime.ExecutionContext{
			BundleName: pn,
			Targets:    targets,
			Metadata:   labels,
			Action:     deprovisionAction,
			Image:      instance.Spec.Image,
			Account:    serviceAccount,
			Location:   namespace,
		}
		ec, err = e.executeApb(ec, instance, instance.Parameters)
		defer func() {
			if err := e.stateManager.DeleteState(e.stateManager.Name(instance.ID.String())); err != nil {
				log.Errorf("failed to delete state for instance %s : %v ", instance.ID.String(), err)
			}
			runtime.Provider.DestroySandbox(
				ec.BundleName,
				ec.Location,
				ec.Targets,
				clusterConfig.Namespace,
				clusterConfig.KeepNamespace,
				clusterConfig.KeepNamespaceOnError,
			)
		}()
		if err != nil {
			log.Errorf("Problem executing bundle [%s] deprovision", ec.BundleName)
			e.actionFinishedWithError(err)
			return
		}

		err = runtime.Provider.WatchRunningBundle(ec.BundleName, ec.Location, e.updateDescription)
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
