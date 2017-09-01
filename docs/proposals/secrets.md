## Motivation

We are provisioning services in Amazon that require highly sensitive credentials to deploy.
Currently, the only way to get credentials to the APBs is by passing them as parameters. These
parameters show up as plain text in the service catalog, broker, and APBs. This workflow is also
very arduous for users, as they have to copy and paste in boilerplate credentials every time they
kick off a deployment.

In addition, a successful deployment of an Amazon Web Service requires intricate knowledge of the
environment, existing networking infrastructure resources.  This is information a typical end user
will not possess.   We would like this configuration information to be entered once by the cluster
administrator and then reused by the broker.


#### Examples ##### Sensitive Information
- ACCESS_KEY SECRET_KEY
##### AWS Environment specifics
- Region VPC_ID Subnets T.B.D. network related variables for RDS, etc.

## Proposed Solution

The Cluster Administrator will create a Secret ahead of time with the data they want to hide from
end user.  The Secret will exist in the namespace of the Broker which the end user likely does not
have access to.  The Broker will have a new section in its configuration file, called `secrets`, to
list Secrets and associate to APBs.  When the Broker runs an APB that has an associated Secret, the
Secret is mounted as a file to the running Pod and the APB is run in the Broker’s namespace.  The
APB base image will have logic to automatically read in mounted secrets and pass to all Ansible
playbook invocations.

Additionally, the APB base image will guard against these secret parameters accidently leaked in
logs through an ansible plugin monitoring stdout/logs. Lastly, the parameters of the ServiceClass
advertised to the Service Catalog will be updated to remove those parameters covered in a Secret for
the APB.  This will remove the ‘secret’ parameters from showing up in the WebUI when provisioning an
APB.


### Details

#### Broker Perspective:

A `secrets` section will be added to the broker's configuration file. The `secrets` section will
contain a list of Secrets referred to by name which must exist in the Broker namespace. Each
referenced Secret will have a list of regexes for whitelisted APBs to apply the Secret to (for first
iteration it will just be a 1-1 mapping of secret to apb name). It is possible for an APB to be
referenced in multiple secrets.

If the Broker is running an APB that matches an entry in the Secret config:
- Its pod is assigned ALL of the Secrets which match the configuration.
- The Secrets are mounted as files to a known location on the APB Pod.
- Secrets are mounted for all methods on the APB The APB runs in the namespace of the Secret.  (For
    first implementation this will be Broker namespace)
- The generated artifacts will continue to be populated in the namespace the user selects at
    provision time, it is only the ABP pod itself which will run in the namespace of the secret.
- During the /catalog operation the Broker will programmatically update the parameters associated
    with a ServiceClass/Plan.
- If the APB has an associated Secret AND the Secret exists, the Broker will read the ‘keys’ from
    the Secret and remove the matching parameters from being associated with the ServiceClass.
    - This is done to remove the WebUI from prompting for these values which will be obtained from
        Secret, yet it allows the APB author to continue to declare the parameters as they normally
        would, and in if no Secret is available the APB is still usable with those parameters
        entered.
- It is possible for a Secret to not yet exist, the APB workflow will continue as is until the
    Secret is created.  The next invocation of the /catalog operation will update the parameters to
    reflect what Secrets exist at that point in time.


#### APB Perspective

Desire is to _not_ impact an APB author with this workflow.  APB author will
declare the parameters required for the application to function as they normally would.  Details for
obtaining the Secret info is handled in the base image.

If the Secret exists for the APB, then the parameters will come from the Secret and the end user
will not be prompted for them. Otherwise, the APB will function as normal and it will be expected
end user provides values for each parameter declared.

#### APB Base Image

Secrets will be mounted to a known location on the Pod. The base image will pass secret data in
through the ansible mechanism for passing in extra variables in a file:
`-e @${PATH_TO_SECRETS}/my_secret_vars.yml`
Ensuring no leakage of parameters contained in a secret through debug/log statements We will write a
layer between the playbook and the output logs that removes any secrets from the log, guaranteeing
that they will never show up in plain text This could either be a script or a custom ansible
callback plugin.


### Summary

This approach requires no changes to the service catalog, and no changes to the APB specification.
The Broker will access the referenced Secrets to determine the ‘key’ names to update the published
parameters, yet it will not use the values, i.e. we won’t be passing in parameter values with
potential for leaking values in logs.  The Secret values will only be accessed inside of the pod
running the ansible code and mounted as a file.



### Limitations
- Cluster Administrator is required to keep Secret data accurate.
- Assumes only a single set of credentials is required, i.e. this workflow doesn't support needing
    different credentials for production and stage environments.
- Future changes may want to move the selection of secret to WebUI to support multiple environment
    configurations.
- Assumes same Secret data is supplied to all Plans of each APB Requires Secret to be the superset
    of all parameters that APB will ever need. The end user will not have a good means for seeing
    the progress of APB as it deploys.

Known Issues/Blockers:
- Open Issues on Service Catalog related to ‘syncing Broker Information’:
  - https://github.com/kubernetes-incubator/service-catalog/issues/1086
    - Mitigation Strategy:
        - Prior to creating Broker resource and instantiating
        - Have APBs published to registry
        - Create required Secrets and update Broker configuration
