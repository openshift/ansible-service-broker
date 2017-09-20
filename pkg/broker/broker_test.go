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

package broker

import (
	"os"
	"testing"

	logging "github.com/op/go-logging"
	"github.com/openshift/ansible-service-broker/pkg/apb"
	ft "github.com/openshift/ansible-service-broker/pkg/fusortest"
)

var log = logging.MustGetLogger("handler")

func init() {
	colorFormatter := logging.MustStringFormatter(
		"%{color}[%{time}] [%{level}] %{message}%{color:reset}",
	)
	backend := logging.NewLogBackend(os.Stdout, "", 1)
	backendFormatter := logging.NewBackendFormatter(backend, colorFormatter)
	logging.SetBackend(backend, backendFormatter)
}

func TestAddNameAndIDForSpecStripsTailingDash(t *testing.T) {
	spec1 := apb.Spec{FQName: "1234567890123456789012345678901234567890-"}
	spec2 := apb.Spec{FQName: "org/hello-world-apb"}
	spcs := []*apb.Spec{&spec1, &spec2}
	addNameAndIDForSpec(spcs, "h")
	ft.AssertEqual(t, "h-1234567890123456789012345678901234567890", spcs[0].FQName)
	ft.AssertEqual(t, "h-org-hello-world-apb", spcs[1].FQName)
}

func TestAddIdForPlan(t *testing.T) {
	plan1 := apb.Plan{Name: "default"}
	plans := []apb.Plan{plan1}
	addIDForPlan(plans, "dh-sns-apb")
	ft.AssertNotEqual(t, plans[0].ID, "", "plan id not updated")
}
