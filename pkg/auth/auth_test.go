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

package auth

import (
	"io/ioutil"
	"os"
	"strings"
	"testing"

	"github.com/openshift/ansible-service-broker/pkg/config"
	ft "github.com/openshift/ansible-service-broker/pkg/fusortest"
)

func TestNewFusa(t *testing.T) {
	username := []byte("admin")
	password := []byte("admin")
	ioutil.WriteFile("/tmp/username", username, 0644)
	ioutil.WriteFile("/tmp/password", password, 0644)

	defer os.Remove("/tmp/username")
	defer os.Remove("/tmp/password")

	fusa, err := NewFileUserServiceAdapter("/tmp/")
	if err != nil {
		t.Fatal(err.Error())
	}
	adminuser, _ := fusa.FindByLogin("admin")
	ft.AssertEqual(t, adminuser.Username, "admin", "username does not match")
	ft.AssertEqual(t, adminuser.Password, "admin", "password does not match")
	ft.AssertTrue(t, fusa.ValidateUser("admin", "admin"), "validation failed")
	ft.AssertFalse(t, fusa.ValidateUser("notme", "admin"), "validation passed, expected failure")
	ft.AssertFalse(t, fusa.ValidateUser("", ""), "expected failure on empty string")
}

func TestErrorBuild(t *testing.T) {
	fusa, err := NewFileUserServiceAdapter("")
	if fusa != nil {
		t.Fatal("fusa is not nil")
	}
	ft.AssertNotNil(t, err, "expected an error")
	ft.AssertTrue(t, strings.Contains(err.Error(), "directory is empty,"))
}

func TestFusaError(t *testing.T) {
	_, err := NewFileUserServiceAdapter("/var/tmp")
	ft.AssertNotNil(t, err, "should have gotten an error")
	ft.AssertTrue(t, strings.Contains(err.Error(), "no such file or directory"), "mismatch error message")
}

func TestUser(t *testing.T) {
	user := User{Username: "admin", Password: "password"}
	ft.AssertEqual(t, user.GetType(), "user", "type doesn't match user")
	ft.AssertEqual(t, user.GetName(), user.Username, "get name and username do not match")
}

func TestGetProviders(t *testing.T) {
	t.Skip("requires /var/run/asb-auth/{username,password} to be present")
	config, err := config.CreateConfig("testdata/test-config.yaml")
	if err != nil {
		t.Fatalf("Unable to create config - %v", err)
	}

	testproviders := GetProviders(config)

	t.Log(len(testproviders))
	ft.AssertEqual(t, len(testproviders), 1, "providers not parsed correctly")
}
