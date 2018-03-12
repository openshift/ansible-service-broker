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
	"strings"

	logging "github.com/op/go-logging"

	"github.com/ghodss/yaml"
	"github.com/openshift/ansible-service-broker/pkg/apb"
	"k8s.io/helm/pkg/chartutil"
	"k8s.io/helm/pkg/repo"
)

const (
	helmName      = "helm"
	helmIndexPath = "/index.yaml"
)

// HelmAdapter - Helm Registry Adapter
type HelmAdapter struct {
	Config Configuration
	Log    *logging.Logger
	Charts map[string][]*repo.ChartVersion
}

// RegistryName - Retrieve the registry name
func (r *HelmAdapter) RegistryName() string {
	return helmName
}

// GetImageNames - retrieve the images
func (r *HelmAdapter) GetImageNames() ([]string, error) {
	var imageNames []string

	r.Charts = map[string][]*repo.ChartVersion{}

	index, err := r.getHelmIndex()
	if err != nil {
		return imageNames, err
	}

	for name, entry := range index.Entries {
		if len(entry) == 0 {
			continue
		}

		r.Charts[name] = entry
		imageNames = append(imageNames, name)
	}

	return imageNames, nil
}

// FetchSpecs - retrieve the spec for the image names.
func (r *HelmAdapter) FetchSpecs(imageNames []string) ([]*apb.Spec, error) {
	var (
		specs  []*apb.Spec
		values string
	)

	for _, name := range imageNames {
		var chartVersions []string
		charts, ok := r.Charts[name]
		if !ok {
			continue
		}

		for _, chart := range charts {
			chartVersions = append(chartVersions, chart.Version)
		}
		// Use the latest chart for creating the bundle
		chart := charts[0]

		resp, err := http.Get(chart.URLs[0])
		if err != nil {
			return specs, err
		}
		defer resp.Body.Close()

		helmChart, err := chartutil.LoadArchive(resp.Body)
		if err != nil {
			return specs, err
		}

		if helmChart.Values != nil {
			values = helmChart.Values.Raw
		}

		// Convert chart to Bundle Spec
		spec := &apb.Spec{
			Runtime:     2,
			Version:     "1.0",
			Async:       "optional",
			Bindable:    false,
			Image:       r.Config.BaseImage,
			FQName:      chart.Name,
			Tags:        chart.Keywords,
			Description: chart.Description,
			Metadata: map[string]interface{}{
				//"longDescription":  chart.Description,
				"displayName":      fmt.Sprintf("%s (Helm)", chart.Name),
				"documentationUrl": chart.Home,
				"dependencies":     chart.Sources,
				"imageUrl":         chart.Icon,
			},
			Plans: []apb.Plan{
				apb.Plan{
					Name:        "default",
					Description: "Default plan for running helm charts",
					Parameters: []apb.ParameterDescriptor{
						apb.ParameterDescriptor{
							Name:      "repo",
							Title:     "Helm Chart Repository URL",
							Type:      "string",
							Default:   r.Config.URL.String(),
							Pattern:   fmt.Sprintf("^%s$", r.Config.URL.String()),
							Updatable: false,
							Required:  false,
						},
						apb.ParameterDescriptor{
							Name:      "chart",
							Title:     "Helm Chart",
							Type:      "string",
							Default:   chart.Name,
							Pattern:   fmt.Sprintf("^%s$", chart.Name),
							Updatable: false,
							Required:  false,
						},
						apb.ParameterDescriptor{
							Name:      "version",
							Title:     "Helm Chart Version",
							Type:      "enum",
							Enum:      chartVersions,
							Default:   chart.Version,
							Updatable: true,
							Required:  false,
						},
						apb.ParameterDescriptor{
							Name:        "values",
							Title:       "Values",
							Type:        "string",
							DisplayType: "textarea",
							Default:     values,
							Updatable:   true,
							Required:    false,
						},
					},
				},
			},
		}

		specs = append(specs, spec)
	}

	return specs, nil
}

// getHelmIndex returns an helm repository IndexFile object
func (r *HelmAdapter) getHelmIndex() (*repo.IndexFile, error) {
	index := &repo.IndexFile{}

	url := strings.TrimSuffix(r.Config.URL.String(), "/") + helmIndexPath
	resp, err := http.Get(url)
	if err != nil {
		return index, err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return index, err
	}

	err = yaml.Unmarshal(body, index)
	if err != nil {
		return index, err
	}
	index.SortEntries()
	if index.APIVersion == "" {
		return index, fmt.Errorf("No APIVersion on Index file")
	}

	return index, nil
}
