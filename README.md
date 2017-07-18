# Ansible Service Broker
[![Build Status](https://travis-ci.org/openshift/ansible-service-broker.svg?branch=master)](https://travis-ci.org/openshift/ansible-service-broker)
[![Go_Report_Card](https://goreportcard.com/badge/github.com/openshift/ansible-service-broker)](https://goreportcard.com/report/github.com/openshift/ansible-service-broker)

Ansible Service Broker is an implementation of the [Open Service Broker API](https://github.com/openservicebrokerapi/servicebroker) that will manage applications defined by [Ansible Playbook Bundles](https://github.com/fusor/ansible-playbook-bundle).


An Ansible Playbook Bundle (APB) is a new method for defining and distributing container applications in OpenShift consisting of a bundle of Ansible Playbooks built into a container with an Ansible runtime.

Read more about the Ansible Service Broker and Ansible Playbook Bundles in this [introduction](docs/introduction.md).

**NOTE**: this broker will be based on the [open-service-broker-sdk](https://github.com/openshift/open-service-broker-sdk) project in the future.

## Project Related Links
* Email:  ansible-service-broker@redhat.com
* IRC (Freenode): #asbroker
* [Trello](https://trello.com/b/50JhiC5v/ansible-service-broker)
* Github:
    * [ansible service broker](https://github.com/openshift/ansible-service-broker)
    * [ansible playbook bundle](https://github.com/fusor/ansible-playbook-bundle)
* [Demo environment with oc cluster up - Catalog & Ansible Service Broker 'catasb'](https://github.com/fusor/catasb)
* [Library of example APBs](https://github.com/fusor/apb-examples)
    * ManageIQ
    * PostgreSQL
    * Wordpress
    * Hello-World
* [Red Hat Summit 2017](https://www.youtube.com/playlist?list=PLZ7osZ-J70IaVc0NVyLs7tLO1hbhBdxHe)
  * [Keynote Demo](https://youtu.be/8MCbJmZQM9c?list=PLEGSLwUsxfEh4TE2GDU4oygCB-tmShkSn&t=4732)
  * [Amazon Web Services deployed into OpenShift via Ansible Service Broker](https://www.youtube.com/watch?v=EKo3khfmhi8&index=2&list=PLZ7osZ-J70IaVc0NVyLs7tLO1hbhBdxHe)
  * [Presentation OpenService Broker API + Ansible Service Broker/Ansible Playbook Bundles](https://www.youtube.com/watch?v=BaPMFZZ5lsc&index=1&list=PLZ7osZ-J70IaVc0NVyLs7tLO1hbhBdxHe)
* [YouTube Channel](https://www.youtube.com/channel/UC04eOMIMiV06_RSZPb4OOBw):
    * [Using the Service Catalog to Bind a PostgreSQL APB to a Python Web App](https://www.youtube.com/watch?v=xmd52NhEjCk)
    * [Service Catalog deploying ManageIQ APB onto OpenShift](https://www.youtube.com/watch?v=J6rDssVEZuQ)
    * [OpenShift Commons Briefing #74: Deploying Multi-Container Applications with Ansible Service Broker](https://www.youtube.com/watch?v=kDJveLN5UOs&list=PLZ7osZ-J70IYBvqTdHt6Lt91I46k-FJI2&index=1)
* [Docker hub published APBs](https://hub.docker.com/u/ansibleplaybookbundle/)

## Documentation
* [Ansible Service Broker - Introduction](docs/introduction.md)
* [Ansible Service Broker - Design](docs/design.md)
* [Other Documentation](docs/README.md)

## Prerequisites

[glide](https://glide.sh/) is used for dependency management. Binaries are available on the
[releases page](https://github.com/Masterminds/glide/releases).

**Packages**

Our dependencies currently require development headers for btrfs and dev-mapper.

CentOS/RHEL/Fedora (sub dnf for Fedora):

`sudo yum install device-mapper-devel btrfs-progs-devel etcd`

## Setup

```
sudo /sbin/service etcd restart # start etcd
mkdir -p $GOPATH/src/github.com/openshift
git clone https://github.com/openshift/ansible-service-broker.git $GOPATH/src/github.com/openshift/ansible-service-broker
cd $GOPATH/src/github.com/openshift/ansible-service-broker
make vendor
```

**Config**
A broker is configured via the `config.yaml` file. It's recommended to
copy over `etc/example-config.yaml` to `etc/ansible-service-broker/config.yaml`, and edit
as desired.

See the [Broker Configuration](docs/config.md) doc for other example
configurations.

## Published Images
The ansible-service-broker community publishes images in (dockerhub)[https://hub.docker.com/u/ansibleplaybookbundle].
Image tags:
- **canary** - The newest source build
- **latest** - The newest release build
- **<release_number>** - The stable release of an image

## Targets
### Broker Targets
* `make vendor`: Installs or updates the dependencies
* `make build`: Builds the binary from source
* `make install`: Installs the built binary.
* `make run`: Runs the broker with the default profile, configured via `/etc/dev.config.yaml`
* `make uninstall` Deletes the installed binary and config.yaml
  * Notes for install, run, and uninstall:
    * The default install prefix is /usr/local. Use `make build && sudo make install` to build and install.
    * Alternatively you can alter the installation directory by using PREFIX, e.g if you don't want to install somewhere that requires escalated privileges. `make build && PREFIX=~ make install`

### Docker Development Build Targets
* `make build-image`: Builds a docker container of the current source

### Docker Release Build Targets
* `make release` Builds a docker container using the latest rpm from [Copr](https://copr.fedorainfracloud.org/coprs/g/ansible-service-broker/ansible-service-broker/)
* `make push` Push the built image

### Misc Targets
* `make clean`: Delete binaries built from source
* `make run`: Runs the broker with the default profile, configured via `/etc/config.yaml`
* `make install`: Builds the source and installs in `$GOPATH/bin`
* `make test`: Runs the test suite.
* `make vendor`: Updates the dependencies
* `make build`: Builds a docker container of the current source
* `make deploy`: Deploys the currently build container into your cluster
* `make test`: Runs the test suite.

**Note**

Scripts found in `/test` can act as manual Service Catalog requests until a larger
user scenario can be scripted.
