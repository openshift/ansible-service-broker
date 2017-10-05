# Troubleshooting

## Introduction

The purpose of this document is to provide troubleshooting steps for different
scenarios. Where possible sections and sub-section should be created to
classify different types of troubles.

## Errors related to Service Catalog communicating with the Broker

### Certificate Authority

Looking at the output below, we can see that the broker is running but the service-catalog is unable to communicate with it
because `certificate signed by unknown authority`:

```
$ oc get pods
NAME          READY     STATUS    RESTARTS   AGE
asb-1-xzjqx   2/2       Running   0          4s

$ oc describe servicebroker
Name:           ansible-service-broker
Namespace:
Labels:         <none>
Annotations:    openshift.io/generated-by=OpenShiftNewApp
API Version:    servicecatalog.k8s.io/v1alpha1
Kind:           ServiceBroker
Metadata:
  Creation Timestamp:   2017-10-05T17:21:01Z
  Finalizers:
    kubernetes-incubator/service-catalog
  Generation:           1
  Resource Version:     21
  Self Link:            /apis/servicecatalog.k8s.io/v1alpha1/servicebrokers/ansible-service-broker
  UID:                  8f3782b5-a9f1-11e7-ab6a-0242ac110006
Spec:
  Auth Info:
    Bearer:
      Secret Ref:
        Kind:           Secret
        Name:           ansibleservicebroker-client
        Namespace:      ansible-service-broker
  Relist Behavior:      Duration
  Relist Duration:      15m0s
  Relist Requests:      0
  URL:                  https://asb.ansible-service-broker.svc:1338/ansible-service-broker/
Status:
  Conditions:
    Last Transition Time:       2017-10-05T17:22:01Z
    Message:                    Error fetching catalog. Error getting broker catalog for broker "ansible-service-broker": Get https://asb.ansible-service-broker.svc:1338/ansible-service-broker/v2/catalog: x509: certificate signed by unknown authority
    Reason:                     ErrorFetchingCatalog
    Status:                     False
    Type:                       Ready
  Operation Start Time:         2017-10-05T17:22:02Z
  Reconciled Generation:        0
Events:
  FirstSeen     LastSeen        Count   From                                    SubObjectPath   Type            Reason                  Message
  ---------     --------        -----   ----                                    -------------   --------        ------                  -------
  6s            6s              1       service-catalog-controller-manager                      Warning         ErrorFetchingCatalog    Error getting broker catalog for broker "ansible-service-broker": Get https://asb.ansible-service-broker.svc:1338/ansible-service-broker/v2/catalog: net/http: request canceled while waiting for connection (Client.Timeout exceeded while awaiting headers)
  6s            2s              9       service-catalog-controller-manager                      Warning         ErrorFetchingCatalog    Error getting broker catalog for broker "ansible-service-broker": Get https://asb.ansible-service-broker.svc:1338/ansible-service-broker/v2/catalog: x509: certificate signed by unknown authority
```

#### Resolution: Provide caBundle to service-catalog

We can get the caBundle with the following command:

```
$ oc get secret -n kube-service-catalog -o go-template='{{ range .items }}{{ if eq .type "kubernetes.io/service-account-token" }}{{ index .data "service-ca.crt" }}{{end}}{{"\n"}}{{end}}' | tail -n 1
LS0t...really long string...
```

And update the servicebroker object, adding `caBundle` to the ansible-service-broker's `Spec`:

```
$ oc edit servicebroker ansible-service-broker
servicebroker "ansible-service-broker" edited
```

### Problem: Service Catalog Invalid Credentials

In the error condition message below you can see that the Service Catalog fails
to get the catalog for the "ansible-service-broker" because of "invalid
credentials":

```
$ oc describe servicebroker
Name:           ansible-service-broker
Namespace:
Labels:         <none>
Annotations:    openshift.io/generated-by=OpenShiftNewApp
API Version:    servicecatalog.k8s.io/v1alpha1
Kind:           ServiceBroker
Metadata:
  Creation Timestamp:   2017-10-05T17:21:01Z
  Finalizers:
    kubernetes-incubator/service-catalog
  Generation:           2
  Resource Version:     40
  Self Link:            /apis/servicecatalog.k8s.io/v1alpha1/servicebrokers/ansible-service-broker
  UID:                  8f3782b5-a9f1-11e7-ab6a-0242ac110006
Spec:
  Auth Info:
    Bearer:
      Secret Ref:
        Kind:           Secret
        Name:           ansibleservicebroker-client
        Namespace:      ansible-service-broker
  Ca Bundle:            LS0t...
  Relist Behavior:      Duration
  Relist Duration:      15m0s
  Relist Requests:      0
  URL:                  https://asb.ansible-service-broker.svc:1338/ansible-service-broker/
Status:
  Conditions:
    Last Transition Time:       2017-10-05T17:22:01Z
    Message:                    Error fetching catalog. Error getting broker catalog for broker "ansible-service-broker": Status: 401; ErrorMessage: <nil>; Description: invalid credentials, corrupt header; ResponseError: <nil>
    Reason:                     ErrorFetchingCatalog
    Status:                     False
    Type:                       Ready
  Operation Start Time:         2017-10-05T17:22:02Z
  Reconciled Generation:        0
Events:
  FirstSeen     LastSeen        Count   From                                    SubObjectPath   Type            Reason                  Message
  ---------     --------        -----   ----                                    -------------   --------        ------                  -------
  38m           38m             1       service-catalog-controller-manager                      Warning         ErrorFetchingCatalog    Error getting broker catalog for broker "ansible-service-broker": Get https://asb.ansible-service-broker.svc:1338/ansible-service-broker/v2/catalog: net/http: request canceled while waiting for connection (Client.Timeout exceeded while awaiting headers)
  38m           10m             107     service-catalog-controller-manager                      Warning         ErrorFetchingCatalog    Error getting broker catalog for broker "ansible-service-broker": Get https://asb.ansible-service-broker.svc:1338/ansible-service-broker/v2/catalog: x509: certificate signed by unknown authority
  10m           17s             34      service-catalog-controller-manager                      Warning         ErrorFetchingCatalog    Error getting broker catalog for broker "ansible-service-broker": Status: 401; ErrorMessage: <nil>; Description: invalid credentials, corrupt header; ResponseError: <nil>
```

#### Resolution: Disable basic auth

What you may notice in the output of `oc describe servicebroker` is that the service-catalog is being configured to use token based authentication
to communicate with the Ansible Service Broker. However, because of [this bug](https://bugzilla.redhat.com/show_bug.cgi?id=1498992), the broker
is configured to use basic auth by default. All that we need to do is 1) update the broker's configmap and 2) redeploy the broker.

##### Update Broker's ConfigMap

We need to update the Broker's configuration to disable basic auth. Use:

```
$ oc edit configmap broker-config
```

Modifying the configuration like what you see below:

```diff
data:
  broker-config: |
    registry:
      - type: "dockerhub"
        name: "dh"
        url: "https://registry.hub.docker.com"
        user: "changeme"
        pass: "changeme"
        org: "ansibleplaybookbundle"
        tag: "latest"
        white_list:
          - ".*-apb$"
    dao:
      etcd_host: 0.0.0.0
      etcd_port: 2379
    log:
      logfile: /var/log/ansible-service-broker/asb.log
      stdout: true
      level: debug
      color: true
    openshift:
      host: ""
      ca_file: ""
      bearer_token_file: ""
      image_pull_policy: "IfNotPresent"
      sandbox_role: "edit"
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

##### Redeploy the Broker

We must redeploy the broker in order for these changes to be used:

```
$ oc deploy asb --latest
Command "deploy" is deprecated, Use the `rollout latest` and `rollout cancel` commands instead.
Flag --latest has been deprecated, use 'oc rollout latest' instead
Started deployment #2
Use 'oc logs -f dc/asb' to track its progress.
```
