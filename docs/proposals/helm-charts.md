# Helm Chart Registry Adapter Proposal

## Introduction

Add a new registry adapter type, `helm`, that supports Helm chart
repositories. With this change, an admin, could configure the broker by adding
a `helm` type registry and have the Helm charts hosted in that Helm chart
repository converted into objects the broker knows how to manage for use via
the Service Catalog.

## Problem Description

Currently, our broker only supports docker container image type registries.
However, with [Helm](https://helm.sh) being the "package manager for Kubernetes",
there is already a strong collection of applications defined in the form of
Helm charts. Adding the ability to the broker to consume these Helm charts as
Service Bundles allows work already done on Helm charts to be made available in
a cluster.

## Phase 1 - Adding a Basic Helm Chart Registry Adapter

### Creating a Base Image for Executing Helm Charts

Ansible Playbook Bundles (APBs) have already been described as an instance of
the more generic Service Bundle. Here, we will be creating a new type of
Service Bundle, Helm Bundles, that respect the Service Bundle contract but use
Helm as the runtime. We already have a [Helm Bundle
Base](https://github.com/ansibleplaybookbundle/helm-bundle-base) and the
changes shown in [Helm Bundle Base PR 2](https://github.com/ansibleplaybookbundle/helm-bundle-base/pull/2)
show the work required to implement the Service Bundle contract using Helm.

### Broker Config Options

It is expected that a broker admin who wants to expose Helm charts would update
the broker's config to look something like:

```
registry:
  - type: helm
    name: stable
    url: "https://kubernetes-charts.storage.googleapis.com"
    base_image: "docker.io/djzager/helm-bundle-base:latest"
    white_list:
      - ".*"
```

1. Type: refers to the registry adapter to handle this registry
1. Name: gives a name to this registry item
1. URL: the URL for the Helm chart repository
1. Base Image: the container image to use when provisioning/deprovisioning a
   Helm chart
1. White List/Black List: allows the Admin to filter out Helm charts

#### Changes to the Registry Package

- Add the ability to read the `base_image` from the broker config.
- Pass the URL and BaseImage to the registry adapter.
- Instantiate a Helm registry adapter to handle Helm Chart registries in the
  config.

#### Creating a Helm Adapter

The `HelmAdapter` struct will look something like, the addition of `Charts`
makes it possible to save our work and only ready the registries index file
once:

```
// import "k8s.io/helm/pkg/repo"
// HelmAdapter - Helm Registry Adapter
type HelmAdapter struct {
    Config Configuration
    Log    *logging.Logger
    Charts map[string]*repo.ChartVersion
}
```

When the broker calls `GetImageNames()` the Helm adapter will read the
`index.yaml` found at the Helm Chart Repository URL based on the [chart
repository
structure](https://github.com/kubernetes/helm/blob/master/docs/chart_repository.md#the-chart-repository-structure)
and add the latest version of each chart to the `Charts` field as well as
its name to the list of `imageNames` to be returned.

After these image names are filtered, the subsequent call to `FetchSpecs()`
will find the chart by name in the `Charts` field and create a `Spec` object
for it. For example:

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
                    Name:      "name",
                    Title:     "Release Name",
                    Type:      "string",
                    Default:   "helmrunner",
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

Most of the data required to create a useful `Spec` object is contained in
the Chart object. However, in order to retrieve the default values we will use
[Helm's Chartutil
pkg](https://github.com/kubernetes/helm/blob/master/pkg/chartutil/load.go) to
load the archive specified at `Chart.URLs[0]` and read the values.

## Phase 2 - Support Authenticated Chart Repositories

Helm Chart Repositories can be authenticated, [this
issue](https://github.com/kubernetes/helm/issues/1038) provides more
information. The broker should support authenticated chart repositories.

## Phase 3 - Chart Versions as Parameter

In phase 1, only the latest chart version will be used. The broker should
support all versions of a chart and allow the consumer to specify which version
of a chart they wish to install. The base image would need to be updated to
handle specifying a specific version at provision time.
