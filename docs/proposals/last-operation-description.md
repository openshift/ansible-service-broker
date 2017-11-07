# Last Operation Description

## Introduction
As per the OpenService Broker API spec, the last operation response can contain
a description as well as a status [last operation response](https://github.com/openservicebrokerapi/servicebroker/blob/v2.12/spec.md#body).
This can provide useful information about what is happening during a broker action along with details of the overall progress.
UI implementors could make valuable use of this information
to provide a progress bar or just simply a log of actions to show progress.
```
60%: succesfully created realm in keycloak
```

## Problem Description
Currently we are only making use of the status field: 
[last operation in broker.go](https://github.com/openshift/ansible-service-broker/blob/master/pkg/broker/broker.go#L1311)
The difficulty is around how we get that detail back from the pod.
Without the description, we limit the feedback that can be provided to the user.


## Proposed Implementation
Create an append only log file in a specified location ```/var/log/last_operation```, which is collected periodically and also a final time after
the apb has completed but before the pod is deleted. 
In the bind workflow, we already do something similar by execing into the container in order to collect
the bind credentials using  [monitor output in ext_cred.go](https://github.com/openshift/ansible-service-broker/blob/master/pkg/apb/ext_creds.go#L53). 
A similar workflow could be used to gather and store the last operation description.

### Description Format
If we use an append only log, then we can do something like this:
```
10%: creating deployment and routes,30%: waiting for service to become available, 50%: retrieved token from API,60%: Unexpected error creating realm in keycloak
```
This could then easily be consumed by a polling client.

### Broker changes
Below are a set of changes that I believe are in line with the current design. The exact implementation would likely differ but the gist would 
be the same.

- A new field would be added to the JobState type and also to the different message types, ProvisionMsg for example:
```Description string ``` 
The string value gathered from the file in the apb container would be stored here.

- Modify ExecuteApb in to add the new volume mount ```/var/log/last_operation```.

- Modify or add a new method similar to monitor output that would gather the information in the background as the apb pod was running

- Add a new method or refactor ExtractCredentials to extract the last operation log. Likely a new method as we would want to send this information back
each time we collected it.

- Pass the log in the msg buffer worker chanel or add a new channel specifically for the last operation log (this channel would need to be passed into the different action such as provision)

- In the subscribers where the state is updated, pull out the description and add it to the stored jobState.

- Modify last operation handler to pull the description out of the Job state and send it back as part of the response.

### APB changes
Add a module that would handle putting the content from a last operation description into the right place on disk. 

Likely something very similar to the encode binding module:
https://github.com/ansibleplaybookbundle/ansible-asb-modules/blob/master/library/asb_encode_binding.py

Something like the following may make sense:

```
 asb_last_operation_description:
   description:"10%: creating deployment and routes"
```   

***Note not very familiar with how the ansible apb works under the hood so would need some guidance here***
