# Deployment

## Parameters
The following are the configurable parameters that you can set for deploying the Ansible Service Broker (ASB):

Name | Default Value | Description
---|---|---
PROJECT | ansible-service-broker | Project Namespace
BROKER_IMAGE | ansibleplaybookbundle/origin-ansible-service-broker:latest| Container Image to use for Ansible Service Broker in format of imagename:tag
ETCD_IMAGE | quay.io/coreos/etcd:latest | Container Image to use for etcd in format of imagename:tag
BROKER_CONFIG | /etc/ansible-service-broker/config.yaml | Configuration file path for Ansible Service Broker
DOCKERHUB_ORG | ansibleplaybookbundle | Dockerhub organization
DOCKERHUB_USER | "" | Dockerhub user Name
DOCKERHUB_PASS | "" | Dockerhub user Password
OPENSHIFT_TARGET | https://kubernetes.default | OpenShift Target URL
REGISTRY_TYPE | dockerhub | Registry Type
REGISTRY_URL | docker.io | Registry URL
DEV_BROKER | true | Include Broker Development Endpoint (true/false)
AUTO_ESCALATE | true | Auto escalate the users permissions when running an APB
BROKER_CA_CERT | "" | Tells the broker that the ca that has signed the SSL Cert and Key
KEEP_NAMESPACE | false | Always keep the APB namespace after execution.
KEEP_NAMESPACE_ON_ERROR | true | Keeps the APB namespace after an error during execution.

## Template
The following is the template used to deploy the ASB:
 * [deploy-ansible-service-broker.template.yaml](../templates/deploy-ansible-service-broker.template.yaml)

### Launch APB on Bind Parameter
Currently, we are ***disabling*** running an APB on `Bind()` due to the lack of async support of bind in the Open Service Broker API.  This is achieved via `launchapbonbind` which is currently set to `false`.  You may enable launching of APB on Bind by changing it to `true`.  However, since there is a timeout of ~15 seconds asserted from the Service Catalog, you will likely see failures running an APB.

## Run Deployment Script
The script below sets the parameter values that the deployment template expects, and deploys the Ansible Service Broker to the cluster.
 * [run_template.sh](../templates/run_template.sh)

To run the script, edit the script file, and modify the parameter values, then execute the script
```bash
./templates/run_template.sh
```
