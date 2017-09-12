# Administration Examples

The Ansible service broker has three main use cases from an administration perspective. Below we will describe the main use cases and discuss the configurations for the broker that would make sense of each one. 

**Note: If you are running on a OpenShift 3.6 Cluster you can only use `auto_escalate: true`** 

More information about the broker configuration can be found [here](#config.md)

The three main use cases are:

1. [Heavy Multi-Tenant Deployment](#heavy-multi-tenant)
2. [Light Multi-Tenant Deployment](#light-multi-tenant)
3. [Very Limited Tenant Deployment](#very-limited-tenant)


## Deployments

### Heavy Multi-Tenant 
The heavy multi-tenant deployment is defined by the having many users with many different permission sets. This environments canonical example is [OpenShift Online](https://manage.openshift.com/). This deployment requires that the broker will enforce the user's permissions when attempting to run APBs for a target name space or project. The configuration values that matter are `openshift.sandbox_role` and `broker.auto_escalate`. The `sandbox_role` will be used to determine what permissions the user will need to run. The `auto_escalate` will tell the broker whether or not to run with out checking the user's permissions. **Note: `auto_escalate` being set to false is the default configuration for the broker**

#### Example Configuration  
```yaml
...
openshift:
  ...
  sandbox_role: "edit"
broker:
  ...
  auto_escalate: false
```

### Light Multi-Tenant
The light multi-tenant deployment is defined by having powerful end users that are expected to have high levels of permissions. This deployment will give the cluster administrator the choice if they want the broker to check the permissions. In this scenario, we suggest that an audit is done of the APBs that will be available. The administrator can use the registry filter configuration to explicitly remove or approve APBs.

#### Example Configuration
```yaml
registry:
- ...
  white_list:
    - "^approved-APB$"
  black_list:
    - "removed-APB$"
...
openshift:
  ...
  sandbox_role: "edit"
broker:
  ...
  auto_escalate: true # will allow all users to deploy the approved APB. Could be false if the administrator would still like the broker to check the permissions.
```

### Very Limited Tenant
The very limited tenant deployment is defined by having end users with very limited rights. This deployment will use the broker to give these users the ability to run certain actions that the cluster administrator has blessed. This will allow the cluster administrator to expose slightly more functionality without giving away more permissions than they would like. Here the cluster administrator should do a thorough analysis of the APBs that they will be offering.

#### Example Configuration
```yaml
registry:
- ...
  white_list:
    - "^approved-APB$"
  black_list:
    - "removed-APB$"
...
openshift:
  ...
  sandbox_role: "edit"
broker:
  ...
  auto_escalate: true
```