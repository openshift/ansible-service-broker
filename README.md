Ansible Service Broker
======================

[![Build Status](https://travis-ci.org/openshift/ansible-service-broker.svg?branch=master)](https://travis-ci.org/openshift/ansible-service-broker)
[![Go_Report_Card](https://goreportcard.com/badge/github.com/openshift/ansible-service-broker)](https://goreportcard.com/report/github.com/openshift/ansible-service-broker)
[![Coverage Status](https://coveralls.io/repos/github/openshift/ansible-service-broker/badge.svg?branch=coveralls)](https://coveralls.io/github/openshift/ansible-service-broker?branch=coveralls)
[![Join the chat at freenode:asbroker](https://img.shields.io/badge/irc-freenode%3A%20%23asbroker-blue.svg)](http://webchat.freenode.net/?channels=%23asbroker)
[![Licensed under Apache License version 2.0](https://img.shields.io/github/license/openshift/origin.svg?maxAge=2592000)](https://www.apache.org/licenses/LICENSE-2.0)

Ansible Service Broker is an implementation of the [Open Service Broker API](https://github.com/openservicebrokerapi/servicebroker)
that manages applications defined in [Ansible Playbook Bundles](https://github.com/fusor/ansible-playbook-bundle).
Ansible Playbook Bundles (APB) are a method of defining applications via a collection of Ansible Playbooks built into a container
with an Ansible runtime with the playbooks corresponding to a type of request specified in the
[Open Service Broker API Specification](https://github.com/openservicebrokerapi/servicebroker/blob/master/spec.md#api-overview).

Check out the [Keynote Demo from Red Hat Summit 2017](https://youtu.be/8MCbJmZQM9c?list=PLEGSLwUsxfEh4TE2GDU4oygCB-tmShkSn&t=4732)

**Features**

- Easily define, distribute, and provision microservice(s), like [RocketChat](https://github.com/fusor/apb-examples/tree/master/rocketchat-apb)
  and [PostgreSQL](https://github.com/ansibleplaybookbundle/postgresql-apb), via ansible playbooks packaged in
  [Ansible Playbook Bundles](https://github.com/fusor/ansible-playbook-bundle).
- Easily bind microservice(s) provisioned through [Ansible Playbook Bundles](https://github.com/fusor/ansible-playbook-bundle),
  for example: [Using the Service Catalog to Bind a PostgreSQL APB to a Python Web App](https://www.youtube.com/watch?v=xmd52NhEjCk).

**Learn More:**

- [Documentation](docs/README.md)
- Our [Trello Board](https://trello.com/b/50JhiC5v/ansible-service-broker)
- Chat with us on [IRC (Freenode): #asbroker](http://webchat.freenode.net/?channels=%23asbroker)
- Email us at ansible-service-broker@redhat.com and subscribe to the Ansible Service Broker's
  [mailing list](https://www.redhat.com/mailman/listinfo/ansible-service-broker)
- Our [YouTube Channel](https://www.youtube.com/channel/UC04eOMIMiV06_RSZPb4OOBw)

**Important Links**
- Check out the [ansible playbook bundle](https://github.com/fusor/ansible-playbook-bundle) project
   and our [library of example APBs](https://github.com/fusor/apb-examples)
- [catasb](https://github.com/fusor/catasb) gives you more control over your development environment
- [Amazon Web Services deployed into OpenShift via Ansible Service Broker](https://www.youtube.com/watch?v=EKo3khfmhi8&index=2&list=PLZ7osZ-J70IaVc0NVyLs7tLO1hbhBdxHe)
- [Presentation Open Service Broker API + Ansible Service Broker/Ansible Playbook Bundles](https://www.youtube.com/watch?v=BaPMFZZ5lsc&index=1&list=PLZ7osZ-J70IaVc0NVyLs7tLO1hbhBdxHe)

# Getting Started with the Ansible Service Broker

There are a few different ways to quickly get up and running with a cluster + ansible-service-broker:

* `oc cluster up`
* Alternatively, [you can use minishift and install the broker with our addon, documented here](docs/minishift.md).

Let's walk through an `oc cluster up` based setup.

## Prerequisites
1. You will need a system setup for local [OpenShift Origin Cluster Management](https://github.com/openshift/origin/blob/master/docs/cluster_up_down.md)
    * Your OpenShift Client binary (`oc`) must be `>=` [v3.7.0-rc.0](https://github.com/openshift/origin/releases/tag/v3.7.0-rc.0)

## Deploy an OpenShift Origin Cluster with the Ansible Service Broker

[![Watch the full asciicast](docs/images/run_latest.gif)](https://asciinema.org/a/134509)

1. Download and execute our [run_latest_build.sh](https://raw.githubusercontent.com/openshift/ansible-service-broker/master/scripts/run_latest_build.sh) script

    Origin Version 3.7:
    ```
    wget https://raw.githubusercontent.com/openshift/ansible-service-broker/master/scripts/run_latest_build.sh
    chmod +x run_latest_build.sh
    ./run_latest_build.sh
    ```

1. At this point you should have a running cluster with the [service-catalog](https://github.com/kubernetes-incubator/service-catalog/) and the Ansible Service Broker running.

**Provision an instance of MediaWiki and PostgreSQL**
1. Log into OpenShift Web Console
1. Create a new project 'apb-demo'
1. Provision [MediaWiki APB](https://github.com/ansibleplaybookbundle/mediawiki-apb)
    * Select the 'apb-demo' project
    * Enter a 'MediaWiki Admin User Password': 's3curepw'
    * Click 'Create'
1. Provision [PostgreSQL APB](https://github.com/ansibleplaybookbundle/postgresql-apb)
    * Select the 'apb-demo' project
    * Leave 'PostgreSQL Password' blank, a random password will be generated
    * Choose a 'PostgreSQL Version'; either version will work.
    * Click 'Next'
    * Select 'Do not bind at this time' and then 'Create'
1. Wait until both APBs have finished deploying, and you see pods running for MediaWiki and PostgreSQL

**Bind MediaWiki to PostgreSQL**
1. Bind MediaWiki to PostgreSQL
    * Click on kebab menu for PostgreSQL
    * Select 'Create Binding' and then 'Bind'
    * Click on the link to the created secret
    * Click 'Add to Application'
    * Select 'mediawiki123' and 'Environment variables'
    * Click 'Save'
1. View the route for MediaWiki and verify the wiki is up and running.
    * Observe that mediawiki123 is on deployment '#2', having been automatically redeployed

# Contributing

First, **start with the** [Contributing Guide](CONTRIBUTING.md).

Contributions are welcome. Open issues for any bugs or problems you may run into,
ask us questions on [IRC (Freenode): #asbroker](http://webchat.freenode.net/?channels=%23asbroker),
or see what we are working on at our [Trello Board](https://trello.com/b/50JhiC5v/ansible-service-broker).

If you want to run the test suite, when you are ready to submit a PR for example,
make sure you have your development environment setup, and from the root of the
project run:

```
# Check your go source files (gofmt, go vet, golint), build the broker, and run unit tests
make check

# Get helpful information about our make targets
make help
```

# License

Ansible Service Broker is licensed under the [Apache License, Version 2.0](http://www.apache.org/licenses/).
