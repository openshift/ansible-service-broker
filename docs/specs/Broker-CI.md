# Broker CI

The Broker CI spec will outline how the project is going create CI.

## Introduction
The Broker CI will involve multiple projects: ansible-service-broker
and apb-examples repo. The goal is to outline the possible CI for
each repo to achieve a successful CI run outlined in the next section.

## Successful CI Run
A successful CI run is spawning the broker and using the latest
rhscl-postgresapb and mediawiki-apb images to build a MediaWiki page.

## Environment
The expected running service in the environment in all test cases will be:
- Docker
- Openshift Cluster
- Service-Catalog

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

## Trigger
Jenkins has a github plugin that will track when a PR has been pushed to a repo.
So the trigger will be handled entirely by Jenkin.

There will be triggers on both apb-examples and ansible-service-broker repos.

## Test Process
The test process will execute the following tests in order:
- Deploy the ansible-service-broker
- Provision MediaWiki
- Provision rhscl-postgresql
- Bind rhscl-postgresql to MediaWiki
- Pull information from the MediaWiki page

### Local Testing
Locally running that gate would be a huge advantage. It would allow for faster
failures and would put a lot less strain on the gating jobs.

To achieve local gating, the script executing the [Test Process](#test-process)
needs to be a workflow only involving locally available tools: service-catalog,
ansible-service-broker, Openshift cli, Docker, and Ansible/GO/Python/Bash.

## Images
There are six Docker images that will be *latest* in every CI run: etcd,
ansible-service-broker, MediaWiki-apb, MediaWiki, rhscl-postgresql-apb,
postgresql.

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
already installed and configured to point at the service-catalog. In addition,

## Work Items
- Build a script that will contact the service-catalog to perform operations
- Build ```make ci``` into the Makefile
- The ```make ci``` recipe will perform the actions outlined [above](#make-test)
- Build a Jenkins job triggering ```make ci```
- Make Automated Builds for each APB in apb-examples
