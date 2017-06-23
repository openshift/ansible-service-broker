package clients

import (
	d "github.com/fsouza/go-dockerclient"

	"github.com/coreos/etcd/client"
	"k8s.io/client-go/rest"
	"k8s.io/kubernetes/pkg/client/clientset_generated/clientset"
)

var Clients struct {
	EtcdClient       client.Client
	KubernetesClient *clientset.Clientset
	DockerClient     *d.Client
	RESTClient       rest.Interface
}
