## Broker authentication

The broker now supports authentication. This means that when connecting to the
broker, the caller needs to supply the basic auth credentials for each request.
Using curl it is as simple as supplying -u username:password. The service
catalog will need to be configured with a secret containing the username and
password combinations.

### Configuration
In order to use the broker with auth enabled, it needs to be enabled in the
broker configuration.

```yaml
broker:
   ...
   auth:
     - type: basic
       enabled: true
```

The type field specifies the type of authentication to use. At the moment we
only support basic auth. There is a desire to support other types like oauth,
bearer token, certificates, etc.

The enabled field allows you to disable a particular auth. This keeps you from
having to delete the entire section of auth just to disable it.

#### Deployment template
Typically the broker is configured via a
[ConfigMap](https://docs.openshift.com/container-platform/3.5/dev_guide/configmaps.html) in a deployment template.
You supply the auth configuration the same way as in the file configuration.

[deployment template](https://github.com/openshift/ansible-service-broker/blob/master/templates/deploy-ansible-service-broker.template.yaml#L195-L197)

### Basic auth secret
There is another part to basic auth, that is the username and password used to
authenticate against the broker. While the basic auth implementation can be
backed by different backend services, the currently supported one is backed by a
[Secret](https://docs.openshift.com/container-platform/3.5/dev_guide/secrets.html).
The secret needs to be injected into the pod via a volume mount at the
`/var/run/asb_auth` location. This is where the broker will read the username
and password from.

In the [deployment template](https://github.com/openshift/ansible-service-broker/blob/master/templates/deploy-ansible-service-broker.template.yaml#L195-L197) a secret needs to be specified. See the example below:

```yaml
- apiVersion: v1
  kind: Secret
  metadata:
    name: asb-auth-secret
    namespace: ansible-service-broker
  data:
    username: ${BROKER_USER}
    password: ${BROKER_PASS}
```

The secret needs to contain username and password. The values are base64
encoded. The easiest way to generate the values for those entries is to use the
echo and base64 commands:

```
$ echo -n admin | base64
YWRtaW4=
```

This secret now needs to be injected to the pod via a volume mount. This is
configured in the deployment template as well.

```yaml
spec:
  serviceAccount: asb
  containers:
  - image: ${BROKER_IMAGE}
    name: asb
    imagePullPolicy: IfNotPresent
    volumeMounts:
      ...
      - name: asb-auth-volume
        mountPath: /var/run/asb-auth
...
```

Then in the `volumes` section mount the secret:

```yaml
volumes:
  ...
  - name: asb-auth-volume
    secret:
      secretName: asb-auth-secret
...
```
So the above will have created a volume mount located at `/var/run/asb-auth`.
This volume will have two files: username and password written by the
`asb-auth-secret` secret.

### Configure the service catalog to communicate with broker
Now that we have the broker configured to use basic auth, we need to tell the
service catalog how to communicate with the broker. This is accomplished by the
`authInfo` section of the broker resource.

Here is an example of creating a broker resource in the service catalog. The
`spec` tells the service catalog what url the broker is listening at. The
`authInfo` tells it what secret to read to get the authentication information.

```yaml
apiVersion: servicecatalog.k8s.io/v1alpha1
kind: Broker
metadata:
  name: ansible-service-broker
spec:
  url: https://asb-1338-ansible-service-broker.172.17.0.1.nip.io
  authInfo:
    basicAuthSecret:
      namespace: ansible-service-broker
      name: asb-auth-secret
```


## Developer section

* how to configure the broker
  * enable basic auth
  * disable basic auth
* how to configure the broker resource
* how to update the username and password in the secret

    Developer section
* how to add a new auth to the broker
