package clients

import (
	"encoding/json"
	"errors"
	"fmt"

	logging "github.com/op/go-logging"
	"github.com/openshift/ansible-service-broker/pkg/origin/copy/authorization"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/pkg/api/v1"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/util/homedir"
	rbac "k8s.io/kubernetes/pkg/apis/rbac/v1beta1"
)

type OpenshiftClient struct {
	restClient     rest.Interface
	restClientAuth rest.Interface
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
type OptionalScopes []string

func (t OptionalScopes) String() string {
	return fmt.Sprintf("%v", []string(t))
}

// SubjectRulesReview is a resource you can create to determine which actions another user can perform in a namespace
type SubjectRulesReview struct {
	metav1.TypeMeta `json:",inline"`

	// Spec adds information about how to conduct the check
	Spec SubjectRulesReviewSpec `json:"spec" protobuf:"bytes,1,opt,name=spec"`

	// Status is completed by the server to tell which permissions you have
	Status SubjectRulesReviewStatus `json:"status,omitempty" protobuf:"bytes,2,opt,name=status"`
}

// SubjectRulesReviewSpec adds information about how to conduct the check
type SubjectRulesReviewSpec struct {
	// User is optional.  At least one of User and Groups must be specified.
	User string `json:"user" protobuf:"bytes,1,opt,name=user"`
	// Groups is optional.  Groups is the list of groups to which the User belongs.  At least one of User and Groups must be specified.
	Groups []string `json:"groups" protobuf:"bytes,2,rep,name=groups"`
	// Scopes to use for the evaluation.  Empty means "use the unscoped (full) permissions of the user/groups".
	Scopes []string `json:"scopes" protobuf:"bytes,3,opt,name=scopes"`
}

// SubjectRulesReviewStatus is contains the result of a rules check
type SubjectRulesReviewStatus struct {
	// Rules is the list of rules (no particular sort) that are allowed for the subject
	Rules []rbac.PolicyRule `json:"rules" protobuf:"bytes,1,rep,name=rules"`
	// EvaluationError can appear in combination with Rules.  It means some error happened during evaluation
	// that may have prevented additional rules from being populated.
	EvaluationError string `json:"evaluationError,omitempty" protobuf:"bytes,2,opt,name=evaluationError"`
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
	if err := setConfigDefaultsAuth(&config); err != nil {
		return nil, err
	}
	clientAuth, err := rest.RESTClientFor(&config)
	if err != nil {
		return nil, err
	}
	return &OpenshiftClient{restClient: client, restClientAuth: clientAuth}, nil
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

func setConfigDefaultsAuth(config *rest.Config) error {
	gv := v1.SchemeGroupVersion
	config.GroupVersion = &gv
	config.APIPath = "/apis/authorization.openshift.io"
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

func (o OpenshiftClient) SubjectRulesReview(user, namespace string, log *logging.Logger) (result *authorization.SubjectRulesReview, err error) {
	body := &SubjectRulesReview{
		Spec: SubjectRulesReviewSpec{
			User: "admin",
		},
	}
	body.Kind = "SubjectRulesReview"
	body.APIVersion = "authorization.openshift.io/v1"
	b, _ := json.Marshal(body)
	r := &SubjectRulesReview{}
	res, err := o.restClientAuth.Post().
		Namespace(namespace).
		Resource("subjectrulesreviews").
		Body(b, log).
		DoRaw()
	err = json.Unmarshal(res, r)
	if err != nil {
		log.Errorf("error - %v\n unmarshall - %q", err, res)
	}
	result = &authorization.SubjectRulesReview{
		Spec: authorization.SubjectRulesReviewSpec{
			User:   r.Spec.User,
			Groups: r.Spec.Groups,
			Scopes: r.Spec.Scopes,
		},
		Status: authorization.SubjectRulesReviewStatus{
			EvaluationError: r.Status.EvaluationError,
			Rules:           authorization.Convert_rbac_PolicyRules_To_authorization_PolicyRules(r.Status.Rules),
		},
	}
	result.Kind = r.Kind
	result.APIVersion = r.APIVersion
	return
}
