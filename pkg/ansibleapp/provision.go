package ansibleapp

import (
	"fmt"
	"github.com/fsouza/go-dockerclient"
	"github.com/op/go-logging"
	"os"
	"os/exec"
)

// HACK: really need a better way to do docker run
func runCommand(cmd string, args ...string) ([]byte, error) {
	output, err := exec.Command(cmd, args...).CombinedOutput()
	return output, err
}

/*
parameters will be 2 keys

answers {}
kubecfg {}

deprovision - delete the namespace and it tears the whole thing down.

oc delete?


route will be hardcoded, need to determine how to get that from the ansibleapp.


need to pass in cert through parameters


First cut might have to pass kubecfg from broker. FIRST SPRINT broker passes username and password.

admin/admin
*/
func pullImage(client *docker.Client, spec *Spec) error {
	// TODO: need to figure out where to send the output
	client.PullImage(docker.PullImageOptions{Repository: spec.Name, OutputStream: os.Stdout}, docker.AuthConfiguration{})
	return nil
}

func runImage(client *docker.Client, spec *Spec, parameters *Parameters) ([]byte, error) {
	// $ docker run [OPTIONS] IMAGE[:TAG|@DIGEST] [COMMAND] [ARG...]
	// docker run
	// -e "OPENSHIFT_TARGET=cap.example.com:8443"
	// -e "OPENSHIFT_USER=admin"
	// -e "OPENSHIFT_PASS=admin"
	// ansibleapp/etherpad-ansibleapp deprovision

	output, err := runCommand("docker", "run",
		"-e", "OPENSHIFT_TARGET=cap.example.com:8443",
		"-e", "OPENSHIFT_USER=admin",
		"-e", "OPENSHIFT_PASS=admin",
		"ansibleapp/etherpad-ansibleapp", "provision")

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

	// TODO: get real endpoint from somewhere
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
