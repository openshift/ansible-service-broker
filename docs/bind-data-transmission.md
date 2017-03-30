# Sharing Data Between the Broker and the APB on a Bind Operation

## Options
| Option            | Notes     |
| ----------------- | --------- |
| Trailing Pod Logs | Credentials are posted to stdout when they are ready. |
| Output on Exit    | Outputs the credentials on exit. |
| Shared Volume     | Store credentials in a shared volume. |
| Web Hook          | Gives the broker credentials through an endpoint. |

## Trailing Pod Logs
Trailing pod logs is a temporary solution for corralling bind credentials from
the bind APB. After spawning the bind container, the Broker waits for output
by doing an `oc logs -f <pod>`.

### Implementation details
| Requirements |
| ------------ |
| None         |

This solutions is easy to implement and doesn't add any complexity or have any
requirements.

### Downside
This isn't a good solution because container logs are readable by anyone
on that host. Since we're passing sensitive information, we don't want any
process on the system to have access to the credentials.

## Output on Exit
This solution adds to the pod definition a location where the bind credentials
can be picked up by Kubernetes with the `terminationMessagePath` parameter.

The APB will run to gather bind credentials and store them in a file and exit.
Upon termination, Kubernetes will gather the log message in the file and set
that to be the pod termination message.  The Broker would look for these
credentials in the terminating pod message [1].

### Implementation details
| Requirements |
| ------------ |
| Kubernetes >= 1.6 |

### Downside
The downside is that anyone in the same project can access these credentials by
reading the log message.

## Shared Volume
Sharing physical volumes in Kubernetes/OpenShift is a common pattern for data
sharing between multiple pods and containers.

The process of sharing a pv between the Broker and the APB is to create a
bind-shared-pv when the Broker is started. Then, when a APB is started, it will
also use the bind-shared-pv volume.

When the bind occurs, the APB will gather credentials and store them in a file
in the mounted physical volume and exit. Then, the Broker will read the contents
of the file and store the bind credentials.

### Implementation details
| Requirements |
| ------------ |
| Shared Storage when using multinode   |
| Ansible module for creating PVs |
| Additional security around Volumes (selinux policy) |
| Broker will only be allowed to bind one app at a time |

### Downside
Having a shared volume can be a vulnerability as long as anyone can mount the
volume and read and write to it.  But, SELinux policy can prevent the file
from being read or written from inside the container.  Also, the Broker can
delete the file after it reads the bind credentials and block any further binds
until the pending bind is completed.

There's also the concern of having shared volumes when the Broker and APB are in
different clusters. The shared volume solution would fail here.

## Web Hook
Using a web hook requires the Broker creating an endpoint that is reachable by
the APB to transmit data. When the APB gathers bind credentials, it will contact
the Broker on the endpoint and pass the data off to the Broker.

### Implementation details
| Requirements |
| ------------ |
| New handler/route |
| Token authentication |
| APBs must always be able to contact the Broker |
| Docs explaining the required networking and firewall rules |

### Downside
Operators with production clusters will have any number of network customizations
and firewall rules.  This solution would require the operator allow the APB to
contact the Broker on an endpoint. But, we could document this to ensure
operators have instructions to make this work.

[1] - https://kubernetes.io/docs/tasks/debug-application-cluster/determine-reason-pod-failure/