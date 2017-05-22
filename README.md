# Ansible Service Broker
[![Build Status](https://travis-ci.org/fusor/ansible-service-broker.svg?branch=master)](https://travis-ci.org/fusor/ansible-service-broker)
[![Code Climate](https://codeclimate.com/github/fusor/ansible-service-broker/badges/gpa.svg)](https://codeclimate.com/github/fusor/ansible-service-broker)
[![Issue Count](https://codeclimate.com/github/fusor/ansible-service-broker/badges/issue_count.svg)](https://codeclimate.com/github/fusor/ansible-service-broker)

Ansible Service Broker is an implementation of the [Open Service Broker API](https://github.com/openservicebrokerapi/servicebroker) that will manage applications defined by [Ansible Playbook Bundles](https://github.com/fusor/ansible-playbook-bundle).  


An Ansible Playbook Bundle (APB) is a new method for defining and distributing container applications in OpenShift consisting of a bundle of Ansible Playbooks built into a container with an Ansible runtime.

Read more about the Ansible Service Broker and Ansible Playbook Bundles in this [introduction](docs/introduction.md).

**NOTE**: this broker will be based on the [open-service-broker-sdk](https://github.com/openshift/open-service-broker-sdk) project in the future.

## Project Related Links
* Email:  ansible-service-broker@redhat.com
* IRC (Freenode): #asbroker
* [Trello](https://trello.com/b/50JhiC5v/ansible-service-broker)
* Github:
    * [ansible service broker](https://github.com/fusor/ansible-service-broker)
    * [ansible playbook bundle](https://github.com/fusor/ansible-playbook-bundle)
* [Demo environment with oc cluster up](https://github.com/fusor/catasb)
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

`sudo yum install device-mapper-devel btrfs-progs-devel jq etcd`

## Setup

```
sudo /sbin/service etcd restart # start etcd
mkdir -p $GOPATH/src/github.com/fusor
git clone https://github.com/fusor/ansible-service-broker.git $GOPATH/src/github.com/fusor/ansible-service-broker
cd $GOPATH/src/github.com/fusor/ansible-service-broker && glide install
```

**Config**

A broker is configured via a `$ENV.config.yaml` file. Example files can be
found under `etc/`. It's recommended to simply copy over `etc/ex.dev.config.yaml`
to `etc/dev.config.yaml`, and edit as desired. `scripts/runbroker.sh` should
handle providing the location to this file. Of course, this can be customized
or the configuration file can be specified by cli args as well.

## Targets

* `make run`: Runs the broker with the default profile, configured via `/etc/dev.config.yaml`
* `make run-mock-registry`: Mock registry. Entirely separate binary.
* `make test`: Runs the test suite.

**Note**

Scripts found in `/test` can act as manual Service Catalog requests until a larger
user scenario can be scripted.
