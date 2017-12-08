# Network Isolation

## Introduction
This [redhat/ovs-multitenant network plugin](https://docs.openshift.com/container-platform/3.6/architecture/additional_concepts/sdn.html#architecture-additional-concepts-sdn) will restrict namespaces from having network traffic flow between them. 

> The ovs-multitenant plug-in provides OpenShift Container Platform project level isolation for pods and services. Each project receives a unique Virtual Network ID (VNID) that identifies traffic from pods assigned to the project. Pods from different projects cannot send packets to or receive packets from pods and services of a different project.

The ASB creates a transient namespace while running an APB and grants the correct access to the target namespace. This network plugin will cause some APBs to fail, if the APB assumes it can access pods in the target namespace over the network.

## Problem Description
The transient namespace does not have network access to the target namespace, leaving the APB pod unable to perform all the tasks that it should be able to perform.

## Implementation Details
With no network access to the target namespace, APBs will be limited in some of the more advanced tasks they can perform. We need a mechanism to give the APBs in the transient namespace access to the target namespace. We should be able to add the transient namespace to the same network as the target namespace. See the [managing networking](https://docs.openshift.com/container-platform/3.6/admin_guide/managing_networking.html) document for different ways of working with networking.

One of the big things is we want this to be easily expandable to [kubernetes SDN's](https://kubernetes.io/docs/concepts/cluster-administration/networking) that could implement the same basic structure. This proposal will not address each networking option but rather create a common structure for implementing more SDN's in the future.


### Steps to take.
1. Inside the runtime package, determine if adding networks is necessary.
    1. The first implementation of this is for openshift, if the network plugin is "redhat/openshift-ovs-multitenant" then we should be joining the networks. 
    2. This can be determined from an openshift [rest call](https://github.com/openshift/origin/blob/1f270ca122306656b228faa92bc71d2136e0f97a/pkg/oc/admin/network/project_options.go#L90)
    3. This should be determined at runtime [initilization](https://github.com/openshift/ansible-service-broker/blob/master/pkg/runtime/runtime.go#L54).
    
    example:
    ```go
    type NetworkIsolation interface {
        JoinNetworks(...)
        SeperateNetworks(...)
    }

    type MultitenantNetwork struct
    }

    func (m MultitenantNetwork) JoinNetworks(...) {
        ... See Step 2 for implementation of this.

    }

    func (m MultitenantNetwork) SeperateNetworks(...) {
        ... See Step 2 for implementation of this 
    }

    type provider struct {
        ...
        networkIsolation NetworkIsolation 

    }

    func NewRuntime() {
         networkIsolation := shouldJoinNetworks(...)

    }

    func shouldJoinNetworks(....) NetworkIsolation {
        n, _ := openshift.Get().Network()
        if n.Plugin == redhatMultitenant {
            return MultitenantNetwork{} 
        }
        return nil
    }

    func (p provider) CreateSandbox(...) ... {
            ...
            if p.networkIsolation != nil {
                if err := p.networkIsolation(...); err != nil {
                    .... Log Statements ...
                    return err
                }
            }
    }
    ```
2. If we do need to add networks together, then during the apb sandbox creation, we should join the networks.
    1. Examples of how to do this are in the `oadm` client in origin. [here](https://github.com/openshift/origin/blob/1f270ca122306656b228faa92bc71d2136e0f97a/pkg/oc/admin/network/project_options.go#L157) and the [update](https://github.com/openshift/origin/blob/master/pkg/network/netid.go#L73) to the annotation.
3. If we should join the networks together but are not able to, then we should error out of the provision.
