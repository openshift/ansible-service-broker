package apb

import (
	logging "github.com/op/go-logging"
	"github.com/pkg/errors"
)

// Deprovision - runs the abp with the deprovision action.
func Deprovision(
	instance *ServiceInstance, clusterConfig ClusterConfig, log *logging.Logger,
) (string, error) {
	log.Notice("============================================================")
	log.Notice("                      DEPROVISIONING                        ")
	log.Notice("============================================================")
	log.Noticef("ServiceInstance.Id: %s", instance.Spec.ID)
	log.Noticef("ServiceInstance.Name: %v", instance.Spec.FQName)
	log.Noticef("ServiceInstance.Image: %s", instance.Spec.Image)
	log.Noticef("ServiceInstance.Description: %s", instance.Spec.Description)
	log.Notice("============================================================")

	// Explicitly error out if image field is missing from instance.Spec
	// was introduced as a change to the apb instance.Spec to support integration
	// with the broker and still allow for providing an img path
	// Legacy ansibleapps will hit this.
	// TODO: Move this validation to a Spec creation function (yet to be created)
	if instance.Spec.Image == "" {
		log.Error("No image field found on the apb instance.Spec (apb.yaml)")
		log.Error("apb instance.Spec requires [name] and [image] fields to be separate")
		log.Error("Are you trying to run a legacy ansibleapp without an image field?")
		return "", errors.New("No image field found on instance.Spec")
	}

	// Might need to change up this interface to feed in instance ids
	podName, err := ExecuteApb(
		"deprovision", clusterConfig, instance.Spec,
		instance.Context, instance.Parameters, log,
	)
	if err != nil {
		log.Error("Problem executing apb %s", err)
		return podName, err
	}

	podOutput, err := watchPod(podName, instance.Context.Namespace, log)
	if err != nil {
		log.Errorf("Error returned from watching pod\nerror: %s", err.Error())
		log.Errorf("output: %s", podOutput)
	}

	return podName, err
}
