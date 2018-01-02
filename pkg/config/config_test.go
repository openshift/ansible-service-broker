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

package config

import (
	"fmt"
	"os"
	"testing"

	ft "github.com/openshift/ansible-service-broker/pkg/fusortest"
)

/*func TestCreateConfig(t *testing.T) {
	config, err := CreateConfig("testdata/test-config.yaml")
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
*/

var config *Config

func TestMain(m *testing.M) {
	/*
		config = Config{config: map[string]interface{}{
			"registry": []interface{}{
				map[string]interface{}{
					"type": "dockerhub",
					"name": "dh",
					"url":  "https://registry.hub.docker.com",
					"user": "shurley",
					"pass": "testingboom",
					"org":  "shurley",
				}, map[string]interface{}{
					"pass": "testingboom",
					"org":  "ansibleplaybookbundle",
					"type": "dockerhub",
					"name": "play",
					"url":  "https://registry.hub.docker.com",
					"user": "shurley",
				},
			},
			"broker": map[string]interface{}{
				"launch_apb_on_bind":   "false",
				"bootstrap_on_startup": true,
				"recovery":             true,
				"output_request":       true,
				"ssl_cert_key":         "/var/run/secrets/kubernetes.io/serviceaccount/tls.key",
				"ssl_cert":             "/var/run/secrets/kubernetes.io/serviceaccount/tls.crt",
				"refresh_interval":     "600s",
				"dev_broker":           true,
				"testInt":              100,
				"testFloat32":          32.87,
				"testFloat64":          45677.0799485958595,
			},
		},
			mutex: sync.RWMutex{},
		}
	*/
	c, err := CreateConfig("testdata/generated_local_development.yaml")
	if err != nil {
		fmt.Printf("Unable to create config - %v", err)
	}
	config = c
	retCode := m.Run()
	os.Exit(retCode)
}

func TestConfigGetInt(t *testing.T) {
	testInt := config.GetInt("broker.testInt")
	testInvalidInt := config.GetInt("makes.no.sense")
	ft.AssertEqual(t, 100, testInt)
	ft.AssertEqual(t, 0, testInvalidInt)
}

func TestConfigGetString(t *testing.T) {
	testString := config.GetString("registry.dh.user")
	testInvalidString := config.GetString("makes.no.sense")
	ft.AssertEqual(t, "shurley", testString)
	ft.AssertEqual(t, "", testInvalidString)
}

func TestConfigGetSliceString(t *testing.T) {
	testString := config.GetSliceOfStrings("registry.dh.black_list")
	whiteList := config.GetSliceOfStrings("registry.dh.white_list")
	value := []string{"malicious.*-apb$", "^specific-blacklist-apb$"}
	whiteListValue := []string{"*-apb$"}
	if len(testString) == 0 || len(whiteList) == 0 {
		t.Fail()
	}
	for i, str := range testString {
		ft.AssertEqual(t, value[i], str)
	}
	for i, str := range whiteList {
		ft.AssertEqual(t, whiteListValue[i], str)
	}
}

func TestConfigGetFloat32(t *testing.T) {
	testFloat32 := config.GetFloat32("broker.testFloat32")
	testInvalidFloat32 := config.GetFloat32("makes.no.sense")
	var defaultFloat32 float32
	ft.AssertEqual(t, float32(32.87), testFloat32)
	ft.AssertEqual(t, defaultFloat32, testInvalidFloat32)
}

func TestConfigGetFloat64(t *testing.T) {
	testFloat64 := config.GetFloat64("broker.testFloat64")
	testInvalidFloat64 := config.GetFloat64("makes.no.sense")
	var defaultFloat64 float64
	ft.AssertEqual(t, float64(45677.0799485958595), testFloat64)
	ft.AssertEqual(t, defaultFloat64, testInvalidFloat64)
}

func TestConfigGetBool(t *testing.T) {
	testBoolTrue := config.GetBool("broker.recovery")
	testInvalidBool := config.GetBool("makes.no.sense")
	ft.AssertTrue(t, testBoolTrue)
	ft.AssertFalse(t, testInvalidBool)
}

func TestConfigGetSubMap(t *testing.T) {
	testInvalidSubMap := config.GetSubConfig("makes.no.sense")
	testNoNameArray := config.GetSubConfig("dh_no_names")
	testSubMap := config.GetSubConfig("registry")
	ft.AssertEqual(t, testSubMap.GetString("dh.name"), "dh")
	ft.AssertEqual(t, testSubMap.GetString("play.user"), "shurley")
	ft.AssertEqual(t, testSubMap.GetString("dh.url"), "https://registry.hub.docker.com")
	ft.AssertEqual(t, testSubMap.GetString("dh.pass"), "testingboom")
	ft.AssertEqual(t, testSubMap.GetString("dh.org"), "shurley")
	ft.AssertEqual(t, testSubMap.GetString("play.org"), "ansibleplaybookbundle")
	ft.AssertEqual(t, testSubMap.GetString("dh.type"), "dockerhub")
	ft.AssertEqual(t, testSubMap.GetString("play.type"), "dockerhub")
	ft.AssertTrue(t, testInvalidSubMap.Empty())
	ft.AssertTrue(t, testNoNameArray.Empty())
}

func TestConfigGetMap(t *testing.T) {
	testMap := config.GetSubConfig("registry").ToMap()
	_, ok := testMap["dh"]
	ft.AssertTrue(t, ok)
	_, ok = testMap["dockerhub"]
	ft.AssertFalse(t, ok)
}
