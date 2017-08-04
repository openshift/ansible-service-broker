# Move to Upstream Kubernetes

## Introduction
The ansible-service-broker and all of the tooling around it
(apb-examples, catasb, and ansible-playbook-bundles) specifically target
OpenShift as the underlying COE (container orchestration engine) when it
should be cluster agnostic.

## Problem Description
Issues have been coming up during testing because the broker does not target
the latest code for either the service-catalog or the underlying cluster.
Kubernetes is the upstream community for OpenShift and the broker should be
consumuing the latest upstream code for development purposes.

This change will improve the Broker development process by having the Broker
constantly being developed and tested against the latest upstream code.

The Broker's functionality won't change, but all the tooling around it will.

## Split Out All Cluster logic
OpenShift is Kubernetes plus a few features, API name changes, and tools.
In the core ansible-service-broker code, ```pkg/cluster/kubernetes.go```
should be able to account for about 80% of the cluster opertations required
by the broker. The remaining ~20% will be held in ```pkg/cluster/openshift.go```.

### Code
Start by creating a new pkg for cluster objects, ```pkg/cluster/...```. From
there, the cluster pkg will hold ```pkg/cluster/cluster.go```,
```pkg/cluster/kubernetes.go```, and ```pkg/cluster/openshift.go```.

The file ```pkg/cluster/cluster.go```  will provide an object that will abstract
cluster logic.

Using an abstration is an advantage because the broker code will only have a
single path. It will call the cluster object asking for an operation. For
example, to create a new Pod:
```diff
- k8scli.CoreV1().Pods(ns).Create(pod)
+ Cluster.CreatePod(name, image, extraVars, pullPolicy, serviceAccountName, ns)
```

The Cluster object will recieve the request and call the correct API.
```diff
+ func CreatePod(name string, image spec.Image, extraVars string,
+                pullPolicy string, serviceAccountName string, ns string) {
+	pod := &v1.Pod{
+		ObjectMeta: metav1.ObjectMeta{
+			Name: apbID,
+		},
+		Spec: v1.PodSpec{
+			Containers: []v1.Container{
+				{
+					Name:  "apb",
+					Image: spec.Image,
+					Args: []string{
+						action,
+						"--extra-vars",
+						extraVars,
+					},
+					ImagePullPolicy: pullPolicy,
+				},
+			},
+			RestartPolicy:      v1.RestartPolicyNever,
+			ServiceAccountName: serviceAccountName,
+		},
+	}
+
+       k8scli.CoreV1().Pods(ns).Create(pod)
+
+ }

```

If there are identical API objects between COEs, check which cluster is
being used.
```diff
+ func CreateClusterRoleBinding(roleBindingName string, svcAccountName string,
+                               ApbRole string, ns string) {
+	roleBindingM := map[string]interface{}{
+		"apiVersion": "v1",
+		"kind":       "RoleBinding",
+		"metadata": map[string]string{
+			"name":      roleBindingName,
+			"namespace": ns,
+		},
+		"subjects": []map[string]string{
+			map[string]string{
+				"kind":      "ServiceAccount",
+				"name":      svcAccountName,
+				"namespace": ns,
+			},
+		},
+		"roleRef": map[string]string{
+			"name": ApbRole,
+		},
+      }
+
+      if OpenShift {
+         OCcli.CoreV1().ClusterRoleBinding(ns).Create(roleBindingM)
+      } elif Kubernetes {
+         k8scli.CoreV1().ClusterRoleBinding(ns).Create(roleBindingM)
+      }
+ }

```

### Hierarchy Diagram
The diagram represents the code split into two files and the API objects
in each of them

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
identifing the $COE.  Based on $COE, the login will use the ```kubectl``` or
```oc``` client.

## Tools and Docs that reference the OpenShift client
**Tools**

 - scripts/prep_local_devel_env.sh
 - scripts/broker-ci/setup.sh
 - scripts/run_latest_build.sh
 - scripts/deploy.sh
 - scripts/broker-ci/wait-for-pods.sh
 - .travis.yml

**Docs**

 - docs/bind-data-transmission.md
 - docs/local_development.md
 - README.md

## CI
The community has two tools for CI: Jenkins and Travis. When adding
Kubernetes support to the broker, the upstream Travis CI job should test the
broker and catalog on top of the upstream COE, Kubernetes.  Jenkins will become
the downstream CI testing on OpenShift.

## Phased Plan
Moving to a cluster agnostic architecure will require a large about of effort.
We'll use a phased approach to accomplish the move so things don't get too
chaotic all at once.

### Phase One
Infrastructure setup - Make sure there's a repeatable infrastracture setup
and begin transitioning broker code.

**catasb**

 1. Create playbook for Kubernetes setup.
 2. Add the ability to spawn the latest service-catalog in the cluster.

**Ansible-service-broker**

 3a. Create ```pkg/cluster/kubernetes-hack.go``` that will run client commands
     via  ```kubectl```.
 3b. Create ```pkg/cluster/openshift-hack.go``` that will run client commands
     via  ```oc```.
 4.  Split out the cluster logic into a ```pkg/cluster/openshift.go``` and
     ```pkg/cluster/kubernetes.go```.

### Phase Two
Tool Transition - Get all the environment tools working and the gate green.

**Ansible-service-broker**

 5. Create a Kubernetes solution for ClusterRoleBindings, ClusterRoles, and
    Projects, using the clients.
 6. Convert any broker tools and scripts to target either OpenShift or
    Kubernetes.

**apb-examples**

 7. Add support for Kubernetes to ```oc-login.sh``` script. Also, it should
    be renamed.

**CI**

 8. Convert upstream CI, Travis, to using catasb to setup Kubernetes and the
    latest service-catalog so the gate can be used to test new PRs.

### Phase Three
Broker Overhaul - Convert all the CLI commands to API calls and round out any
edges.

**Ansible-service-broker**

 9. Convert all the client commands to API [calls](https://github.com/openshift/ansible-service-broker/search?p=1&q=%22oc%22&type=&utf8=%E2%9C%93).

**apb-examples**
 10. Address any packaging concerns.

## Work Items
### Ansible-service-broker
 - Create ```pkg/cluster/kubernetes-hack.go``` that will run client commands
    via  ```kubectl```.
 - Create ```pkg/cluster/openshift-hack.go``` that will run client commands
    via  ```oc```.
 - Split out all the cluster logic into a ```pkg/cluster/openshift.go``` and
    ```pkg/cluster/kubernetes.go```.
 - Convert any broker tools to target either OpenShift or Kubernetes.
 - Create a Kubernetes solution for ClusterRoleBindings, ClusterRoles, and
    Projects using the clients.
 - Convert all the client commands to API [calls](https://github.com/openshift/ansible-service-broker/search?p=1&q=%22oc%22&type=&utf8=%E2%9C%93).

### apb-examples
 - Add support for Kubernetes to ```oc-login.sh``` script. Also, it should
   be renamed.
 - Address any packaging concerns.
 - Additional [comments](https://github.com/fusor/apb-examples/issues/60).

### catasb
 - Create playbook for Kubernetes setup.
 - Add the ability to spawn the latest service-catalog in the cluster.

### CI
 - Convert upstream CI, Travis, to using Kubernetes
