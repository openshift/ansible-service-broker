# Ansible Service Broker Configuration Examples

The behavior of the broker is largely dictated by the broker's configuration
file loaded on startup and contains:

- [Registry Configuration](#registry-configuration)
- [DAO Configuration](#dao-configuration)
- [Log Configuration](#log-configuration)
- [OpenShift Configuration](#openshift-configuration)
- [Broker Configuration](#broker-configuration)
- [Secrets Configuration](#secrets-configuration)

## Registry Configuration

The Registry section will allow you to define the registries that the broker should look at
for APBs. All the registry config options are defined below

| field         | description                                                                                                                      | Required |
|---------------|----------------------------------------------------------------------------------------------------------------------------------|----------|
| name          | The name of the registry. Used by the broker to identify APBs from this registry.                                                |     Y    |
| auth_type     | How the broker should read the credentials                                                                                       |     N    |
| auth_name     | Name of secret/file credentials should be read from. Used when auth_type is set to `secret` or `file`                            |     N    |
| user          | The username for authenticating to the registry                                                                                  |     N    |
| pass          | The password for authenticating to the registry                                                                                  |     N    |
| org           | The namespace/organization that the image is contained in                                                                        |     N    |
| type          | The type of registry. The only adapters so far are mock, RHCC, openshift, dockerhub, and local_openshift.                        |     Y    |
| namespaces    | The list of namespaces to configure the local_openshift registry adapter with. By default a user should use `openshift`          |     N    |
| url           | The url that is used to retrieve image information. Used extensively for RHCC while the docker hub adapter uses hard-coded URLs. |     N    |
| fail_on_error | Should this registry fail the bootstrap request if it fails. will stop the execution of other registries loading.                |     N    |
| white_list    | The list of regular expressions used to define which image names should be allowed through. Must have a white list to allow APBs to be added to the catalog. The most permissive regular expression that you can use is `.*-apb$` if you would want to retrieve all APBs in a registry.                                     |     N    |
| black_list    | The list of regular expressions used to define which images names should never be allowed through.                               |     N    |
| images        | The list of images to be used with OpenShift Registry.                                                                           |     N    |

For filter please look at the [filtering documentation](filtering_apbs.md).

### Production

The Production broker configuration is designed to be pointed at a trusted
container distribution registry.

```
registry:
  - name: rhcc
    type: rhcc
    url: http://rhcc.redhat.com/api
    user: USER
    pass: PASS
  - type: local_openshift
    name: lo
    namespaces:
      - openshift
```

### Storing registry credentials in a secret/file

```
registry:
  - name: rhcc
    type: rhcc
    url: registry.access.redhat.com
    auth_type: secret
    auth_name: asb-auth-secret
```

```
registry:
  - name: rhcc
    type: rhcc
    url: registry.access.redhat.com
    auth_type: file
    auth_name: /tmp/auth-credentials
```

The associated secret should have the values `username` and `password` defined. The following is an example of using a YAML file to define credentials.
```
$ cat /tmp/auth-credentials
---
username: foo
password: foo
```

### Development

The developer configuration is primarily used by developers working on the
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
Using the dockerhub registry will allow you to load APBs from  a specific organization dockerhub. A good example is the examples [organization](https://hub.docker.com/u/ansibleplaybookbundle/).

```yaml
registry:
  - name: dockerhub
    type: dockerhub
    org: ansibleplaybookbundle
    user: user
    pass: password
    white_list:
      - ".*-apb$"
```

## Local OpenShift Registry
Using the local openshift registry will allow you to load APBs from the internal registry. The administrator can configure which namespaces they want to look for published APBs.
```yaml
registry:
  - type: local_openshift
    name: lo
    namespaces:
      - openshift
    white_list:
      - ".*-apb$"
```

### Red Hat Container Catalog (RHCC) Registry
Using the RHCC (Red Hat Container Catalog) registry will allow you to load APBs that are published to this type of [registry](https://access.redhat.com/containers).

```yaml
registry:
  - name: rhcc
    type: rhcc
    url: <rhcc url>
    white_list:
      - ".*-apb$"
```

### OpenShift Registry
Using the OpenShift registry will allow you to load APBs that are published to this type of [registry](http://www.projectatomic.io/registry/).

```yaml
registry:
  - name: openshift
    type: openshift
    user: <RH_user>
    pass: <RH_pass>
    url: <openshift_url>
    images:
      - <image_1>
      - <image_2>
    white_list:
      - ".*-apb$"
```

There is a limitation when working with the OpenShift Registry right now. We have no capability to search the registry so we require that the user configure the broker with a list of images they would like to source from for when the broker bootstraps. The image name must be the fully qualified name without the registry URL.

### Multiple Registries Example
You can use more then one registry to separate APBs into logical organizations and be able to manage them from the same broker. The main thing here is that the registries must have a unique non-empty name. If there is no unique name the service broker will fail to start with an error message alerting you to the problem.

```yaml
registry:
  - name: dockerhub
    type: dockerhub
    org: ansibleplaybookbundle
    user: user
    pass: password
    white_list:
      - ".*-apb$"
  - name: rhcc
    type: rhcc
    url: <rhcc url>
    white_list:
      - ".*-apb$"
```

## DAO Configuration

| field  | description | Required |
|--------|-------------|----------|
| etcd_host | The url of the etcd host. | Y |
| etcd_port | The port to use when communicating with `etcd_host` | Y |

## Log Configuration

| field  | description | Required |
|--------|-------------|----------|
| logfile | Where to write the broker's logs | Y |
| stdout | Write logs to stdout | Y |
| level | Level of the log output | Y |
| color | Color the logs | Y |

## OpenShift Configuration

| field  | description | Required |
|--------|-------------|----------|
| host | OpenShift host  | N |
| ca_file | Location of the certificate authority file | N |
| bearer_token_file | Location of bearer token to be used | N |
| image_pull_policy | When to pull the image | Y |
| sandbox_role | Role to give to apb sandbox environment | Y |
| keep_namespace | Always keep namespace after apb execution | N |
| keep_namespace_on_error | Keep namespace after apb execution has an error | N |

## Broker Configuration
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
ssl_cert_key|Tells the broker where to find the tls key file. If not set the [apiserver](https://github.com/kubernetes/apiserver) will attempt to create one.|""|N
ssl_cert|Tells the broker where to find the tls crt file. If not set the [apiserver](https://github.com/kubernetes/apiserver) will attempt to create one.|""|N
refresh_interval|The interval to query registries for new image specs|"600s"|N
auto_escalate|Allows the broker to escalate the permissions of a user while running the APB [read more](administration.md)|false|N
cluster_url|Sets the prefix for the url that the broker is expecting|ansible-service-broker|N

## Secrets Configuration
The secrets config section will create associations between secrets in the broker's namespace and apbs the broker runs.
The broker will use these rules to mount secrets into running apbs, allowing the user to use secrets to pass parameters
without exposing them to the catalog or users. The config section is a list where each entry has the following structure:

| field         | description                                                                                                                 | Required |
|---------------|-----------------------------------------------------------------------------------------------------------------------------|----------|
| title         | The title of the rule. This is just for display/output purposes.                                                            |     Y    |
| apb_name      | The name of the APB to associate with the specified secret. This is the fully qualified name (registry_name-image). |     Y    |
| secret        | The name of the secret to pull parameters from.                                                                             |     Y    |

You can use the script in scripts/create_broker_secret.py to create and format this configuration section.

### Secrets Example
```yaml
secrets:
- title: Database credentials
  secret: db_creds
  apb_name: dh-rhscl-postgresql-apb
```
