# Ansible Service Broker

Work in progress.

## Usage

In terminal 1:

```bash
go get -u github.com/fusor/ansible-service-broker
$GOPATH/bin/broker
```

In terminal 2:

```bash
cd $GOPATH/src/github.com/fusor/ansible-service-broker
test/catalog.sh
test/provision.sh
test/bind.sh
test/unbind.sh
test/deprovision.sh
```

## Links

- [OpenShift Origin](https://github.com/openshift/origin)
- [Open Service Broker API](https://github.com/openservicebrokerapi/servicebroker)
