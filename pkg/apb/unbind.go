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
func Unbind(instance *ServiceInstance, clusterConfig ClusterConfig, log *logging.Logger, parameters *Parameters) error {
	log.Notice("============================================================")
	log.Notice("                       UNBINDING                              ")
	log.Notice("============================================================")

	// podName, err
	_, err := ExecuteApb(
		"unbind", clusterConfig, instance.Spec,
		instance.Context, parameters, log,
	)

	if err != nil {
		log.Error("Problem executing APB unbind", err)
	}

	return err
}
