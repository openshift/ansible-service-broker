# Ansible Playbook Bundle

## Overview
An [Ansible Playbook Bundle (APB)](https://github.com/fusor/ansible-playbook-bundle)
borrows several concepts from the [Nulecule](https://github.com/projectatomic/nulecule)
or [Atomicapp](http://www.projectatomic.io/docs/atomicapp/) project, namely the concept of a short
lived container with the sole purpose of orchestrating the deployment of the intended application. For the case
of APB, this short lived container is the APB; a container with an Ansible runtime environment
plus any files required to assist in orchestration such as playbooks, roles, and extra dependencies.
Specification of an APB is intended to be lightweight, consisting of several named playbooks and a
metadata file to capture information such as parameters to pass into the application.


For more information on APBs
* [Ansible Playbook Bundle Repo](https://github.com/fusor/ansible-playbook-bundle)
* [Getting started with APBs](https://github.com/fusor/ansible-playbook-bundle/blob/master/docs/getting_started.md).

## Workflow
APB is broken up into the following steps.

  1. [Preparation](#preparation)
  1. [Build](#build)
  1. [Deploy](#deploy)

### Preparation
The first step to creating an APB is preparing the files required to manage the application’s lifecycle.
Two methods of preparing the needed files are supported, a [guided approach](#guided-approach) that uses tooling to handle the majority of cases and makes the experience easier as well as an [advanced approach](#advanced-approach) that allows an experienced user full control to generate the few required files by hand.

![Prepare](images/apb-prepare.png)

#### Guided Approach
The guided approach leverages and extends [ansible-container](https://github.com/ansible/ansible-container) to provide a
solution for building all referenced images, generating a deployment role, and populating the named playbooks an APB requires.

The use of `ansible-container` allows a user to create `yaml` files to express image building and container
deployment instructions for multiple environments. A translation step looks at this “single source of truth”
(`main.yml`) and translates it to a deployment role targeted for a specific platform, in our case of OpenShift
this is powered by the [Kompose](https://github.com/kubernetes-incubator/kompose) project.

As the deployment role is generated from a translation step we recognize that the approach is unlikely to handle
all of the possible use cases for managing applications on OpenShift. The guided approach may be suitable for
many use cases yet an alternative method is needed to address the trickier problems.

#### Advanced Approach
As an alternative to the guided approach, a user can package an APB reusing their existing Ansible playbooks and
roles. Translating a working Ansible deployment role to an APB requires adding a few named playbooks and a
metadata file.

Requirements:
 * provision.yaml
   * Playbook called to handle installing application to the cluster
 * deprovision.yaml
   * Playbook called to handle uninstalling
 * bind.yaml
   * Playbook to grant access to another service to use this service, i.e. generates credentials
 * unbind.yaml
   * Playbook to revoke access to this service
 * apb.yaml
   * Metadata file, exposes parameters the application accepts

The required named playbooks correspond to methods defined by the Open Service Broker API. For example, when the
Ansible Service Broker needs to provision an APB it will execute the provision.yaml.

After the required named playbooks have been generated the files can be used directly to test management of the
application. A developer may want to work with this directory of files, make tweaks, run, repeat until they are
happy with the behavior. They can test the playbooks by invoking Ansible directly with the playbook and any
required variables.

### Build
The build step is responsible for building a container image from the named playbooks for distribution.
Packaging combines a base image containing an Ansible runtime with Ansible artifacts and any dependencies required
to run the playbooks. The result is a container image with an ENTRYPOINT set to take in several arguments, one of
which is the method to execute, such as provision, deprovision, etc.

![Package](images/apb-package.png)

### Deploy
Deploying an APB means invoking the container and passing in the name of the playbook to execute along with any
required variables. It’s possible to invoke the APB directly without going through the Ansible Service Broker.
Each APB is packaged so it’s ENTRYPOINT will invoke Ansible when run. The container is intended to be short-lived,
coming up to execute the Ansible playbook for managing the application then exiting.

In a typical APB deploy, the APB container will provision an application by running the provision.yaml playbook which
executes a deployment role. The deployment role is responsible for creating the OpenShift resources, perhaps through
calling oc create commands or leveraging Ansible modules. The end result is that the APB runs Ansible to talk to
OpenShift to orchestrate the provisioning of the intended application.

![Deploy](images/apb-deploy.png)
