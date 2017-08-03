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

## Published Images
The ansible-service-broker community publishes images in [Docker Hub](https://hub.docker.com/u/ansibleplaybookbundle).
Image tags:
- **canary** - The newest source build
- **latest** - The newest release build
- **<release_number>** - The stable release of an image

## QuickStart - Running Ansible Service Broker

### Running

The following will use `oc cluster up` to bring up a cluster with Ansible Service Broker installed.
The script will run from start to finish in under 2 minutes, giving you an easy way to see the Ansible Service Broker in action.

  1. Ensure that your system is setup to run `oc cluster up`
      * Follow [these instructions](https://github.com/openshift/origin/blob/master/docs/cluster_up_down.md) if you haven't already seen success with `oc cluster up`
  1. Download [run_latest_build.sh](https://raw.githubusercontent.com/openshift/ansible-service-broker/master/scripts/run_latest_build.sh)
      ```
      wget https://raw.githubusercontent.com/openshift/ansible-service-broker/master/scripts/run_latest_build.sh
      chmod +x ./run_latest_build.sh
      ```
  1. Execute [run_latest_build.sh](https://raw.githubusercontent.com/openshift/ansible-service-broker/master/scripts/run_latest_build.sh), this will take ~90 seconds.
      ```
      ./run_latest_build.sh
      ```
  1. You now have a cluster running with the Service Catalog and Ansible Service Broker ready

### Sample Workflow

A basic test to see the capabilities of the Ansible Service Broker:
  1. Provision [Mediawiki APB](https://github.com/fusor/apb-examples/tree/master/mediawiki123-apb)
  1. Provision [PostgreSQL APB](https://github.com/fusor/apb-examples/tree/master/rhscl-postgresql-apb)
  1. Bind Mediawiki to PostgreSQL

Steps to accomplish this are:
  1. Log into OpenShift Web Console
  1. Create a new project 'demo'
  1. Select 'Mediawiki(APB)' to Provision
      * Select the 'demo' project
      * Enter a 'Mediawiki Admin User Password': 's3curepw'
      * Select 'Create'
  1. Go Back to Catalog main page
  1. Select 'PostgreSQL(APB)' to Provision
      * Select the 'demo' project
      * Leave 'PostgreSQL Password' blank, a random password will be generated
      * Chose a 'PostgreSQL Version', either version will work.
      * Select 'Create'
  1. View the 'demo' project
  1. Wait till both APBs have finished deploying and you see pods running for mediawiki and postgres
  1. Right click on the kebab menu for mediawiki
  1. Select 'Create Binding'
  1. Select the Postgres service and complete creating the Binding
  1. Redeploy mediawiki so the pod is able to consume the credentials for the database.
  1. View the route for mediawiki and verify the wiki is up and running.

## Developer Focused Information Below
### Prerequisites

[glide](https://glide.sh/) is used for dependency management. Binaries are available on the
[releases page](https://github.com/Masterminds/glide/releases).

**Packages**

Our dependencies currently require development headers for btrfs and dev-mapper.

CentOS/RHEL/Fedora (sub dnf for Fedora):

`sudo yum install device-mapper-devel btrfs-progs-devel etcd`


### Setup

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


### Targets
#### Broker Targets
* `make vendor`: Installs or updates the dependencies
* `make build`: Builds the binary from source
* `make install`: Installs the built binary.
* `make prepare-local-env`: will set up the local environemt to test and run the broker against a local deployment of [catasb](https://github.com/fusor/catasb)
* `make run`: Runs the broker with the default profile, configured via `etc/generated_local_development.yaml`
  * make run can be run without catasb and prepare-local-env by setting the `BROKER_INSECURE="true"` variable to true in `scripts/my_local_dev_vars`
* `make uninstall` Deletes the installed binary and config.yaml
  * Notes for install, run, and uninstall:
    * The default install prefix is /usr/local. Use `make build && sudo make install` to build and install.
    * Alternatively you can alter the installation directory by using PREFIX, e.g if you don't want to install somewhere that requires escalated privileges. `make build && PREFIX=~ make install`

#### CI Target
* `make ci`: Run the CI workflow that gets executed by travis, locally.
   * Workflow:
     - Provision Mediawiki
     - Provision Postgresql
     - Bind Postgresql and Mediawiki
     - Curl the Mediawiki endpoint to check for success
   * Requires:
     - Cluster
     - Service Catalog
     - Ansible-service-broker either running locally or in the cluster
     - DOCKERHUB_ORG="ansibleplaybookbundle"

#### Docker Development Build Targets
* `make build-image`: Builds a docker container of the current source

#### Docker Release Build Targets
* `make release` Builds a docker container using the latest rpm from [Copr](https://copr.fedorainfracloud.org/coprs/g/ansible-service-broker/ansible-service-broker/)
* `make push` Push the built image

#### Misc Targets
* `make clean`: Delete binaries built from source
* `make run`: Runs the broker with the default profile, configured via `etc/generated_local_development.yaml`
* `make install`: Builds the source and installs in `$GOPATH/bin`
* `make test`: Runs the test suite.
* `make vendor`: Updates the dependencies
* `make build`: Builds a docker container of the current source
* `make deploy`: Deploys the currently build container into your cluster
* `make test`: Runs the test suite.

**Note**

Scripts found in `/test` can act as manual Service Catalog requests until a larger
user scenario can be scripted.
