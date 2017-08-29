## Configuration Changes

### Goal
Create a package for creating and managing configuration values as a value store. This will be used to remove the necessity of each package to be loaded just for configuration values. The other benefit is that this package could be a more robust solution to the configuration, Allowing us to do a better job of watching for updates to the configmap and managing subscriptions to those changes. This value store should also be usable by other parts of the application and therefore will require less usage of hard coded configuration structures and will allow the addition or removal of configuration values to not create such an issue. 

### Potential Library to user: [Viper](https://github.com/spf13/viper)
Seems to be the most like what we were thinking and is used already in some large projects.
#### Pros
* Interface seems to be the same idea of what we wanted.
* Sub maps that can be retrieved
* Can read from a YAML
* Optionally can set file watchers, if we eventually want to. 
* Under active development.
* No new package needed.

#### Cons
* More permissive interface.
* Another library that we will need to the vendor and keep up to date.
* May be a little overkill for our use case.

### Work Items
* Create a specific package for managing configuration 
* Create a package mutex map to be used to store configuration values.
* Take structured yaml and create a map of map of map.... saving the key as the name of the config value
```go
m["broker"] = map["dev_broker": true...]
m["registry"] = [map["name": "dh"...], map["name": "play"...]]
```
* Create a syntax, and document, for retrieving a single configuration value from the list. I.E `config.MustGetBoolConfig("broker.dev_broker")`
* Create a public API that will give the caller multiple ways to retrieve data.
    - `Get<Type>(key)` -> Allows for typed returns so the caller does not have to worry about type casting. The default for the type will be returned if the value is not found.
    - `Put<Type>(key, value)` - Allows for programmatic setting of configuration values.
    - Ability to specify a `map[string]interface{}` as a return value so one could say `config.MustGetMapConfig("broker")` and would retrieve the map of values for the broker configuration section. 
* Update the all of the initialization code to take a map[string]interface{} to create the structures with the correct values. At this point, we can use the "kind" argument for the adapters in the registry addressing [this issue](https://github.com/openshift/ansible-service-broker/issues/49).
* Allows for the initialization code to print warnings if configuration values are not present, taking care of [this issue](https://github.com/openshift/ansible-service-broker/issues/270).
* This will also allow us to use the per package logger that we have decided on.

Go Map of generated config
```go
map[
    registry:[
        map[
            type:dockerhub 
            name:dh 
            url:https://registry.hub.docker.com 
            user:shurley 
            pass:testingboom 
            org:shurley
        ], 
        map[
            pass:testingboom 
            org:ansibleplaybookbundle 
            type:dockerhub 
            name:play 
            url:https://registry.hub.docker.com 
            user:shurley
        ]
    ] 
    dao:map[
        etcd_host:asb-etcd-ansible-service-broker.172.17.0.1.nip.io 
        etcd_port:80
    ] 
    log:map[
        logfile:/tmp/ansible-service-broker-asb.log 
        stdout:true 
        level:debug 
        color:true
    ] 
    openshift:map[
        host:172.17.0.1 
        bearer_token_file:<nil> 
        ca_file:<nil> 
        image_pull_policy:<nil>
    ] 
    broker:map[
        launch_apb_on_bind:false 
        bootstrap_on_startup:true 
        recovery:true 
        output_request:true 
        ssl_cert_key:/var/run/secrets/kubernetes.io/serviceaccount/tls.key 
        ssl_cert:/var/run/secrets/kubernetes.io/serviceaccount/tls.crt 
        refresh_interval:600s 
        dev_broker:true
    ]
]
```