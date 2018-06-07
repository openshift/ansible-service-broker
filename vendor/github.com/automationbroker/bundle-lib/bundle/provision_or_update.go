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
	"errors"
	"fmt"

	"github.com/automationbroker/bundle-lib/runtime"
	"github.com/pborman/uuid"
	log "github.com/sirupsen/logrus"
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

	// Create namespace name that will be used to generate a name.
	ns := fmt.Sprintf("%s-%.4s-", instance.Spec.FQName, method)

	// Determine if we should be using the context namespace from the executor config.
	if e.skipCreateNS {
		ns = instance.Context.Namespace
	}
	// Create the podname
	pn := fmt.Sprintf("bundle-%s", uuid.New())
	targets := []string{instance.Context.Namespace}
	labels := map[string]string{
		"bundle-fqname":   instance.Spec.FQName,
		"bundle-action":   string(method),
		"bundle-pod-name": pn,
	}
	serviceAccount, namespace, err := runtime.Provider.CreateSandbox(pn, ns, targets, clusterConfig.SandboxRole, labels)
	if err != nil {
		log.Errorf("Problem executing bundle create sandbox [%s] %v", pn, method)
		e.actionFinishedWithError(err)
		return err
	}
	ec := runtime.ExecutionContext{
		BundleName: pn,
		Targets:    targets,
		Metadata:   labels,
		Action:     string(method),
		Image:      instance.Spec.Image,
		Account:    serviceAccount,
		Location:   namespace,
	}
	ec, err = e.executeApb(ec, instance, instance.Parameters)
	defer runtime.Provider.DestroySandbox(
		ec.BundleName,
		ec.Location,
		ec.Targets,
		clusterConfig.Namespace,
		clusterConfig.KeepNamespace,
		clusterConfig.KeepNamespaceOnError,
	)
	if err != nil {
		log.Errorf("Problem executing bundle [%s] %v", ec.BundleName, method)
		e.actionFinishedWithError(err)
		return err
	}

	if instance.Spec.Runtime >= 2 || !instance.Spec.Bindable {
		log.Debugf("watching pod for serviceinstance %#v", instance.Spec)
		err := runtime.Provider.WatchRunningBundle(ec.BundleName, ec.Location, e.updateDescription)
		if err != nil {
			log.Errorf("Provision or Update action failed - %v", err)
			return err
		}
	}

	// pod execution is complete so transfer state back
	err = e.stateManager.CopyState(
		ec.BundleName,
		e.stateManager.MasterName(instance.ID.String()),
		ec.Location,
		e.stateManager.MasterNamespace(),
	)
	if err != nil {
		return err
	}

	if !instance.Spec.Bindable {
		return nil
	}

	credBytes, err := runtime.Provider.ExtractCredentials(
		ec.BundleName,
		ec.Location,
		instance.Spec.Runtime,
	)
	if err != nil {
		log.Errorf("bundle::%v error occurred - %v", method, err)
		return err
	}

	creds, err := buildExtractedCredentials(credBytes)
	if err != nil {
		log.Errorf("bundle::%v error occurred - %v", method, err)
		return err
	}

	e.extractedCredentials = creds
	return nil
}
