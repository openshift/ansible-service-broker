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

package dao

import (
	"github.com/openshift/ansible-service-broker/pkg/apb"
)

// MapJobStatesWithMethod - takes a slice of JobState stucts and returns a slice containing
// only JobStates that match the specified JobMethod.
func MapJobStatesWithMethod(jobs []apb.JobState, method apb.JobMethod) []apb.JobState {
	filteredJobStates := []apb.JobState{}
	for _, js := range jobs {
		if method == js.Method {
			filteredJobStates = append(filteredJobStates, js)
		}
	}
	return filteredJobStates
}
