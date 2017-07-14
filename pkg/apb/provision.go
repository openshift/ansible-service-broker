package apb

import (
	"errors"
	"fmt"

	logging "github.com/op/go-logging"
)

// Provision - will run the abp with the provision action.
// TODO: Figure out the right way to allow apb to log
// It's passed in here, but that's a hard coupling point to
// github.com/op/go-logging, which is used all over the broker
// Maybe apb defines its own interface and accepts that optionally
// Little looser, but still not great
func Provision(
	instance *ServiceInstance,
	clusterConfig ClusterConfig, log *logging.Logger,
) (string, *ExtractedCredentials, error) {
	log.Notice("============================================================")
	log.Notice("                       PROVISIONING                         ")
	log.Notice("============================================================")
	log.Notice(fmt.Sprintf("Spec.ID: %s", instance.Spec.ID))
	log.Notice(fmt.Sprintf("Spec.Name: %s", instance.Spec.Name))
	log.Notice(fmt.Sprintf("Spec.Image: %s", instance.Spec.Image))
	log.Notice(fmt.Sprintf("Spec.Description: %s", instance.Spec.Description))
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
		return "", nil, errors.New("No image field found on instance.Spec")
	}

	ns := instance.Context.Namespace
	log.Info("Checking if project %s exists...", ns)
	if !projectExists(ns) {
		log.Error("Project %s does NOT exist! Cannot provision requested %s", ns, instance.Spec.Name)
		return "", nil, errors.New(fmt.Sprintf("Project %s does not exist", ns))
	}

	podName, err := ExecuteApb(
		"provision", clusterConfig, instance.Spec,
		instance.Context, instance.Parameters, log,
	)

	if err != nil {
		log.Errorf("Problem executing apb [%s]", podName)
		log.Error(err.Error())
		return podName, nil, err
	}

	creds, err := ExtractCredentials(podName, instance.Context.Namespace, log)
	// We should not save credentials from an app that finds them and isn't
	// bindable
	if creds != nil && !instance.Spec.Bindable {
		log.Warningf("APB %s is not bindable", instance.Spec.Name)
		log.Warningf("Ignoring Credentials")
		creds = nil
	}
	return podName, creds, err
}

func projectExists(project string) bool {
	_, _, code := RunCommandWithExitCode("oc", "get", "project", project)
	return code == 0
}
