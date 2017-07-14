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

```yaml
name: < The name of the registry. Used by the broker to identify APB's from this registry. MUST BE DEFINED>
user: < The username for authenticating to the registry>
pass: <The password for authenticating to the registry>
org: <The namespace/organization that the image is contained in>
type: <The type of registry. The only adapters so far are mock, RHCC, and dockerhub.  MUST BE DEFINED>
URL: <The URL that is used to retrieve image information. Used extensively for RHCC while the docker hub adapter uses hardcoded URLs.>
fail_on_error: <Should this registry fail the bootstrap request if it fails. will stop the execution of other registries loading.>
white_list: <The list of regular expressions used to define which image names should be allowed through.>
black_list: <The list of regular expressions used to define which images names should neve be allowed through.>
```

For filter please look at the [filtering documentation](apb-filter-design.md).


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
```yaml
registry:
  - name: dockerhub
    type: dockerhub
    org: ansibleplaybookbundle
    user: user
    pass: password
```

### RHCC Registry
```yaml
registry:
  - name: rhcc
    type: rhcc
    url: <rhcc url>
```

### Multiple Registries Example
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
