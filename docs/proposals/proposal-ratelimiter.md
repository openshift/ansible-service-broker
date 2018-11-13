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
namespaces.

We will need a way to keep track of the namespaces previously used for the jobs.
Maybe we can find previous jobs.

### Rate limit the API

Today the broker will respond to an API call as many times as you call it. We
would like to limit the number of times per minute an API is called. We should
probably have a limit per API.

New configuration items:

```yaml
rate_limit:
  timeout: in seconds, default to 300
  handler: ["all", "provision", "deprovision", "bind", "unbind" ]
```
