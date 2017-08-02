# Move to Upstream Kubernetes

## Introduction
The ansible-service-broker and all of the tooling around it
(apb-examples, catasb, adn ansible-playbook-bundles) specifically target
OpenShift as the underlying COE (container orchestration engine) when it
should be cluster agnostic.

## Problem Description
Issues having been coming up during testing because the broker does not target
the latest code for either the service-catalog or the underlying cluster.
Kubernetes is the upstream community for OpenShift and the broker should be
consumuing the latest upstream code for development purposes.

This change will improve the Broker development process by having the Broker
constantly being developed and tested against the latest upstream code. The
Broker's functionality won't change, but all the tooling around it will.

## Split Out All Cluster logic
The target is to make the Broker cluster agnostic. Where the code path for
any COE can be added to the Broker.

Split out all Cluster logic into ```pkg/cluster/openshift.go``` and
```pkg/cluster/kubernetes.go```.

### Upstream and Downstream Code Hierarchy
OpenShift is Kubernetes plus a few features, API name changes, and tools.
In the core ansible-service-broker code, ```pkg/cluster/kubernetes.go```
should be able to account for about 80% of the cluster opertations required
by the broker. The remaining ~20% will be held in ```pkg/cluster/openshift.go```.

### Hierarchy Diagram
This diagram represents if the current state of code were split into two
files.

**pkg/cluster/kubernetes.go**

| API Resource | Uses API | Uses CLI | Different between OpenShift and Kubernetes |
|:---:|:---:|:---:|:---:|
| Pods | X |  |  |
| Exec |  | X |  |
| Login |  | X |  |
| ServiceAccount |  | X |  |
| RBAC ClusterRoleBindings |  | X | X |
| RBAC ClusterRoles |  | X | X |
| Namespace |  | X | X |

**pkg/cluster/openshift.go**

| API Resource | Uses API | Uses CLI | Different between OpenShift and Kubernetes |
|:---:|:---:|:---:|:---:|
| ClusterRoleBindings |  | X | X |
| ClusterRoles |  | X | X |
| Project |  | X | X |

## APBs
The APBs only target OpenShift as a cluster. The APBs will differ in two ways:
 - Packages
 - oc-login.sh

### Packages
The main issue with packaging is that folks using Kubernetes may only want the
Kubernetes client. The package ```origin-clients``` installs both the OpenShift
and Kubernetes clients.  If this the container size isn't an issue, then there
may not be a problem using ```origin-clients```.

### Script
The ```oc-login.sh``` script should be changed to accept an environment varible
identifing the $COE.  Based on $COE, the login will user the ```kubectl``` or
```oc``` client.

## Work Items
### Ansible-service-broker
 - Split out all Cluster logic into a ```pkg/cluster/openshift.go``` and
   ```pkg/cluster/kubernetes.go```.
 - Create ```pkg/cluster/kubernetes-hack.go``` that will run client commands
   via  ```kubectl```.
 - Create a Kubernetes solution for ClusterRoleBindings, ClusterRoles, and
   Projects using the clients.
 - Convert all the client commands to API [calls](https://github.com/openshift/ansible-service-broker/search?p=1&q=%22oc%22&type=&utf8=%E2%9C%93).
 - Convert any broker tools to target either OpenShift or Kubernetes.

### apb-examples
 - Add support for Kubernetes to ```oc-login.sh``` script. Also, it should
   be renamed.
 - Address any packaging concerns.
 - Additional [comments](https://github.com/fusor/apb-examples/issues/60).

### catasb
 - Create playbook for Kubernetes setup.
 - Add the ability to spawn the latest service-catalog in the cluster.
