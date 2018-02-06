## Proposal: Move To Custom Resources Definitions For Persisting Data
The slides that describe the problem in more detail can be found [here](https://docs.google.com/presentation/d/1UQU4BWlLGw70KQEwqNTZghxQVl0dTzyWLGtQ9z1DLx8/edit?usp=sharing).

### Problem Description
Currently, the Service broker is deployed with an etcd pod, that the broker connects to over the cluster network using mutual TLS Auth. The etcd pod also needs a persistent volume to store its data. This current deployment strategy makes production deployments and upgrades too complicated and requires a lot of edge case handling just for the Ansible Service Broker.

To solve this problem we are proposing to use [CRDs (Custom Resource Definitions)](https://kubernetes.io/docs/tasks/access-kubernetes-api/extend-api-custom-resource-definitions/) to store the broker's data. The CRDs will mimic the broker types(ServiceInstance, BindingInstance, JobState, Specs...). We will also need to create a `dao` interface and CRD dao implementation. The CRD dao implementation will need to connect to the defined APIs that the k8s cluster will now know, once the CRDs are created. This client can be generated or we could do something similar to what we did with the openshift client. 

This will improve the broker because it will no longer be necessary to create a PV during deployment, we will not need to worry about data migration during updates. This will also remove cert generation for mutual TLS auth for our own etcd. 

### Work Items

- [ ] 1. Create a DAO interface PR
- [ ] 2. Design and make CRDs encapsulate our data objects should be a declarative resource with a Spec and Status field. Add to templates (deploy templates for both k8s and openshift. Should also add to openshift-ansible)
- [ ] 3. Generate or create client for new CRDs
- [ ] 4. Create new implementation of DAO
- [ ] 5. Create config value for DAO to specify type default to new DAO implementation


### Outstanding Questions
1. Should we generate our own client or write our own. [This](https://blog.openshift.com/kubernetes-deep-dive-code-generation-customresources/) is a nice overview of client generation. [This](https://github.com/kubernetes/sample-controller) is an example of client generation. [This](https://github.com/openshift/ansible-service-broker/blob/master/pkg/clients/openshift.go) is an example of creating our own.
2. We should discuss where this client lives. Should we create a new repo for this client so that other users could interact with our resources? This may eventually be a nice to have but could also cause headaches.  
