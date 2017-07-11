package registries

import (
	"os"
	"strings"
	"testing"

	ft "github.com/openshift/ansible-service-broker/pkg/fusortest"
)

const testWhitelistFile = "whitelist.yaml"
const testBlacklistFile = "blacklist.yaml"
const testBlacklistOverrideFile = "blacklist_override.yaml"

func testFilePath(file string) string {
	gopath := os.Getenv("GOPATH")
	if gopath == "" {
		panic("GOPATH not set!!")
	}

	return strings.Join([]string{
		gopath, "src", "github.com", "openshift",
		"ansible-service-broker", "pkg", "apb", "testdata", file,
	}, "/")

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

	for k, _ := range as {
		if _, ok := bs[k]; !ok {
			return false
		}
	}

	return true
}

func TestInit(t *testing.T) {
	filter := NewFilter(
		testFilePath(testWhitelistFile),
		testFilePath(testBlacklistFile),
	)

	err := filter.Init()
	ft.AssertNil(t, err)
}

func TestNoList(t *testing.T) {
	filter := NewFilter("", "")
	err := filter.Init()
	ft.AssertNotNil(t, err)
}

func TestOnlyBlacklist(t *testing.T) {
	filter := NewFilter("", testFilePath(testBlacklistFile))
	err := filter.Init()
	ft.AssertNil(t, err)

	expectedValidNames := []string{
		"legitimate-postgresql-apb",
		"legitimate-mediawiki-apb",
		"foo-apb",
		"bar-apb",
		"rhscl-postgresql-apb",
		"baz-apb",
		"foobar-apb",
	}

	expectedFilteredNames := []string{
		"totally-not-malicious-apb",
		"malicious-bar-apb",
		"specific-blacklist-apb",
	}

	validNames, filteredNames := filter.Run(testNames())

	expectedTotal := append(expectedValidNames, expectedFilteredNames...)
	ft.AssertNil(t, err)
	ft.AssertTrue(t, testSetEq(expectedValidNames, validNames))
	ft.AssertTrue(t, testSetEq(expectedFilteredNames, filteredNames))
	ft.AssertTrue(t, testSetEq(expectedTotal, testNames()))
}

func TestOnlyWhitelist(t *testing.T) {
	filter := NewFilter(testFilePath(testWhitelistFile), "")
	err := filter.Init()
	ft.AssertNil(t, err)

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
	ft.AssertNil(t, err)
	ft.AssertTrue(t, testSetEq(expectedValidNames, validNames))
	ft.AssertTrue(t, testSetEq(expectedFilteredNames, filteredNames))
	ft.AssertTrue(t, testSetEq(expectedTotal, testNames()))
}

func TestBlackAndWhitelistOverride(t *testing.T) {
	// If both a black and whitelist are present and both contain matches,
	// blacklist will override anything that has passed the whitelist match and
	// get excluded
	filter := NewFilter(
		testFilePath(testWhitelistFile),
		testFilePath(testBlacklistOverrideFile),
	)
	err := filter.Init()
	ft.AssertNil(t, err)

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
	ft.AssertNil(t, err)
	ft.AssertTrue(t, testSetEq(expectedValidNames, validNames))
	ft.AssertTrue(t, testSetEq(expectedFilteredNames, filteredNames))
	ft.AssertTrue(t, testSetEq(expectedTotal, testNames()))
}

func TestBlackAndWhitelistNoOverride(t *testing.T) {
	// Both lists set, but no overlap between sets. foo-apb should *not*
	// be filtered compared to the Override case
	filter := NewFilter(
		testFilePath(testWhitelistFile),
		testFilePath(testBlacklistFile),
	)
	err := filter.Init()
	ft.AssertNil(t, err)

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
	ft.AssertNil(t, err)
	ft.AssertTrue(t, testSetEq(expectedValidNames, validNames))
	ft.AssertTrue(t, testSetEq(expectedFilteredNames, filteredNames))
	ft.AssertTrue(t, testSetEq(expectedTotal, testNames()))
}
