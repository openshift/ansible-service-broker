//
// Copyright (c) 2018 Red Hat, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//    http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//

package clients

import (
	b64 "encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	authoapi "github.com/openshift/api/authorization/v1"
	networkoapi "github.com/openshift/api/network/v1"
	log "github.com/sirupsen/logrus"
	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/util/homedir"
)

const (
	// ChangePodNetworkAnnotation - Annotation used for changing a pod
	// network, used to join networks together.
	ChangePodNetworkAnnotation = "pod.network.openshift.io/multitenant.change-network"
)

// OpenshiftClient - Client to interact with openshift api
type OpenshiftClient struct {
	authRestClient  rest.Interface
	imageRestClient rest.Interface
	networkClient   rest.Interface
}

type imageLabel struct {
	Spec    string `json:"com.redhat.apb.spec"`
	Runtime string `json:"com.redhat.apb.runtime"`
}

type containerConfig struct {
	Labels imageLabel `json:"Labels"`
}

type imageMetadata struct {
	ContainerConfig containerConfig `json:"ContainerConfig"`
}

type image struct {
	DockerImage string        `json:"dockerImageReference"`
	Metadata    imageMetadata `json:"dockerImageMetadata"`
}

// ImageList is a resource you can create to determine which actions another user can perform in a namespace
type ImageList struct {
	metav1.TypeMeta `json:",inline"`
	// Items holds the image data
	Items []image `json:"items"`
}

// FQImage is a struct to map FQNames to Imagestreams
type FQImage struct {
	Name        string
	DecodedSpec []byte
	Runtime     string
}

// Openshift - Create a new openshift client if needed, returns reference
func Openshift() (*OpenshiftClient, error) {
	errMsg := "Something went wrong while initializing openshift client!\n"
	once.Openshift.Do(func() {
		client, err := newOpenshift()
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
		return nil, errors.New("OpenShift client instance is nil")
	}
	return instances.Openshift, nil
}

func newOpenshift() (*OpenshiftClient, error) {
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
	authConfig := *c
	imageConfig := *c
	networkConfig := *c
	if err := setConfigDefaults(&authConfig, "/apis/authorization.openshift.io"); err != nil {
		return nil, err
	}
	authClient, err := rest.RESTClientFor(&authConfig)
	if err != nil {
		return nil, err
	}

	if err := setConfigDefaults(&imageConfig, "/apis/image.openshift.io"); err != nil {
		return nil, err
	}
	imageClient, err := rest.RESTClientFor(&imageConfig)
	if err != nil {
		return nil, err
	}
	if err := setConfigDefaults(&networkConfig, "/apis/network.openshift.io"); err != nil {
		return nil, err
	}
	networkClient, err := rest.RESTClientFor(&networkConfig)
	if err != nil {
		return nil, err
	}
	return &OpenshiftClient{authRestClient: authClient, imageRestClient: imageClient, networkClient: networkClient}, nil
}

func setConfigDefaults(config *rest.Config, APIPath string) error {
	gv := v1.SchemeGroupVersion
	config.GroupVersion = &gv
	config.APIPath = APIPath
	//	config.APIPath = "/apis/authorization.openshift.io"
	config.NegotiatedSerializer = serializer.DirectCodecFactory{CodecFactory: scheme.Codecs}

	if config.UserAgent == "" {
		config.UserAgent = rest.DefaultKubernetesUserAgent()
	}

	return nil
}

// SubjectRulesReview - create and run a OpenShift Subject Rules Review
func (o OpenshiftClient) SubjectRulesReview(user string, groups []string,
	scopes []string, namespace string) (result []authoapi.PolicyRule, err error) {

	body := &authoapi.SubjectRulesReview{
		Spec: authoapi.SubjectRulesReviewSpec{
			User:   user,
			Groups: groups,
			Scopes: scopes,
		},
	}
	body.Kind = "SubjectRulesReview"
	body.APIVersion = "authorization.openshift.io/v1"
	b, _ := json.Marshal(body)
	r := &authoapi.SubjectRulesReview{}
	err = o.authRestClient.Post().
		Namespace(namespace).
		Resource("subjectrulesreviews").
		Body(b).
		Do().
		Into(r)
	if err != nil {
		log.Errorf("error - %v\n", err)
		return
	}
	return r.Status.Rules, nil
}

// ConvertRegistryImagesToSpecs - Return APB specs from internal OCP registry
func (o OpenshiftClient) ConvertRegistryImagesToSpecs(imageList []string) ([]FQImage, error) {
	var fqList []FQImage
	fqImage := FQImage{}
	var err error
	r := &ImageList{}
	err = o.imageRestClient.Get().
		Resource("images").
		Do().
		Into(r)
	if err != nil {
		return nil, err
	}

	for _, image := range r.Items {
		var imageName = strings.Split(image.DockerImage, "@")[0]
		for _, providedImage := range imageList {
			if providedImage == imageName {
				encodedSpec := image.Metadata.ContainerConfig.Labels.Spec
				decodedSpec, err := b64.StdEncoding.DecodeString(encodedSpec)
				if err != nil {
					return nil, fmt.Errorf("Failed to grab encoded spec label")
				}
				fqImage.Name = imageName
				fqImage.DecodedSpec = decodedSpec
				fqImage.Runtime = image.Metadata.ContainerConfig.Labels.Runtime
				fqList = append(fqList, fqImage)
			}
		}
	}
	if err != nil {
		log.Errorf("error - %v\n", err)
		return nil, err
	}

	return fqList, nil
}

// ListRegistryImages - List images in internal OpenShift registry
func (o OpenshiftClient) ListRegistryImages() (images []string, err error) {
	var imageList []string
	r := &ImageList{}
	err = o.imageRestClient.Get().
		Resource("images").
		Do().
		Into(r)
	if err != nil {
		log.Errorf("error - %v\n", err)
		return
	}

	for _, image := range r.Items {
		imageList = append(imageList, strings.Split(image.DockerImage, "@")[0])
	}
	return imageList, nil
}

// GetClusterNetworkPlugin - Get cluster network
func (o OpenshiftClient) GetClusterNetworkPlugin() (string, error) {
	net := &networkoapi.ClusterNetwork{}
	err := o.networkClient.Get().Resource("clusternetworks").Name(networkoapi.ClusterNetworkDefault).Do().Into(net)
	if err != nil {
		return "", err
	}
	return net.PluginName, nil
}

// GetNetNamespace - Get Net Namespace.
func (o OpenshiftClient) GetNetNamespace(nsName string) (*networkoapi.NetNamespace, error) {
	netNamespace := &networkoapi.NetNamespace{}
	err := o.networkClient.Get().Resource("netnamespaces").Name(nsName).Do().Into(netNamespace)
	if err != nil {
		return nil, err
	}
	return netNamespace, nil
}

// JoinNamespacesNetworks - Will take the net namespace to be added to a network,
// and the namespace ID of the network that is being added to.
func (o OpenshiftClient) JoinNamespacesNetworks(netns *networkoapi.NetNamespace,
	targetNS string,
) (*networkoapi.NetNamespace, error) {
	if netns.Annotations == nil {
		netns.Annotations = make(map[string]string)
	}
	netns.Annotations[ChangePodNetworkAnnotation] = fmt.Sprintf("%s:%s", "join", targetNS)
	result := &networkoapi.NetNamespace{}
	err := o.networkClient.Put().Resource("netnamespaces").Name(netns.Name).Body(netns).Do().Into(result)
	if err != nil {
		return nil, err
	}
	return result, nil
}

// IsolateNamespacesNetworks - Will take the net namespace to be added to a network,
// and the namespace ID of the network that is being added to..
func (o OpenshiftClient) IsolateNamespacesNetworks(netns *networkoapi.NetNamespace,
	targetNS string,
) (*networkoapi.NetNamespace, error) {
	if netns.Annotations == nil {
		netns.Annotations = make(map[string]string)
	}
	netns.Annotations[ChangePodNetworkAnnotation] = fmt.Sprintf("%s:%s", "isolate", targetNS)
	result := &networkoapi.NetNamespace{}
	err := o.networkClient.Put().Resource("netnamespaces").Name(netns.Name).Body(netns).Do().Into(result)
	if err != nil {
		return nil, err
	}
	return result, nil
}
