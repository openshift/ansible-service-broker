package apb

import (
	"fmt"

	logging "github.com/op/go-logging"
)

// Bind - Will run the APB with the bind action.
func Bind(
	instance *ServiceInstance,
	parameters *Parameters,
	clusterConfig ClusterConfig,
	log *logging.Logger,
) (string, *ExtractedCredentials, error) {
	log.Notice("============================================================")
	log.Notice("                       BINDING                              ")
	log.Notice("============================================================")
	log.Notice(fmt.Sprintf("ServiceInstance.ID: %s", instance.Spec.ID))
	log.Notice(fmt.Sprintf("ServiceInstance.Name: %v", instance.Spec.FQName))
	log.Notice(fmt.Sprintf("ServiceInstance.Image: %s", instance.Spec.Image))
	log.Notice(fmt.Sprintf("ServiceInstance.Description: %s", instance.Spec.Description))
	log.Notice("============================================================")

	podName, err := ExecuteApb(
		"bind", clusterConfig, instance.Spec,
		instance.Context, parameters, log,
	)

	if err != nil {
		log.Error("Problem executing apb [%s]:", podName)
		return podName, nil, err
	}

	creds, err := ExtractCredentials(podName, instance.Context.Namespace, log)
	return podName, creds, err
}
