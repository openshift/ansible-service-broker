# Broker Authentication

This document attempts to outline the high-level plan for adding authentication
to the broker api routes.

When talking to service brokers, the [Service Catalog uses basic auth](#service-catalog-basic-auth). The Ansible Service Broker will implement Basic Auth using a secret shared between the catalog and the broker. The authentication will be configurable allowing a user i.e. developer to disable it.

The plan for this sprint is to do the following:

* make auth routes configurable by adding an auth subsection to the broker section
```
broker:
  ...
  auth:
    - type: basic | INSERT OTHER TYPE HERE
    - enable: true | false
```
  * update the following files to reflect the new entry:
      * [deploy-ansible-service-broker.template.yaml](https://github.com/openshift/ansible-service-broker/blob/master/templates/deploy-ansible-service-broker.template.yaml)
      * [prep_local_devel_env.sh](https://github.com/openshift/ansible-service-broker/blob/master/scripts/prep_local_devel_env.sh) script
      * [example-config.yaml](https://github.com/openshift/ansible-service-broker/blob/master/etc/example-config.yaml)
      * [README.md](https://github.com/openshift/ansible-service-broker/blob/master/README.md)
      * docs directory:
          * [deployment.md](https://github.com/openshift/ansible-service-broker/blob/master/docs/deployment.md)
          * [config.md](https://github.com/openshift/ansible-service-broker/blob/master/docs/config.md)
          * [local_deployment.md](https://github.com/openshift/ansible-service-broker/blob/master/docs/local_deployment.md)l
      * [`make deploy`](https://github.com/openshift/ansible-service-broker/blob/master/Makefile#L57-L58) does not need updating because it uses the above template.
  * [catasb](https://github.com/fusor/catasb/tree/dev)
      * pulls down the `deploy-ansible-service-broker.template.yaml`
      * may have to update all.yml in case we need to override the auth setting

* create a secret that both the broker and service catalog can use

* secret rotation is desired but needs more research, see [#open-issues](#open-issues)

* use [`http.BasicAuth`](https://golang.org/pkg/net/http/#Request.BasicAuth) to handle the basic auth case
  * BasicAuth returns (username, password string, ok bool)
  * Compare against the values from the secret
      * if they match, proceed with call
      * if no match, return [401 Unauthorized](https://golang.org/pkg/net/http/#pkg-constants)

## Future improvements to authentication
* the [OpenServiceBroker API proposes](https://github.com/openservicebrokerapi/servicebroker/pull/223) making auth optional, but also allow things like [OAuth](https://oauth.net/2/), or other mechanism.
* the Service Catalog has an issue open to implement [JSON Web Token (JWT)](https://jwt.io/): [issue #990](https://github.com/kubernetes-incubator/service-catalog/issues/990)

## Open Issues

* reloading secrets is different between kubernetes and openshift. [see updating secrets](#updating-secrets)
    * might need to reconcile the secret on an interval if not automatic
* shared secrets are mysterious. [OpenShift docs](https://github.com/openshift/openshift-docs/blob/master/dev_guide/secrets.adoc#creating-secrets) state:
> Secret data can be shared within a namespace
Except that the broker and the service-catalog run in different namespaces. How
do we *share* a secret in this case?

## Relevant information

### Updating secrets

[Kubernetes Secrets](https://kubernetes.io/docs/concepts/configuration/secret/#creating-your-own-secrets)

> Mounted Secrets are updated automatically
>
> When a secret being already consumed in a volume is updated, projected keys are eventually updated as well. Kubelet is checking whether the mounted secret is fresh on every periodic sync. However, it is using its local ttl-based cache for getting the current value of the secret. As a result, the total delay from the moment when the secret is updated to the moment when new keys are projected to the pod can be as long as kubelet sync period + ttl of secrets cache in kubelet

But the latest [OpenShift docs](https://github.com/openshift/openshift-docs/blob/master/dev_guide/secrets.adoc#dev-guide-secrets-using-secrets) states something different entirely:

> When you modify the value of a secret, the value (used by an already running pod) will not dynamically change. To change a secret, you must delete the original pod and create a new pod (perhaps with an identical PodSpec).
>
> Updating a secret follows the same workflow as deploying a new container image. You can use the kubectl rolling-update command.

It would be really useful if we could do the automatic updates, then the broker
could always read the secret and always get the latest.


### Broker spec authentication
The current [OpenServiceBroker API authentication](https://github.com/openservicebrokerapi/servicebroker/blob/master/spec.md#authentication) uses basic auth or nothing. The marketplace in our world is the service catalog.
> The marketplace MUST authenticate with the service broker using HTTP basic authentication (the Authorization: header) on every request. The broker is responsible for validating the username and password and returning a 401 Unauthorized message if credentials are invalid. It is RECOMMENDED that brokers support secure communication from platform marketplaces over TLS.


### Service Catalog Basic Auth

The Service Catalog calls [req.SetBasicAuth(c.username, c.password)](https://github.com/kubernetes-incubator/service-catalog/blob/master/pkg/brokerapi/openservicebroker/open_service_broker_client.go#L121) when calling the broker.

We will create a secret with the username and password in it. Then
specify said secret in the `authInfo` field when creating the `Broker`
resource in the Service Catalog.

```yaml
apiVersion: servicecatalog.k8s.io/v1alpha1
kind: Broker
metadata:
  name: test-broker
spec:
  url: http://beefco.de
  # put the basic auth for the broker in a secret, and reference the secret here.
  # service-catalog will use the contents of the secret. The secret should have "username"
  # and "password" keys
  authInfo:
    basicAuthSecret:
      namespace: some-namespace
      name: secret-name
```

### Kubernetes Authentication

Not sure if this is relevant. It's regarding authenticating to the kubernetes
cluster.
[Kubernetes Authentication
Strategies](https://kubernetes.io/docs/admin/authentication/#authentication-strategies)
