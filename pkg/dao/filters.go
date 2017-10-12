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

package dao

import (
	"github.com/openshift/ansible-service-broker/pkg/apb"
)

func GetJobStatesForAPBMethod(jobs []apb.JobState, method apb.JobStateAPBMethodType) []apb.JobState {
	filteredJobStates := []apb.JobState{}
	for _, js := range jobs {
		if method == js.APBMethodType {
			filteredJobStates = append(filteredJobStates, js)
		}
	}
	return filteredJobStates
}
