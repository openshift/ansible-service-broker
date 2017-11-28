package runtime

import (
	"github.com/openshift/ansible-service-broker/pkg/clients"

	logging "github.com/op/go-logging"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	apicorev1 "k8s.io/kubernetes/pkg/api/v1"
	rbac "k8s.io/kubernetes/pkg/apis/rbac/v1beta1"
)

// Runtime - Abstraction for broker actions
type Runtime interface {
	CreateSandbox(string, string, []string, string)
}

// Provider - Variables for interacting with runtimes
type Provider struct {
	Log *logging.Logger

	coe
}

// Abstraction for actions that are different between runtimes
type coe interface{}

// Different runtimes
type openshift struct{}
type kubernetes struct{}

// CreateSandbox - Translate the broker CreateSandbox call into cluster resource calls
func (p Provider) CreateSandbox(podName string, namespace string, targets []string, apbRole string) (string, error) {
	k8scli, err := clients.Kubernetes(p.Log)
	if err != nil {
		return "", err
	}

	serviceAccount := &apicorev1.ServiceAccount{
		ObjectMeta: metav1.ObjectMeta{
			Name: podName,
		},
	}
	_, err = k8scli.Client.CoreV1().ServiceAccounts(namespace).Create(serviceAccount)
	if err != nil {
		return "", err
	}

	p.Log.Debug("Trying to create apb sandbox: [ %s ], with %s permissions in namespace %s", podName, apbRole, namespace)

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

	p.Log.Info("Successfully created apb sandbox: [ %s ], with %s permissions in namespace %s", podName, apbRole, namespace)

	return podName, nil
}
