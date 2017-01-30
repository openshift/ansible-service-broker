package ansibleapp

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/fsouza/go-dockerclient"
	"github.com/op/go-logging"
)

// HACK: really need a better way to do docker run
func runCommand(cmd string, args ...string) ([]byte, error) {
	output, err := exec.Command(cmd, args...).CombinedOutput()
	return output, err
}

func pullImage(client *docker.Client, spec *Spec) error {
	// TODO: need to figure out where to send the output
	client.PullImage(docker.PullImageOptions{Repository: spec.Name, OutputStream: os.Stdout}, docker.AuthConfiguration{})
	return nil
}

func runImage(client *docker.Client, spec *Spec, parameters *Parameters) ([]byte, error) {
	output, err := runCommand("docker", "run", spec.Name)
	return output, err
}

// TODO: Figure out the right way to allow ansibleapp to log
// It's passed in here, but that's a hard coupling point to
// github.com/op/go-logging, which is used all over the broker
// Maybe ansibleapp defines its own interface and accepts that optionally
// Little looser, but still not great
func Provision(spec *Spec, parameters *Parameters, log *logging.Logger) error {
	log.Notice("============================================================")
	log.Notice("                       PROVISIONING                         ")
	log.Notice("============================================================")
	log.Notice(fmt.Sprintf("Spec.Id: %s", spec.Id))
	log.Notice(fmt.Sprintf("Spec.Name: %s", spec.Name))
	log.Notice(fmt.Sprintf("Spec.Description: %s", spec.Description))
	log.Notice(fmt.Sprintf("Parameters: %v", parameters))
	log.Notice("============================================================")

	endpoint := "unix:///var/run/docker.sock"
	client, err := docker.NewClient(endpoint)
	if err != nil {
		log.Error("Could not load docker client")
		return err
	}

	// pull image
	pullImage(client, spec)
	output, err := runImage(client, spec, parameters)
	if err != nil {
		log.Error("Problem running image")
		return err
	}
	log.Info(string(output))

	return nil
}
