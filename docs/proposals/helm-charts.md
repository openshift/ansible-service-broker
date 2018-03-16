# Basic Helm Chart Registry Adapter Proposal

## Introduction

Add a new registry adapter type, `helm`, that supports Helm Chart
Repositories. With this change, an admin could configure the broker by adding
a `helm` registry type and have the Helm charts hosted in that Helm Chart
Repository converted into objects the broker knows how to manage for use via
the Service Catalog.

## Problem Description

Currently, our broker only supports docker container image type registries.
However, with [Helm](https://helm.sh) being the "package manager for Kubernetes",
there is already a strong collection of applications defined in the form of
Helm charts. Modifying the broker to consume these Helm charts as Service
Bundles allows work already done in Helm charts to be made available in
a cluster.

## Creating a Base Image for Executing Helm Charts

Ansible Playbook Bundles (APBs) have served as an initial implementation of the
more generic [Service Bundle Contract](../service-bundle.md). Here, we will be
creating a new type of Service Bundle, Helm Bundles, that respect the Service
Bundle contract but use Helm as the runtime. We already have a
[Helm Bundle Base](https://github.com/ansibleplaybookbundle/helm-bundle-base)
and this [PR](https://github.com/ansibleplaybookbundle/helm-bundle-base/pull/2)
shows the work required to implement the Service Bundle contract generically
using Helm.

## Broker Modifications

### Broker Config Options

It is expected that a broker admin who wants to expose Helm charts would update
the broker's config to look something like:

```
registry:
  - type: helm
    name: stable
    url: "https://kubernetes-charts.storage.googleapis.com"
    runner: "docker.io/djzager/helm-runner:latest"
    white_list:
      - ".*"
```

1. Type: refers to the registry adapter to handle Helm Chart Repositories
1. Name: gives a name to this registry item
1. URL: the URL for the Helm Chart Repository
1. Runner: the container image to use when interacting with a given
   Helm Chart
1. White List/Black List: allows the Admin to filter out Helm Charts

### Changes to the Registry Package

- Add the ability to read the `runner` from the broker config.
- Pass the URL and BaseImage to the registry adapter.
- Instantiate a Helm registry adapter to handle Helm Chart repositories in the
  config.

### Creating a Helm Adapter

The `HelmAdapter` struct will look something like the example below:

```
// import "k8s.io/helm/pkg/repo"
// HelmAdapter - Helm Registry Adapter
type HelmAdapter struct {
    Config Configuration
    Log    *logging.Logger
    Charts map[string][]*repo.ChartVersion
}
```

The addition of the `Charts` field makes it possible to save our work and only
read the registries' index file once.

When the registry package calls `GetImageNames()` the Helm adapter will read the
`index.yaml` found at the Helm Chart Repository URL based on the
[chart repository structure](https://github.com/kubernetes/helm/blob/master/docs/chart_repository.md#the-chart-repository-structure).
Inside `GetImageNames()`, the `Charts` fields will be initialized and we will
save the `ChartVersion`s into the `Charts` field based on the Chart's name. For
example:

```
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
```

This way we only have to evaluate the Chart Repository on the call to
`GetImageNames()` and the subsequent call to `FetchSpecs()` can work
to transform Charts:

- Find the chart by name in the `Charts` field
- Save all the versions of the chart:

```
var chartVersions []string
for _, chart := range charts {
    chartVersions = append(chartVersions, chart.Version)
}
```

- Use the latest version of the chart to create a bundle `Spec` object:

```
// Use the latest chart for creating the bundle
chart := charts[0]
```

- Load the chart in memory to grab the content of the values file:

```
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
```

- Create a bundle `Spec` object:

```
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
                    Name:      "name",
                    Title:     "Release Name",
                    Type:      "string",
                    Default:   "helm-sb",
                    Updatable: false,
                    Required:  false,
                },
                apb.ParameterDescriptor{
                    Name:        "values",
                    Title:       "Values",
                    Type:        "string",
                    DisplayType: "textarea",
                    Default:     values,
                    Updatable:   false,
                    Required:    false,
                },
            },
        },
    },
}
```

## Work Not Covered

Helm Chart Repositories can be authenticated, [this
issue](https://github.com/kubernetes/helm/issues/1038) provides more
information. The broker should support authenticated chart repositories.
A separate proposal will be created for this work.
