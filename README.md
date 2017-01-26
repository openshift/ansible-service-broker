# Ansible Service Broker

[![Codacy Badge](https://api.codacy.com/project/badge/Grade/9b0f6bca11c040d2ad7894d97353cd37)](https://www.codacy.com/app/eriknelson/ansible-service-broker?utm_source=github.com&utm_medium=referral&utm_content=fusor/ansible-service-broker&utm_campaign=badger)

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
