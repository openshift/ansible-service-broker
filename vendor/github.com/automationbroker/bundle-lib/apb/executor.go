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

package apb

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strings"
	"sync"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/automationbroker/bundle-lib/clients"
	"github.com/automationbroker/bundle-lib/runtime"
	"github.com/pborman/uuid"
	log "github.com/sirupsen/logrus"
	"k8s.io/api/core/v1"
)

// ExecutorAccessors - Accessors for Executor state.
type ExecutorAccessors interface {
	PodName() string
	LastStatus() StatusMessage
	DashboardURL() string
	ExtractedCredentials() *ExtractedCredentials
}

// ExecutorAsync - Main interface used for running APBs asynchronously.
type ExecutorAsync interface {
	Provision(*ServiceInstance) <-chan StatusMessage
	Deprovision(instance *ServiceInstance) <-chan StatusMessage
	Bind(instance *ServiceInstance, parameters *Parameters, bindingID string) <-chan StatusMessage
	Unbind(instance *ServiceInstance, parameters *Parameters, bindingID string) <-chan StatusMessage
	Update(instance *ServiceInstance) <-chan StatusMessage
}

//go:generate mockery -name=Executor -case=underscore -inpkg -note=Generated

// Executor - Composite executor interface.
type Executor interface {
	ExecutorAccessors
	ExecutorAsync
}

type executor struct {
	extractedCredentials *ExtractedCredentials
	dashboardURL         string
	podName              string
	lastStatus           StatusMessage
	statusChan           chan StatusMessage
	mutex                sync.Mutex
}

// NewExecutor - Creates a new Executor for running an APB.
func NewExecutor() Executor {
	exec := &executor{
		statusChan: make(chan StatusMessage),
		lastStatus: StatusMessage{State: StateNotYetStarted},
	}
	return exec
}

// PodName - Returns the name of the pod running the APB
func (e *executor) PodName() string {
	return e.podName
}

// LastStatus - Returns the last known status of the APB
func (e *executor) LastStatus() StatusMessage {
	return e.lastStatus
}

// DashboardURL - Returns the dashboard URL of the APB
func (e *executor) DashboardURL() string {
	return e.dashboardURL
}

// ExtractedCredentials - Credentials extracted from the APB while running,
// if they were discovered.
func (e *executor) ExtractedCredentials() *ExtractedCredentials {
	return e.extractedCredentials
}

func (e *executor) actionStarted() {
	log.Debug("executor::actionStarted")
	e.lastStatus.State = StateInProgress
	e.lastStatus.Description = "action started"
	e.statusChan <- e.lastStatus
}

func (e *executor) actionFinishedWithSuccess() {
	e.mutex.Lock()
	defer e.mutex.Unlock()

	log.Debug("executor::actionFinishedWithSuccess")

	if e.statusChan != nil {
		e.lastStatus.State = StateSucceeded
		e.lastStatus.Description = "action finished with success"
		e.statusChan <- e.lastStatus
		close(e.statusChan)
		e.statusChan = nil
	} else {
		log.Warning("executor::actionFinishedWithSuccess was called, but the statusChan was already closed!")
	}
}

func (e *executor) actionFinishedWithError(err error) {
	e.mutex.Lock()
	defer e.mutex.Unlock()

	log.Debug("executor::actionFinishedWithError[ %v ]", err.Error())

	if e.statusChan != nil {
		e.lastStatus.State = StateFailed
		e.lastStatus.Error = err
		e.lastStatus.Description = "action finished with error"
		e.statusChan <- e.lastStatus
		close(e.statusChan)
		e.statusChan = nil
	} else {
		log.Warning("executor::actionFinishedWithError was called, but the statusChan was already closed!")
	}
}

func (e *executor) updateDescription(newDescription string, dashboardURL string) {
	if newDescription != "" {
		status := e.lastStatus
		status.Description = newDescription
		e.lastStatus = status
		e.statusChan <- status
	}
	if dashboardURL != "" {
		e.dashboardURL = dashboardURL
	}
}

// executeApb - Runs an APB Action with a provided set of inputs
func (e *executor) executeApb(
	action string, spec *Spec, context *Context, p *Parameters,
) (ExecutionContext, error) {
	log.Debug("ExecutingApb:")
	log.Debug("name:[ %s ]", spec.FQName)
	log.Debug("image:[ %s ]", spec.Image)
	log.Debug("action:[ %s ]", action)
	log.Debug("pullPolicy:[ %s ]", clusterConfig.PullPolicy)
	log.Debug("role:[ %s ]", clusterConfig.SandboxRole)

	executionContext := ExecutionContext{ProxyConfig: GetProxyConfig()}

	extraVars, err := createExtraVars(context, p)

	if err != nil {
		return executionContext, err
	}
	// It's a critical error if a Namespace is not provided to the
	// broker because its required to know where to execute the pods and
	// sandbox them based on that Namespace. Should fail fast and loud,
	// with controlled error handling.
	if context.Namespace == "" {
		errStr := "Namespace not found within request context. Cannot perform requested " + action
		log.Error(errStr)
		return executionContext, errors.New(errStr)
	}

	pullPolicy, err := checkPullPolicy(clusterConfig.PullPolicy)
	if err != nil {
		return executionContext, err
	}

	secrets := getSecrets(spec)

	k8scli, err := clients.Kubernetes()
	if err != nil {
		return executionContext, err
	}

	executionContext.PodName = fmt.Sprintf("apb-%s", uuid.New())
	labels := map[string]string{
		"apb-fqname":   spec.FQName,
		"apb-action":   action,
		"apb-pod-name": executionContext.PodName,
	}

	e.podName = executionContext.PodName

	executionContext.Targets = append(executionContext.Targets, context.Namespace)
	// Create namespace.
	namespace := v1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Labels:       labels,
			GenerateName: fmt.Sprintf("%s-%.4s-", spec.FQName, action),
		},
	}
	ns, err := k8scli.Client.CoreV1().Namespaces().Create(&namespace)
	if err != nil {
		return executionContext, err
	}
	//Sandbox (i.e Namespace) was created.
	executionContext.Namespace = ns.ObjectMeta.Name
	err = copySecretsToNamespace(executionContext, clusterConfig, k8scli, secrets)
	if err != nil {
		log.Errorf("unable to copy secrets: %v to  new namespace", secrets)
		return executionContext, err
	}

	executionContext.ServiceAccount, err = runtime.Provider.CreateSandbox(executionContext.PodName, executionContext.Namespace, executionContext.Targets, clusterConfig.SandboxRole)
	if err != nil {
		log.Error(err.Error())
		return executionContext, err
	}
	volumes, volumeMounts := buildVolumeSpecs(secrets)

	pod := &v1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:   executionContext.PodName,
			Labels: labels,
		},
		Spec: v1.PodSpec{
			Containers: []v1.Container{
				{
					Name:  ApbContainerName,
					Image: spec.Image,
					Args: []string{
						action,
						"--extra-vars",
						extraVars,
					},
					Env:             createPodEnv(executionContext),
					ImagePullPolicy: pullPolicy,
					VolumeMounts:    volumeMounts,
				},
			},
			RestartPolicy:      v1.RestartPolicyNever,
			ServiceAccountName: executionContext.ServiceAccount,
			Volumes:            volumes,
		},
	}

	log.Infof(fmt.Sprintf("Creating pod %q in the %s namespace", pod.Name, executionContext.Namespace))
	_, err = k8scli.Client.CoreV1().Pods(executionContext.Namespace).Create(pod)

	return executionContext, err
}

// GetProxyConfig - Returns a ProxyConfig based on the presence of a proxy
// configuration in the broker's environment. HTTP_PROXY, HTTPS_PROXY, and
// NO_PROXY are the relevant environment variables. If no proxy is found,
// GetProxyConfig will return nil.
func GetProxyConfig() *ProxyConfig {
	httpProxy, httpProxyPresent := os.LookupEnv(httpProxyEnvVar)
	httpsProxy, httpsProxyPresent := os.LookupEnv(httpsProxyEnvVar)
	noProxy, noProxyPresent := os.LookupEnv(noProxyEnvVar)

	// TODO: Probably some more permutations of these that should be validated?

	if !noProxyPresent && !httpProxyPresent && !httpsProxyPresent {
		log.Debug("No proxy env vars found to be configured.")
		return nil
	}

	if noProxyPresent && !httpProxyPresent && !httpsProxyPresent {
		log.Info("NO_PROXY env var set, but no proxy has been found via HTTP_PROXY, or HTTPS_PROXY")
		return nil
	}

	return &ProxyConfig{
		HTTPProxy:  httpProxy,
		HTTPSProxy: httpsProxy,
		NoProxy:    noProxy,
	}
}

func buildVolumeSpecs(secrets []string) ([]v1.Volume, []v1.VolumeMount) {
	var optional bool
	var mountName string
	volumes := []v1.Volume{}
	volumeMounts := []v1.VolumeMount{}

	for _, secret := range secrets {
		mountName = "apb-" + secret
		volumes = append(volumes, v1.Volume{
			Name: mountName,
			VolumeSource: v1.VolumeSource{
				Secret: &v1.SecretVolumeSource{
					SecretName: secret,
					Optional:   &optional,
					// Eventually, we can include: Items: []v1.KeyToPath here to specify specific keys in the secret
				},
			},
		})
		volumeMounts = append(volumeMounts, v1.VolumeMount{
			Name:      mountName,
			MountPath: "/etc/apb-secrets/" + mountName,
			ReadOnly:  true,
		})
	}
	return volumes, volumeMounts
}

// TODO: Instead of putting namespace directly as a parameter, we should create a dictionary
// of apb_metadata and put context and other variables in it so we don't pollute the user
// parameter space.
func createExtraVars(context *Context, parameters *Parameters) (string, error) {
	var paramsCopy Parameters
	if parameters != nil && *parameters != nil {
		paramsCopy = *parameters
	} else {
		paramsCopy = make(Parameters)
	}

	if context != nil {
		paramsCopy[NamespaceKey] = context.Namespace
	}

	paramsCopy[ClusterKey] = runtime.Provider.GetRuntime()
	extraVars, err := json.Marshal(paramsCopy)
	return string(extraVars), err
}

func createPodEnv(executionContext ExecutionContext) []v1.EnvVar {
	podEnv := []v1.EnvVar{
		v1.EnvVar{
			Name: "POD_NAME",
			ValueFrom: &v1.EnvVarSource{
				FieldRef: &v1.ObjectFieldSelector{
					FieldPath: "metadata.name",
				},
			},
		},
		v1.EnvVar{
			Name: "POD_NAMESPACE",
			ValueFrom: &v1.EnvVarSource{
				FieldRef: &v1.ObjectFieldSelector{
					FieldPath: "metadata.namespace",
				},
			},
		},
	}

	if executionContext.ProxyConfig != nil {
		conf := executionContext.ProxyConfig

		log.Info("Proxy configuration present. Applying to APB before execution:")
		log.Infof("%s=\"%s\"", httpProxyEnvVar, conf.HTTPProxy)
		log.Infof("%s=\"%s\"", httpsProxyEnvVar, conf.HTTPSProxy)
		log.Infof("%s=\"%s\"", noProxyEnvVar, conf.NoProxy)

		podEnv = append(podEnv, []v1.EnvVar{
			v1.EnvVar{
				Name:  httpProxyEnvVar,
				Value: conf.HTTPProxy,
			},
			v1.EnvVar{
				Name:  httpsProxyEnvVar,
				Value: conf.HTTPSProxy,
			},
			v1.EnvVar{
				Name:  noProxyEnvVar,
				Value: conf.NoProxy,
			},
			v1.EnvVar{
				Name:  strings.ToLower(httpProxyEnvVar),
				Value: conf.HTTPProxy,
			},
			v1.EnvVar{
				Name:  strings.ToLower(httpsProxyEnvVar),
				Value: conf.HTTPSProxy,
			},
			v1.EnvVar{
				Name:  strings.ToLower(noProxyEnvVar),
				Value: conf.NoProxy,
			}}...)
	}

	return podEnv
}

// Verify PullPolicy is acceptable
func checkPullPolicy(policy string) (v1.PullPolicy, error) {
	n := map[string]v1.PullPolicy{
		"always":       v1.PullAlways,
		"never":        v1.PullNever,
		"ifnotpresent": v1.PullIfNotPresent,
	}
	p := strings.ToLower(policy)
	value, _ := n[p]
	if value == "" {
		return "", fmt.Errorf("ImagePullPolicy: %s not found in [%s, %s, %s,]", policy, v1.PullAlways, v1.PullNever, v1.PullIfNotPresent)
	}

	return value, nil
}

func copySecretsToNamespace(
	executionContext ExecutionContext,
	clusterConfig ClusterConfig,
	k8scli *clients.KubernetesClient,
	secrets []string,
) error {
	for _, secrectName := range secrets {
		secretData, err := k8scli.Client.CoreV1().Secrets(clusterConfig.Namespace).Get(secrectName, metav1.GetOptions{})
		if err != nil {
			return err
		}
		oldMeta := secretData.ObjectMeta
		secretData.ObjectMeta = metav1.ObjectMeta{Name: oldMeta.Name, Namespace: executionContext.Namespace, Labels: oldMeta.Labels, Annotations: oldMeta.Annotations}
		_, err = k8scli.Client.CoreV1().Secrets(executionContext.Namespace).Create(secretData)
		if err != nil {
			return err
		}
	}
	return nil
}
