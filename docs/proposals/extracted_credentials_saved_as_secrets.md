## Extracted Credentials Saved As Secrets

Extracted Credentials are currently saved in our etcd for the service broker. This is not desirable for many reasons, but the two biggest are kubernetes already has a built-in way to manage this data, secrets, and when moving to CRDs we don't want to create a resource for extracted credentials.

### Problem Description
The problem is that we should not manage data that is of a sensitive nature if we do not have to. This proposal is limited in scope and only interested in how we save the extracted credentials. It is worth noting that we should eventually be better about how we transmit this data to APBs. 

In the secret, we will save the data in the following format.
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
  labels:
    <labels provided>
```

[Gob encoding](https://godoc.org/encoding/gob) will allow us to save arbitrary data in the secret for a key. The secrets keys will look rational to a user who looks at the created secret. This user would need permissions to see the secret, but if someone is looking at the secret making it obvious what data is in there will be helpful. 

The functions for saving and retrieving will be in the `clients` package. This means the callers will be required to use the underlying extracted credentials type `map[string]interface{}` because we do not want a circular dependency between `apb` package and `clients` package. 

We will interact with the secrets from the namespace defined in the configuration by the `openshift.namespace` value. 

The APB package will now be required to do all CRUD operations for extracted credentials. The APB package will expose a single retrieve extracted credentials method, that will take a UUID (either service instance id or binding instance id) and returns an `apb` package extracted credentials object.

Runtime package should be used to encapsulate the `clients` package calls. This will mean we have a default function for CRUD operations with extracted credentials. These default functions will be set to function vars at init of runtime. The function vars are then overrideable in the future. example:
```go
var SaveExtractedCredentials SaveExtractedCredentialsFunc
...

func saveExtractedCredentials(...) {
    ...
    k8scli, err := clients.Kubernetes()
    if err != nil {
        ...
    }
    k8scli.Clients.CoreV1().Secrets()...

}

init {
    SaveExtractedCredentials = saveExtractedCredentials
    ....
}
```


### Work Items
- [ ] Add kubernetes client methods to interact with extracted credentials in the [namespace](https://github.com/openshift/ansible-service-broker/blob/master/docs/config.md#openshift-configuration). 
- [ ] Add runtime methods for interacting with extracted credentials. These methods should be overridable. 
- [ ] Remove all dao implementation and interface methods regarding extracted credentials.
- [ ] Remove all instances of interacting with dao extracted credentials in the `broker` package. Add back call to APB package to get extracted credentials when needed.
- [ ] Update APB package to create/save/delete extracted credentials for the correct actions. this should call the correct `runtime` package methods.
- [ ] Add exposed method on APB  that will retrieve the extracted credentials.