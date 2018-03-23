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

package registries

import (
	"io/ioutil"
	"os"
	"strings"
	"testing"

	yaml "gopkg.in/yaml.v1"

	log "github.com/sirupsen/logrus"
	ft "github.com/stretchr/testify/assert"
)

const testWhitelistFile = "whitelist.yaml"
const testBlacklistFile = "blacklist.yaml"
const emptyWhiteListFile = "empty_whitelist.yaml"
const testBlacklistOverrideFile = "blacklist_override.yaml"

func testGetRegexFromFile(file string) []string {
	gopath := os.Getenv("GOPATH")
	if gopath == "" {
		panic("GOPATH not set!!")
	}

	filePath := strings.Join([]string{
		gopath, "src", "github.com", "automationbroker",
		"bundle-lib", "registries", "testdata", file,
	}, "/")
	contents, err := ioutil.ReadFile(filePath)
	if err != nil {
		panic(err)
	}
	var regex []string
	if yaml.Unmarshal(contents, &regex); err != nil {
		panic(err)
	}
	return regex
}

func testNames() []string {
	return []string{
		"legitimate-postgresql-apb",
		"legitimate-mediawiki-apb",
		"totally-not-malicious-apb",
		"malicious-bar-apb",
		"specific-blacklist-apb",
		"foo-apb",
		"bar-apb",
		"rhscl-postgresql-apb",
		"baz-apb",
		"foobar-apb",
	}
}

func testSetEq(a []string, b []string) bool {
	if a == nil && b == nil {
		return true
	}
	if a == nil || b == nil {
		return false
	}
	if len(a) != len(b) {
		return false
	}

	toSet := func(s []string) map[string]bool {
		m := make(map[string]bool)
		for _, k := range s {
			m[k] = true
		}
		return m
	}

	as := toSet(a)
	bs := toSet(b)

	for k := range as {
		if _, ok := bs[k]; !ok {
			return false
		}
	}

	return true
}

func TestOnlyBlacklist(t *testing.T) {
	filter := Filter{whitelist: []string{},
		blacklist: testGetRegexFromFile(testBlacklistFile)}
	filter.Init()

	expectedValidNames := []string{}

	expectedFilteredNames := []string{
		"legitimate-postgresql-apb",
		"legitimate-mediawiki-apb",
		"foo-apb",
		"bar-apb",
		"rhscl-postgresql-apb",
		"baz-apb",
		"foobar-apb",
		"totally-not-malicious-apb",
		"malicious-bar-apb",
		"specific-blacklist-apb",
	}

	validNames, filteredNames := filter.Run(testNames())

	expectedTotal := append(expectedValidNames, expectedFilteredNames...)
	ft.True(t, testSetEq(expectedValidNames, validNames))
	ft.True(t, testSetEq(expectedFilteredNames, filteredNames))
	ft.True(t, testSetEq(expectedTotal, testNames()))
}

func TestOnlyWhitelist(t *testing.T) {
	filter := Filter{whitelist: testGetRegexFromFile(testWhitelistFile),
		blacklist: []string{}}
	filter.Init()

	expectedValidNames := []string{
		"legitimate-postgresql-apb",
		"legitimate-mediawiki-apb",
		"foo-apb",
		"bar-apb",
		"rhscl-postgresql-apb",
	}

	expectedFilteredNames := []string{
		"totally-not-malicious-apb",
		"malicious-bar-apb",
		"specific-blacklist-apb",
		// Not explicitly whitelisted, so should be filtered
		"baz-apb",
		"foobar-apb",
	}

	validNames, filteredNames := filter.Run(testNames())

	expectedTotal := append(expectedValidNames, expectedFilteredNames...)
	ft.True(t, testSetEq(expectedValidNames, validNames))
	ft.True(t, testSetEq(expectedFilteredNames, filteredNames))
	ft.True(t, testSetEq(expectedTotal, testNames()))
}

func TestEmptyWhitelist(t *testing.T) {
	filter := Filter{whitelist: testGetRegexFromFile(emptyWhiteListFile),
		blacklist: []string{}}
	filter.Init()

	expectedValidNames := []string{}

	expectedFilteredNames := []string{
		"totally-not-malicious-apb",
		"malicious-bar-apb",
		"specific-blacklist-apb",
		"baz-apb",
		"foobar-apb",
		"legitimate-postgresql-apb",
		"legitimate-mediawiki-apb",
		"foo-apb",
		"bar-apb",
		"rhscl-postgresql-apb",
	}

	validNames, filteredNames := filter.Run(testNames())

	expectedTotal := append(expectedValidNames, expectedFilteredNames...)
	ft.True(t, testSetEq(expectedValidNames, validNames))
	ft.True(t, testSetEq(expectedFilteredNames, filteredNames))
	ft.True(t, testSetEq(expectedTotal, testNames()))
}

func TestBlackAndWhitelistOverride(t *testing.T) {
	// If both a black and whitelist are present and both contain matches,
	// blacklist will override anything that has passed the whitelist match and
	// get excluded
	filter := Filter{
		whitelist: testGetRegexFromFile(testWhitelistFile),
		blacklist: testGetRegexFromFile(testBlacklistOverrideFile),
	}
	filter.Init()

	expectedValidNames := []string{
		"legitimate-postgresql-apb",
		"legitimate-mediawiki-apb",
		"bar-apb",
		"rhscl-postgresql-apb",
	}

	expectedFilteredNames := []string{
		"foo-apb", // Appears in white and blacklist, therefore, excluded
		"totally-not-malicious-apb",
		"malicious-bar-apb",
		"specific-blacklist-apb",
		"foobar-apb",
		"baz-apb",
	}

	validNames, filteredNames := filter.Run(testNames())

	expectedTotal := append(expectedValidNames, expectedFilteredNames...)
	ft.True(t, testSetEq(expectedValidNames, validNames))
	ft.True(t, testSetEq(expectedFilteredNames, filteredNames))
	ft.True(t, testSetEq(expectedTotal, testNames()))
}

func TestBlackAndWhitelistNoOverride(t *testing.T) {
	// Both lists set, but no overlap between sets. foo-apb should *not*
	// be filtered compared to the Override case
	filter := Filter{
		whitelist: testGetRegexFromFile(testWhitelistFile),
		blacklist: testGetRegexFromFile(testBlacklistFile),
	}
	filter.Init()

	expectedValidNames := []string{
		"foo-apb", // NOTE: expected
		"legitimate-postgresql-apb",
		"legitimate-mediawiki-apb",
		"bar-apb",
		"rhscl-postgresql-apb",
	}

	expectedFilteredNames := []string{
		"totally-not-malicious-apb",
		"malicious-bar-apb",
		"specific-blacklist-apb",
		"foobar-apb",
		"baz-apb",
	}

	validNames, filteredNames := filter.Run(testNames())

	expectedTotal := append(expectedValidNames, expectedFilteredNames...)
	ft.True(t, testSetEq(expectedValidNames, validNames))
	ft.True(t, testSetEq(expectedFilteredNames, filteredNames))
	ft.True(t, testSetEq(expectedTotal, testNames()))
}

func TestOnlyWhitelistWithEmptyString(t *testing.T) {
	filter := Filter{whitelist: []string{""},
		blacklist: []string{}}
	filter.Init()

	log.Debugf("filter regex: %#v", filter.whiteRegexp)
	log.Debugf("filter regex: %#v", filter.blackRegexp)

	expectedFilteredNames := []string{
		"totally-not-malicious-apb",
		"malicious-bar-apb",
		"specific-blacklist-apb",
		// Not explicitly whitelisted, so should be filtered
		"baz-apb",
		"foobar-apb",
		"legitimate-postgresql-apb",
		"legitimate-mediawiki-apb",
		"foo-apb",
		"bar-apb",
		"rhscl-postgresql-apb",
	}

	validNames, filteredNames := filter.Run(testNames())

	expectedTotal := expectedFilteredNames
	log.Debugf("validNames: %#v", validNames)
	log.Debugf("filteredNames: %#v", filteredNames)
	log.Debugf("expectedTotal: %#v", expectedTotal)
	ft.True(t, testSetEq(nil, validNames))
	ft.True(t, testSetEq(expectedFilteredNames, filteredNames))
	ft.True(t, testSetEq(expectedTotal, testNames()))
}
