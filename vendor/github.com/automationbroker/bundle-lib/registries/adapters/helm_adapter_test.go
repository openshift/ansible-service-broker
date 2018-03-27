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

package adapters

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	ft "github.com/stretchr/testify/assert"
)

const MariaDB = "mariadb"
const MariaDBPath = "/mariadb-2.1.4.tgz"
const RepoResponse = `
apiVersion: v1
entries:
  mariadb:
  - appVersion: 10.1.31
    created: 2018-03-19T13:33:01.77344858-04:00
    description: Fast, reliable, scalable, and easy to use open-source relational
      database system. MariaDB Server is intended for mission-critical, heavy-load
      production systems as well as for embedding into mass-deployed software.
    digest: bb3900d825e06e63bebd35fd05d16b05f6be030abdf02513b9affa3fd17f67b4
    engine: gotpl
    home: https://mariadb.org
    icon: https://bitnami.com/assets/stacks/mariadb/img/mariadb-stack-220x234.png
    keywords:
    - mariadb
    - mysql
    - database
    - sql
    - prometheus
    maintainers:
    - email: containers@bitnami.com
      name: bitnami-bot
    name: mariadb
    sources:
    - https://github.com/bitnami/bitnami-docker-mariadb
    - https://github.com/prometheus/mysqld_exporter
    urls:
    - mariadb-2.1.4.tgz
    version: 2.1.4
generated: 2018-03-19T13:33:01.771496768-04:00
`

func TestHelmRegistryName(t *testing.T) {
	ha := HelmAdapter{}
	ft.Equal(t, ha.RegistryName(), "helm", "Helm adapter name mismatch")
}

func TestHelmFetchSpecs(t *testing.T) {
	// Set up a fake http server
	serv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" {
			t.Errorf("Expected 'Get' request, got '%s'", r.Method)
		}

		if strings.HasPrefix(r.URL.EscapedPath(), "/index.yaml") {
			w.WriteHeader(http.StatusOK)
			fmt.Fprintf(w, RepoResponse)
		} else if r.URL.EscapedPath() == MariaDBPath {
			response, err := ioutil.ReadFile("testdata" + MariaDBPath)
			if err != nil {
				t.Fatal("ERROR: ", err)
			}
			w.WriteHeader(http.StatusOK)
			w.Write(response)
		} else {
			t.Errorf("Expected '/index.yaml' URL Path, got '%s'", r.URL.EscapedPath())
		}

	}))
	url, err := url.Parse(serv.URL)
	if err != nil {
		t.Fatal("ERROR: ", err)
	}

	// Create a helm adapter
	ha := HelmAdapter{
		Config: Configuration{
			URL:    url,
			Runner: "runner_image",
		},
	}

	// Get image names
	imageNames, err := ha.GetImageNames()
	if err != nil {
		t.Fatal("ERROR: ", err)
	}
	ft.Equal(t, len(imageNames), 1)
	ft.Equal(t, imageNames[0], MariaDB)
	ft.NotNil(t, imageNames)

	// Override Chart URL to point to our http server
	ha.Charts[MariaDB][0].URLs[0] = url.String() + MariaDBPath

	// Fetch Specs
	specs, err := ha.FetchSpecs(imageNames)
	if err != nil {
		t.Fatal("ERROR: ", err)
	}
	ft.Equal(t, len(specs), 1)
	ft.NotNil(t, specs)

	spec := specs[0]
	ft.Equal(t, spec.Runtime, 2)
	ft.Equal(t, spec.Version, "1.0")
	ft.Equal(t, spec.Image, "runner_image")
	ft.Equal(t, spec.Metadata["displayName"], "mariadb (Helm)")
}
