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
	"github.com/automationbroker/bundle-lib/runtime"
	log "github.com/sirupsen/logrus"
)

// Provision - will run the apb with the provision action.
func (e *executor) Provision(instance *ServiceInstance) <-chan StatusMessage {
	log.Infof("============================================================")
	log.Infof("                       PROVISIONING                         ")
	log.Infof("============================================================")
	log.Infof("Spec.ID: %s", instance.Spec.ID)
	log.Infof("Spec.Name: %s", instance.Spec.FQName)
	log.Infof("Spec.Image: %s", instance.Spec.Image)
	log.Infof("Spec.Description: %s", instance.Spec.Description)
	log.Infof("============================================================")

	go func() {
		e.actionStarted()
		err := e.provisionOrUpdate(executionMethodProvision, instance)
		if err != nil {
			log.Errorf("Provision APB error: %v", err)
			e.actionFinishedWithError(err)
			return
		}
		// Provision can not have extracted credentials.
		if e.extractedCredentials != nil {
			labels := map[string]string{"apbAction": string(executionMethodProvision), "apbName": instance.Spec.FQName}
			err := runtime.Provider.CreateExtractedCredential(instance.ID.String(), clusterConfig.Namespace, e.extractedCredentials.Credentials, labels)
			if err != nil {
				log.Errorf("apb::%v error occurred - %v", executionMethodProvision, err)
				e.actionFinishedWithError(err)
				return
			}
		}
		e.actionFinishedWithSuccess()
	}()
	return e.statusChan
}
