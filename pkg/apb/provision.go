package apb

import (
	"errors"
	"fmt"

	logging "github.com/op/go-logging"
)

// TODO: Figure out the right way to allow apb to log
// It's passed in here, but that's a hard coupling point to
// github.com/op/go-logging, which is used all over the broker
// Maybe apb defines its own interface and accepts that optionally
// Little looser, but still not great
func Provision(
	spec *Spec, context *Context, parameters *Parameters,
	clusterConfig ClusterConfig, log *logging.Logger,
) (string, *ExtractedCredentials, error) {
	log.Notice("============================================================")
	log.Notice("                       PROVISIONING                         ")
	log.Notice("============================================================")
	log.Notice(fmt.Sprintf("Spec.Id: %s", spec.Id))
	log.Notice(fmt.Sprintf("Spec.Name: %s", spec.Name))
	log.Notice(fmt.Sprintf("Spec.Image: %s", spec.Image))
	log.Notice(fmt.Sprintf("Spec.Description: %s", spec.Description))
	log.Notice(fmt.Sprintf("Parameters: %v", parameters))
	log.Notice("============================================================")

	// Explicitly error out if image field is missing from spec
	// was introduced as a change to the apb spec to support integration
	// with the broker and still allow for providing an img path
	// Legacy ansibleapps will hit this.
	if spec.Image == "" {
		log.Error("No image field found on the apb spec (apb.yaml)")
		log.Error("apb spec requires [name] and [image] fields to be separate")
		log.Error("Are you trying to run a legacy ansibleapp without an image field?")
		return "", nil, errors.New("No image field found on Spec")
	}

	ns := context.Namespace
	log.Info("Checking if project %s exists...", ns)
	if !projectExists(ns) {
		log.Info("Project %s does NOT exist, creating project...", ns)
		output, err := createProject(ns)
		if err != nil {
			log.Error("Something went wrong creating project %s!", ns)
			log.Error(err.Error())
			return "", nil, err
		} else {
			log.Info("Successfully created project %s", ns)
			log.Debug("oc new-project output:")
			log.Debug(string(output))
		}
	} else {
		log.Info("Project %s already exists!", ns)
	}

	var client *Client
	var err error

	if client, err = NewClient(log); err != nil {
		return "", nil, err
	}

	if err = client.PullImage(spec.Image); err != nil {
		return "", nil, err
	}

	podName, err := client.RunImage("provision", clusterConfig, spec, context, parameters)

	if err != nil {
		log.Error("Problem running image")
		log.Error(string(podName))
		log.Error(err.Error())
		return podName, nil, err
	}

	creds, err := extractCredentials(podName, context.Namespace, log)
	return podName, creds, err
}

func projectExists(project string) bool {
	_, _, code := RunCommandWithExitCode("oc", "get", "project", project)
	return code == 0
}

func createProject(project string) ([]byte, error) {
	return RunCommand("oc", "new-project", project)
}
