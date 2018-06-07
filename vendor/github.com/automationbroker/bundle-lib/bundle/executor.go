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

package bundle

import (
	"encoding/json"
	"errors"
	"os"
	"sync"

	"github.com/automationbroker/bundle-lib/runtime"
	log "github.com/sirupsen/logrus"
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
	stateManager         runtime.StateManager
	skipCreateNS         bool
}

// ExecutorConfig - configuration for the executor.
type ExecutorConfig struct {
	// This will tell the executor to use the context namespace as the
	// namespace for the bundle to be created in.
	SkipCreateNS bool
}

// NewExecutor - Creates a new Executor for running an APB.
func NewExecutor(config ExecutorConfig) Executor {
	return &executor{
		statusChan:   make(chan StatusMessage),
		lastStatus:   StatusMessage{State: StateNotYetStarted},
		skipCreateNS: config.SkipCreateNS,
		stateManager: runtime.Provider,
	}
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

	log.Debugf("executor::actionFinishedWithError[ %v ]", err.Error())

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
	exContext runtime.ExecutionContext, instance *ServiceInstance, parameters *Parameters,
) (runtime.ExecutionContext, error) {
	log.Debug("ExecutingApb:")
	log.Debugf("name:[ %s ]", instance.Spec.FQName)
	log.Debugf("image:[ %s ]", exContext.Image)
	log.Debugf("action:[ %s ]", exContext.Action)
	log.Debugf("pullPolicy:[ %s ]", clusterConfig.PullPolicy)
	log.Debugf("role:[ %s ]", clusterConfig.SandboxRole)

	// It's a critical error if a Namespace is not provided to the
	// broker because its required to know where to execute the pods and
	// sandbox them based on that Namespace. Should fail fast and loud,
	// with controlled error handling.
	if exContext.Location == "" || len(exContext.Targets) == 0 {
		errStr := "Namespace not found within request context. Cannot perform requested " + exContext.Action
		log.Error(errStr)
		return exContext, errors.New(errStr)
	}

	extraVars, err := createExtraVars(exContext.Targets[0], parameters)
	if err != nil {
		return exContext, err
	}

	secrets := getSecrets(instance.Spec)
	exContext.ProxyConfig = getProxyConfig()
	exContext.Secrets = secrets
	exContext.ExtraVars = extraVars
	exContext.Policy = clusterConfig.PullPolicy

	err = runtime.Provider.CopySecretsToNamespace(exContext, clusterConfig.Namespace, secrets)
	if err != nil {
		log.Errorf("unable to copy secrets: %v to  new namespace", secrets)
		return exContext, err
	}
	masterStateName := e.stateManager.MasterName(instance.ID.String())
	present, err := e.stateManager.StateIsPresent(masterStateName)
	if err != nil {
		return exContext, err
	}
	if present {
		log.Info("state: present for service instance copying to bundle namespace")
		// copy from master ns to execution namespace
		if err := e.stateManager.CopyState(masterStateName, exContext.BundleName, e.stateManager.MasterNamespace(), exContext.Location); err != nil {
			return exContext, err
		}
		exContext.StateName = exContext.BundleName
		exContext.StateLocation = e.stateManager.MountLocation()
	}

	exContext, err = runtime.Provider.RunBundle(exContext)
	if err != nil {
		log.Errorf("error running bundle - %v", err)
		return exContext, err
	}
	return exContext, nil
}

// TODO: Instead of putting namespace directly as a parameter, we should create a dictionary
// of apb_metadata and put context and other variables in it so we don't pollute the user
// parameter space.
func createExtraVars(targetNamespace string, parameters *Parameters) (string, error) {
	var paramsCopy Parameters
	if parameters != nil && *parameters != nil {
		paramsCopy = *parameters
	} else {
		paramsCopy = make(Parameters)
	}

	if targetNamespace != "" {
		paramsCopy[NamespaceKey] = targetNamespace
	}

	paramsCopy[ClusterKey] = runtime.Provider.GetRuntime()
	extraVars, err := json.Marshal(paramsCopy)
	return string(extraVars), err
}

// getProxyConfig - Returns a ProxyConfig based on the presence of a proxy
// configuration in the broker's environment. HTTP_PROXY, HTTPS_PROXY, and
// NO_PROXY are the relevant environment variables. If no proxy is found,
// GetProxyConfig will return nil.
func getProxyConfig() *runtime.ProxyConfig {
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

	return &runtime.ProxyConfig{
		HTTPProxy:  httpProxy,
		HTTPSProxy: httpsProxy,
		NoProxy:    noProxy,
	}
}
