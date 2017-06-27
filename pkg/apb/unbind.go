package apb

import (
	logging "github.com/op/go-logging"
)

// Unbind - runs the abp with the unbind action.
// TODO: Figure out the right way to allow apb to log
// It's passed in here, but that's a hard coupling point to
// github.com/op/go-logging, which is used all over the broker
// Maybe apb defines its own interface and accepts that optionally
// Little looser, but still not great
func Unbind(instance *ServiceInstance, clusterConfig ClusterConfig, log *logging.Logger) error {
	log.Notice("============================================================")
	log.Notice("                       UNBINDING                              ")
	log.Notice("============================================================")

	var client *Client
	var err error

	if client, err = NewClient(log); err != nil {
		return err
	}

	output, err := client.RunImage("unbind", clusterConfig, instance.Spec, instance.Context, instance.Parameters)
	log.Debugf("Output from unbind call to APB: %s", string(output))

	if err != nil {
		log.Error("Problem running image", err)
	}

	return err
}
