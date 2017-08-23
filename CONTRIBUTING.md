Contributing to Ansible Service Broker
======================================

This article explains how to set up a development environment and get involved
with Ansible Service Broker development.

Before anything else, [fork](https://help.github.com/articles/fork-a-repo) the
[ansible service broker project](https://github.com/openshift/ansible-service-broker).


# Set Up Development Environment

## Install Prerequisites

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

## Install OpenShift Origin Client

The broker relies heavily on services provided by OpenShift Origin and it's tools.
You will need to setup your system for local
[OpenShift Origin Cluster Management](https://github.com/openshift/origin/blob/master/docs/cluster_up_down.md).
Your OpenShift Client binary (`oc`) must be `>=` [v3.6.0-rc.0](https://github.com/openshift/origin/releases/tag/v3.6.0-rc.0).

## Clone the Repository

At this point you should have dependencies installed (git, docker, golang, etc.).
[Golang.org has excellent documentation](https://golang.org/doc/code.html) to get you started
developing in Go.

Next, you will want to [clone the repository](https://help.github.com/articles/cloning-a-repository/).

```
mkdir -p $GOPATH/src/github.com/openshift
git clone https://github.com/openshift/ansible-service-broker.git $GOPATH/src/github.com/openshift/ansible-service-broker
cd $GOPATH/src/github.com/openshift/ansible-service-broker
# Install or update project dependencies
make vendor
```

## Deploy an OpenShift Origin Cluster with the Ansible Service Broker

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

**Congratulations**, you should now have a running OpenShift Origin Cluster with the Service Catalog
and Ansible Service Broker running inside. If you are having issues, have a look at
[fusor/catasb](https://github.com/fusor/catasb) for some troubleshooting steps.

### Reset OpenShift Origin Cluster

**NOTE** if you ever want to reset your environment, look at [fusor/catasb](https://github.com/fusor/catasb)
because that project makes it as easy as `./reset_environment.sh`, **or** take the cluster down and `run_latest_build.sh`
again:

```
oc cluster down
cd $GOPATH/src/github.com/openshift/ansible-service-broker
./scripts/run_latest_build.sh
```


# Build the Ansible Service Broker

Building the Ansible Service Broker is as simple as running `make build` from the root
of the project:

```
cd $GOPATH/src/github.com/openshift/ansible-service-broker
make build
```

Now you can [run your broker locally](#run-your-broker-locally) with `make run` or
[package your broker using docker](#package-your-broker-using-docker) with `make build-image`.

## Package Your Broker Using Docker

You can also package your built Ansible Service Broker binary into a Docker image through
`make build-image`.

```
cd $GOPATH/src/github.com/openshift/ansible-service-broker
export ORG=${YOUR_DOCKERHUB_ID}
export TAG=my_custom_tag
make build-image
```

If you have a look at the [Dockerfile](build/Dockerfile-canary) used when building a Docker
image with the broker, there are a few things worth pointing out:

- The `broker` image is copied to `/usr/bin/asbd`.
- The home directory in the image is `/opt/ansibleservicebroker`.
- The directory `/etc/ansible-service-broker` is where, by default, the broker's configuration
  file is mounted when deployed in OpenShift. Looking at the
  [deploy-ansible-service-broker template](templates/deploy-ansible-service-broker.template.yaml)
  you can find the `broker-config` included  as `config-volume` in the `asb` container and mounted at
  `/etc/ansible-service-broker`.
- The entrypoint to the image is [entrypoint.sh](build/entrypoint.sh). This script is what starts
  the broker when deployed as a container.

Once you have a built Docker image, you can [deploy broker from image](#deploy-broker-from-image)
with `make deploy` **BUT** only after you have pushed your image to the registry
(ie. `docker push ${REGISTRY}/${ORG}/origin-ansible-service-broker:${TAG}`).


# Run Your Broker

Once you have [built the Ansible Service Broker](#build-the-ansible-service-broker), there
are a couple of options for running, testing, and debugging the broker. You can
[run the broker binary locally](#run-your-broker-locally) **or**, assuming you
[built a Docker image](#package-your-broker-using-docker), you can
[deploy your broker](#deploy-broker-from-image) into a running cluster.

## Run Your Broker Locally

If you are interested in rapidly iterating over your broker changes or want to easily keep
tabs on what the broker is doing, running the broker locally is an excellent idea.

Before attempting to run the broker locally, you will need to have deployed an
[OpenShift Origin Cluster with the Ansible Service Broker](#deploy_an_openshift_origin_cluster_with_the_ansible_service_broker).
You can verify this by running `oc get all --all-namespaces` and looking for the `ansible-service-broker`
`service-catalog` running in the cluster.

**Configuration**

Next, you will need configure your local development variables:

```
cp scripts/my_local_dev_vars.example scripts/my_local_dev_vars
```

Now you can modify `scripts/my_local_dev_vars` with things like your `DOCKERHUB_USERNAME`
or use an insecure broker with `BROKER_INSECURE="true"`.

**Prepare Local Environment**

Running `make prepare-local-env` will do several things on your behalf:

* Remove the running `ansible-service-broker` from the cluster, leaving only
  etcd running in the namespace.
* Modify the `asb` service to point where our locally running broker will be.
* Generate a configuration file for the broker (`etc/generated_local_development.yaml`)

This would be a good time to have a look at the broker's configuration and make
changes. Have a look at [the broker configuration examples](docs/config.md).
By default, the registry section of the config will look something like:

```
registry:
  - type: dockerhub
    name: dh
    url: https://registry.hub.docker.com
    user: changeme
    pass: changeme
    org: ansibleplaybookbundle
```

It may server your purposes to append additional registries, for example, if
you wanted to develop APB's in your organization while still seeing those in
the ansibleplaybookbundle organization. That may look something like:

```
registry:
  - type: dockerhub
    name: dh
    url: https://registry.hub.docker.com
    user: changeme
    pass: changeme
    org: ansibleplaybookbundle
  - type: dockerhub
    name: example
    url: https://registry.hub.docker.com
    user: changeme
    pass: changeme
    org: example
```

**Start the Broker**

With the environment prepared, running your broker is as simple as running
`make run`.

## Deploy Broker From Image

Once you have [built a Docker image with your broker binary](#package-your-broker-using-docker),
you can run your broker inside the cluster. Now is an excellent time to point out that
[fusor/catasb](https://github.com/fusor/catasb) gives you the ability to deploy
an OpenShift Origin Cluster with the Service Catalog and **your** broker image. If that
still feels like overkill you can use `make deploy` to replace the running broker with
one of your choosing.

```
cd $GOPATH/src/github.com/openshift/ansible-service-broker
export ORG=${YOUR_DOCKERHUB_ID}
export TAG=my_custom_tag
make deploy
```

The `deploy` target runs [scripts/deploy.sh](scripts/deploy.sh) to:

- Remove the existing deployment of the Ansible Service Broker.
- Create any broker clusterrolebindings from a previous run (they will be
  created again when the template is processed).
- Process the [deploy-ansible-service-broker template](templates/deploy-ansible-service-broker.template.yaml)
  that we introduced before when we [packaged your broker using docker](#package-your-broker-using-docker).
- Create the objects from the template in OpenShift Origin.


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
