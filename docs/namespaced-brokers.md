# Namespaced Automation Brokers

In the Service Catalog, brokers that conform to the Open Service Broker spec
can now be registered with the catalog as either a cluster-scoped
`ClusterServiceBroker`, or a namespace-scoped `ServiceBroker` kind. Namespaced
brokers will have their services and plans created and offered, likewise, on
a cluster wide, or namespaced basis. This enables a number of interesting use
cases. For example, a cluster administrator may want to control access to
certain services and plans, so they are able to do so by creating a namespaced
service broker that will only expose these to users that have access to that
namespace. For APB developers, they have the opportunity to install a private
namespaced broker within their own namespace so that their APBs will not display
as available to the rest of the cluster during the course of their development.

Namespaced brokers are strictly a Kubernetes Service Catalog concept. Therefore,
the Automation Broker (AB) requires no additional configuration beyond registration
as a `ServiceBroker` kind to leverage this feature.

For exhaustive documentation, [please refer to the catalog](https://github.com/kubernetes-incubator/service-catalog/blob/master/docs/namespaced-broker-resources.md).

# Installing a namespaced Automation Broker

Namespaced ABs can be installed alongside a cluster-scoped 
broker without issue. There are a few options for installing the AB as a
namespaced broker:

**IMPORTANT: A note about Automation Broker privileges**
The AB requires certain cluster level permissions that the average user may be
unable to install themselves due to limited access rights. In this case, it is
recommended that the user request a cluster administrator to install the broker
into a namespace that is owned by the user.

The admin should audit all of the roles and bindings before executing the broker's
APB on behalf of a user! To inspect the resources that the APB will grant,
the APB can be downloaded and extracted with the following:

```bash
# Tweak this image based on where your broker's APB is published
docker pull docker.io/automationbroker/automation-broker-apb
docker cp $(docker create docker.io/automationbroker/automation-broker-apb):/opt/ansible/roles/automation-broker-apb/templates /tmp`
```

The resources can then be audited from within your `/tmp/templates` directory
(or wherever you have decided to extract them).

## Installing via job

The broker's APB can be run via an install job using the following resource file:

```yaml
---
apiVersion: v1
kind: Namespace
metadata:
  name: automation-broker-apb
---
apiVersion: v1
kind: Pod
metadata:
  name: automation-broker-apb
  namespace: automation-broker-apb
spec:
  serviceAccount: automation-broker-apb
  containers:
    - name: apb
      image: docker.io/automationbroker/automation-broker-apb:latest
      args: [ "provision", "-e broker_kind=ServiceBroker" ]
      imagePullPolicy: IfNotPresent
restartPolicy: Never
```

This is a slightly modified `install.yaml` that can be [found here](../apb/install.yaml). Notice
the `broker_kind=ServiceBroker` argument. This tells the APB to install the broker
as a namespaced broker, rather than a cluster-scoped `ClusterServiceBroker`.

By default, the broker will be installed to the `automation-broker` namespace,
although it's likely users may wish to specify the namespace where the broker
will be installed. This can be specified by tweaking the APB's arguments:

`args: [ "provision", "-e broker_kind=ServiceBroker", "-e broker_namespace=my_namespace" ]`

Additionally, if the namespace does not already exist, the APB can automatically
create it as part of provision:

`args: [ "provision", "-e broker_kind=ServiceBroker", "-e broker_namespace=my_namespace", "-e create_broker_namespace=true"]`

## Usage

Once installed, you can confirm your namespaced broker has been successfully
installed by running:

`kubectl get servicebrokers -n $YOUR_BROKER_NAMESPACE` to list the namespaced
brokers that exist within that namespace.

Similarly, you can list the classes and plans available with the following:

```
kubectl get serviceclasses -n $YOUR_BROKER_NAMESPACE
kubectl get serviceplans -n $YOUR_BROKER_NAMESPACE
```

Interacting with these namespaced classes and plans is very similar to
standard cluster-scoped resources. Here is an example `ServiceInstance` that
will provision a `ServiceClass`:

```yaml
apiVersion: servicecatalog.k8s.io/v1beta1
kind: ServiceInstance
metadata:
  name: postgresql
  namespace: foo # Where the service will be provisioned
spec:
  serviceClassExternalName: dh-postgresql-apb
  servicePlanExternalName: dev
  parameters:
    app_name: "postgresql"
    postgresql_database: "admin"
    postgresql_password: "admin"
    postgresql_user: "admin"
    postgresql_version: "9.6"
```

The only difference here is the specification of the class and plan using the
`serviceClassExternalName` and `servicePlanExternalName` fields (as opposed
to the `clusterServiceClassExternalName` and `clusterServicePlanExternalName`
fields that would be used to provision a cluster class/plan).

There is no change to the way you normally deprovision, bind, and unbind.
