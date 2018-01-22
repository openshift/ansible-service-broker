# Bind Job State

## Introduction

Asynchronous binding and unbinding is proposed as a set of
[changes](https://github.com/openservicebrokerapi/servicebroker/pull/334/) to
the OSB API. This proposal to our broker enables it to track the state of a
Service Binding, whether its creation is in-progress or complete. Knowing this
state is a requirement for implementing the proposed OSB API changes.

## Problem Description

Our broker has a limitation that when a request to create a Service Binding
comes in, we can tell if there is already a ``BindInstance`` in our data store
with the same ID, but we don't know if it is already completed, or if it is
in-progress.

This inability to distinguish a completed vs. in-progress ``BindInstance`` led
in part to a
[bug](https://github.com/openshift/ansible-service-broker/issues/670) that
always assumes a ``BindInstance`` is completed if it is found in etcd.

Our broker also likely behaves incorrectly in response to the "Get Service
Binding" endpoint. When a GET request for a Service Binding is received, the
OSB API spec says that a 404 Not Found response "MUST be returned if the
Service Binding does not exist or if a binding operation is still in progress."
That requires the ability to distinguish when the operation is in-progress or
complete.

### Data Stored

When a request to create a ServiceBinding executes asynchronously, two records
are stored in etcd before the initial HTTP response is sent. The schema for
those two records is below:

```
BindInstance (/bind_instance/:binding_id)
* ID
* ServiceID
* Parameters

JobState (/state/:service_id/job/:token)
* Token
* State
* Podname
* Method
* Error 
```

If a second identical request comes in, it is easy to find the ``BindInstance``
created above. But there is no information in the request that can be used to
find the ``JobState`` record. Neither record has any reference to the other.
Thus it is not possible to determine if the ``BindInstance`` is complete or
still in-progress.

## Solutions

### JobState.Resource

We could add a field to the ``JobState`` called something like ``Resource``,
and make it a canonical reference to a resource such as a Service Binding.
Using the etcd key for the binding would be a good option.

A downside is that in order to find the correct ``JobState`` record, the broker
would have to iterate through the records for a given Service Instance, and
there is no limit to how many such records there could be.

### BindInstance.CreateJob

We could add a field on the ``BindInstance`` whose value is a reference to the
``JobState`` for the job that created it.

This would be an efficient model, because there would be no need for iterating
through numerous records. It fits how the data gets used through the API; there
will always be a known ``binding_id``, and there will then be a need to look up
the associated create job.

### BindInstance.State

We could store the state of a ``BindInstance`` as a new field on the object
itself. There would be no need to look up a ``JobState``.

The downside is that the state would effectively be duplicated in both the
``BindInstance`` and the ``JobState``. We would have to be careful to keep the
two in-sync with each other.

