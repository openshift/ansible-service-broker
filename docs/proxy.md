# Running ASB behind a proxy

There are a few prerequisites necessary for running the broker in a proxied cluster:

**NO_PROXY**

The cluster must be configured to *not* proxy internal cluster requests. This
is typically configured with a `NO_PROXY` setting of ".cluster.local,.svc", in addition
to any other desired `NO_PROXY` settings. This is because the broker must be able
to directly communicate with its etcd instance.

**Adapter Whitelists**

The configured adapters must be able to communicate with their external registries
in order to bootstrap successfully and load remote APB manifests. These requests
can be made via the proxy, however, the proxy must ensure that the required remote
hosts are accessible.

Example required whitelisted hosts:

* rhcc - `registry.access.redhat.com`, `access.redhat.com`
* dh - `docker.io`
