//
// Copyright (c) 2017 Red Hat, Inc.
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

package runtime

import (
	"encoding/json"
	"fmt"

	"github.com/openshift/ansible-service-broker/pkg/clients"
	"github.com/openshift/ansible-service-broker/pkg/metrics"

	logutil "github.com/openshift/ansible-service-broker/pkg/util/logging"
	apicorev1 "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"
	rbac "k8s.io/api/rbac/v1beta1"
	kapierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	kubeversiontypes "k8s.io/apimachinery/pkg/version"
)

var log = logutil.NewLog()

// Provider - Variable for accessing provider functions
var Provider Runtime

// Runtime - Abstraction for broker actions
type Runtime interface {
	ValidateRuntime() error
	GetRuntime() string
	CreateSandbox(string, string, []string, string) (string, error)
	DestroySandbox(string, string, []string, string, bool, bool)
	AddPostCreateSandbox(f PostSandboxCreate)
	AddPostDestroySandbox(f PostSandboxDestroy)
}

// Variables for interacting with runtimes
type provider struct {
	coe
	postSandboxCreate  []PostSandboxCreate
	postSandboxDestroy []PostSandboxDestroy
}

// Abstraction for actions that are different between runtimes
type coe interface {
	getRuntime() string
	shouldJoinNetworks() (bool, PostSandboxCreate, PostSandboxDestroy)
}

// Different runtimes
type openshift struct{}
type kubernetes struct{}

// NewRuntime - Initialize provider variable
func NewRuntime() {
	k8scli, err := clients.Kubernetes()
	if err != nil {
		log.Error(err.Error())
		panic(err.Error())
	}

	// Identify which cluster we're using
	restclient := k8scli.Client.CoreV1().RESTClient()
	body, err := restclient.Get().AbsPath("/version/openshift").Do().Raw()

	var cluster coe
	switch {
	case err == nil:
		var kubeServerInfo kubeversiontypes.Info
		err = json.Unmarshal(body, &kubeServerInfo)
		if err != nil && len(body) > 0 {
			log.Error(err.Error())
			panic(err.Error())
		}
		log.Info("OpenShift version: %v", kubeServerInfo)
		cluster = newOpenshift()
	case kapierrors.IsNotFound(err) || kapierrors.IsUnauthorized(err) || kapierrors.IsForbidden(err):
		cluster = newKubernetes()
	default:
		log.Error(err.Error())
		panic(err.Error())
	}

	Provider = &provider{coe: cluster}
	if ok, postCreateHook, postDestroyHook := cluster.shouldJoinNetworks(); ok {
		log.Debugf("adding posthook to provider now.")
		if postCreateHook != nil {
			Provider.AddPostCreateSandbox(postCreateHook)
		}
		if postDestroyHook != nil {
			Provider.AddPostDestroySandbox(postDestroyHook)
		}
	}
}

func newOpenshift() coe {
	return openshift{}
}

func newKubernetes() coe {
	return kubernetes{}
}

// ValidateRuntime - Translate the broker cluster validation check into specfici runtime checks
func (p provider) ValidateRuntime() error {
	k8scli, err := clients.Kubernetes()
	if err != nil {
		return err
	}

	restclient := k8scli.Client.CoreV1().RESTClient()
	body, err := restclient.Get().AbsPath("/version").Do().Raw()

	switch {
	case err == nil:
		var kubeServerInfo kubeversiontypes.Info
		err = json.Unmarshal(body, &kubeServerInfo)
		if err != nil && len(body) > 0 {
			return err
		}
		log.Info("Kubernetes version: %v", kubeServerInfo)
	case kapierrors.IsNotFound(err) || kapierrors.IsUnauthorized(err) || kapierrors.IsForbidden(err):
		log.Error("the server could not find the requested resource")
		return err
	default:
		return err
	}
	return nil
}

// CreateSandbox - Translate the broker CreateSandbox call into cluster resource calls
func (p provider) CreateSandbox(podName string,
	namespace string,
	targets []string,
	apbRole string) (string, error) {

	k8scli, err := clients.Kubernetes()
	if err != nil {
		return "", err
	}

	err = k8scli.CreateServiceAccount(podName, namespace)
	if err != nil {
		return "", err
	}

	log.Debug("Trying to create apb sandbox: [ %s ], with %s permissions in namespace %s", podName, apbRole, namespace)

	subjects := []rbac.Subject{
		rbac.Subject{
			Kind:      "ServiceAccount",
			Name:      podName,
			Namespace: namespace,
		},
	}

	roleRef := rbac.RoleRef{
		APIGroup: "rbac.authorization.k8s.io",
		Kind:     "ClusterRole",
		Name:     apbRole,
	}

	// targetNamespace and namespace are the same
	err = k8scli.CreateRoleBinding(podName, subjects, namespace, namespace, roleRef)
	if err != nil {
		return "", err
	}

	for _, target := range targets {
		err = k8scli.CreateRoleBinding(podName, subjects, namespace, target, roleRef)
		if err != nil {
			return "", err
		}
	}

	// Check to see if there are already namespaces available before
	// creating ours
	policies, err := k8scli.Client.NetworkingV1().NetworkPolicies(targets[0]).List(metav1.ListOptions{})
	if err != nil {
		return "", err
	}

	// If there are already network policies, let's add one to allow for
	// communication from the APB pod to the target namespace
	if len(policies.Items) > 0 {
		networkPolicy := &networkingv1.NetworkPolicy{
			ObjectMeta: metav1.ObjectMeta{
				Name: podName,
			},
			Spec: networkingv1.NetworkPolicySpec{
				PodSelector: metav1.LabelSelector{},
				Ingress: []networkingv1.NetworkPolicyIngressRule{
					networkingv1.NetworkPolicyIngressRule{
						From: []networkingv1.NetworkPolicyPeer{
							networkingv1.NetworkPolicyPeer{
								NamespaceSelector: metav1.AddLabelToSelector(
									&metav1.LabelSelector{}, "apb-pod-name", podName),
							},
						},
					},
				},
			},
		}

		log.Debugf("Creating network policy for pod: %v to grant network access to ns: %v", podName, targets[0])
		_, err = k8scli.Client.NetworkingV1().NetworkPolicies(targets[0]).Create(networkPolicy)
		if err != nil {
			log.Errorf("unable to create network policy object - %v", err)
			return "", err
		}
		log.Debugf("Successfully created network policy for pod: %v to grant network access to ns: %v", podName, targets[0])
	} else {
		log.Info("No network policies found. Assuming things are open, skip network policy creation")
	}

	log.Info("Successfully created apb sandbox: [ %s ], with %s permissions in namespace %s", podName, apbRole, namespace)
	log.Info("Running post create sandbox fuctions if defined.")
	for i, f := range p.postSandboxCreate {
		log.Debugf("Running post create sandbox function: %v", i+1)
		err := f(podName, namespace, targets, apbRole)
		if err != nil {
			// Log the error and continue processing hooks. Expect hook to
			// clean up after itself.
			log.Warningf("Post create sandbox function failed with err: %v", err)
		}
	}

	return podName, nil
}

// DestroySandbox - Translate the broker DestorySandbox call into cluster resource calls
func (p provider) DestroySandbox(podName string,
	namespace string,
	targets []string,
	configNamespace string,
	keepNamespace bool,
	keepNamespaceOnError bool) {

	log.Info("Destroying APB sandbox...")
	if podName == "" {
		log.Info("Requested destruction of APB sandbox with empty handle, skipping.")
		return
	}
	k8scli, err := clients.Kubernetes()
	if err != nil {
		log.Error("Something went wrong getting kubernetes client")
		log.Errorf("%s", err.Error())
		return
	}
	pod, err := k8scli.Client.CoreV1().Pods(namespace).Get(podName, metav1.GetOptions{})
	if err != nil {
		log.Errorf("Unable to retrieve pod - %v", err)
	}
	if shouldDeleteNamespace(keepNamespace, keepNamespaceOnError, pod, err) {
		if configNamespace != namespace {
			log.Debug("Deleting namespace %s", namespace)
			k8scli.Client.CoreV1().Namespaces().Delete(namespace, &metav1.DeleteOptions{})
			// This is keeping track of namespaces.
			metrics.SandboxDeleted()
		} else {
			// We should not be attempting to run pods in the ASB namespace, if we are, something is seriously wrong.
			panic(fmt.Errorf("Broker is attempting to delete its own namespace"))
		}

	} else {
		log.Debugf("Keeping namespace alive due to configuration")
	}
	log.Debugf("Deleting rolebinding %s, namespace %s", podName, namespace)

	err = k8scli.DeleteRoleBinding(podName, namespace)
	if err != nil {
		log.Error("Something went wrong trying to destroy the rolebinding! - %v", err)
		return
	}
	log.Notice("Successfully deleted rolebinding %s, namespace %s", podName, namespace)

	for _, target := range targets {
		log.Debugf("Deleting rolebinding %s, namespace %s", podName, target)
		err = k8scli.DeleteRoleBinding(podName, target)
		if err != nil {
			log.Error("Something went wrong trying to destroy the rolebinding!")
			return
		}
		log.Notice("Successfully deleted rolebinding %s, namespace %s", podName, target)
	}

	log.Debugf("Deleting network policy for pod: %v to grant network access to ns: %v", podName, targets[0])
	// Must clean up the network policy that allowed comunication from the APB pod to the target namespace.
	err = k8scli.Client.NetworkingV1().NetworkPolicies(targets[0]).Delete(podName, &metav1.DeleteOptions{})
	if err != nil {
		log.Errorf("unable to delete the network policy object - %v", err)
		return
	}
	log.Debugf("Successfully deleted network policy for pod: %v to grant network access to ns: %v", podName, targets[0])

	log.Debugf("Running post sandbox destroy hooks")
	for i, f := range p.postSandboxDestroy {
		log.Debugf("Running post sandbox destroy:  %v", i+1)
		f(podName, namespace, targets)
	}
	return
}

// GetRuntime - Return a string value of the runtime
func (p provider) GetRuntime() string {
	return p.coe.getRuntime()
}

func shouldDeleteNamespace(keepNamespace bool,
	keepNamespaceOnError bool,
	pod *apicorev1.Pod,
	getPodErr error) bool {

	if keepNamespace {
		return false
	}

	if keepNamespaceOnError {
		if pod.Status.Phase == apicorev1.PodFailed || pod.Status.Phase == apicorev1.PodUnknown || getPodErr != nil {
			return false
		}
	}
	return true
}
