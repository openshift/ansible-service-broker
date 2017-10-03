# Keeping transient pod name spaces alive

## Introduction
The goal of this proposal is to give administrators options to keep namespaces and therefore APB pods after the execution. The use cases for this feature are demos and debugging APBs.

## Problem Description
[Bug 1497766](https://bugzilla.redhat.com/show_bug.cgi?id=1497766) titled "APB Pods are deleted even when an error occurs" occurred when we moved to creating transient namespaces during the APB pod execution. This bug makes it difficult to debug APB development, debug issues in a production cluster using ASB, and creating demos because the pods do not live long enough to capture useful information.

## Implementation Details.

### Configuration Values

* `keep_namespace` or `ClusterConfig.KeepNamespace` - Parameter to always keep name space no matter what.
* `keep_namespace_on_error` or `ClusterConfig.KeepNamespaceOnError` -  parameter to keep name space around if an error occurs in the play book that is running.

Example:
```yaml
...
openshift:
  ....
  keep_namespace: false
  keep_namespace_on_error: true
...
```

**NOTE: Both configuration options will default to false in [openshift-ansible](https://github.com/openshift/openshift-ansible). When Deploying with CATASB or run_latest.sh we will set keep_namespace_on_error to true**

### Major Code Change

We will use the [DestroySandbox](https://github.com/openshift/ansible-service-broker/blob/34f643eec5349f58300e4e802581a65f4120976c/pkg/apb/svc_acct.go#L225) method to determine if we should delete the sandbox. This method is used already used by all methods that run the APB.

Here we can [get the pod](https://godoc.org/k8s.io/client-go/kubernetes/typed/core/v1#PodInterface). Then we can use the [PodPhase](https://godoc.org/k8s.io/api/core/v1#PodStatus) from the [PodSatus](https://godoc.org/k8s.io/api/core/v1#Pod). The [PodPhase](https://godoc.org/k8s.io/api/core/v1#PodPhase) will be in error if it is `PodFailed` or `PodUnknown`. 


The logic will be in the `pkg/apb/svc_account.go` file in the `DestroySandbox` method:
```golang
....
pod, err := k8scli.CoreV1().Pods(executionContext.Namespace).Get(executionContext.PodName, metav1.GetOptions{})
if err != nil {
        s.log.Errorf("Unable to retrieve pod - %v", err)     
}
if shouldDeleteNamespace(clusterConfig, pod, err)  {
    ... Delete Namespace 
}
...Delete role bindings.
```

```golang
func shouldDeleteNamespace(clusterConfig ClusterConfig, pod *apiv1.Pod, getPoderr error) bool {
    if clusterConfig.KeepNamespaces {
        return false
    }
    if clusterConfig.KeepNamespacesOnError {
        if pod.Status.Phase == apiv1.PodFailed || pod.Status.Phase == apiv1.PodUnknown || getPoderr != nil {
           return false
        }
    }
    return true
}
```

The outcome of keeping a namespace alive is that the RoleBindings and Service Accounts should be deleted. The only thing that should be left is the pod and the namespace.

## Work Items
- Add Code Above
- Add broker config values above during creation of Service Account.
- update the deployed template to set default  values
- CATASB change to allow for overriding the default values.
- doc updates for config, deployment
