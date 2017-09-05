# User Impresonation Proposal

## Introduction
We will need to add some security settings to help administrators lock down the broker. For this we are going to check for privillage escalation, and will move the location of where the APB pod will run.

## Problem Description
The problem is that currently privillage escalation is a concern for users who have access to the broker. We want to give the cluster admin who is setting up the service broker to have assurances that the broker can have some saftey assurances. 

## <Implementation Details>
We will need to do three things to satisfy the requirements.
1. Check if the user has the premissions to cover the cluster role to be used.
2. If the cluster admin while setting up the broker has set a config value to auto escalate we will imdediatly continue and not check the users permissions.
3. Create transiant namespace
4. Create a service account with permissions to the target namespace.
5. Create a pod running as the service account created above.

## Work Items
- Create a new config value called auto escalate.
- Use the SubjectRulesReview api to retrieve the rules for a user.
- Use the k8s api's to retrieve the rules from the cluster role.
- Use the conversion from origin to convert to same type of rules
- Use the cover api from origin to check if the users rules cover the service accounts
- Create a new project
- Create service account in the new project, create role binding for cluster role and service account in the target namespace
- Create pod running as service account.
