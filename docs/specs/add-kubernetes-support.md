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

The file ```pkg/cluster/cluster.go``` will provide an object that will abstract
cluster logic.
```diff
+type COE interface {
+     CreateSandbox()
+     Run(APBActionProvision APBAction, req Context)
+     DestroySandbox()
+}

+struct Cluster {
+     ...
+}

+func (c *Cluster) ApbAction(...) {
+     c.cluster.CreateSandbox()
+     c.cluster.Run(APBActionProvision, req)
+     c.cluster.DeleteSandbox()
+}

+func (k *Kubernetes) Run(action, req) {
+     if action == provison {
+         k.CreatePod(req)
+     }
+}
```

The ```COE``` interface will hold all the public functions whose details are
hidden away inside the cluster files, kubernetes.go and openshift.go. Adding a
new COE operation requires adding an abstraced function that describes what
the interface does so it remains agnostic.

Adding a secret:
```diff
type COE interface {
     CreateSandbox()
     Run(APBActionProvision APBAction, req Context)
     DestroySandbox()
+    UpdatePermissions(req Context)
}
```

Using an abstration is an advantage because the broker code will only have a
single path and have no knowlege of cluster resources. It will call the cluster
object using broker native information, which will translate into cluster
resource allocation.

To provision an APB:
```diff
- k8scli.CoreV1().Pods(ns).Create(pod)
+ Cluster.ApbAction(APBActionProvision, req)
```

Down inside the cluster delegates, kubernetes.go and openshift.go, there will be
functions that handle the resource specific work.
```diff
+ func (k *Kubernetes) CreatePod(name string, image spec.Image, extraVars string,
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
+       k.client.CoreV1().Pods(ns).Create(pod)
+ }

```

```diff
+ func (o OpenShift) CreateClusterRoleBinding(roleBindingName string, svcAccountName string,
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
+      o.client.CoreV1().ClusterRoleBinding(ns).Create(roleBindingM)
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
| ClusterRoleBindings |  | X | X |n
| ClusterRoles |  | X | X |
| Project |  | X | X |

## Cluster Identification
Since the plan is for the Broker to run on two clusters, there needs to be a way
to identify which is being used.

### Broker Configuration and Validation
The broker will have configuration settings for the cluster.
```ClusterVersion``` will accept the special character '+'
to account for multiple versions.

```diff
broker:
+ Cluster: OpenShift
+ ClusterVersion: "1.6+"
```
or
```diff
broker:
+ Cluster: Kubernetes
+ ClusterVersion: "1.7"
```

When the broker is started, there will be a validation test to make sure
the cluster setting is correct.

### APB Spec
APBs will be written only for a single cluster and should identify which cluster
they work with. If not cluster is specified in the APB, assume it's meant for
OpenShift all the current examples work.

```diff
bindable: false
+ cluster: kubernetes
async: optional
```

## APB Cluster Identification
One of the trickiest parts of having one APB per cluster is identification.
It needs to be obvious for a user to know that an APB only works with OpenShift
or Kubernetes. Currently, the only annotation outlined is the 'cluster' option
in apb.yml, but there needs to be further identification.

### New APB Level Directory
Add a new directory layer into each APB.

```
mediawiki123-apb/
├── kubernetes
│   ├── apb.yml
│   ├── Dockerfile
│   ├── Dockerfile-dev
│   ├── playbooks
│   │   ├── deprovision.yml
│   │   └── provision.yml
│   └── roles
│       └── provision-mediawiki123-apb
│           ├── defaults
│           │   └── main.yml
│           └── tasks
│               └── main.yml
└── openshift
    ├── apb.yml
    ├── Dockerfile
    ├── Dockerfile-dev
    ├── playbooks
    │   ├── deprovision.yml
    │   └── provision.yml
    └── roles
        └── provision-mediawiki123-apb
            ├── defaults
            │   └── main.yml
            └── tasks
                └── main.yml
```

### New Top Level Directory
Add a new directory layer at the top level.

```
apb-examples/
├── kubernetes
│   └── mediawiki123-apb
│       ├── apb.yml
│       ├── Dockerfile
│       ├── Dockerfile-dev
│       └── playbooks
│           ├── deprovision.yml
│           └── provision.yml
│               └── roles
│                    └── provision-mediawiki123-apb
│                        ├── defaults
│                        │   └── main.yml
│                        └── tasks
│                            └── main.yml
│  
└── openshift
    └── mediawiki123-apb
        ├── apb.yml
        ├── Dockerfile
        ├── Dockerfile-dev
        ├── playbooks
        │   ├── deprovision.yml
        │   └── provision.yml
        └── roles
            └── provision-mediawiki123-apb
                ├── defaults
                │   └── main.yml
                └── tasks
                    └── main.yml
```

### Combine Repos
A bit more complex scenario, but apb-examples could be folded into the
ansible-service-broker repo under the 'examples' or 'apb' directory. That
directory can use either the 'New APB Level Directory' strategy or the
'New Top Level Directory' strategy.

## APB Containers
In order to target multiple clusters the APB containers require two changes:
 - Packages
 - oc-login.sh

### Packages
The main issue with packaging is that folks using Kubernetes may **only** want
the Kubernetes client. The package ```origin-clients``` installs both the
OpenShift and Kubernetes clients.  If this the container size isn't an issue,
then there may not be a problem using ```origin-clients```.

### Script
The ```oc-login.sh``` script should be changed to accept an environment varible
identifing the $COE.  Based on $COE, the login will use the ```kubectl``` or
```oc``` client.  Also, it should be renamed to ```login.sh```

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

## Documentation Impact
The apb-examples repo will be most impacted by the documentation change.
There needs to be a section in the README.md or another document detailing
the difference between writing an APB for Kubernetes pr OpenShift.

## CI
The community has two tools for CI: Jenkins and Travis. When adding
Kubernetes support to the broker, the upstream Travis CI job should test the
broker and catalog on top of both Kubernetes and OpenShift.  Jenkins will do
testing on OpenShift.

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
 7. Hook up the broker to the Cluster pkg.
 8. Accept Cluster and ClusterVersion parameters and create validations.
 9. Pass $CLUSTER as a variable to the APB.

**apb-examples**

 10. Add support for Kubernetes to ```oc-login.sh``` script. Also, it should
    be renamed.

**CI**

 11. Convert upstream CI, Travis, to using catasb to setup Kubernetes and the
     latest service-catalog so the gate can be used to test new PRs.

### Phase Three
Broker Overhaul - Convert all the CLI commands to API calls and round out any
edges.

**Ansible-service-broker**

 12. Convert all the client commands to API [calls](https://github.com/openshift/ansible-service-broker/search?p=1&q=%22oc%22&type=&utf8=%E2%9C%93).
 13. Update broker documentation.

**apb-examples**
 14. Address any packaging concerns.
 15. Update apb documentation.

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
 - Update documentation

### apb-examples
 - Add support for Kubernetes to ```oc-login.sh``` script. Also, it should
   be renamed.
 - Address any packaging concerns.
 - Additional [comments](https://github.com/fusor/apb-examples/issues/60).
 - Update documentation

### catasb
 - Create playbook for Kubernetes setup.
 - Add the ability to spawn the latest service-catalog in the cluster.
 - Update documentation

### CI
 - Convert upstream CI, Travis, to using Kubernetes
