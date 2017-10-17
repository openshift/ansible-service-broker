//
// Copyright (c) 2017 Red Hat, Inc.
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
// Red Hat trademarks are not licensed under Apache License, Version 2.
// No permission is granted to use or replicate Red Hat trademarks that
// are incorporated in this software or its documentation.
//

package apb

import (
	"errors"
	"fmt"
	logging "github.com/op/go-logging"
	"github.com/openshift/ansible-service-broker/pkg/runtime"
)

type executionMethod string

const (
	executionMethodProvision executionMethod = "provision"
	executionMethodUpdate    executionMethod = "provision"
)

func provision_or_update(
	method executionMethod,
	instance *ServiceInstance,
	clusterConfig ClusterConfig,
	log *logging.Logger,
) (string /*podName*/, *ExtractedCredentials, error) {
	// Explicitly error out if image field is missing from instance.Spec
	// was introduced as a change to the apb instance.Spec to support integration
	// with the broker and still allow for providing an img path
	// Legacy ansibleapps will hit this.
	// TODO: Move this validation to a Spec creation function (yet to be created)
	if instance.Spec.Image == "" {
		log.Error("No image field found on the apb instance.Spec (apb.yaml)")
		log.Error("apb instance.Spec requires [name] and [image] fields to be separate")
		log.Error("Are you trying to run a legacy ansibleapp without an image field?")
		return "", nil, errors.New("No image field found on instance.Spec")
	}

	ns := instance.Context.Namespace
	log.Info("Checking if project %s exists...", ns)
	if !projectExists(ns) {
		log.Error("Project %s does NOT exist! Cannot provision requested %s", ns, instance.Spec.FQName)
		return "", nil, fmt.Errorf("Project %s does not exist", ns)
	}

	executionContext, err := ExecuteApb(
		string(method), clusterConfig, instance.Spec,
		instance.Context, instance.Parameters, log,
	)

	if err != nil {
		log.Errorf("Problem executing apb [%s]", executionContext.PodName)
		log.Error(err.Error())
		return executionContext.PodName, nil, err
	}

	creds, err := ExtractCredentials(executionContext.PodName, executionContext.Namespace, log)

	// We should not save credentials from an app that finds them and isn't bindable
	if creds != nil && !instance.Spec.Bindable {
		log.Warningf("APB %s is not bindable", instance.Spec.FQName)
		log.Warning("Ignoring Credentials")
		creds = nil
	}

	sm := NewServiceAccountManager(log)
	sm.DestroyApbSandbox(executionContext, clusterConfig)
	if err != nil {
		log.Errorf("apb::%s error occurred", string(method))
		log.Error("%s", err.Error())
		return executionContext.PodName, creds, err
	}

	return executionContext.PodName, creds, err
}

func projectExists(project string) bool {
	_, _, code := runtime.RunCommandWithExitCode("kubectl", "get", "project", project)
	return code == 0
}
