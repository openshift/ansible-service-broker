# Recommended Local Development and Debugging Workflow

## Overview

We recommend 2 modes of running the broker for development work.

1. Build the broker executable and run it locally connected to a local oc cluster up from [fusor/catasb 'dev' branch](https://github.com/fusor/catasb/tree/dev)
2. Build a local image and deploy from image

## Running the broker executable locally

1. Build the Ansible Service Broker executable with the command: ```make build```
2. Ensure you have a local oc cluster up environment running with [fusor/catasb 'dev' branch](https://github.com/fusor/catasb/tree/dev)
3. ```cp scripts/my_local_dev_vars.example scripts/my_local_dev_vars```
4. Edit 'scripts/my_local_dev_vars' 
5. Prepare your local development environment by running: ```make prepare-local-env```
    * This will remove the running 'asb' pod and replace the endpoint of the asb route with the locally running broker executable.
    * You __must__ rerun this whenever you have reset your cluster.
    * It is safe to run this multiple times
6. Run the broker executable locally:  ```make run```
    * Use the cluster as you normally would for testing.
    * When you want to make a change and rebuild the broker.  CTRL-C, rebuild and re-run

## Build a local image and deploy from image

1. Build the Ansible Service Broker executable with the command: ```make build ```
3. Build a development image:  ```make build-image BROKER_IMAGE_NAME=asb-dev TAG=local```
4. Deploy from [templates/deploy-ansible-service-broker.template.yaml](https://github.com/openshift/ansible-service-broker/blob/master/templates/deploy-ansible-service-broker.template.yaml) setting the parameter "-p BROKER_IMAGE=asb-dev:local" to your local image name/tag 
