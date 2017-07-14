package apb

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	logging "github.com/op/go-logging"
	"github.com/openshift/ansible-service-broker/pkg/clients"
	"github.com/pborman/uuid"
	"k8s.io/kubernetes/pkg/api/v1"
)

// ExecuteApb - Runs an APB Action with a provided set of inputs
func ExecuteApb(
	action string,
	clusterConfig ClusterConfig,
	spec *Spec,
	context *Context,
	p *Parameters,
	log *logging.Logger,
) (string, error) {
	extraVars, err := createExtraVars(context, p)

	if err != nil {
		return "", err
	}

	log.Debug("ExecutingApb:")
	log.Debug("name:[ %s ]", spec.FQName)
	log.Debug("image:[ %s ]", spec.Image)
	log.Debug("action:[ %s ]", action)
	log.Debug("pullPolciy:[ %s ]", clusterConfig.PullPolicy)

	// It's a critical error if a Namespace is not provided to the
	// broker because its required to know where to execute the pods and
	// sandbox them based on that Namespace. Should fail fast and loud,
	// with controlled error handling.
	if context.Namespace == "" {
		errStr := "Namespace not found within request context. Cannot perform requested " + action
		log.Error(errStr)
		return "", errors.New(errStr)
	}

	pullPolicy, err := checkPullPolicy(clusterConfig.PullPolicy)
	if err != nil {
		return "", err
	}

	ns := context.Namespace
	apbID := fmt.Sprintf("apb-%s", uuid.New())

	sam := NewServiceAccountManager(log)
	serviceAccountName, err := sam.CreateApbSandbox(ns, apbID)

	if err != nil {
		log.Error(err.Error())
		return apbID, err
	}

	pod := &v1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name: apbID,
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
				},
			},
			RestartPolicy:      v1.RestartPolicyNever,
			ServiceAccountName: serviceAccountName,
		},
	}

	log.Notice(fmt.Sprintf("Creating pod %q in the %s namespace", pod.Name, ns))
	k8scli, err := clients.Kubernetes(log)
	if err != nil {
		return apbID, err
	}
	_, err = k8scli.CoreV1().Pods(ns).Create(pod)
	return apbID, err
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
