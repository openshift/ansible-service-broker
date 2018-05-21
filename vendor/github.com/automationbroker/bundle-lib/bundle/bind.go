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
	bindAction = "bind"
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
		// Create namespace name that will be used to generate a name.
		ns := fmt.Sprintf("%s-%.4s-", instance.Spec.FQName, bindAction)
		// Create the podname
		pn := fmt.Sprintf("bundle-%s", uuid.New())
		targets := []string{instance.Context.Namespace}
		labels := map[string]string{
			"bundle-fqname":   instance.Spec.FQName,
			"bundle-action":   bindAction,
			"bundle-pod-name": pn,
		}

		serviceAccount, namespace, err := runtime.Provider.CreateSandbox(pn, ns, targets, clusterConfig.SandboxRole, labels)
		ec := runtime.ExecutionContext{
			BundleName: pn,
			Targets:    targets,
			Metadata:   labels,
			Action:     bindAction,
			Image:      instance.Spec.Image,
			Account:    serviceAccount,
			Location:   namespace,
		}
		if err != nil {
			log.Errorf("Problem executing bundle create sandbox [%s] bind", ec.BundleName)
			e.actionFinishedWithError(err)
			return
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
			log.Errorf("Problem executing bundle [%s] bind", ec.BundleName)
			e.actionFinishedWithError(err)
			return
		}

		if instance.Spec.Runtime >= 2 {
			err := runtime.Provider.WatchRunningBundle(ec.BundleName, ec.Location, e.updateDescription)
			if err != nil {
				log.Errorf("Bind action failed - %v", err)
				e.actionFinishedWithError(err)
				return
			}
		}

		// pod execution is complete so transfer state back
		err = e.stateManager.CopyState(
			ec.BundleName,
			e.stateManager.Name(instance.ID.String()),
			ec.Location, e.stateManager.MasterNamespace())
		if err != nil {
			e.actionFinishedWithError(err)
			return
		}

		credBytes, err := runtime.Provider.ExtractCredentials(
			ec.BundleName,
			ec.Location,
			instance.Spec.Runtime,
		)
		if err != nil {
			log.Errorf("apb::bind error occurred - %v", err)
			e.actionFinishedWithError(err)
			return
		}

		creds, err := buildExtractedCredentials(credBytes)
		if err != nil {
			log.Errorf("apb::bind error occurred - %v", err)
			e.actionFinishedWithError(err)
			return
		}

		labels = map[string]string{"bundleAction": "bind", "bundleName": instance.Spec.FQName}
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
