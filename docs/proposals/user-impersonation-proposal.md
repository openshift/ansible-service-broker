# User Impersonation Proposal

## Introduction
We will need to add some security settings to help administrators lock down the broker. For this, we are going to check for privilege escalation and will move the location of where the APB pod will run.

## Problem Description
The problem is that currently privilege escalation is a concern for users who have access to the broker. We want to give the cluster admin who is setting up the service broker to have assurances that the broker can have some safety assurances.  

## Implementation Details
We will need to do three things to satisfy the requirements.
1. Check if the user has the permissions to cover the cluster role to be used.
2. If the cluster admin while setting up the broker has set a config value to auto escalate we will immediately continue and not check the user's permissions.
3. Create transient namespace/project
4. Create a service account with permissions to the target namespace/project.
5. Create a pod running as the service account created above.

## Work Items
- Create a new config value called auto escalate.
- Use the [SubjectRulesReview](https://docs.openshift.org/latest/rest_api/apis-authorization.openshift.io/v1.SubjectRulesReview.html) API to retrieve the rules for a user.
- Use the [k8s API's](https://godoc.org/k8s.io/client-go/kubernetes/typed/rbac/v1#ClusterRoleInterface) to retrieve the [rules](https://godoc.org/k8s.io/api/rbac/v1#PolicyRule) from the [cluster role](https://godoc.org/k8s.io/api/rbac/v1#ClusterRole).
- Use the conversion from origin to convert to the same type of rules
- Use the cover API from origin to check if the user's rules cover the service accounts
- Create a new project
- Create service account in the new project, create role binding for cluster role and service account in the target namespace/project
- Create pod running as the service account.
