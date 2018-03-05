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
	"errors"
	"fmt"

	"github.com/automationbroker/bundle-lib/clients"
	"github.com/automationbroker/bundle-lib/runtime"
	log "github.com/sirupsen/logrus"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type executionMethod string

const (
	executionMethodProvision executionMethod = "provision"
	executionMethodUpdate    executionMethod = "update"
)

// returns PodName, ExtractedCredentials, error
func (e *executor) provisionOrUpdate(method executionMethod, instance *ServiceInstance) error {
	// Explicitly error out if image field is missing from instance.Spec
	// was introduced as a change to the apb instance.Spec to support integration
	// with the broker and still allow for providing an img path
	// Legacy ansibleapps will hit this.
	// TODO: Move this validation to a Spec creation function (yet to be created)
	if instance.Spec.Image == "" {
		log.Error("No image field found on the apb instance.Spec (apb.yaml)")
		log.Error("apb instance.Spec requires [name] and [image] fields to be separate")
		log.Error("Are you trying to run a legacy apb without an image field?")
		return errors.New("No image field found on instance.Spec")
	}

	k8scli, err := clients.Kubernetes()
	if err != nil {
		log.Error("Something went wrong getting kubernetes client")
		return err
	}

	ns := instance.Context.Namespace
	log.Info("Checking if namespace %s exists.", ns)
	_, err = k8scli.Client.CoreV1().Namespaces().Get(ns, metav1.GetOptions{})
	if err != nil {
		return fmt.Errorf("Project %s does not exist", ns)
	}

	executionContext, err := e.executeApb(string(method), instance.Spec,
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
		log.Errorf("Problem executing apb [%s] provision - err: %v ", executionContext.PodName, err)
		return err
	}

	if instance.Spec.Runtime >= 2 || !instance.Spec.Bindable {
		log.Debugf("watching pod for serviceinstance %#v", instance.Spec)
		err := runtime.WatchPod(executionContext.PodName, executionContext.Namespace,
			k8scli.Client.CoreV1().Pods(executionContext.Namespace), e.updateDescription)
		if err != nil {
			log.Errorf("Provision or Update action failed - %v", err)
			return err
		}
	}

	if !instance.Spec.Bindable {
		return nil
	}

	creds, err := ExtractCredentials(
		executionContext.PodName,
		executionContext.Namespace,
		instance.Spec.Runtime,
	)

	if err != nil {
		log.Errorf("apb::%v error occurred - %v", method, err)
		return err
	}

	e.extractedCredentials = creds
	return nil
}
