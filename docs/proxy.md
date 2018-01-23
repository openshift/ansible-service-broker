# Running ASB behind a proxy

It is recommended to review the [OpenShift Documentation ](https://docs.openshift.com/container-platform/3.7/install_config/http_proxies.html)
and to configure a cluster accordingly before attempting to run the broker behind
a proxy. When running an Ansible Broker inside of a proxied OpenShift cluster,
it's important to understand its core concepts and consider them within the context
of a proxy used for external network access.

As an overview, the broker itself runs as a pod within the cluster. It has a requirement
for external network access depending on how its registries have been configured.
To configure the broker for external access via proxy, the cluster operator must
[edit the broker's DeploymentConfig and set the HTTP_PROXY, HTTPS_PROXY, and NO_PROXY environment variables](https://docs.openshift.com/container-platform/3.7/install_config/http_proxies.html#setting-environment-variables-in-pods).
It's common that APB pods themselves may require external access via proxy as well.
If the broker recognizes it has a proxy configuration, it will transparently
apply these env vars to the APB pods that it spawns. As long as the modules used
within the APB respect proxy configuration via environment variable, the APB
will also use these settings to perform its work. 

Finally, it's possible the services spawned by the APB may also require external
network access via proxy. The APB *must* be authored to set these environment variables
explicitly if recognizes them in its own execution environment, or the cluster
operator must manually mutate the required services to inject them into their environments.

Some important prerequisites necessary for running the broker behind a proxy:

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
