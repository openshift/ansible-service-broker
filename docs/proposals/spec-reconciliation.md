# Spec Reconciliation Proposal

## Introduction
Add a way for broker to sync the list of specs it has with registries while handling the following corner cases:
 1. Temporary loss of connection with registry
 2. A spec removed from the registry but has provisioned instance/s
 3. The service catalog sends a provision request to the broker after the broker has started bootstrapping but before it has completed bootstrapping (untimely provision request)
 4. The service catalog sends a provision request to the broker for specs that has been marked deleted in the broker (out of sync provision request)

## Problem Description

The current approach of syncing the spec from registries is as follows:
 1. Delete all the specs currently in the datastore
 2. Fetch new specs
 3. Add all the fetched specs to the datastore
 
This leads to a bunch of bugs:
 - [Bug 1583495](https://bugzilla.redhat.com/show_bug.cgi?id=1583495)
 - [Bug 1577810](https://bugzilla.redhat.com/show_bug.cgi?id=1583495) 
 - [Bug 1586345](https://bugzilla.redhat.com/show_bug.cgi?id=1586345)

The underlying problem with all those bugs is that the broker is deleting specs even if it is provisioned. When the catalog tries to deprovision, bind, unbind with these dangling instances (instances with no specs) it will fail. Hence the broker needs a way to safely delete the specs.

### Approaches to safely deleting specs

The broker can take the following actions to safely delete the specs:
#### Approach#1
 1. Fetch the specs from registry
    1. If the fetch fails increment a failed connection counter. If the failed connection counter reaches a limit delete all the specs that do not have a service instance
    2. If the fetch succeeds:
        1. add new specs to the datastore
        2. delete that specs that are removed from datastore and do not have a service instance.
        3. mark the specs that are removed from datastore and has a service instance
 2. During deprovision, if the instance getting deprovisioned is the last one delete the spec from datastore
 3. Serve the catalog request with specs that are not marked
 
This will solve the problem but the task of deleting the specs is spread across time (when fetching specs and when deprovisioning)

#### Approach#2
 1. Fetch the specs from registry
    1. If the fetch fails mark all the specs for deletion
    2. If the fetch succeeds,
       1. add new specs to the datastore
       2. mark the specs that are removed from the registry.
 3. Run a background task to delete all the marked specs that does not have a service instance currently provisioned. The interval of running this task can be configurable
 4. Serve the catalog request with specs that are not marked
 
This solves the problem but introduces a new parameter of configuration. 

If in future, other conditions of deleting a spec crop up, just mark it for deletion and background task will take care of deleting it safely.


Either approach will take care of the all objectives and bugs. 

### Approaches to handle untimely provision request


To to handle a case where provision is requested after the broker has started bootstrapping but before it has completed bootstrapping, the broker can take the following approaches:

#### Approach#1
The broker always sends an error code if it is busy provisioning. OSB [spec](https://github.com/openservicebrokerapi/servicebroker/blob/master/spec.md#response-2) does not have an appropriate error code for busy.

#### Approach#2 
It can block the request and when the bootstrap completes:
 1. provision the instance if the spec is still there in datastore and return a success code
 2. send an error code - 400 Bad request, if the spec is deleted in the registry

#### Approach #3
It can send a 202 Accepted status, suggesting that provisioning is in progress. The service catalog will when keep polling to check if the service instance is fully provisioned. The [last operation](https://github.com/openservicebrokerapi/servicebroker/blob/master/spec.md#response-1) has a mechanism where broker can tell if the provision is success of failure.

### Approaches to handle out of sync provision request

To handle the case where service catalog sends a request to provision an instance of a spec that has been marked for deletion the broker can take the following approaches:

#### Approach#1
Successfully serves the provision request (very similar to scenario #1): This approach might lead to unexpected problems during execution.

#### Approach#2
Error out the provision request. It is clear with the polling approach of the Service Catalog to update the [specs](https://github.com/openservicebrokerapi/servicebroker/blob/master/spec.md#response-2), that this scenario will eventually popup, but the specs doesnt have an error code specifically meant for this scenario.

#### Approach#3
It can send a 202 Accepted status, suggesting that provisioning is in progress. The service catalog will when keep polling to check if the service instance is fully provisioned. The [last operation](https://github.com/openservicebrokerapi/servicebroker/blob/master/spec.md#response-1) has a mechanism where broker can tell if the provision is success of failure.


## Other ideas

These ideas are not directly related to the objectives of this proposal but are mentioned here and might become independent proposals of their own when addressing those issues.

#### APB Versioning

Currently, tags are the only way an apb developer can version an apb. If the developer is given a way to do an explicit versioning of the apb irrespective of the tags, there might never be a need to allow multiple tags

A rough idea of what can be done:

 1. Add a new field to apb.yml call bundleVersion
 2. During provision save the value of "bundleVersion" with the service instance
 3. During de-provision pass the value of "bundleVersion" in the args.
 
This allows:
 1. the developer to have a versioning scheme that is independent of the tags.
 2. the developer can push v1.1 of that apb even if v1.0 is provisioned. The version number passed during binding, unbinding and deprovision would still be v1.0
 3. the admin can know what version of apb is provisioned in the cluster
 

#### Multiple tags

The requirement supporting multiple tags of apb is largely because versioning of an apb is tied to the tag right now. There needs to be a way in which admins can have multiple versions of the same apb. If the apb versioning scheme mentioned above is implemented, there might not be a need to support multiple tags

If there is, the registry section of configmap of the broker can be modified as follows:

    registry:
      - type: dockerhub
        name: dh
        url: https://registry.hub.docker.com
        org: ansibleplaybookbundle
        tag:
          default: latest
          named: 
            - hello-apb: 
              - v1.0
              - v2.0
            - postgres-apb: 
              - v3.9
        white_list:
          - '.*-apb$'


This would require modification to the bootstrap function where after fetching all the specs, the broker needs to iterate through the list of named specs and fetch specific tags from the registry

Adding a method to adaptor interface that takes two inputs, image name and image tag, and fetches its spec. The broker will then call this method iteratively for all named pair of specs and tags
