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

package app

import (
	"testing"

	ft "github.com/openshift/ansible-service-broker/pkg/fusortest"
)

func TestCreateConfig(t *testing.T) {
	config, err := CreateConfig("testdata/test-config.yaml", "testdata/test-registry-auth.yaml")
	if err != nil {
		t.Fatal(err.Error())
	}

	ft.AssertEqual(t, config.Registry[0].Type, "dockerhub", "mismatch registry type")
	ft.AssertEqual(t, config.Registry[0].Name, "docker", "mismatch registry name")
	ft.AssertEqual(t, config.Registry[0].URL, "https://registry.hub.docker.com",
		"mismatch registry url")
	ft.AssertEqual(t, config.Registry[0].User, "DOCKERHUB_USER", "mismatch registry user")
	ft.AssertEqual(t, config.Registry[0].Pass, "DOCKERHUB_PASS", "mismatch registry pass")
	ft.AssertEqual(t, config.Registry[0].Org, "DOCKERHUB_ORG", "mismatch registry org")
	ft.AssertFalse(t, config.Registry[0].Fail, "mismatch registry fail")
	ft.AssertEqual(t, config.Registry[1].WhiteList[0], "^legitimate.*-apb$",
		"mismatch whitelist entry")
	ft.AssertEqual(t, config.Registry[1].BlackList[1], "^specific-blacklist-apb$",
		"mismatch blacklist entry")

	ft.AssertEqual(t, config.Dao.EtcdHost, "localhost", "")
	ft.AssertEqual(t, config.Dao.EtcdPort, "2379", "")

	ft.AssertEqual(t, config.Log.LogFile, "/var/log/ansible-service-broker/asb.log", "")
	ft.AssertTrue(t, config.Log.Stdout, "")
	ft.AssertTrue(t, config.Log.Color, "")
	ft.AssertEqual(t, config.Log.Level, "debug", "")

	ft.AssertEqual(t, config.Openshift.Host, "", "")
	ft.AssertEqual(t, config.Openshift.CAFile, "", "")
	ft.AssertEqual(t, config.Openshift.BearerTokenFile, "", "")
	ft.AssertEqual(t, config.Openshift.PullPolicy, "IfNotPresent", "")
	ft.AssertEqual(t, config.Openshift.SandboxRole, "edit", "")

	ft.AssertTrue(t, config.Broker.BootstrapOnStartup, "mismatch bootstrap")
	ft.AssertTrue(t, config.Broker.DevBroker, "mismatch dev")
	ft.AssertTrue(t, config.Broker.Recovery, "mismatch recovery")
	ft.AssertTrue(t, config.Broker.OutputRequest, "mismatch output")
	ft.AssertFalse(t, config.Broker.LaunchApbOnBind, "mismatch launch")
	ft.AssertEqual(t, config.Broker.SSLCert, "/path/to/cert", "mismatch cert")
	ft.AssertEqual(t, config.Broker.SSLCertKey, "/path/to/key", "mismatch key")

	ft.AssertEqual(t, config.Broker.Auth[0].Type, "basic", "mismatch basic")
	ft.AssertTrue(t, config.Broker.Auth[0].Enabled, "mismatch enable")
	ft.AssertEqual(t, config.Broker.Auth[1].Type, "oauth", "mismatch basic")
	ft.AssertFalse(t, config.Broker.Auth[1].Enabled, "mismatch enable")
}

func TestAdminRole(t *testing.T) {
	config, err := CreateConfig("testdata/test-config-admin.yaml", "testdata/test-registry-auth.yaml")
	if err != nil {
		t.Fatal(err.Error())
	}

	ft.AssertEqual(t, config.Openshift.SandboxRole, "admin", "")
}
