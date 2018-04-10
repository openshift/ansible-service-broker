# Service Bundle Specification

A Service Bundle is a container image that the broker can use to manage
the deployment of a service in a cluster. It adheres to the properties and
behaviors described below.

## Label

Every service bundle container image has a label ‘com.redhat.apb.spec’ that
contains a base64 encoded string. This label holds [important
metadata](https://github.com/ansibleplaybookbundle/ansible-playbook-bundle/blob/master/docs/developers.md#apb-spec-file)
related to the service bundle. Example:

```
version: 1.0        # The version of the Service Bundle Spec being conformed to
name: example-apb   # The name of the Service Bundle
description:        # A short description of the Service Bundle
bindable: True      # Whether this Service Bundle can be bound to other services
async: optional     # Does this Service Bundle support asynchronous provision
metadata:
  documentationUrl: <link to documentation>
  imageUrl: <link to URL of image>
  dependencies:
  - '<registry>/<organization>/<dependency-name-1>'
  - '<registry>/<organization>/<dependency-name-2>'
  displayName: Example App (APB)
  longDescription: A longer description of what this APB does
  providerDisplayName: "Red Hat, Inc."
plans: # An array of plans supported by this Service Bundle (must have at least 1)
  - name: default
    description: A short description of what this plan does
    free: true
    metadata:
      displayName: Default
      longDescription: A longer description of what this plan deploys
      cost: $0.00
    Parameters: # Parameters for this plan
      - name: parameter_one
        required: true
        default: foo_string
        type: string
        title: Parameter One
        maxlength: 63
      - name: parameter_two
        required: true
        default: true
        title: Parameter Two
        type: boolean
```

## Input

The container is run with three arguments.

* Action: a single word such as “provision”. The full list is below.
* --extra-vars: This is ansible-specific and always has exactly the value
  “--extra-vars”. An Ansible Playbook Bundle uses this and the third argument
  as-is when calling ansible by appending them to the ansible command.
* JSON: a json document of information that is useful to the bundle. Currently
  it is expressed in a form that is useful for populating variables in ansible
  that can be used by roles.

### Actions

Possible actions include
* provision
* deprovision
* bind
* unbind
* update
* test

### JSON
Example of JSON document passed to __provision__:

```
{
    "cluster": "openshift",
    "_apb_plan_id": "default",
    "_apb_service_class_id": "c23ec213bb8dea1577230c5ce005b9c2",
    "_apb_service_instance_id": "54636ad3-0378-49a1-a494-f97d0a0daf8e",
    "_apb_last_requesting_user": "admin",
    "mediawiki_admin_pass": "pass",
    "mediawiki_admin_user": "admin",
    "mediawiki_db_schema": "mediawiki",
    "mediawiki_site_lang": "en",
    "mediawiki_site_name": "MediaWiki",
    "namespace": "bazproject"
}
```

* cluster: the type of cluster on which the service bundle is being run
  (currently either ‘openshift’ or ‘kubernetes’).
* _apb_plan_id: the name of a “plan”, as defined by the OSB API, that is
  available on the service class.
* _apb_service_class_id: the “service_id”, as defined through the OSB API, that
  is being used for the current action.
* _apb_service_instance_id: the “instance_id”, as defined through the OSB API,
  that uniquely identifies the service instance being acted upon.
* namespace: the target namespace in which resources should be created,
  deleted, or acted upon.

Other keys represent parameter names from the Plan, the values of which have
the type specified in the Plan.

Example of JSON document passed to __bind__:

```
{
    "_apb_plan_id": "dev",
    "_apb_provision_creds": {
        "DB_ADMIN_PASSWORD": "yiX2hXH7RU24RTVRXuye",
        "DB_HOST": "postgresql",
        "DB_NAME": "foo",
        "DB_PASSWORD": "letmein",
        "DB_PORT": "5432",
        "DB_TYPE": "postgres",
        "DB_USER": "bar"
    },
    "_apb_service_class_id": "1dda1477cace09730bd8ed7a6505607e",
    "_apb_service_instance_id": "d6ac6b10-fbff-4944-8e6b-3478313e20d1",
    "_apb_service_binding_id": "54636ad3-0378-49a1-a494-f97d0a0daf8e",
    "_apb_last_requesting_user": "admin",
    "cluster": "kubernetes",
    "namespace": "default"
}
```

* _apb_provision_creds: A set of credentials and related data created during a
  provision action for this service instance. Details are below in the Output
  section.

#### Last Request User

The requesting username of the [actions](#actions) _provision_, _deprovision_, _bind_, _unbind_, and _update_ is available in the `_apb_last_requesting_user` parameter. The parameter is set to the `UID` if the  `username` of the action is empty (e.g. auto escalation). However, this parameter may be completely empty as it's not a requirement to send the requesting user information. The user for the current action may be different from the users that initiated any previous actions on the same resource. The _apb_last_requesting_user will always reflect the username of this current action.

This field was introduced in version 1.2

### Environment Variables

The following environment variables are set by the broker:

* __POD_NAMESPACE__: the namespace in which the service bundle is being run
* __POD_NAME__: the name of the pod

## Output

### stdout / stderr

These are ignored.

### Exit Code

* 0 - success
* 1 - error
* 8 - action not implemented/supported

### Encoded Credentials

Binding credentials may be returned during a provision or bind operation by
creating a named secret (POD_NAME) in the sandbox namespace (POD_NAMESPACE).
The secret should have a single key called “fields” whose value is a
base64-encoded JSON document.

Ansible Playbook Bundles use the “asb_encode_binding” module to do this.

```
apiVersion: v1
data:
  fields: eyJEQl9OQU1FIjogImZvbyIsICJEQl9QQVNTV09SRCI6ICJsZXRtZWluIiwgIkRCX1RZUEUiOiAicG9zdGdyZXMiLCAiREJfUE9SVCI6ICI1NDMyIiwgIkRCX1VTRVIiOiAiYmFyIiwgIkRCX0hPU1QiOiAicG9zdGdyZXNxbCJ9
kind: Secret
metadata:
  creationTimestamp: 2018-02-26T18:34:50Z
  name: apb-874ae78d-acfe-4181-b064-124d6e13c592
  namespace: dh-postgresql-apb-prov-md5xt
  resourceVersion: "11846"
  selfLink: /api/v1/namespaces/dh-postgresql-apb-prov-md5xt/secrets/apb-874ae78d-acfe-4181-b064-124d6e13c592
  uid: ba79e27a-1b23-11e8-aebf-94b96c9ea8ca
type: Opaque
```

The value for “fields” is a base64-encoded JSON document whose keys and values
represent any data the APB wants to return to the broker and ultimately to a
client that retrieves a binding.

__Provision Credentials__ - If a service bundle saves such a secret during the
provision operation, the broker saves the passed data as “provision
credentials”. That data gets passed into all further actions as part of the
JSON document under the key ``_apb_provision_creds``, as a sub-document. As such,
it can be useful for making root-level credentials created during provision
available to other operations that need to use them. If a request for non-async
binding is made and the bundle does not run, any “provision credentials” that
were saved will be returned to the client as binding credentials.

__Binding Credentials__ - If a service bundle saves such a secret during the bind
operation, the broker saves the passed data as “binding credentials”. These
credentials get returned to the client that initiated that specific binding
operation.

## Runtime

The container will be run in a pod by itself with ``restartPolicy: Never``. The
pod will run in a sandbox namespace created by the broker for that specific
operation only. The namespace is not intended to persist, and in normal
operation it will be deleted as soon as the pod completes. The broker can
optionally be configured to retain the namespace for debugging purposes.

The pod is run with a service account granted access to the ``POD_NAMESPACE`` and
to the “namespace” that the pod is to deploy to. Networks between the two
namespaces are tied together so that you can reliably find your services in the
target namespace. That only works for
[NetworkPolicy](https://kubernetes.io/docs/concepts/services-networking/network-policies/)
or for multi tenant openshift-sdn.

### Proxy

If there are proxy settings for the broker, they will be forwarded to the pod
by setting environment variables for ``HTTP_PROXY``, ``HTTPS_PROXY``,
``NO_PROXY``, ``http_proxy``, ``https_proxy``, and ``no_proxy``.

No action is taken to apply proxy settings to pods that a service bundle
creates. Each service bundle is repsonsible for configuring its provisioned
resources to use a proxy as appropriate.

### Service Account

Details of a kubernetes service account are mounted at
``/var/run/secrets/kubernetes.io/serviceaccount`` by Kubernetes. That mount is
documented [here](https://kubernetes.io/docs/admin/service-accounts-admin/).

Clients running in a Service Bundle will usually need a ``.kube/config`` file
such as the following to use the service account:

```
apiVersion: v1
clusters:
- cluster:
    certificate-authority: /var/run/secrets/kubernetes.io/serviceaccount/ca.crt
    server: https://kubernetes.default:443
  name: kubernetes
contexts:
- context:
    cluster: kubernetes
    user: apb-user
  name: /kubernetes/apb-user
current-context: /kubernetes/apb-user
kind: Config
preferences: {}
users:
- name: apb-user
  user:
    tokenFile: /var/run/secrets/kubernetes.io/serviceaccount/token
```
