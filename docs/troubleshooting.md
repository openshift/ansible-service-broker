# Troubleshooting

## Introduction

The purpose of this document is to provide troubleshooting steps for different
scenarios. Where possible, sections and sub-section should be created to
classify different types of troubles.

## Service Catalog and Broker Communication Issues

### Problem: Certificate Authority

Sometimes the service-catalog is unable to communicate with the broker because
of an unknown certificate authority.

Looking at the output below, we see the broker is running.

```
$ oc get pods
NAME          READY     STATUS    RESTARTS   AGE
asb-1-xzjqx   2/2       Running   0          4s
```

However, in the "Status" field of the `ansible-service-broker` description
we can see there is a `certificate signed by unknown authority` preventing
the service-catalog from fetching the broker's catalog.

```
$ oc describe servicebroker ansible-service-broker
Name:           ansible-service-broker
Namespace:
Labels:         <none>
Annotations:    openshift.io/generated-by=OpenShiftNewApp
API Version:    servicecatalog.k8s.io/v1alpha1
Kind:           ServiceBroker
...
Status:
  Conditions:
    Last Transition Time:       2017-10-05T17:22:01Z
    Message:                    Error fetching catalog. Error getting broker catalog for broker "ansible-service-broker": Get https://asb.ansible-service-broker.svc:1338/ansible-service-broker/v2/catalog: x509: certificate signed by unknown authority
    Reason:                     ErrorFetchingCatalog
    Status:                     False
    Type:                       Ready
  Operation Start Time:         2017-10-05T17:22:02Z
  Reconciled Generation:        0
...
```

#### Resolution: Provide caBundle to service-catalog

We need to provide the service-catalog with the caBundle so that it can
validate the certificate signing chain. We can get the caBundle with
the following command:

```
$ oc get secret -n kube-service-catalog -o go-template='{{ range .items }}{{ if eq .type "kubernetes.io/service-account-token" }}{{ index .data "service-ca.crt" }}{{end}}{{"\n"}}{{end}}' | tail -n1
```

Once we have the `caBundle` we can update the servicebroker object, adding
`caBundle` to the ansible-service-broker's `Spec`.
Use `oc edit servicebroker ansible-service-broker` to make the change:

```diff
apiVersion: servicecatalog.k8s.io/v1alpha1
kind: ServiceBroker
...
spec:
  authInfo:
    bearer:
      secretRef:
        kind: Secret
        name: ansibleservicebroker-client
        namespace: ansible-service-broker
+ caBundle: LS0tLS1CRUdJTiBDRVJUSUZJQ0FURS0tLS0tCk1JSUM2akNDQWR...
  relistBehavior: Duration
  relistDuration: 15m0s
  relistRequests: 0
  url: https://asb.ansible-service-broker.svc:1338/ansible-service-broker/
```

### Problem: Service Catalog Invalid Credentials

If the service-catalog does not have the proper credentials, it will not be
able to communicate with the broker.

Looking at the output below, we see the broker is running.

```
$ oc get pods
NAME          READY     STATUS    RESTARTS   AGE
asb-1-xzjqx   2/2       Running   0          4s
```

However, in the "Status" field of the `ansible-service-broker` description
we can see the service-catalog is using `invalid credentials`, preventing
the service-catalog from fetching the broker's catalog. The "Spec" field
shows that the service-catalog is configured to use token based authentication
to communicate with the broker.

```
$ oc describe servicebroker ansible-service-broker
Name:           ansible-service-broker
Namespace:
Labels:         <none>
Annotations:    openshift.io/generated-by=OpenShiftNewApp
API Version:    servicecatalog.k8s.io/v1alpha1
Kind:           ServiceBroker
...
Spec:
  Auth Info:
    Bearer:
      Secret Ref:
        Kind:           Secret
        Name:           ansibleservicebroker-client
        Namespace:      ansible-service-broker
  Ca Bundle:            LS0tLS1CRUdJTiBDRVJUSUZJQ0FURS0tLS0tCk1JSUM2akNDQWR...
...
Status:
  Conditions:
    Last Transition Time:       2017-10-05T17:22:01Z
    Message:                    Error fetching catalog. Error getting broker catalog for broker "ansible-service-broker": Status: 401; ErrorMessage: <nil>; Description: invalid credentials, corrupt header; ResponseError: <nil>
    Reason:                     ErrorFetchingCatalog
    Status:                     False
    Type:                       Ready
  Operation Start Time:         2017-10-05T17:22:02Z
  Reconciled Generation:        0
...
```

Look back at the "Spec" field of the `ansible-service-broker` description to
see that the service-catalog is configured to use token based authentication
when communicating with the broker and the "auth" field of the `broker-config`
ConfigMap confirms the broker has basic auth enabled.

```
$ oc get configmap broker-config -o yaml
apiVersion: v1
kind: ConfigMap
data:
  broker-config: |
  ...
    broker:
      dev_broker: True
      bootstrap_on_startup: true
      refresh_interval: "600s"
      launch_apb_on_bind: False
      output_request: False
      recovery: True
      ssl_cert_key: /etc/tls/private/tls.key
      ssl_cert: /etc/tls/private/tls.crt
      auto_escalate: True
      auth:
        - type: basic
          enabled: True
...
```

#### Resolution: Disable Basic Auth

All that we need to do is 1) update the broker's ConfigMap and 2) redeploy the broker.

Update the `broker-config` ConfigMap using `oc edit configmap broker-config` to
disable basic auth by setting the "enabled" field to `false`.

```diff
data:
  broker-config: |
    ...
    broker:
      dev_broker: true
      bootstrap_on_startup: true
      refresh_interval: "600s"
      launch_apb_on_bind: false
      output_request: true
      recovery: true
      ssl_cert_key: /etc/tls/private/tls.key
      ssl_cert: /etc/tls/private/tls.crt
      auto_escalate: true
      auth:
        - type: basic
+           enabled: false
-           enabled: true
```

Redeploy the broker using origin clients `rollout latest` command.

```
$ oc rollout latest asb
deploymentconfig "asb" rolled out
```
