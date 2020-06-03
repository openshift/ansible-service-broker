# Using as a library

The Automation Broker implements the Open Service Broker API. As a broker it
provides quite a bit of behavior that might want to be used by other broker
authors, for example, asynchronous job support, etc.

## Vendoring

Here is an excerpt from `Gopkg.toml` used by [golang/dep](https://golang.github.io/dep/)
that adds the Automation Broker as a dependency. We can use any of the types
supported by the `dep` tool: branches, tags, or versions. Here we are using the master
branch.

```toml
[[constraint]]
  branch = "master"
  name = "github.com/openshift/ansible-service-broker"
```

## Extending

Once we have the dependency installed in our project, how do we use it? The
Automation Broker will do most, if not all, of the heavy lifting, but allows
the custom registry adapters to be supplied. We can also have our own
configuration, custom parameters, just about anything we want since we will
have our own `main.go`.

### CreateApp

At the most basic, we need to import the app so we can create a new instance
and start it.

```golang
package main

import (
    "os"

    "github.com/openshift/ansible-service-broker/pkg/app"
)

func main() {
    var args app.Args
    var err error

    // capture the args, this is also a way we can
    // handle our own arguments.
    if args, err = app.CreateArgs(); err != nil {
        os.Exit(1)
    }

    // Create app passing in args
    app := app.CreateApp(args, nil)
    app.Start()
}
```

### Custom Registry

Supplying our own custom registry is probably the most compelling feature of
using the Automation Broker as a library.

Starting with the above main example, we'll want to pass our custom registry to
the broker. Let's assume we have a custom registry that reads a set of bundles
from a file, FileAdapter. We will create a configuration for the Automation
Broker to use for the registry adapter. We will also create a new instance of
the registry adapter to pass in at startup.

```golang
import (
    //...
    "github.com/automationbroker/bundle-lib/registries"
    "github.com/jmrodri/samplebroker/pkg/registries/adapters"
)

    //...
    // To add our custom registries, define an entry in this array
    regs := []registries.Registry{}

    // Create config
    c := registries.Config{
        URL:        "",
        User:       "",
        Pass:       "",
        Org:        "jmrodri",
        Tag:        "latest",
        Type:       "file",
        Name:       "foo",
        Images:     []string{"hello-world-db-apb"},
        Namespaces: []string{"openshift"},
        Fail:       false,
        AllowList:  []string{".*-apb$"},
        DenyList:   []string{},
        AuthType:   "",
        AuthName:   "",
        Runner:     "",
    }

    // Instantiate our custom registry adapter
    fadapter := adapters.FileAdapter{Name: "sampleadapter"}

    // Now create a new registry using our customer adapter
    // ignoring the errors for this example
    reg, _ := registries.NewCustomRegistry(c, fadapter, "openshift")

    // Add to array
    regs = append(regs, reg)

    //...

```

[`FileAdapter`](https://github.com/jmrodri/samplebroker/blob/master/pkg/registries/adapters/file_adapter.go)
is simple registry adapter that implements the `Adapter` interface
from the `bundle-lib` project. `FileAdapter` contains an APB yaml string that it
returns when `FetchSpecs` is called.

```golang
// Adapter - Adapter will wrap the methods that a registry needs to
// fully manage images.
type Adapter interface {
    // RegistryName will return the registry prefix for the adapter.
    // Example is docker.io for the dockerhub adapter.
    RegistryName() string
    // GetImageNames will return all the image names for the adapter configuration.
    GetImageNames() ([]string, error)
    // FetchSpecs will retrieve all the specs for the list of images names.
    FetchSpecs([]string) ([]*apb.Spec, error)
}

```

### Putting things together

We have a custom registry and we know how to create and start the Automation
Broker. Let's put them together to see what our custom broker would look like:

```golang
package main

import (
    "fmt"
    "os"

    "github.com/automationbroker/bundle-lib/registries"
    "github.com/jmrodri/samplebroker/pkg/registries/adapters"
    "github.com/openshift/ansible-service-broker/pkg/app"
)

func main() {

    var args app.Args
    var err error

    // capture the args, this is also a way we can
    // handle our own arguments.
    if args, err = app.CreateArgs(); err != nil {
        os.Exit(1)
    }

    // To add our custom registries, define an entry in this array
    regs := []registries.Registry{}

    // Create config
    c := registries.Config{
        URL:        "",
        User:       "",
        Pass:       "",
        Org:        "jmrodri",
        Tag:        "latest",
        Type:       "file", // bundle-lib registry.go needs to change for this
        Name:       "foo",
        Images:     []string{"hello-world-db-apb"},
        Namespaces: []string{"openshift"},
        Fail:       false,
        AllowList:  []string{".*-apb$"},
        DenyList:   []string{},
        AuthType:   "",
        AuthName:   "",
        Runner:     "",
    }

    // Instantiate our custom registry adapter
    fadapter := adapters.FileAdapter{Name: "foobar"}

    // Now create a new registry using our customer adapter
    // ignoring the errors for this example
    reg, err := registries.NewCustomRegistry(c, fadapter, "openshift")
    if err != nil {
        fmt.Printf(
            "Failed to initialize foo Registry err - %v \n", err)
        os.Exit(1)
    }

    // Add to array
    regs = append(regs, reg)

    // CreateApp passing in the args and registries
    // NOTE: instead of passing nil, we now pass regs
    app := app.CreateApp(args, regs)
    app.Start()
}
```

We capture the args, create a configuration, a registry adapter instance, new
custom registry instance, then a new broker instance.
