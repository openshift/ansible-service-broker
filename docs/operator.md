# Operator Management

The Automation Broker may be installed and managed by its operator, which is built using
[operator-sdk](https://github.com/operator-framework/operator-sdk) ansible support. The operator source can be found in [operator/](..operator)
at the root of our project. Once this image is built and deployed, either manually or
through OLM, the broker can be deployed by creating a CR of the `automationbroker.osb.openshift.io`
CRD.

When deploying a broker that is managed by the operator, users should configure the operator
by setting supported options in the `spec` of the CR. The operator will pick up these values and
intelligently apply them to the operator.

**Note:** if you make any changes to the broker's config
directly, the operator will overwrite them and they will be lost. If you must make changes
directly, you will need to scale down your operator so that the broker is no longer managed.
You may then edit the config directly, and must redeploy the pod so that the config re-mounts
and takes effect.

## Options
| Yaml Key | Description | Default |
|----------|-------------|---------|
| `brokerName` | Name used to identify the broker instance | `ansible-service-broker` |
| `brokerNamespace` | Namespace where the broker resides | `openshift-ansible-service-broker` |
| `brokerImage` | Fully qualified image used for the broker | `docker.io/ansibleplaybookbundle/origin-ansible-service-broker:v3.11` |
| `brokerImagePullPolicy` | Pull policy used for the broker image itself | `IfNotPresent` |
| `brokerNodeSelector` | Node selector string used for the broker's deployment | `''` |
| `registries` | Expressed as a yaml list of broker registry configs, allowing the user to configure the image registries the broker will discover and source its apbs from | See [1]
| `logLevel` | Log level used for the broker's logs | `info` |
| `apbPullPolicy` | The pull policy used for apb pods | `IfNotPresent` |
| `sandboxRole` | The role granted to the service account used to execute apbs | `admin` |
| `keepNamespace` | Controls whether the broker should delete the transient namespace created to run the apb or not after the conclusion of the apb, regardless of the result | `false` |
| `keepNamespaceOnError` | Similar to `keepNamespace`, but just controls whether or not the namespace is deleted in the event of an error result from the apb | `false` |
| `bootstrapOnStartup` | Indicates whether or not the broker should run its bootstrap routine on startup | `true` |
| `refreshInterval` | The interval of time between broker bootstraps, refreshing its inventory of apbs | `600s` |
| `launchApbOnBind` | **Experimental**: Toggles the broker executing apbs on bind operations | `false` |
| `autoEscalate` | Automatically tells the broker to escalate the permissions of a user while running the apb. Typically should remain false, since the broker will perform originating user authorization to ensure the user has permissions granted to the apb sandbox | `false` |
| `outputRequest` | Will output the low level HTTP requests the broker receives | `false` |

[1] Default registries array:
```
- type: dockerhub
  name: dh
  url: https://registry.hub.docker.com
  org: ansibleplaybookbundle
  tag: latest
  white_list:
  - ".*-apb$"
  black_list:
  - ".*automation-broker-apb$"
```

## Example CR with option overrides

```
apiVersion: osb.openshift.io/v1alpha1
kind: AutomationBroker
metadata:
  name: ansible-service-broker
  namespace: openshift-ansible-service-broker
spec:
  keepNamespace: true
  sandboxRole: edit
```
