# User Impersonation Proposal

## Introduction
We will need to add some security settings to help administrators lock down the broker. For this, we are going to check for privilege escalation and will move the location of where the APB pod will run.

## Problem Description
The problem is that currently privilege escalation is a concern for users who have access to the broker. We want to give the cluster admin who is setting up the service broker the ability to not allow users to have privilege escalation.

## Implementation Details
We will need to do 5 things to satisfy the requirements.
1.  Check if the user has the permissions to cover the cluster role to be used. If the `auto_escalate` option is enabled in the broker config, we will immediately continue and not check the user's permissions.
2. Create transient namespace/project.
3. Create a service account with permissions to the target namespace/project.
4. Create a pod running as the service account created above.
5. Delete service account, role binding, and transient namespace/project.

## Work Items
- Create a new config value called auto escalate.
- Use the [SubjectRulesReview](https://docs.openshift.org/latest/rest_api/apis-authorization.openshift.io/v1.SubjectRulesReview.html) API to retrieve the rules for a user.
- Use the [k8s API's](https://godoc.org/k8s.io/client-go/kubernetes/typed/rbac/v1#ClusterRoleInterface) to retrieve the [rules](https://godoc.org/k8s.io/api/rbac/v1#PolicyRule) from the [cluster role](https://godoc.org/k8s.io/api/rbac/v1#ClusterRole).
- Use the conversion from origin to convert to the same type of rules.
- Use the cover API from origin to check if the user's rules cover the service accounts.
- Create a new project.
- Create service account in the new project, create role binding for cluster role and service account in the target namespace/project.
- Create pod running as the service account.


## Of Importance To Note
The current proposal will achieve the re-use of the conversion and cover methods from origin by copying files to the broker. This particular path will require vendor being updated. This also means that we **not** be vendoring all of origin to get the functions that we need, but will be making copies.

#### Pros Of Not Vendoring Origin
* The broker still remains independent of origin.
* We could have written the functions ourselves, but they already had them. Updating wording to address PR comments.
* The vendor for ASB is still manageable. If we bring in origin we will bring in many dependencies that we don not even need.

#### Cons OF Not Vendoring Origin
* Code is copied, which never feels like the right decision.
* We need to change the code that is copied slightly to make it work correctly.
