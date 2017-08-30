package clients

import (
	"errors"

	logging "github.com/op/go-logging"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/pkg/api/v1"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/util/homedir"
)

type OpenshiftClient struct {
	restClient rest.Interface
}

type Project struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty" protobuf:"bytes,1,opt,name=metadata"`
	Spec              ProjectSpec   `json:"spec,omitempty" protobuf:"bytes,2,opt,name=spec"`
	Status            ProjectStatus `json:"status,omitempty" protobuf:"bytes,3,opt,name=status"`
}

type ProjectSpec struct {
	// Finalizers is an opaque list of values that must be empty to permanently remove object from storage.
	// More info: https://git.k8s.io/community/contributors/design-proposals/namespaces.md#finalizers
	// +optional
	Finalizers []string `json:"finalizers,omitempty" protobuf:"bytes,1,rep,name=finalizers"`
}

type ProjectStatus struct {
	Phase string `json:"phase,omitempty" protobuf:"bytes,1,opt,name=phase"`
}

// Openshift - Create a new openshift client if needed, returns reference
func Openshift(log *logging.Logger) (*OpenshiftClient, error) {
	errMsg := "Something went wrong while initializing kubernetes client!\n"
	once.Openshift.Do(func() {
		client, err := newOpenshift(log)
		if err != nil {
			log.Error(errMsg)
			// NOTE: Looking to leverage panic recovery to gracefully handle this
			// with things like retries or better intelligence, but the environment
			// is probably in a unrecoverable state as far as the broker is concerned,
			// and demands the attention of an operator.
			panic(err.Error())
		}
		instances.Openshift = client
	})
	if instances.Openshift == nil {
		return nil, errors.New("Kubernetes client instance is nil")
	}
	return instances.Openshift, nil
}

func newOpenshift(log *logging.Logger) (*OpenshiftClient, error) {
	// NOTE: Both the external and internal client object are using the same
	// clientset library. Internal clientset normally uses a different
	// library
	clientConfig, err := rest.InClusterConfig()
	if err != nil {
		log.Warning("Failed to create a InternalClientSet: %v.", err)

		log.Debug("Checking for a local Cluster Config")
		clientConfig, err = createClientConfigFromFile(homedir.HomeDir() + "/.kube/config")
		if err != nil {
			log.Error("Failed to create LocalClientSet")
			return nil, err
		}
	}

	clientset, err := newForConfig(clientConfig)
	if err != nil {
		log.Error("Failed to create LocalClientSet")
		return nil, err
	}

	return clientset, err
}

func newForConfig(c *rest.Config) (*OpenshiftClient, error) {
	config := *c
	if err := setConfigDefaults(&config); err != nil {
		return nil, err
	}
	client, err := rest.RESTClientFor(&config)
	if err != nil {
		return nil, err
	}
	return &OpenshiftClient{client}, nil
}

func setConfigDefaults(config *rest.Config) error {
	gv := v1.SchemeGroupVersion
	config.GroupVersion = &gv
	config.APIPath = "/oapi"
	config.NegotiatedSerializer = serializer.DirectCodecFactory{CodecFactory: scheme.Codecs}

	if config.UserAgent == "" {
		config.UserAgent = rest.DefaultKubernetesUserAgent()
	}

	return nil
}

func (o OpenshiftClient) CreateProject(name string) (result *Project, err error) {
	body := &Project{}
	body.Name = name
	result = &Project{}
	err = o.restClient.Post().
		Resource("projects").
		Body(body).
		Do().
		Into(result)
	return
}
