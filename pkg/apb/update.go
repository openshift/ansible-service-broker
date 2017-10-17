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
	logging "github.com/op/go-logging"
)

// Update - will run the abp with the provision action.
func Update(
	instance *ServiceInstance,
	clusterConfig ClusterConfig,
	log *logging.Logger,
) (string, *ExtractedCredentials, error) {
	log.Notice("============================================================")
	log.Notice("                       UPDATING                             ")
	log.Notice("============================================================")
	log.Noticef("Spec.ID: %s", instance.Spec.ID)
	log.Noticef("Spec.Name: %s", instance.Spec.FQName)
	log.Noticef("Spec.Image: %s", instance.Spec.Image)
	log.Noticef("Spec.Description: %s", instance.Spec.Description)
	log.Notice("============================================================")

	// Nearly all of the logic for provisioning or updating is shared between
	// provision and update, save for passing through the method type. Update
	// provides a nice public interface, but the bulk of the work is passed to
	// provision_or_update as an implementation detail.
	return provision_or_update(
		executionMethodUpdate, instance, clusterConfig, log,
	)
}
