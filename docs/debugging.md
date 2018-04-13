# Debugging

A debugging guide for the [Ansible service broker (ASB)](https://github.com/openshift/ansible-service-broker/)

For known issues, visit the [troubleshooting guide](./troubleshooting.md)

## Verify that the ASB Pods are Ready and in the `Running` state

Change to the ASB's namespace (e.g. `ansible-service-broker`)

```bash
$ oc project ansible-service-broker
Now using project "ansible-service-broker" on server "https://172.17.0.1:8443".
```

Run the `oc get pods` command in the ASB's name space to list the pods

```bash
$ oc get pods
NAME               READY     STATUS    RESTARTS   AGE
asb-1-wtd6f        1/1       Running   0          6m
asb-etcd-1-j9zgz   1/1       Running   0          6m
```

All pods should be in the ready and should be 'Running'.  If not, investigate the log of the specific pod that is failing.

## Verify that the ASB Retrieved a list of APBs from the Docker Org

The ASB will first get a list of the APBs from the specified org upon bootstrap.  For example, if the [`ansibleplaybookbundle`](https://hub.docker.com/u/ansibleplaybookbundle/dashboard/) org was used, the ASB pod logs will show some like this in the very beginning of the log

```bash
$ oc  logs asb-1-wtd6f | grep ansibleplaybookbundle
[2018-01-22T21:22:11.942Z] [DEBUG] - Loading image list for org: [ ansibleplaybookbundle ]
[2018-01-22T21:22:13.005Z] [DEBUG] - Trying to load ansibleplaybookbundle/s2i-apb
[2018-01-22T21:22:13.005Z] [DEBUG] - Trying to load ansibleplaybookbundle/hello-world-apb
[2018-01-22T21:22:13.005Z] [DEBUG] - Trying to load ansibleplaybookbundle/pyzip-demo-apb
[2018-01-22T21:22:13.005Z] [DEBUG] - Trying to load ansibleplaybookbundle/jenkins-apb
[2018-01-22T21:22:13.005Z] [DEBUG] - Trying to load ansibleplaybookbundle/hello-world-db-apb
[2018-01-22T21:22:13.005Z] [DEBUG] - Trying to load ansibleplaybookbundle/rocketchat-apb
[2018-01-22T21:22:13.006Z] [DEBUG] - Trying to load ansibleplaybookbundle/hastebin-apb
[2018-01-22T21:22:13.006Z] [DEBUG] - Trying to load ansibleplaybookbundle/wordpress-ha-apb
[2018-01-22T21:22:13.006Z] [DEBUG] - Trying to load ansibleplaybookbundle/rds-postgres-apb
[2018-01-22T21:22:13.006Z] [DEBUG] - Trying to load ansibleplaybookbundle/etherpad-apb
[2018-01-22T21:22:13.006Z] [DEBUG] - Trying to load ansibleplaybookbundle/pyzip-demo-db-apb
[2018-01-22T21:22:13.006Z] [DEBUG] - Trying to load ansibleplaybookbundle/thelounge-apb
[2018-01-22T21:22:13.006Z] [DEBUG] - Trying to load ansibleplaybookbundle/manageiq-apb
[2018-01-22T21:22:13.006Z] [DEBUG] - Trying to load ansibleplaybookbundle/nginx-apb
[2018-01-22T21:22:13.006Z] [DEBUG] - Trying to load ansibleplaybookbundle/postgresql-apb
[2018-01-22T21:22:13.006Z] [DEBUG] - Trying to load ansibleplaybookbundle/mediawiki-apb
[2018-01-22T21:22:13.006Z] [DEBUG] - Trying to load ansibleplaybookbundle/mysql-apb
[2018-01-22T21:22:13.006Z] [DEBUG] - Trying to load ansibleplaybookbundle/mariadb-apb
[2018-01-22T21:22:13.006Z] [DEBUG] - Trying to load ansibleplaybookbundle/origin-ansible-service-broker
[2018-01-22T21:22:13.006Z] [DEBUG] - Trying to load ansibleplaybookbundle/ansible-service-broker
[2018-01-22T21:22:13.006Z] [DEBUG] - Trying to load ansibleplaybookbundle/mediawiki123
[2018-01-22T21:22:13.006Z] [DEBUG] - Trying to load ansibleplaybookbundle/apb-base
[2018-01-22T21:22:13.006Z] [DEBUG] - Trying to load ansibleplaybookbundle/hello-world
[2018-01-22T21:22:13.006Z] [DEBUG] - Trying to load ansibleplaybookbundle/origin-service-catalog
[2018-01-22T21:22:13.006Z] [DEBUG] - Trying to load ansibleplaybookbundle/apb-tools
[2018-01-22T21:22:13.006Z] [DEBUG] - Trying to load ansibleplaybookbundle/py-zip-demo
[2018-01-22T21:22:13.006Z] [DEBUG] - Trying to load ansibleplaybookbundle/apb-assets-base
[2018-01-22T21:22:13.006Z] [DEBUG] - Trying to load ansibleplaybookbundle/asb-installer
[2018-01-22T21:22:13.006Z] [DEBUG] - Trying to load ansibleplaybookbundle/deploy-broker
[2018-01-22T21:22:13.006Z] [DEBUG] - Trying to load ansibleplaybookbundle/origin
[2018-01-22T21:22:13.006Z] [DEBUG] - Trying to load ansibleplaybookbundle/origin-deployer
[2018-01-22T21:22:13.006Z] [DEBUG] - Trying to load ansibleplaybookbundle/origin-docker-registry
[2018-01-22T21:22:13.006Z] [DEBUG] - Trying to load ansibleplaybookbundle/origin-haproxy-router
[2018-01-22T21:22:13.006Z] [DEBUG] - Trying to load ansibleplaybookbundle/origin-pod
[2018-01-22T21:22:13.007Z] [DEBUG] - Trying to load ansibleplaybookbundle/origin-sti-builder
[2018-01-22T21:22:13.007Z] [DEBUG] - Trying to load ansibleplaybookbundle/origin-recycler
[2018-01-22T21:22:13.007Z] [DEBUG] - -> ansibleplaybookbundle/pyzip-demo-apb
[2018-01-22T21:22:13.007Z] [DEBUG] - -> ansibleplaybookbundle/pyzip-demo-db-apb
[2018-01-22T21:22:13.007Z] [DEBUG] - -> ansibleplaybookbundle/nginx-apb
[2018-01-22T21:22:13.007Z] [DEBUG] - -> ansibleplaybookbundle/manageiq-apb
[2018-01-22T21:22:13.007Z] [DEBUG] - -> ansibleplaybookbundle/jenkins-apb
[2018-01-22T21:22:13.007Z] [DEBUG] - -> ansibleplaybookbundle/hello-world-db-apb
[2018-01-22T21:22:13.007Z] [DEBUG] - -> ansibleplaybookbundle/rds-postgres-apb
[2018-01-22T21:22:13.007Z] [DEBUG] - -> ansibleplaybookbundle/etherpad-apb
[2018-01-22T21:22:13.007Z] [DEBUG] - -> ansibleplaybookbundle/mediawiki-apb
[2018-01-22T21:22:13.007Z] [DEBUG] - -> ansibleplaybookbundle/mysql-apb
[2018-01-22T21:22:13.007Z] [DEBUG] - -> ansibleplaybookbundle/mariadb-apb
[2018-01-22T21:22:13.008Z] [DEBUG] - -> ansibleplaybookbundle/s2i-apb
[2018-01-22T21:22:13.008Z] [DEBUG] - -> ansibleplaybookbundle/hello-world-apb
[2018-01-22T21:22:13.008Z] [DEBUG] - -> ansibleplaybookbundle/rocketchat-apb
[2018-01-22T21:22:13.008Z] [DEBUG] - -> ansibleplaybookbundle/hastebin-apb
[2018-01-22T21:22:13.008Z] [DEBUG] - -> ansibleplaybookbundle/wordpress-ha-apb
[2018-01-22T21:22:13.008Z] [DEBUG] - -> ansibleplaybookbundle/postgresql-apb
[2018-01-22T21:22:13.008Z] [DEBUG] - -> ansibleplaybookbundle/thelounge-apb
[2018-01-22T21:22:13.305Z] [DEBUG] - Successfully converted Image docker.io/ansibleplaybookbundle/pyzip-demo-apb:latest into Spec
[2018-01-22T21:22:13.521Z] [DEBUG] - Successfully converted Image docker.io/ansibleplaybookbundle/pyzip-demo-db-apb:latest into Spec
[2018-01-22T21:22:13.633Z] [DEBUG] - Successfully converted Image docker.io/ansibleplaybookbundle/nginx-apb:latest into Spec
[2018-01-22T21:22:13.831Z] [DEBUG] - Successfully converted Image docker.io/ansibleplaybookbundle/manageiq-apb:latest into Spec
[2018-01-22T21:22:13.938Z] [DEBUG] - Successfully converted Image docker.io/ansibleplaybookbundle/jenkins-apb:latest into Spec
[2018-01-22T21:22:14.039Z] [DEBUG] - Successfully converted Image docker.io/ansibleplaybookbundle/hello-world-db-apb:latest into Spec
[2018-01-22T21:22:14.156Z] [DEBUG] - Successfully converted Image docker.io/ansibleplaybookbundle/rds-postgres-apb:latest into Spec
[2018-01-22T21:22:14.272Z] [DEBUG] - Successfully converted Image docker.io/ansibleplaybookbundle/etherpad-apb:latest into Spec
[2018-01-22T21:22:14.389Z] [DEBUG] - Successfully converted Image docker.io/ansibleplaybookbundle/mediawiki-apb:latest into Spec
[2018-01-22T21:22:14.569Z] [DEBUG] - Successfully converted Image docker.io/ansibleplaybookbundle/mysql-apb:latest into Spec
[2018-01-22T21:22:14.679Z] [DEBUG] - Successfully converted Image docker.io/ansibleplaybookbundle/mariadb-apb:latest into Spec
[2018-01-22T21:22:15.045Z] [DEBUG] - Successfully converted Image docker.io/ansibleplaybookbundle/hello-world-apb:latest into Spec
[2018-01-22T21:22:15.29Z] [DEBUG] - Successfully converted Image docker.io/ansibleplaybookbundle/rocketchat-apb:latest into Spec
[2018-01-22T21:22:15.395Z] [DEBUG] - Successfully converted Image docker.io/ansibleplaybookbundle/hastebin-apb:latest into Spec
[2018-01-22T21:22:16.311Z] [DEBUG] - Successfully converted Image docker.io/ansibleplaybookbundle/wordpress-ha-apb:latest into Spec
[2018-01-22T21:22:16.759Z] [DEBUG] - Successfully converted Image docker.io/ansibleplaybookbundle/postgresql-apb:latest into Spec
[2018-01-22T21:22:16.853Z] [DEBUG] - Successfully converted Image docker.io/ansibleplaybookbundle/thelounge-apb:latest into Spec
```

If you do NOT see any APBs as shown in the logs above, verify that the Docker Org used is correct, and that the permissions for the org are properly set (i.e. public).

## Verify the ASB connection to the Service Catalog

The ASB may get a list of APBs from the Docker Org, but it still needs to connect to the service catalog to use them in the cluster.

First check the ASB's logs and making sure that no connection errors occurred (e.g. 'TLS handshake errors'). If you see connection errors in the logs, review the [Service Catalog and ASB Communication Troubleshooting Guide](https://github.com/openshift/ansible-service-broker/blob/master/docs/troubleshooting.md#service-catalog-and-broker-communication-issues) for further troubleshooting steps.

You can verify ASB's connection to the service catalog via the `curl` command below, which also provides all the APB information:

```bash
curl -k -H "Authorization: Bearer $(oc whoami -t)" https://$(oc get routes --no-headers | awk '{print $2}')/ansible-service-broker/v2/catalog
```
## Forcing a relist

The catalog's inventory of `ServiceClasses` for the Ansible Broker can be
manually refreshed by using the `relist` feature. To force a relist, run the
commmand `oc edit clusterservicebroker ansible-service-broker`. This will open
the document in your `$EDITOR`. Search for the field `relistRequests`; it will
be an integer number. Incremement this value by one, save and quit the document.
This will force the Service Catalog to make a `/catalog` request against the
broker, refreshing all of the available `ServiceClasses.

## Debugging APBs

Once the ASB's communication with the service catalog has been verified, you may view information about the APBs. For example, you may want to get a list of all of your APBs (services), along with all of their service plans.

### WebUI

Logon to the OpenShift WebUI (e.g. <https://172.17.0.1:8443>). You should see a list of APBs and their plan offerings.

### CLI

#### APB List

As shown earlier, you can retrieve all of the APB information via the `curl` command below:

```bash
curl -k -H "Authorization: Bearer $(oc whoami -t)" https://$(oc get routes --no-headers | awk '{print $2}')/ansible-service-broker/v2/catalog
```

#### Service Plans

You can get a list of all of the service plans via the `oc get clusterserviceplans` command.  However, the output of that command does not show anything useful since it only shows the 'NAME' and 'AGE'. To get a better list, issue the following command

```bash
export IFS=$'\n'; for i in $(oc get clusterserviceplans -o go-template='{{ range $k, $v := .items }}{{ if $v.spec.externalMetadata.displayName }}{{ printf "%s %s %s\n" $v.metadata.name $v.spec.clusterServiceClassRef.name $v.spec.externalMetadata.displayName }}{{ end }}{{ end }}') ; do j=$(echo $i | awk '{print $2}'); k=$(oc get clusterserviceclass $j --no-headers -o custom-columns=name:.spec.externalMetadata.displayName) ; l=$(printf "%-30s" $k); echo -ne "$(echo $i | awk '{print $1}')\t$l\t$(echo $i | awk '{print $3}')\n"; done | sort -k2
```

The above command may produce an output that looks something like this:

```bash
6acd95356d01ab1753458097d249bff3	Amazon RDS - PostgreSQL (APB) 	Default
a7f2eb136c88bfde4b33339c255a64e1	Etherpad (APB)                	Default
2529017a538fda00903782ddd68124c9	Hello World (APB)             	Default
db1cdb40646cd408a924445e2b95a1bf	Hello World Database (APB)    	Default
35b7492c6704a6771fa895ebad2dbb1a	Jenkins (APB)                 	Default
e2bedf1aeec9ebd41b9308d831fcdf47	ManageIQ (APB)                	Default
7f88be6129622f72554c20af879a8ce0	MariaDB (APB)                 	Development
a180af14f32f36f62f03d1fc83215bb6	MariaDB (APB)                 	Production
76b2bdf5381b809657c90350726595e5	Mediawiki (APB)               	Default
583f053f9ba165125a16cf9aff768017	MySQL (APB)                   	Development
53bd38d78b2d279f6524ac4f271e9b76	MySQL (APB)                   	Production
7f4a5e35e4af2beb70076e72fab0b7ff	PostgreSQL (APB)              	Development
ea4c99bb7d7d0d492ce55a8ac8c75373	PostgreSQL (APB)              	Production
edb27dcda66700646529749cb69bd4de	Pyzip Demo (APB)              	Default
6446f514c7e2d2aa95c335dc166489cd	Pyzip Demo Database (APB)     	Default
b2bba601df39d5774c2313bc716981e0	RocketChat (APB)              	Default
b0a54fc269e1d2391641df450cd35cac	Wordpress-HA (APB)            	Default
```

The above shows you the internal name, the APB (external) name, and its plan name, for all of the available APB Service Plans.

## Debugging APB Provision

### Monitor APB Logs

The provisioning of an APB will occur in a temporary namespace/pod which will run the playbook for the APB's provision steps. If the provision is successful, this namespace/pod will be removed, and your APB will be available for use in the project/namespace that you've specified.

However if the provision was not successful, reviewing the logs of the pod that's temporarily launched would be help debugging what happened. It's best to open a terminal and issue a `watch` command to get pods in all of the namespaces before provisioning an APB.

Below is an example output of the command:  `watch 'oc get pods --all-namespaces'`

```bash
Every 2.0s: oc get pods --all-namespaces
NAMESPACE                           NAME                                  READY     STATUS      RESTARTS   AGE
ansible-service-broker              asb-1-wtd6f                           1/1       Running     1          1h
ansible-service-broker              asb-etcd-1-j9zgz                      1/1       Running     0          1h
default                             docker-registry-1-j2z9t               1/1       Running     0          1h
default                             persistent-volume-setup-l8d4d         0/1       Completed   0          1h
default                             router-1-zrx7f                        1/1       Running     0          1h
kube-service-catalog                apiserver-d5d7846f7-wd4pc             2/2       Running     0          1h
kube-service-catalog                controller-manager-78987f457c-cx4dz   1/1       Running     2          1h
openshift-template-service-broker   apiserver-psbqt                       1/1       Running     0          1h
openshift-web-console               webconsole-6678586555-qgcr9           1/1       Running     0          1h
```

After provisioning your APB, you will see another namespace appear (e.g. Hello APB was provisioned)

```bash
Every 2.0s: oc get pods --all-namespaces
NAMESPACE                           NAME                                  READY     STATUS              RESTARTS   AGE
ansible-service-broker              asb-1-wtd6f                           1/1       Running             1          1h
ansible-service-broker              asb-etcd-1-j9zgz                      1/1       Running             0          1h
default                             docker-registry-1-j2z9t               1/1       Running             0          1h
default                             persistent-volume-setup-l8d4d         0/1       Completed           0          1h
default                             router-1-zrx7f                        1/1       Running             0          1h
helloapb                            hello-world-1-deploy                  1/1       Running             0          10s
helloapb                            hello-world-1-fzhfl                   0/1       ContainerCreating   0          9s
kube-service-catalog                apiserver-d5d7846f7-wd4pc             2/2       Running             0          1h
kube-service-catalog                controller-manager-78987f457c-cx4dz   1/1       Running             2          1h
openshift-template-service-broker   apiserver-psbqt                       1/1       Running             0          1h
openshift-web-console               webconsole-6678586555-qgcr9           1/1       Running             0          1h
```

The above shows that the temporary pod `hello-world-1-deploy` was created.  Review the logs in that pod to further investigate any errors.
