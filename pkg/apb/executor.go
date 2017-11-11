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

package apb

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	logging "github.com/op/go-logging"
	"github.com/openshift/ansible-service-broker/pkg/clients"
	"github.com/openshift/ansible-service-broker/pkg/metrics"
	"github.com/pborman/uuid"
	"k8s.io/kubernetes/pkg/api/v1"
	"k8s.io/kubernetes/pkg/client/clientset_generated/clientset"
)

// ExecuteApb - Runs an APB Action with a provided set of inputs
func ExecuteApb(
	action string,
	clusterConfig ClusterConfig,
	spec *Spec,
	context *Context,
	p *Parameters,
	log *logging.Logger,
) (ExecutionContext, error) {
	executionContext := ExecutionContext{}
	extraVars, err := createExtraVars(context, p)

	if err != nil {
		return executionContext, err
	}

	log.Debug("ExecutingApb:")
	log.Debug("name:[ %s ]", spec.FQName)
	log.Debug("image:[ %s ]", spec.Image)
	log.Debug("action:[ %s ]", action)
	log.Debug("pullPolciy:[ %s ]", clusterConfig.PullPolicy)
	log.Debug("role:[ %s ]", clusterConfig.SandboxRole)
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

	secrets := GetSecrets(spec)
	k8scli, err := clients.Kubernetes(log)
	if err != nil {
		return executionContext, err
	}

	labels := map[string]string{
		"apb-fqname": spec.FQName,
		"apb-action": action,
	}

	executionContext.Targets = append(executionContext.Targets, context.Namespace)
	// Create namespace.
	namespace := v1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Labels:       labels,
			GenerateName: fmt.Sprintf("%s-%.4s-", spec.FQName, action),
		},
	}
	ns, err := k8scli.CoreV1().Namespaces().Create(&namespace)
	if err != nil {
		return executionContext, err
	}
	//Sandbox (i.e Namespace) was created.
	metrics.SandboxCreated()
	executionContext.Namespace = ns.ObjectMeta.Name
	executionContext.PodName = fmt.Sprintf("apb-%s", uuid.New())
	err = copySecretsToNamespace(executionContext, clusterConfig, k8scli, secrets)
	if err != nil {
		log.Errorf("unable to copy secrets: %v to  new namespace", secrets)
		return executionContext, err
	}

	sam := NewServiceAccountManager(log)
	executionContext.ServiceAccount, err = sam.CreateApbSandbox(executionContext, clusterConfig.SandboxRole)
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
					Name:  "apb",
					Image: spec.Image,
					Args: []string{
						action,
						"--extra-vars",
						extraVars,
					},
					ImagePullPolicy: pullPolicy,
					VolumeMounts:    volumeMounts,
				},
			},
			RestartPolicy:      v1.RestartPolicyNever,
			ServiceAccountName: executionContext.ServiceAccount,
			Volumes:            volumes,
		},
	}

	log.Notice(fmt.Sprintf("Creating pod %q in the %s namespace", pod.Name, executionContext.Namespace))
	_, err = k8scli.CoreV1().Pods(executionContext.Namespace).Create(pod)
	return executionContext, err
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
		paramsCopy["namespace"] = context.Namespace
	}
	extraVars, err := json.Marshal(paramsCopy)
	return string(extraVars), err
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
	k8scli *clientset.Clientset,
	secrets []string,
) error {
	for _, secrectName := range secrets {
		secretData, err := k8scli.CoreV1().Secrets(clusterConfig.Namespace).Get(secrectName, metav1.GetOptions{})
		if err != nil {
			return err
		}
		oldMeta := secretData.ObjectMeta
		secretData.ObjectMeta = metav1.ObjectMeta{Name: oldMeta.Name, Namespace: executionContext.Namespace, Labels: oldMeta.Labels, Annotations: oldMeta.Annotations}
		_, err = k8scli.CoreV1().Secrets(executionContext.Namespace).Create(secretData)
		if err != nil {
			return err
		}
	}
	return nil
}
