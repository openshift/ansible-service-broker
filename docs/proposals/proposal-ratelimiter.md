# Rate Limiter

## Problem

There are times that APIs are getting called for what seems like forever. For
example, a provision fails which will cause the service catalog to issue a
deprovision. If the deprovision fails, this cycle can continue forever.
Normally the broker is configured to save failed APB pods so that they may be
debugged later. At the moment we save ALL APB pods which can create a lot of
unwanted namespaces.

There are 2 parts to this feature:

1) look into only saving the latest failed APB pod. For example, if a
   deprovision call fails, and a subsequent one fails with the same
   parameters, then we should delete the older pod and keep the newer one

2) rate limit the API, such that after a certain period of time we don't
   spawn new pods at all for the API and just return that it failed always.
   Basically becoming a no-op.

### Saving latest failed pod

Today we have a KeepNamespaceOnError configuration that tells the broker to
keep the namespace around if there was an error. This happens on every error.
We would want to keep only the latest one and get rid of the previously failed
namespaces. We will need a way to keep track of the namespaces previously used
for the jobs.

The `Namespace` is stored on the `Context` which allows the current run to be
able to remote the namespace at the end of the action. Because the names are
generated it isn't feasible to go searching for all "action" jobs.

I think we need to add a new attribute, `PreviousNamespaces`, to
`BundleInstanceStatus`. It will be an slice of strings. The list will be ordered
as first in first out. Newer jobs will append to the list, we simply pop the
first ones off.

```
  // BundleInstanceStatus status is a service instance status.
    Bindings           []LocalObjectReference `json:"bindings"`
    State              State                  `json:"state"`
    LastDescription    string                 `json:"lastDescription,omitempty"`
    Jobs               map[string]Job         `json:"jobs"`
    PreviousNamespaces []string               `json:"previousNamespaces"`
  }
```

When we start the action, we should append the namespace to the
`PreviousNamespaces` field. If `KeepNamespaceOnError` is false, we will delete
the sandbox namespace and should remove it from the list. This would probably be
done in a defer similar to how we do the `DestroySandbox` call.

We will keep at most 1 old failed namespace instead of all of them. The idea
here is to minimize wasting resources in the cluster and still allow developers
& admins to have access to the failed action. During a deprovision & unbind,
once we get the `ServiceInstance` we will see if there are any
`PreviousNamespaces`. If there are old ones, we will delete the first one in the
list, id 0. Before creating a new sandbox namespace.

TBD: not sure if this should be done at the handler level or further down. I'll
leave that as an implementation detail.

NOTE: we could make the number of namespaces kept to be configurable if users
want to keep _n_ number of namespaces. But I think that might add unneeded
complexity.

### Rate limit the API

The broker will respond to an API call as many times as you call it. Often this
means launching a new APB which can result in a ton of cluster resources being
used when the API call was a result of a previous failure. For example, if a
provision call fails, the platform will likely call a deprovision to prevent any
dangling resources. If the deprovision fails, the platform will continue to call
deprovision until it works. If the `KeepNamespaceOnError` is `true`, we will
keep the namespaces for *every* failed call around wasting precious cluster
resources.

The proposal above will mitigate the left over namespaces. But if we now a call
will always fail because it has failed the last n times. There really is no
point to continue to spawn new APBs and namespaces. We should backoff the calls
exponentially. We take n calls with no limit, once those n are exhausted we take
every 3rd call, then we back that off to every 5th call. After m failed attempts
we stop executing APBs and simply return failure always.

New configuration items:

```yaml
rate_limit:
  timeout: in seconds, default to 300
  handler: ["all", "provision", "deprovision", "bind", "unbind" ]
```
