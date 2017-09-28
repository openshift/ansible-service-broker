## Broker authentication

The broker now supports authentication. This means that when connecting to the
broker, the caller needs to supply the [basic auth](#basic-auth) or [bearer](#bearer-auth) auth credentials for each request.
Using curl it is as simple as supplying -u username:password or -h "Authorization: beaer <token>". The service
catalog will need to be configured with a secret containing the username and
password combinations or the bearer token.

**Note: When using OpenShift 3.6 the only option for authentication is Basic Auth. Basic Auth must be enabled to true.**

### Basic Auth 
The below section will focus on the imlementation of basic auth.

#### Configuration
In order to use the broker with basic auth enabled, it needs to be enabled in the broker configuration.

```yaml
broker:
   ...
   auth:
     - type: basic
       enabled: true
```

The type field specifies the type of authentication to use.

The enabled field allows you to disable a particular auth. This keeps you from
having to delete the entire section of auth just to disable it.

##### Deployment template
Typically the broker is configured via a [ConfigMap](https://docs.openshift.com/container-platform/3.5/dev_guide/configmaps.html) in a deployment template.
You supply the auth configuration the same way as in the file configuration.

Here is an example of the [deployment template](https://github.com/openshift/ansible-service-broker/blob/master/templates/deploy-ansible-service-broker.template.yaml#L220-L222):
```
auth:
  - type: basic
    enabled: ${ENABLE_BASIC_AUTH}
```

#### Basic auth secret
There is another part to basic auth, that is the username and password used to
authenticate against the broker. While the basic auth implementation can be
backed by different back end services, the currently supported one is backed by a
[Secret](https://docs.openshift.com/container-platform/3.5/dev_guide/secrets.html).
The secret needs to be injected into the pod via a volume mount at the
`/var/run/asb_auth` location. This is where the broker will read the username
and password from.

In the [deployment template](https://github.com/openshift/ansible-service-broker/blob/master/templates/deploy-ansible-service-broker.template.yaml#L168-L175) a secret needs to be specified. See the example below:

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

The secret needs to contain username and password. The values need to be **base64**
encoded. The easiest way to generate the values for those entries is to use the
echo and base64 commands:

```bash
$ echo -n admin | base64
YWRtaW4=
```

NOTE: the `-n` option is very important

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

#### Configure the service catalog to communicate with broker
Now that we have the broker configured to use basic auth, we need to tell the
service catalog how to communicate with the broker. This is accomplished by the
`authInfo` section of the broker resource.

Here is an example of creating a broker resource in the service catalog. The
`spec` tells the service catalog what URL the broker is listening at. The
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

Starting with v0.0.17 of the service catalog the broker resource configuration changes.

```yaml
apiVersion: servicecatalog.k8s.io/v1alpha1
kind: ServiceBroker
metadata:
  name: ansible-service-broker
spec:
  url: https://asb-1338-ansible-service-broker.172.17.0.1.nip.io
  authInfo:
    basic:
      secretRef:
        namespace: ansible-service-broker
        name: asb-auth-secret
```



*NOTE*: this section is highly dependent on what the service catalog expects. If
the format for the secret changes we will need to create a separate secret for
just the service catalog today OR we need to add a new `UserServiceAdapter` that
understands that secret.

### Bearer Auth
The below section will focus on the bearer token auth.

#### Configuration
By default, if no authentication is specified the broker will use bearer token auth. The bearer token authentication will use delegated auth from the [kubernetes apiserver](https://github.com/kubernetes/apiserver) library.

The configuration is to grant access, through [kubernetes RBAC](https://kubernetes.io/docs/admin/authorization/rbac/) roles and role-bindings, to the url prefix. The broker has added a configuration option `cluster_url` to specify the url_prefx. This value but it will default to `ansible-service-broker`. 

Example cluster role:
```yaml
- apiVersion: authorization.k8s.io/v1
  kind: ClusterRole
  metadata:
    name: access-asb-role
  rules:
  - nonResourceURLs: ["/ansible-service-broker", "/ansible-service-broker/*"]
    verbs: ["get", "post", "put", "patch", "delete"]
```

#### Deployment template and secrets
Here is an example of the using the cluster role example above. [deployment template](https://github.com/openshift/ansible-service-broker/blob/master/templates/deploy-ansible-service-broker.template.yaml#L224-L248):
```yaml
- apiVersion: v1
  kind: ServiceAccount
  metadata:
    name: ansibleservicebroker-client
    namespace: ansible-service-broker

- apiVersion: authorization.openshift.io/v1
  kind: ClusterRoleBinding
  metadata:
    name: ansibleservicebroker-client
  subjects:
  - kind: ServiceAccount
    name: ansibleservicebroker-client
    namespace: ansible-service-broker
  roleRef:
    kind: ClusterRole
    name: access-asb-role

- apiVersion: v1
  kind: Secret
  metadata:
    name: ansibleservicebroker-client
    annotations: 
      kubernetes.io/service-account.name: ansibleservicebroker-client
  type: kubernetes.io/service-account-token
```

Here we are creating a service account, granting access to the `access-asb-role` and [creating a secret](https://kubernetes.io/docs/admin/service-accounts-admin/) for that service accounts token.

#### Configure the service catalog to communicate with broker
Now that we have the broker configured to use bearer token auth, we need to tell the
service catalog how to communicate with the broker. This is accomplished by the
`authInfo` section of the broker resource.

Here is an example of creating a broker resource in the service catalog. The
`spec` tells the service catalog what URL the broker is listening at. The
`authInfo` tells it what secret to read to get the authentication information.

```yaml
apiVersion: servicecatalog.k8s.io/v1alpha1
kind: ServiceBroker 
metadata:
  name: ansible-service-broker
spec:
  url: https://asb.ansible-service-broker.svc:1338${BROKER_URL_PREFIX}/
  authInfo:
    bearer: 
      secretRef:  
        kind: Secret
        namespace: ansible-service-broker,
        name: ansibleservicebroker-client
```

## Developer section

### Basic Auth design

The authentication system is built with a set of interfaces to allow for easily
adding new methods of authentication. The 3 core interfaces are: Provider,
Principal, and UserServiceAdapter. You can see these interfaces below:


```golang
// Provider - an auth provider is an adapter that provides the principal
// object required for authentication. This can be a User, a System, or some
// other entity.
type Provider interface {
    GetPrincipal(*http.Request) (Principal, error)
}

// Principal - principal user or other identity of some kind with access to the
// broker.
type Principal interface {
    GetType() string
    GetName() string
}

// UserServiceAdapter - is the interface for a service that stores Users. It can
// be anything you want: file, database, whatever as long as you can search and
// validate them.
type UserServiceAdapter interface {
    FindByLogin(string) (User, error)
    ValidateUser(string, string) bool
}
```

The `BasicAuth` struct is a `Provider` and takes in a `UserServiceAdapter`.
`BasicAuth` will return a Principal that it gets from the `UserServiceAdapter`.

At current we have one `UserServiceAdapter` implementation, the `FileUserServiceAdapter`.
This `FileUserServiceAdapter` reads from the filesystem, specifically the
username and password files located in the given directory. It knows how to
validate the username and password.

### Extending the auth system

As stated above, there are 2 core concepts `Provider` and `UserServiceAdapter`.
Let's say you want to validate users against a user database. You would create a
`DBUserServiceAdapter` that takes a DB connection, possibly a database table
name.

You could hook that `DBUserServiceAdapter` to the existing `BasicAuth`

```golang
func createProvider(providerType string, log *logging.Logger) (Provider, error) {

    switch providerType {
    case "basic":
       ...
    case "basicdb":
       dbusa := DBUserServiceAdapter{...}
       return NewBasicAuth(dbusa, log)
       ...
    }
}
```

or created a new `Provider`

```golang

// TrustedUserAuth allows for a consumer id to be passed in a clear http header.
// this should be used only if the environment is known to be secure.
type TrustedUserAuth struct {
   usa UserServiceAdapter
}

func (t TrustedUserAuth) GetPrincipal(r *http.Request) (Princpal, error) {
    userHeader := r.getHeader("cp-user")
    user, err := t.usa.FindByLogin(userHeader)
    if err != nil {
        return nil, err
    }

    // some other validation code
    return TrustedUserPrincipal{username: user.Name}, nil
}
```

### Bearer Auth design

Bearer auth is using the [kubernetes apiserver](https://github.com/kubernetes/apiserver) to do delegated authentication and authorization. Kubernetes team will keep this library up to date.

The first thing that you need to do when setting up the generic api server is to tell it the cert, key, port, and address to listen on. You are going to use the [`SecureServingOptions`](https://github.com/kubernetes/apiserver/blob/master/pkg/server/options/serving.go#L38) to set these values.

[`MaybeDefaultsWithSelfSignedCerts`](https://github.com/kubernetes/apiserver/blob/master/pkg/server/options/serving.go#L248) will set the defaults if you do not supply them. The biggest caveat here is if `SSLCert` and `SSLCertKey` are not set then the generic api server will attempt to create them. The cert that is created is also the `ca.crt`.

```golang
secureServing := genericoptions.NewSecureServingOptions()
  secureServing.ServerCert = genericoptions.GeneratableKeyCert{CertKey: genericoptions.CertKey{
    CertFile: config.Broker.SSLCert,
    KeyFile:  config.Broker.SSLCertKey,
  }}
  secureServing.BindPort = 1338
  secureServing.BindAddress = net.ParseIP("0.0.0.0")
  if err := secureServing.MaybeDefaultWithSelfSignedCerts("localhost", nil, []net.IP{net.ParseIP("127.0.0.1")}); err != nil {
    return nil, fmt.Errorf("error creating self-signed certificates: %v", err)
  }
```

You are then going to create a new [server config](https://github.com/kubernetes/apiserver/blob/master/pkg/server/config.go#L79).
```golang
  serverConfig := genericapiserver.NewConfig(Codecs)
  if err := secureServing.ApplyTo(serverConfig); err != nil {
    log.Debug("error applying to %#v", err)
    return nil, err
  }
```

To set up authentication and authorization, you will need to use the 
[`DelegatingAuthenticatorConfig`](https://github.com/kubernetes/apiserver/blob/master/pkg/authentication/authenticatorfactory/delegating.go#L41) and the [`DelegatingAuthorizationOptions`](https://github.com/kubernetes/apiserver/blob/master/pkg/server/options/authorization.go#L33). The biggest thing to notice here is we are using the `go-client` interface's for the `TokenReviews` [call](https://godoc.org/k8s.io/client-go/kubernetes/typed/authentication/v1#TokenReviewInterface). 

The other thing to note here is the `xxx.ApplyTo(serverConfig)` is what is applying the configuration to the actual server configuration.
```golang
    authn := genericoptions.NewDelegatingAuthenticationOptions()
    authenticationConfig := authenticatorfactory.DelegatingAuthenticatorConfig{
      Anonymous:               true,
      TokenAccessReviewClient: client.TokenReviews(),
      CacheTTL:                authn.CacheTTL,
    }
    authenticator, _, err := authenticationConfig.New()
    if err != nil {
      return nil, err
    }
    serverConfig.Authenticator = authenticator

    authz := genericoptions.NewDelegatingAuthorizationOptions()
    if err := authz.ApplyTo(serverConfig); err != nil {
      return nil, err
    }
```

Now you need to create the generic api server.
```golang
  return serverConfig.SkipComplete().New("ansible-service-broker", genericapiserver.EmptyDelegate)
```

The last thing to do is run the generic api server.
```golang
err = genericserver.PrepareRun().Run(wait.NeverStop)
```
