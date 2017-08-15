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

The broker uses etcd as it's backend.

```
# Fedora
$ sudo dnf install etcd

# RHEL/CENTOS
$ sudo yum install etcd
```

To start etcd.
```
# Start/Restart etcd, this is largely dependent on your init system
# For example, systemd
$ sudo systemctl restart etcd
```

Dependency management is handled using [glide](https://glide.sh/) and
binaries are available on the [releases page](https://github.com/Masterminds/glide/releases).

### OpenShift Origin Cluster Setup

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
make vendor
```

## Setup

**Config**
A broker is configured via the `config.yaml` file. It's recommended to
copy over `etc/example-config.yaml` to `etc/ansible-service-broker/config.yaml`, and edit
as desired.

See the [Broker Configuration](docs/config.md) doc for other example
configurations.


**Note**

Scripts found in `/test` can act as manual Service Catalog requests until a larger
user scenario can be scripted.

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

- Chat with us on [IRC (Freenode): #asbroker](https://botbot.me/freenode/openshift-dev/)
- Email us at ansible-service-broker@redhat.com
- Our [YouTube Channel](https://www.youtube.com/channel/UC04eOMIMiV06_RSZPb4OOBw)
