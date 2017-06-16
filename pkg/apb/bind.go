package apb

import (
	"fmt"

	logging "github.com/op/go-logging"
)

// TODO: Figure out the right way to allow apb to log
// It's passed in here, but that's a hard coupling point to
// github.com/op/go-logging, which is used all over the broker
// Maybe apb defines its own interface and accepts that optionally
// Little looser, but still not great
func Bind(
	instance *ServiceInstance,
	parameters *Parameters,
	clusterConfig ClusterConfig, log *logging.Logger,
) (string, *ExtractedCredentials, error) {
	log.Notice("============================================================")
	log.Notice("                       BINDING                              ")
	log.Notice("============================================================")
	log.Notice(fmt.Sprintf("ServiceInstance.Id: %s", instance.Spec.Id))
	log.Notice(fmt.Sprintf("ServiceInstance.Name: %v", instance.Spec.Name))
	log.Notice(fmt.Sprintf("ServiceInstance.Image: %s", instance.Spec.Image))
	log.Notice(fmt.Sprintf("ServiceInstance.Description: %s", instance.Spec.Description))
	log.Notice("============================================================")

	var client *Client
	var err error

	if client, err = NewClient(log); err != nil {
		return "", nil, err
	}

	podName, err := client.RunImage("bind", clusterConfig, instance.Spec, instance.Context, parameters)

	if err != nil {
		log.Error("Problem running image", err)
		return podName, nil, err
	}

	ns := instance.Context.Namespace
	creds, err := ExtractCredentials(podName, ns, log)
	return podName, creds, err
}
