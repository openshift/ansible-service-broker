package apb

import (
	logging "github.com/op/go-logging"
	"github.com/pkg/errors"
)

// Deprovision - runs the abp with the deprovision action.
func Deprovision(instance *ServiceInstance, clusterConfig ClusterConfig, log *logging.Logger) (string, error) {
	log.Notice("============================================================")
	log.Notice("                      DEPROVISIONING                        ")
	log.Notice("============================================================")
	log.Noticef("ServiceInstance.Id: %s", instance.Spec.ID)
	log.Noticef("ServiceInstance.Name: %v", instance.Spec.Name)
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

	var client *Client
	var err error

	if client, err = NewClient(log); err != nil {
		return "", err
	}

	// Might need to change up this interface to feed in instance ids
	podName, err := client.RunImage(
		"deprovision", clusterConfig, instance.Spec, instance.Context, instance.Parameters)

	if err != nil {
		log.Error("Problem running image")
		return podName, err
	}

	log.Info(string(podName))

	// Using ExtractCredentials to display output from the apb run
	// TODO: breakout the output logic from the credentials logic
	_, err = ExtractCredentials(podName, instance.Context.Namespace, log)
	return podName, err
}
