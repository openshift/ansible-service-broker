# Ansible Service Broker Configuration Examples

## Production

The Production broker configuration is designed to be pointed at a trusted
container distribution registry.

```
registry:
  - name: rhcc
    url: http://rhcc.redhat.com/api
    user: USER
    pass: PASS
```

## Development

The Developer configuration is the primarily used by developers working on the
broker. Set the registry name to 'dev' and 'devbroker' field to 'true' to enable
developer settings.

```
registry:
  name: dev
```

```
broker:
  devbroker: true
```
## Registry Configuration

The Registry section will allow you to define the registries that the broker should look at
for APB's. All the registry config options are defined below

| field         | description                                                                                                                     | Required |
|---------------|---------------------------------------------------------------------------------------------------------------------------------|----------|
| name          | The name of the registry. Used by the broker to identify APB's from this registry.                                              |     Y    |
| user          | The username for authenticating to the registry                                                                                 |     N    |
| pass          | The password for authenticating to the registry                                                                                 |     N    |
| org           | The namespace/organization that the image is contained in                                                                       |     N    |
| type          | The type of registry. The only adapters so far are mock, RHCC, and dockerhub.                                                   |     Y    |
| url           | The url that is used to retrieve image information. Used extensively for RHCC while the docker hub adapter uses hardcoded urls. |     N    |
| fail_on_error | Should this registry fail the bootstrap request if it fails. will stop the execution of other registries loading.               |     N    |
| white_list    | The list of regular expressions used to define which image names should be allowed through.                                     |     N    |
| black_list    | The list of regular expressions used to define which images names should neve be allowed through.                               |     N    |

For filter please look at the [filtering documentation](filtering_apbs.md).

### Mock Registry
Using a Mock registry is useful for reading local APB specs. Instead of going
out to a registry to search for image specs, use a list of local specs. Set the
name of the registry to 'mock' to use the Mock registry.

```
registry:
  - name: mock
    type: mock
```

### Dockerhub Registry
Using the dockerhub registry will allow you to load APB's from  a specific organization dockerhub. A good example is the examples [organization](https://hub.docker.com/u/ansibleplaybookbundle/).

```yaml
registry:
  - name: dockerhub
    type: dockerhub
    org: ansibleplaybookbundle
    user: user
    pass: password
```

### Red Hat Container Catalog (RHCC) Registry
Using the RHCC (Red Hat Container Catalog) registry will allow you to load APB's that are published to this type of [registry](https://access.redhat.com/containers).

```yaml
registry:
  - name: rhcc
    type: rhcc
    url: <rhcc url>
```

### Multiple Registries Example
You can use more then one registry to separate apbs into logical organizations and be able to manage them from the same broker. The main thing here is that the registries must have a unique non-empty name. If there is no unique name the service broker will fail to start with an error message alerting you to the problem.

```yaml
registry:
  - name: dockerhub
    type: dockerhub
    org: ansibleplaybookbundle
    user: user
    pass: password
  - name: rhcc
    type: rhcc
    url: <rhcc url>
```

## Broker Config
The broker config section will tell the broker what functionality should be enabled
and disabled. It will also tell the broker where to find files on disk that will
enable the full functionality.

*Note: with the absence of async bind, setting launch_apb_on_bind to true can cause the bind action to timeout and will span a retry. The broker will handle with with 409 Conflicts because it is the same bind request with different parameters.*

**field**|**description**|**default value**|**required**
:-----:|:-----:|:-----:|:-----:
dev_broker|Allow development routes to be accessible|false|N
launch_apb_on_bind|Allow bind be be no op|false|N
bootstrap_on_startup|Allow the broker attempt to bootstrap itself on start up. Will retrieve the APBs from configured registries|false|N
recovery|Allow the broker to attempt to recover itself by dealing with pending jobs noted in etcd|false|N
output_request|Allow the broker to output the requests to the log file as they come in for easier debugging|false|N
ssl_cert_key|Tells the broker where to find the tls key file. Can be ignored if using `--insecure` option but otherwise is required|""|N
ssl_cert|Tells the broker where to find the tls crt file. Can be ignored if using the `--insecure` option but otherwise is required|""|N
refresh_interval|The interval to query registries for new image specs|"600s"|N
