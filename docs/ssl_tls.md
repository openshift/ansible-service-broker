## How the Broker Uses TLS/SSL

When deploying to an OpenShift cluster, the broker will need to be configured
for TLS/SSL. The workflow below uses the CA built into [OpenShift](https://docs.openshift.com/container-platform/3.6/dev_guide/secrets.html#service-serving-certificate-secrets) to sign a generated certificate and key, stored in  ```tls.crt``` and ```tls.key```. respectively.

The OpenShift documentation explains how this works.

>To secure communication to your service, have the cluster generate a signed serving certificate/key pair into a secret in your namespace. To do this, set the service.alpha.openshift.io/serving-cert-secret-name annotation on your service with the value set to the name you want to use for your secret. Then, your PodSpec can mount that secret. When it is available, your pod will run. The certificate will be good for the internal service DNS name, <service.name>.<service.namespace>.svc.
The certificate and key are in PEM format, stored in tls.crt and tls.key respectively. The certificate/key pair is automatically replaced when it gets close to expiration. View the expiration date in the service.alpha.openshift.io/expiry annotation on the secret, which is in RFC3339 format.

First, in the service definition's metadata, we need to add an annotation.
```yaml
...
    kind: Service
    metadata:
      name: asb
      labels:
        app: ansible-service-broker
        service: asb
      annotations:
        service.alpha.openshift.io/serving-cert-secret-name: asb-tls
```
This tells OpenShift that the cert will be stored in secret `asb-tls`.

Next, we need to tell the deployment config to mount the secret for the broker container.
```yaml
...
        spec:
          serviceAccount: asb
          containers:
          - image: ${BROKER_IMAGE}
            name: asb
            imagePullPolicy: IfNotPresent
            volumeMounts:
              - name: config-volume
                mountPath: /etc/ansible-service-broker
              - name: asb-tls
                mountPath: /etc/tls/private # this is where the broker needs to be configured to look for *.crt and *.key.
```

We also need to tell the deployment config about the volume.
```yaml
          volumes:
            - name: etcd
              persistentVolumeClaim:
                claimName: etcd
            - name: config-volume
              configMap:
                name: broker-config
                items:
                - key: broker-config
                  path: config.yaml
            - name: asb-tls # name will be used by the volume mount section above.
              secret:
                secretName: asb-tls # This is the name of the annotation that we mentioned above.
```

Next, we just need to tell the route to use tls. Example:
```yaml
  - apiVersion: v1
    kind: Route
    metadata:
      name: asb-1338
      labels:
        app: ansible-service-broker
        service: asb
    spec:
      to:
        kind: Service
        name: asb
      port:
        targetPort: port-1338
      tls:
        termination: <TERMINATION> # this where you will change how the route terminates.
```

More info on the secured route can be found in the [OpenShift documentation](https://docs.openshift.com/container-platform/3.6/architecture/core_concepts/routes.html#secured-routes).

In the broker config we need to tell the broker where to look for the certificate and key.
```yaml
...
broker:
  ...
  ssl_cert_key: /etc/tls/private/tls.key
  ssl_cert: /etc/tls/private/tls.crt
  ...
```
