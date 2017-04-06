package apb

import (
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
) (*ExtractedCredentials, error) {
	log.Notice("============================================================")
	log.Notice("                       BINDING                              ")
	log.Notice("============================================================")
	log.Notice("Parameters: %v", parameters)
	log.Notice("============================================================")

	var client *Client
	var err error

	if client, err = NewClient(log); err != nil {
		return nil, err
	}

	if err = client.PullImage(instance.Spec.Name); err != nil {
		return nil, err
	}

	output, err := client.RunImage("bind", clusterConfig, instance.Spec, parameters)

	if err != nil {
		log.Error("Problem running image", err)
		return nil, err
	}

	return extractCredentials(output, log)
}
