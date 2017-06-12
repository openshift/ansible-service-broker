package apb

import (
	"fmt"

	"github.com/op/go-logging"
)

func Deprovision(instance *ServiceInstance, log *logging.Logger) (string, error) {
	specJSON, _ := DumpJSON(instance)

	log.Notice("============================================================")
	log.Notice("                      DEPROVISIONING                        ")
	log.Notice("============================================================")
	log.Notice(fmt.Sprintf("ServiceInstance.Id: %s", instance.Id))
	log.Notice(fmt.Sprintf("ServiceInstance.Spec: %v", specJSON))
	log.Notice(fmt.Sprintf("ServiceInstance.Parameters: %v", instance.Parameters))
	log.Notice("============================================================")

	var client *Client
	var err error

	if client, err = NewClient(log); err != nil {
		return "", err
	}

	if err = client.PullImage(instance.Spec.Name); err != nil {
		return "", err
	}

	// Might need to change up this interface to feed in instance ids
	podName, err := client.RunImage(
		"deprovision", HardcodedClusterConfig, instance.Spec, instance.Context, instance.Parameters)

	if err != nil {
		log.Error("Problem running image")
		return podName, err
	}

	log.Info(string(podName))
	return podName, nil
}
