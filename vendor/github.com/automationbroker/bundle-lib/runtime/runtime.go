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

package runtime

import (
	"encoding/json"
	"fmt"

	"github.com/automationbroker/bundle-lib/clients"

	log "github.com/sirupsen/logrus"
	apicorev1 "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"
	rbac "k8s.io/api/rbac/v1beta1"
	kapierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	kubeversiontypes "k8s.io/apimachinery/pkg/version"
)

// Provider - Variable for accessing provider functions
var Provider Runtime

// Configuration - The configuration for the runtime
type Configuration struct {
	// PostCreateSandboxHooks - The sandbox hooks that you would like to run.
	PostCreateSandboxHooks []PostSandboxCreate
	// PostDestroySandboxHooks - The sandbox hooks that you would like to run.
	PostDestroySandboxHooks []PostSandboxDestroy
	// PreCreateSandboxHooks - The sandbox hooks that you would like to run.
	PreCreateSandboxHooks []PreSandboxCreate
	// PreDestroySandboxHooks - The sandbox hooks that you would like to run.
	PreDestroySandboxHooks []PreSandboxDestroy
	// WatchBundle - this is the method that watches the bundle for completion.
	// The UpdateDescriptionFunc in the default case will call this function when the last description
	// annotation on the running bundle is changed.
	WatchBundle WatchRunningBundleFunc
	// RunBundle - This is the method that will run the bundle.
	RunBundle RunBundleFunc
	// CopySecretsToNamespace - This is the method that is used to copy
	// secrets from a namespace to the executionContext namespace.
	CopySecretsToNamespace CopySecretsToNamespaceFunc
	ExtractedCredential
	// StateMountLocation this is where on disk the state will be stored for a bundle
	StateMountLocation string
	// StateMasterNamespace the namespace where state created by bundles will be copied to between actions
	StateMasterNamespace string
}

// Runtime - Abstraction for broker actions
type Runtime interface {
	ValidateRuntime() error
	GetRuntime() string
	CreateSandbox(string, string, []string, string, map[string]string) (string, string, error)
	DestroySandbox(string, string, []string, string, bool, bool)
	ExtractCredentials(string, string, int) ([]byte, error)
	ExtractedCredential
	WatchRunningBundle(string, string, UpdateDescriptionFn) error
	RunBundle(ExecutionContext) (ExecutionContext, error)
	CopySecretsToNamespace(ExecutionContext, string, []string) error
	StateManager
}

// Variables for interacting with runtimes
type provider struct {
	coe
	ExtractedCredential
	postSandboxCreate      []PostSandboxCreate
	preSandboxCreate       []PreSandboxCreate
	postSandboxDestroy     []PostSandboxDestroy
	preSandboxDestroy      []PreSandboxDestroy
	watchBundle            WatchRunningBundleFunc
	runBundle              RunBundleFunc
	copySecretsToNamespace CopySecretsToNamespaceFunc
	state
}

// Abstraction for actions that are different between runtimes
type coe interface {
	getRuntime() string
	shouldJoinNetworks() (bool, PostSandboxCreate, PostSandboxDestroy)
}

// NewRuntime - Initialize provider variable
// extCreds - You can pass an ExtractedCredential conforming object this will
// be used to do CRUD operations. If you want to use the default pass nil
// and we will use the built-in default of saving them as secrets in the
// broker namespace.
func NewRuntime(config Configuration) {
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
		log.Infof("OpenShift version: %v", kubeServerInfo)
		cluster = newOpenshift()
	case kapierrors.IsNotFound(err) || kapierrors.IsUnauthorized(err) || kapierrors.IsForbidden(err):
		cluster = newKubernetes()
	default:
		log.Error(err.Error())
		panic(err.Error())
	}

	var c ExtractedCredential
	if config.ExtractedCredential == nil {
		c = defaultExtractedCredential{}
	} else {
		c = config.ExtractedCredential
	}
	// defaults for state
	if config.StateMasterNamespace == "" {
		config.StateMasterNamespace = defaultNamespace
	}
	if config.StateMountLocation == "" {
		config.StateMountLocation = defaultMountLocation
	}

	defaultStateManager := state{mountLocation: config.StateMountLocation, nsTarget: config.StateMasterNamespace}
	var w WatchRunningBundleFunc
	if config.WatchBundle != nil {
		w = config.WatchBundle
	} else {
		w = defaultWatchRunningBundle
	}
	var r RunBundleFunc
	if config.RunBundle != nil {
		r = config.RunBundle
	} else {
		r = defaultRunBundle
	}
	var s CopySecretsToNamespaceFunc
	if config.CopySecretsToNamespace != nil {
		s = config.CopySecretsToNamespace
	} else {
		s = defaultCopySecretsToNamespace
	}

	p := &provider{coe: cluster,
		ExtractedCredential:    c,
		watchBundle:            w,
		runBundle:              r,
		copySecretsToNamespace: s,
		state: defaultStateManager,
	}

	if len(config.PreCreateSandboxHooks) > 0 {
		p.preSandboxCreate = config.PreCreateSandboxHooks
	}

	if len(config.PostCreateSandboxHooks) > 0 {
		p.postSandboxCreate = config.PostCreateSandboxHooks
	}

	if len(config.PreDestroySandboxHooks) > 0 {
		p.preSandboxDestroy = config.PreDestroySandboxHooks
	}

	if len(config.PostDestroySandboxHooks) > 0 {
		p.postSandboxDestroy = config.PostDestroySandboxHooks
	}

	if ok, postCreateHook, postDestroyHook := cluster.shouldJoinNetworks(); ok {
		log.Debugf("adding posthook to provider now.")
		if postCreateHook != nil {
			p.addPostCreateSandbox(postCreateHook)
		}
		if postDestroyHook != nil {
			p.addPostDestroySandbox(postDestroyHook)
		}
	}
	Provider = p

}

func newOpenshift() coe {
	return openshift{}
}

func newKubernetes() coe {
	return kubernetes{}
}

// ValidateRuntime - Translate the broker cluster validation check into specific runtime checks
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
		log.Infof("Kubernetes version: %v", kubeServerInfo)
	case kapierrors.IsNotFound(err) || kapierrors.IsUnauthorized(err) || kapierrors.IsForbidden(err):
		log.Error("the server could not find the requested resource")
		return err
	default:
		return err
	}
	return nil
}

func isNamespaceInTargets(ns string, targets []string) bool {
	for _, tns := range targets {
		if tns == ns {
			return true
		}
	}
	return false
}

// CreateSandbox - Translate the broker CreateSandbox call into cluster resource calls
func (p provider) CreateSandbox(podName string,
	namespace string,
	targets []string,
	apbRole string,
	metadata map[string]string,
) (string, string, error) {
	k8scli, err := clients.Kubernetes()
	if err != nil {
		return "", "", err
	}
	err = validateTargets(targets)
	if err != nil {
		return "", "", fmt.Errorf("unable to get target namespaces: %v", err)
	}

	// If Location is in the targets then we should not create the namespace.
	if !isNamespaceInTargets(namespace, targets) {
		// Create namespace.
		ns := &apicorev1.Namespace{
			ObjectMeta: metav1.ObjectMeta{
				Labels:       metadata,
				GenerateName: namespace,
			},
		}
		ns, err = k8scli.Client.CoreV1().Namespaces().Create(ns)
		if err != nil {
			return "", "", err
		}
		//Sandbox (i.e Namespace) was created.
		namespace = ns.ObjectMeta.Name
	}

	for i, f := range p.preSandboxCreate {
		log.Debugf("Running pre create sandbox function: %v", i+1)
		err := f(podName, namespace, targets, apbRole)
		if err != nil {
			// Log the error and continue processing hooks. Expect hook to
			// clean up after itself.
			log.Warningf("Pre create sandbox function failed with err: %v", err)
		}
	}

	err = k8scli.CreateServiceAccount(podName, namespace)
	if err != nil {
		return "", "", err
	}

	log.Debugf("Trying to create apb sandbox: [ %s ], with %s permissions in namespace %s", podName, apbRole, namespace)

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
		return "", "", err
	}

	for _, target := range targets {
		// It could be the case that we already added the rolebinding as target and namespace are equal.
		if target != namespace {
			err = k8scli.CreateRoleBinding(podName, subjects, namespace, target, roleRef)
			if err != nil {
				return "", "", err
			}
		}
	}

	// Must create a Network policy to allow for communication from the APB pod to the target namespace.
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
							NamespaceSelector: metav1.AddLabelToSelector(&metav1.LabelSelector{}, "apb-pod-name", podName),
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
		return "", "", err
	}
	log.Debugf("Successfully created network policy for pod: %v to grant network access to ns: %v", podName, targets[0])

	log.Infof("Successfully created apb sandbox: [ %s ], with %s permissions in namespace %s", podName, apbRole, namespace)
	log.Info("Running post create sandbox functions if defined.")
	for i, f := range p.postSandboxCreate {
		log.Debugf("Running post create sandbox function: %v", i+1)
		err := f(podName, namespace, targets, apbRole)
		if err != nil {
			// Log the error and continue processing hooks. Expect hook to
			// clean up after itself.
			log.Warningf("Post create sandbox function failed with err: %v", err)
		}
	}

	return podName, namespace, nil
}

func validateTargets(targets []string) error {
	k8scli, err := clients.Kubernetes()
	if err != nil {
		return err
	}
	for _, ns := range targets {
		_, err = k8scli.Client.CoreV1().Namespaces().Get(ns, metav1.GetOptions{})
		if err != nil {
			return err
		}
	}
	return nil
}

// DestroySandbox - Translate the broker DestorySandbox call into cluster resource calls
func (p provider) DestroySandbox(podName string,
	namespace string,
	targets []string,
	configNamespace string,
	keepNamespace bool,
	keepNamespaceOnError bool) {

	for i, f := range p.preSandboxDestroy {
		log.Debugf("Running pre sandbox destroy:  %v", i+1)
		err := f(podName, namespace, targets)
		if err != nil {
			// Log the error and continue processing hooks. Expect hook to
			// clean up after itself.
			log.Warningf("Pre destroy sandbox function failed with err: %v", err)
		}
	}

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
			log.Debugf("Deleting namespace %s", namespace)
			k8scli.Client.CoreV1().Namespaces().Delete(namespace, &metav1.DeleteOptions{})
			// This is keeping track of namespaces.
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
		log.Errorf("Something went wrong trying to destroy the rolebinding! - %v", err)
		return
	}
	log.Infof("Successfully deleted rolebinding %s, namespace %s", podName, namespace)

	for _, target := range targets {
		log.Debugf("Deleting rolebinding %s, namespace %s", podName, target)
		err = k8scli.DeleteRoleBinding(podName, target)
		if err != nil {
			log.Error("Something went wrong trying to destroy the rolebinding!")
			return
		}
		log.Infof("Successfully deleted rolebinding %s, namespace %s", podName, target)
	}

	log.Debugf("Deleting network policy for pod: %v to grant network access to ns: %v", podName, targets[0])
	// Must clean up the network policy that allowed communication from the APB pod to the target namespace.
	err = k8scli.Client.NetworkingV1().NetworkPolicies(targets[0]).Delete(podName, &metav1.DeleteOptions{})
	if err != nil {
		log.Errorf("unable to delete the network policy object - %v", err)
		return
	}
	log.Debugf("Successfully deleted network policy for pod: %v to grant network access to ns: %v", podName, targets[0])

	log.Debugf("Running post sandbox destroy hooks")
	for i, f := range p.postSandboxDestroy {
		log.Debugf("Running post sandbox destroy:  %v", i+1)
		err := f(podName, namespace, targets)
		if err != nil {
			// Log the error and continue processing hooks. Expect hook to
			// clean up after itself.
			log.Warningf("Post destroy sandbox function failed with err: %v", err)
		}
	}
	return
}

// GetRuntime - Return a string value of the runtime
func (p provider) GetRuntime() string {
	return p.coe.getRuntime()
}

func (p provider) WatchRunningBundle(podName string, namespace string, updateFunc UpdateDescriptionFn) error {
	return p.watchBundle(podName, namespace, updateFunc)
}

func (p provider) CopySecretsToNamespace(ec ExecutionContext, cn string, secrets []string) error {
	return p.copySecretsToNamespace(ec, cn, secrets)
}

func (p provider) RunBundle(ec ExecutionContext) (ExecutionContext, error) {
	return p.runBundle(ec)
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
