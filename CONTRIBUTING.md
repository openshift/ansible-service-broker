Contributing to Ansible Service Broker
======================================

This article explains how to set up a development environment and get involved
with Ansible Service Broker development.

# Ansible Service Broker Development

Before anything else, [fork](https://help.github.com/articles/fork-a-repo) the
[ansible service broker project](https://github.com/openshift/ansible-service-broker).

## Develop locally on your host

### Installing Prerequisites

**Fedora/RHEL/CENTOS**

You are going to need git, docker, golang, and make.

```
# Fedora
$ sudo dnf install git docker-latest golang make

# RHEL/CENTOS
$ sudo yum install git docker-latest golang make
```

Our dependencies currently require development headers for btrfs and dev-mapper.

```
# Fedora
$ sudo dnf install device-mapper-devel btrfs-progs-devel

# RHEL/CENTOS
$ sudo yum install device-mapper-devel btrfs-progs-devel
```

Dependency management is handled using [glide](https://glide.sh/) and
binaries are available on the [releases page](https://github.com/Masterminds/glide/releases).

### Install OpenShift Origin Client

The broker relies heavily on services provided by OpenShift Origin and it's tools.
You will need to setup your system for local
[OpenShift Origin Cluster Management](https://github.com/openshift/origin/blob/master/docs/cluster_up_down.md).
Your OpenShift Client binary (`oc`) must be `>=` [v3.6.0-rc.0](https://github.com/openshift/origin/releases/tag/v3.6.0-rc.0).

### Clone the Repository

At this point you should have dependencies installed (git, docker, golang, etc.).
[Golang.org has excellent documentation](https://golang.org/doc/code.html) to get you started
developing in Go.

```
mkdir -p $GOPATH/src/github.com/openshift
git clone https://github.com/openshift/ansible-service-broker.git $GOPATH/src/github.com/openshift/ansible-service-broker
cd $GOPATH/src/github.com/openshift/ansible-service-broker
# Install or update project dependencies
make vendor
```

### Deploy an OpenShift Origin Cluster with the Ansible Service Broker

At some point, when contributing to the Ansible Service Broker project (or just checking it out),
you are going to want to have a cluster to integrate with. There are two primary methods for
starting an OpenShift Origin Cluster with the Ansible Service Broker:

1. Use [fusor/catasb to gain more control over the environment](https://github.com/fusor/catasb).
1. Use [run_latest_build.sh](../scripts/run_latest_build.sh) to create a cluster with the service
   catalog enabled and the Ansible Service Broker running. We covered this before in
   ["Getting Started with the Ansible Service Broker"](README.md#getting-started-with-the-ansible-service-broker).

   ```
   cd $GOPATH/src/github.com/openshift/ansible-service-broker
   bash -x scripts/run_latest_build.sh
   ```

### Building the Ansible Service Broker from Source

Building the Ansible Service Broker is as simple as running `make build` from the root
of the project:

```
cd $GOPATH/src/github.com/openshift/ansible-service-broker
make build
```

Now you can [run your broker locally](#run-your-broker-locally) with `make run` or
[package your broker using docker](#package-your-broker-using-docker) with `make build-image`.

#### Package Your Broker Using Docker

You can also package your built Ansible Service Broker binary into a Docker image through
`make build-image`.

```
cd $GOPATH/src/github.com/openshift/ansible-service-broker
export ORG=${YOUR_DOCKERHUB_ID}
export TAG=my_custom_tag
make build-image
```

Once you have a built Docker image, you can [deploy broker from image](#deploy-broker-from-image)
with `make deploy` **BUT** only after you have pushed your image to the registry
(ie. `docker push ${REGISTRY}/${ORG}/origin-ansible-service-broker:${TAG}`).

### Run Your Broker Locally

1. Build the Ansible Service Broker executable with the command: ```make build```
1. Ensure you have a local oc cluster up environment running with [fusor/catasb 'master' branch](https://github.com/fusor/catasb)
1. ```cp scripts/my_local_dev_vars.example scripts/my_local_dev_vars```
1. Edit 'scripts/my_local_dev_vars'
1. Prepare your local development environment by running: ```make prepare-local-env```
    * This will remove the running 'asb' pod and replace the endpoint of the asb route with the locally running broker executable.
    * You __must__ rerun this whenever you have reset your cluster.
    * It is safe to run this multiple times
1. Run the broker executable locally:  ```make run```
    * Use the cluster as you normally would for testing.
    * When you want to make a change and rebuild the broker.  CTRL-C, rebuild and re-run

### Deploy Broker From Image

1. Build the Ansible Service Broker executable with the command: ```make build ```
1. Build a development image:  ```make build-image BROKER_IMAGE_NAME=asb-dev TAG=local```
1. Deploy from [templates/deploy-ansible-service-broker.template.yaml](https://github.com/openshift/ansible-service-broker/blob/master/templates/deploy-ansible-service-broker.template.yaml) setting the parameter "-p BROKER_IMAGE=asb-dev:local" to your local image name/tag

# Submitting changes to Ansible Service Broker

## Before making a pull request

There are a few things you should keep in mind before creating a PR.
- New code should have no less than corresponding unit tests (see [broker.go](pkg/broker/broker.go)
  and [broker_test.go](pkg/broker/broker_test.go) as an example.
- Must run `make check` (and it should pass) before you create a PR. Changes that
  do not pass `make check` will not be reviewed until they pass.
- Have a method that demonstrates what this PR accomplishes. As an example,
  [PR 250](https://github.com/openshift/ansible-service-broker/pull/250) clearly
  provides: a test to demonstrate the changes AND the output of those tests.

## Making a pull request

Make a [pull request](https://help.github.com/articles/using-pull-requests) (PR).
See the [OWNERS](OWNERS) for a list of reviewers/approvers.

Use [WIP] at the beginning of the title (ie. [WIP] Add feature to the broker)
to mark a PR as a Work in Progress.

Upon successful review, someone will approve the PR in the review thread.
A reviewer with merge power may merge the PR with or without an approval from
someone else. Or they may wait a business day to get further feedback from other
reviewers.

## Major Features
The ansible-service-broker community uses a proposal process when introducing a
major feature in order to encourage collaboration and building the best solution.

Major features are things that take about 2 weeks of development and introduce
disruptive changes to the code base.

Start the proposal process by reviewing the [proposal template](docs/proposals/proposal-template.md).
Use this document to guide how to write a proposal. Then, submit it as a pull
request where the community will review the plan.  The proposal process will require
two approvals from the community before merging.

# Roadmap

See what work is in progress, upcoming, and planned via our [Trello Board](https://trello.com/b/50JhiC5v/ansible-service-broker).

# Download images from Docker Hub

Docker images are available on [Docker Hub](https://hub.docker.com/r/ansibleplaybookbundle/origin-ansible-service-broker/)

* [Docker Hub published APBs](https://hub.docker.com/u/ansibleplaybookbundle/)

Image [tags](https://hub.docker.com/r/ansibleplaybookbundle/origin-ansible-service-broker/tags):
- **canary** - The newest source build
- **latest** - The newest release build
- **<release_number>** - The stable release of an image

# Stay in Touch

- IRC: Join the conversation on [Freenode: #asbroker](https://botbot.me/freenode/openshift-dev/)
- Email: Subscribe to the Ansible Service Broker's [mailing list](https://www.redhat.com/mailman/listinfo/ansible-service-broker)
- Our [YouTube Channel](https://www.youtube.com/channel/UC04eOMIMiV06_RSZPb4OOBw)
