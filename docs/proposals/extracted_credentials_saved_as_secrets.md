## Extracted Credentials Saved As Secrets

Extracted Credentials are currently saved in our etcd for the service broker. This is not desirable for many reasons, but the two biggest are kubernetes already has a built-in way to manage this data, secrets, and when moving to CRDs we don't want to create a resource for extracted credentials.

### Problem Description
The problem is that we should not manage data that is of a sensitive nature if we do not have to. This proposal is limited in scope and only interested in how we save the extracted credentials. It is worth noting that we should eventually be better about how we transmit this data to APBs. 

In our secret, we will save the data in the following format.
```yaml
data:
  DB_PASSWORD: <GOB_ENCODED_VALUE>
  DB_USERNAME: <GOB_ENCODED_VALUE>
  ....
apiVersion: v1
kind: Secret
metadata:
  name: <Service/Binding id>
  namespace: <Namespace for the broker>
```

[Gob encoding](https://godoc.org/encoding/gob) will allow us to save arbitrary data in a key, and our secrets keys will look rational to a user who looks are our secret.

The functions for saving and retrieving will be in the `clients` package. This means the callers will be required to use the underlying extracted credentials type `map[string]interface{}` because we do not want a circular dependency between `apb` package and `clients` package.

### Work Items
- [ ] Add kubernetes client methods to save and retrieve extracted credentials to the [namespace](https://github.com/openshift/ansible-service-broker/blob/master/docs/config.md#openshift-configuration). 
- [ ] Remove all dao implementation and interface methods regarding extracted credentials.
- [ ] Update all instances of saving and retrieving extracted credentials to use new methods.