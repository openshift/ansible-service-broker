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

var config *Config

func TestMain(m *testing.M) {
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
	testInvalidSubConfigArray := config.GetSubConfig("registry")
	testSubMap := config.GetSubConfig("broker.new_object")
	testSubMapBroker := config.GetSubConfig("broker")
	ft.AssertEqual(t, testSubMap.GetString("key"), "value1")
	ft.AssertEqual(t, testSubMap.GetString("key2"), "value2")
	ft.AssertEqual(t, testSubMapBroker.GetString("new_object.key"), "value1")
	ft.AssertEqual(t, testSubMapBroker.GetString("new_object.key2"), "value2")
	ft.AssertTrue(t, testInvalidSubMap.Empty())
	ft.AssertTrue(t, testInvalidSubConfigArray.Empty())
}

func TestConfigGetMap(t *testing.T) {
	testMap := config.GetSubConfig("broker").ToMap()
	_, ok := testMap["dev_broker"]
	ft.AssertTrue(t, ok)
	_, ok = testMap["recovery"]
	ft.AssertTrue(t, ok)
}

func TestGetSubConfigArray(t *testing.T) {
	testSubConfig := config.GetSubConfigArray("registry")
	testInvalidSubConfig := config.GetSubConfigArray("makes.no_sense")
	ft.AssertEqual(t, len(testInvalidSubConfig), 0)
	ft.AssertEqual(t, len(testSubConfig), 2)
	ft.AssertEqual(t, testSubConfig[0].GetString("name"), "dh")
	ft.AssertEqual(t, testSubConfig[1].GetString("name"), "play")
}
