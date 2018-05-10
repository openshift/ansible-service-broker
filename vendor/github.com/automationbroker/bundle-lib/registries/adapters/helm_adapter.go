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
	"archive/tar"
	"compress/gzip"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"path"
	"sort"
	"strings"
	"time"

	"github.com/Masterminds/semver"
	"github.com/automationbroker/bundle-lib/bundle"
	"github.com/ghodss/yaml"
)

const (
	helmName          = "helm"
	helmIndexPath     = "/index.yaml"
	valuesFilePattern = "*/values.yaml"
)

// HelmAdapter - Helm Registry Adapter
type HelmAdapter struct {
	Config Configuration
	Charts map[string]ChartVersions
}

// ChartVersions is a list of versioned chart references.
// Implements a sorter on Version.
type ChartVersions []*ChartVersion

// Len returns the length of the list of versioned chart references.
func (c ChartVersions) Len() int { return len(c) }

// Swap swaps the position of two items in the versions slice.
func (c ChartVersions) Swap(i, j int) { c[i], c[j] = c[j], c[i] }

// Less returns true if the version of entry a is less than the version of entry b.
func (c ChartVersions) Less(a, b int) bool {
	// Failed parse pushes to the back.
	i, err := semver.NewVersion(c[a].Version)
	if err != nil {
		return true
	}
	j, err := semver.NewVersion(c[b].Version)
	if err != nil {
		return false
	}
	return i.LessThan(j)
}

// IndexFile represents the index file in a chart repository
// https://github.com/kubernetes/helm/blob/48e703997016f3edeb4f0b90e6cfdb3456ce3db0/pkg/repo/index.go#L78
type IndexFile struct {
	APIVersion string                   `json:"apiVersion"`
	Generated  time.Time                `json:"generated"`
	Entries    map[string]ChartVersions `json:"entries"`
	PublicKeys []string                 `json:"publicKeys,omitempty"`
}

// Maintainer describes a Chart maintainer.
// https://github.com/kubernetes/helm/blob/48e703997016f3edeb4f0b90e6cfdb3456ce3db0/pkg/proto/hapi/chart/metadata.pb.go#L37
type Maintainer struct {
	// Name is a user name or organization name
	Name string `protobuf:"bytes,1,opt,name=name" json:"name,omitempty"`
	// Email is an optional email address to contact the named maintainer
	Email string `protobuf:"bytes,2,opt,name=email" json:"email,omitempty"`
	// Url is an optional URL to an address for the named maintainer
	URL string `protobuf:"bytes,3,opt,name=url" json:"url,omitempty"`
}

// ChartVersion represents a chart entry in the IndexFile
// https://github.com/kubernetes/helm/blob/48e703997016f3edeb4f0b90e6cfdb3456ce3db0/pkg/repo/index.go#L216
// https://github.com/kubernetes/helm/blob/48e703997016f3edeb4f0b90e6cfdb3456ce3db0/pkg/proto/hapi/chart/metadata.pb.go#L75
type ChartVersion struct {
	// The name of the chart
	Name string `protobuf:"bytes,1,opt,name=name" json:"name,omitempty"`
	// The URL to a relevant project page, git repo, or contact person
	Home string `protobuf:"bytes,2,opt,name=home" json:"home,omitempty"`
	// Source is the URL to the source code of this chart
	Sources []string `protobuf:"bytes,3,rep,name=sources" json:"sources,omitempty"`
	// A SemVer 2 conformant version string of the chart
	Version string `protobuf:"bytes,4,opt,name=version" json:"version,omitempty"`
	// A one-sentence description of the chart
	Description string `protobuf:"bytes,5,opt,name=description" json:"description,omitempty"`
	// A list of string keywords
	Keywords []string `protobuf:"bytes,6,rep,name=keywords" json:"keywords,omitempty"`
	// A list of name and URL/email address combinations for the maintainer(s)
	Maintainers []*Maintainer `protobuf:"bytes,7,rep,name=maintainers" json:"maintainers,omitempty"`
	// The name of the template engine to use. Defaults to 'gotpl'.
	Engine string `protobuf:"bytes,8,opt,name=engine" json:"engine,omitempty"`
	// The URL to an icon file.
	Icon string `protobuf:"bytes,9,opt,name=icon" json:"icon,omitempty"`
	// The API Version of this chart.
	APIVersion string `protobuf:"bytes,10,opt,name=apiVersion" json:"apiVersion,omitempty"`
	// The condition to check to enable chart
	Condition string `protobuf:"bytes,11,opt,name=condition" json:"condition,omitempty"`
	// The tags to check to enable chart
	Tags string `protobuf:"bytes,12,opt,name=tags" json:"tags,omitempty"`
	// The version of the application enclosed inside of this chart.
	AppVersion string `protobuf:"bytes,13,opt,name=appVersion" json:"appVersion,omitempty"`
	// Whether or not this chart is deprecated
	Deprecated bool `protobuf:"varint,14,opt,name=deprecated" json:"deprecated,omitempty"`
	// TillerVersion is a SemVer constraints on what version of Tiller is required.
	// See SemVer ranges here: https://github.com/Masterminds/semver#basic-comparisons
	TillerVersion string `protobuf:"bytes,15,opt,name=tillerVersion" json:"tillerVersion,omitempty"`
	// Annotations are additional mappings uninterpreted by Tiller,
	// made available for inspection by other applications.
	Annotations map[string]string `protobuf:"bytes,16,rep,name=annotations" json:"annotations,omitempty" protobuf_key:"bytes,1,opt,name=key" protobuf_val:"bytes,2,opt,name=value"`
	// KubeVersion is a SemVer constraint specifying the version of Kubernetes required.
	KubeVersion string    `protobuf:"bytes,17,opt,name=kubeVersion" json:"kubeVersion,omitempty"`
	URLs        []string  `json:"urls"`
	Created     time.Time `json:"created,omitempty"`
	Removed     bool      `json:"removed,omitempty"`
	Digest      string    `json:"digest,omitempty"`
}

// RegistryName - Retrieve the registry name
func (r *HelmAdapter) RegistryName() string {
	return helmName
}

// GetImageNames - retrieve the images
func (r *HelmAdapter) GetImageNames() ([]string, error) {
	var imageNames []string

	r.Charts = map[string]ChartVersions{}

	index, err := r.getHelmIndex()
	if err != nil {
		return imageNames, err
	}

	for name, entry := range index.Entries {
		// Do not add a chart w/o at least one chart version
		if len(entry) == 0 {
			continue
		}

		r.Charts[name] = entry
		imageNames = append(imageNames, name)
	}

	return imageNames, nil
}

// FetchSpecs - retrieve the spec for the image names.
func (r *HelmAdapter) FetchSpecs(imageNames []string) ([]*bundle.Spec, error) {
	var specs []*bundle.Spec

	for _, name := range imageNames {
		var (
			chartVersions []string
			values        string
		)

		charts, ok := r.Charts[name]
		if !ok {
			continue
		}

		for _, chart := range charts {
			chartVersions = append(chartVersions, chart.Version)
		}

		// Use the latest chart for creating the bundle:
		// This works works because we previously sorted the chart's versions
		// and excluded charts w/o at least one chart version.
		chart := charts[0]

		if len(chart.URLs) > 0 {
			resp, err := http.Get(chart.URLs[0])
			if err != nil {
				continue
			}
			defer resp.Body.Close()

			values = r.loadArchive(resp.Body)
		}

		// Convert chart to Bundle Spec
		spec := &bundle.Spec{
			Runtime:     2,
			Version:     "1.0",
			Async:       "optional",
			Bindable:    false,
			Image:       r.Config.Runner,
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
			Plans: []bundle.Plan{
				bundle.Plan{
					Name:        "default",
					Description: "Default plan for running helm charts",
					Parameters: []bundle.ParameterDescriptor{
						bundle.ParameterDescriptor{
							Name:      "repo",
							Title:     "Helm Chart Repository URL",
							Type:      "string",
							Default:   r.Config.URL.String(),
							Pattern:   fmt.Sprintf("^%s$", r.Config.URL.String()),
							Updatable: false,
							Required:  false,
						},
						bundle.ParameterDescriptor{
							Name:      "chart",
							Title:     "Helm Chart",
							Type:      "string",
							Default:   chart.Name,
							Pattern:   fmt.Sprintf("^%s$", chart.Name),
							Updatable: false,
							Required:  false,
						},
						bundle.ParameterDescriptor{
							Name:      "version",
							Title:     "Helm Chart Version",
							Type:      "enum",
							Enum:      chartVersions,
							Default:   chart.Version,
							Updatable: true,
							Required:  false,
						},
						bundle.ParameterDescriptor{
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

// getHelmIndex returns a helm repository IndexFile object
// https://github.com/kubernetes/helm/blob/48e703997016f3edeb4f0b90e6cfdb3456ce3db0/pkg/repo/index.go#L271
func (r *HelmAdapter) getHelmIndex() (*IndexFile, error) {
	index := &IndexFile{}

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
	for _, versions := range index.Entries {
		sort.Sort(sort.Reverse(versions))
	}
	if index.APIVersion == "" {
		return index, fmt.Errorf("No APIVersion on Index file")
	}

	return index, nil
}

// loadArchive returns the Helm Chart values as a string
// https://github.com/kubernetes/helm/blob/48e703997016f3edeb4f0b90e6cfdb3456ce3db0/pkg/chartutil/load.go#L66
func (r *HelmAdapter) loadArchive(in io.Reader) string {
	unzipped, err := gzip.NewReader(in)
	if err != nil {
		return ""
	}
	defer unzipped.Close()

	tr := tar.NewReader(unzipped)
	for {
		hdr, err := tr.Next()
		if err != nil {
			return ""
		}

		valuesMatch, err := path.Match(valuesFilePattern, hdr.Name)
		if err != nil {
			return ""
		}
		if valuesMatch {
			data, err := ioutil.ReadAll(tr)
			if err != nil {
				return ""
			}
			return string(data)
		}
	}

}
