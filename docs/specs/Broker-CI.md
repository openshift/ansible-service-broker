# Broker Continuous Integration

The Broker CI spec will outline how the project is going to create CI.

## Introduction
The Broker CI will involve multiple projects: ansible-service-broker
and apb-examples repo. The goal is to outline the possible CI for
each repo to achieve a successful CI run outlined in the next section.

## Successful CI Run
A successful CI run is spawning the broker and using the latest
rhscl-postgresapb and mediawiki-apb images to build a MediaWiki page.
The MediiaWiki page will be verified by curling information from the page
and checking it's correct.

## Environment
The expected running service in the environment in all test cases will be:
- Docker
- OpenShift Cluster
- [Service-Catalog](https://github.com/kubernetes-incubator/service-catalog)

The services that will be deployed every CI run are:
- ansible-service-broker
- mediawiki
- postgresql

The environment will be deployed using [catasb](https://github.com/fusor/catasb)

## Additional Environment Questions
When to redeploy catasb?
Deploy catasb each CI run?
Redeploy catasb every new service-catalog image?
Redeploy catasb nightly?
Should there be an already existing OpenShift Cluster to land pods on?
Should Jenkins be run in the OpenShift Cluster?

## Tools
There are multiple tools availble to use for CI: Travis and Jenkins.

Travis CI will be used for the simpler and less robust CI. It
will test a provision and bind of an app using code a PR introduces.

Jenkins is going to be used for the more expansive integration CI.
An example of this is CI around third party software consuming the broker
to provision applications. These could be any of Amazon's applications.

## Trigger
- Jenkins has a github plugin that will track when a PR has been pushed to a repo.
- Travis will automatically run when a PR is made to the repo.

There will be triggers on both apb-examples and ansible-service-broker repos.

### Advanced CI Triggers
The more expensive CI operations should only be only started by a trigger.
After a PR has been testing by the cheap CI and recieved an approval, a
CI trigger can by used by commenting in the PR.

## Test Process
There will be different levels of CI that will be used for a PR.

- 'Fast and Cheap' will run with every change to a PR.
- 'Full Test' will be triggered when a PR has been approved and is passing
the Fast and Cheap CI. Commenting ```full-test``` in Git will trigger the 'Full
Test'.

< add any additional CI layers here>

### Fast and Cheap
- Build the Broker, MediaWiki, and rhscl-prosgrsql containers
- Deploy the Broker
- Provision MediaWiki
- Provision rhscl-postgresql
- Bind rhscl-postgresql to MediaWiki
- Pull information from the MediaWiki page

Runtime with caching: ~2 minutes

### Full Test
- Deploy OpenShift & Service-Catalog with catasb
- Build the Broker, MediaWiki, and rhscl-prosgrsql containers
- Deploy the Broker
- Provision MediaWiki
- Provision rhscl-postgresql
- Bind rhscl-postgresql to MediaWiki
- Pull information from the MediaWiki page

Runtime with caching: ~5 minutes

## Local Testing
Locally running that gate would be a huge advantage. It would allow for faster
failures and would put a lot less strain on the gating jobs.

To achieve local gating, the script executing the [Test Process](#test-process)
needs to be a workflow only involving locally available tools:
- service-catalog
- ansible-service-broker
- OpenShift cli
- Docker
- Ansible/GO/Python/Bash

local workflow:
- Build MediaWiki & rhscl-postgresql containers
- ```make run &```
- Provision MediaWiki
- Provision rhscl-postgresql
- Bind rhscl-postgresql to MediaWiki
- Pull information from the MediaWiki page
- kill make run process

## Images
There are six Docker images that will be *latest* in every CI run:
- etcd
- ansible-service-broker
- MediaWiki-apb
- MediaWiki
- rhscl-postgresql-apb
- postgresql

*latest* implies that they are being built locally or pulled using the latest
tag.

## Docker Image Building and Publishing
The docker images used for the CI job will be built from the checkout of the PR
with the ```make prepare-build-image``` command.

After a CI run, the images will remain cached on the machine and won't be
published. When the PR merges, the images will be automatically re-built in
the Dockerhub Registry.

As an example, the broker already has [this](https://hub.docker.com/r/ansibleplaybookbundle/ansible-service-broker-source/builds/)

## Additional Pieces for the Broker
This is the code that still needs to be written in order to complete a CI run.

### Make CI
The CI will be triggered by running ```make ci```. ```make ci``` will be
usable both locally and in remote CI.

```make ci``` is purely going to execute the test workflow. No environmental
work will be handled by ```make ci```.

### Service Catalog Script
The service catalog script will imitate the UI's interaction with the
service-catalog.

#### Using Curl
The service-catalog will be listening for http requests on <ip>:<port> so we
can contact the service-catalog with ```curl```. Script these curl commands
behind a cli that the CI can use.

#### Using Bash
The CI job can execute the tests purely from bash by having an OpenShift client
already installed and configured to point at the service-catalog.

#### Using GO
Using a more powerful language to organize the CI will allow for it to be more
extensible. Ci jobs can be organized behind objects so creating new jobs will
be easy in the future.

## Work Items
[x] Make Automated Builds for each APB in apb-examples
[] Build ```make ci``` into the Makefile
[] Build a Travis job triggering ```make ci```
[] Build a CI framework so it's easy to create new CI jobs
[] Build a script that will contact the service-catalog to perform operations
[] Build a Jenkins job triggering more robust testing
