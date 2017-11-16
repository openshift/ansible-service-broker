# Last Operation Description

## Introduction
As per the OpenService Broker API spec, the last operation response can contain
a description as well as a status: [last operation response](https://github.com/openservicebrokerapi/servicebroker/blob/v2.12/spec.md#body).
This can provide useful information about what is happening during a broker action along with details of the overall progress.
UI implementors could make valuable use of this information to provide an indicator of progress and provide feedback while long running
actions are in progress. An example of a possible description is shown below:
```
60%: succesfully created realm in keycloak
```

## Problem Description
Currently we are only making use of the status field: 
[last operation in broker.go](https://github.com/openshift/ansible-service-broker/blob/master/pkg/broker/broker.go#L1311)
The difficulty is around how we get that detail back from the pod.
Without the description, we limit the feedback that can be provided to the user for actions performed against the service catalog.

## Expectations
- The last_operation demonstrates APB progress.
- No restrictions or requirements for user to demonstrate APB progress.
- No guarantee that last_operation shows every operation.
- The final_operation is the last_operation gathered before the APB is deleted.


## Terms

**Last Operation:**  the most recent operation an apb performed 

**Final operation:** the operation that indicates the apb is 100% complete
                     

## Proposed Solution

Using a new apb module, allow for a description to be added by the apb developer for the last operation the apb took along with the final operation the apb took. This module would take advantage of env vars,  provided to it via the downward api, that reference the pod name and namespace the apb is executing within. 
A PR is already in place to expose this information: https://github.com/openshift/ansible-service-broker/pull/546
When called this apb module would update known annotations on the pod ie: ```apb_last_operation``` and ``` apb_final_operation```
with the description provided by the apb developer. This would be part of the 
[ansible playbook modules](https://github.com/ansibleplaybookbundle/ansible-asb-modules).

In order to collect this information we would use a watch via the Kubernetes client on the pod resource within the temporary namespace [Pod Rest API](https://docs.openshift.com/container-platform/3.5/rest_api/kubernetes_v1.html#list-or-watch-objects-of-kind-pod).
This would allow us to react to changes made (i.e to the annotations) on Pod Object. Whenever a change occurred, an update to the JobStatus would happen. If the ```apb_final_operation``` annotation is present this would take precedence over the last_operation annotation. 
Once the pod was deleted we would stop the watch on the pod and update the JobStatus ```final_operation``` annotation value.

Sudo code example

```go

wi, err := k8client.CoreV1().Pods(ns).Watch(meta_v1.ListOptions{})
changes := wi.ResultChan()
    for ch := range changes{
		if ch.Type == watch.Modified{
			...
		}
		if ch.Type == watch.Deleted{
			close(ch)
			...
		}
	}

```

As this would block, it would need to be done in a background go routine. Using a watch in a background routine should allow us to update the JobState independent of the actual execution of the apb.    


### Broker changes
Below are a set of changes that I believe are in line with the current design. The exact implementation would likely differ but the gist would be the same.

- A new field would be added to the JobState type:
```Description string ``` 
The string value , added by the apb module and gathered from the pod annotation would be stored here.

- Modify the existing provision and deprovision subscribers along with the corresponding work messages.
A namespace field would be added to these messages as the pod name is already present. 
Inside the subscriber, the go routine to watch the pod and update the JobStatus would start when a new work msg was recieved.
This routine would stop once the a deleted change was received in the watch.

- Add a new subscriber for bind
async bind will be part of the service catalog, so having a Status and Description for an async binding will also be needed.
Adding a new subscriber and workmsg for binding operations will allow us to update the JobStatus once async binding arrives.

- Modify last operation handler to pull the description out of the Job state and send it back as part of the response.

### APB changes
Add a module that would handle putting the content from a last operation description into the right place on disk. 

Something like the following may make sense:

```
 asb_last_operation_description:
   description:"10%: creating deployment and routes"

asb_final_operation_description:
   description:"100%: keycloak succesfully provisioned"   
```   

***Note not very familiar with how the ansible apb works under the hood so would need some guidance here***
