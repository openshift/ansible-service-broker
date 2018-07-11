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
	"errors"
	"fmt"

	authoapi "github.com/openshift/api/authorization/v1"
	"github.com/openshift/api/image/v1"
	networkoapi "github.com/openshift/api/network/v1"
	authv1 "github.com/openshift/client-go/authorization/clientset/versioned/typed/authorization/v1"
	imagev1 "github.com/openshift/client-go/image/clientset/versioned/typed/image/v1"
	networkv1 "github.com/openshift/client-go/network/clientset/versioned/typed/network/v1"
	routev1 "github.com/openshift/client-go/route/clientset/versioned/typed/route/v1"
	log "github.com/sirupsen/logrus"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
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
	authClient    authv1.AuthorizationV1Interface
	imageClient   imagev1.ImageV1Interface
	networkClient networkv1.NetworkV1Interface
	routeClient   routev1.RouteV1Interface
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
	authClient, err := authv1.NewForConfig(c)
	if err != nil {
		return nil, err
	}
	imageClient, err := imagev1.NewForConfig(c)
	if err != nil {
		return nil, err
	}
	networkClient, err := networkv1.NewForConfig(c)
	if err != nil {
		return nil, err
	}
	routeClient, err := routev1.NewForConfig(c)
	if err != nil {
		return nil, err
	}
	return &OpenshiftClient{authClient: authClient, imageClient: imageClient, networkClient: networkClient, routeClient: routeClient}, nil
}

// SubjectRulesReview - create and run a OpenShift Subject Rules Review
func (o OpenshiftClient) SubjectRulesReview(user string, groups []string,
	scopes []string, namespace string) (result []authoapi.PolicyRule, err error) {

	body := authoapi.SubjectRulesReview{
		Spec: authoapi.SubjectRulesReviewSpec{
			User:   user,
			Groups: groups,
			Scopes: scopes,
		},
	}
	r, err := o.authClient.SubjectRulesReviews(namespace).Create(&body)
	if err != nil {
		return nil, err
	}
	return r.Status.Rules, nil
}

// ListRegistryImages - List images in internal OpenShift registry
func (o OpenshiftClient) ListRegistryImages() (*v1.ImageList, error) {
	return o.imageClient.Images().List(metav1.ListOptions{})
}

// GetClusterNetworkPlugin - Get cluster network
func (o OpenshiftClient) GetClusterNetworkPlugin() (string, error) {
	net, err := o.networkClient.ClusterNetworks().Get(networkoapi.ClusterNetworkDefault, metav1.GetOptions{})
	if err != nil {
		return "", err
	}
	return net.PluginName, nil
}

// GetNetNamespace - Get Net Namespace.
func (o OpenshiftClient) GetNetNamespace(nsName string) (*networkoapi.NetNamespace, error) {
	return o.networkClient.NetNamespaces().Get(nsName, metav1.GetOptions{})
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
	result, err := o.networkClient.NetNamespaces().Update(netns)
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
	result, err := o.networkClient.NetNamespaces().Update(netns)
	if err != nil {
		return nil, err
	}
	return result, nil
}

// Route - Returns a V1Route Interface
func (o OpenshiftClient) Route() routev1.RouteV1Interface {
	return o.routeClient
}

// Image - Returns a V1Image Interface
func (o OpenshiftClient) Image() imagev1.ImageV1Interface {
	return o.imageClient
}
