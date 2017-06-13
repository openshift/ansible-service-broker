# Ansible Service Broker Configuration Examples

## Production

The Production broker configuration is designed to be pointed at a trusted
container distribution registry.

```
registry:
  name: rhcc
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

## Mock Registry

Using a Mock registry is useful for reading local APB specs. Instead of going
out to a registry to search for image specs, use a list of local specs. Set the
name of the registry to 'mock' to use the Mock registry.

```
registry:
  name: mock
```
