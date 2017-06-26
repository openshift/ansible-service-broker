package clients

import (
	"github.com/coreos/etcd/client"
	"k8s.io/client-go/rest"
	"k8s.io/kubernetes/pkg/client/clientset_generated/clientset"
)

// Clients - Object containing clients that the application should be talking too.
var Clients struct {
	EtcdClient       client.Client
	KubernetesClient *clientset.Clientset
	RESTClient       rest.Interface
}
