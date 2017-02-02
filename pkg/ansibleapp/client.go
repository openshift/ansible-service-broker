package ansibleapp

import (
	"fmt"
	"github.com/fsouza/go-dockerclient"
	"github.com/op/go-logging"
	"os"
	"os/exec"
)

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

var DockerSocket = "unix:///var/run/docker.sock"

type ClusterConfig struct {
	Target   string
	User     string
	Password string
}

type Client struct {
	dockerClient *docker.Client
}

func NewClient(log *logging.Logger) (*Client, error) {
	dockerClient, err := docker.NewClient(DockerSocket)
	if err != nil {
		log.Error("Could not load docker client")
		return nil, err
	}

	client := &Client{
		dockerClient: dockerClient,
	}

	return client, nil
}

func (c *Client) RunImage(
	action string,
	clusterConfig ClusterConfig,
	spec *Spec,
	_ *Parameters,
) ([]byte, error) {
	// HACK: We're expecting to run containers via go APIs rather than cli cmds
	// TODO: Expecting parameters to be passed here in the future as well
	return runCommand("docker", "run",
		"-e", fmt.Sprintf("OPENSHIFT_TARGET=%s", clusterConfig.Target),
		"-e", fmt.Sprintf("OPENSHIFT_USER=%s", clusterConfig.User),
		"-e", fmt.Sprintf("OPENSHIFT_PASS=%s", clusterConfig.Password),
		spec.Name, action)
}

func (c *Client) PullImage(imageName string) error {
	// Under what circumstances does this error out?
	c.dockerClient.PullImage(docker.PullImageOptions{
		Repository:   imageName,
		OutputStream: os.Stdout,
	}, docker.AuthConfiguration{})
	return nil
}

// HACK: really need a better way to do docker run
func runCommand(cmd string, args ...string) ([]byte, error) {
	output, err := exec.Command(cmd, args...).CombinedOutput()
	return output, err
}
