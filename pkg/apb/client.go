package apb

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	docker "github.com/fsouza/go-dockerclient"
	logging "github.com/op/go-logging"
	restclient "k8s.io/client-go/rest"

	"github.com/pborman/uuid"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
	"k8s.io/kubernetes/pkg/client/clientset_generated/clientset"
)

/*
parameters will be 2 keys

answers {}
kubecfg {}

deprovision - delete the namespace and it tears the whole thing down.

oc delete?


route will be hardcoded, need to determine how to get that from the apb.


need to pass in cert through parameters


First cut might have to pass kubecfg from broker. FIRST SPRINT broker passes username and password.

admin/admin
*/

var DockerSocket = "unix:///var/run/docker.sock"

type ClusterConfig struct {
	InCluster bool
	Target    string
	User      string
	Password  string `yaml:"pass"`
}

type Client struct {
	dockerClient  *docker.Client
	ClusterClient *clientset.Clientset
	RESTClient    restclient.Interface
	log           *logging.Logger
}

func createClientConfigFromFile(configPath string) (*restclient.Config, error) {
	clientConfig, err := clientcmd.LoadFromFile(configPath)
	if err != nil {
		return nil, err
	}

	config, err := clientcmd.NewDefaultClientConfig(*clientConfig, &clientcmd.ConfigOverrides{}).ClientConfig()
	if err != nil {
		return nil, err
	}
	return config, nil
}

func NewClient(log *logging.Logger) (*Client, error) {
	dockerClient, err := docker.NewClient(DockerSocket)
	if err != nil {
		log.Error("Could not load docker client")
		return nil, err
	}

	// NOTE: Both the external and internal client object are using the same
	// clientset library. Internal clientset normally uses a different
	// library
	clientConfig, err := restclient.InClusterConfig()
	if err != nil {
		log.Warning("Failed to create a InternalClientSet: %v.", err)

		log.Debug("Checking for a local Cluster Config")
		clientConfig, err = createClientConfigFromFile(homedir.HomeDir() + "/.kube/config")
		if err != nil {
			log.Error("Failed to create LocalClientSet")
			return nil, err
		}
	}

	clientset, err := clientset.NewForConfig(clientConfig)
	if err != nil {
		log.Error("Failed to create LocalClientSet")
		return nil, err
	}

	rest := clientset.CoreV1().RESTClient()

	client := &Client{
		dockerClient:  dockerClient,
		ClusterClient: clientset,
		RESTClient:    rest,
		log:           log,
	}

	return client, nil
}

func (c *Client) RunImage(
	action string,
	clusterConfig ClusterConfig,
	spec *Spec,
	p *Parameters,
) ([]byte, error) {
	// HACK: We're expecting to run containers via go APIs rather than cli cmds
	// TODO: Expecting parameters to be passed here in the future as well

	params, err := json.Marshal(p)
	if err != nil {
		return nil, err
	}

	////////////////////////////////////////////////////////////////////////////////
	// This needs a lot of cleanup. Broker was originally written to run
	// inside a machine that also had a running dockerd available on /var/run/docker.sock
	// If the broker is running inside of a container on ocp or k8s, this docker runtime
	// is not available to the broker's environment in the same way. This requires
	// the broker to run the apb metacontainer remotely...somehow.
	// * options for docker are ugly:
	// -> Nested docker? (bad)
	// -> Remote docker (not a ton of options?) TCP socket?
	// * We know we've got an available ocp cluster _somewhere_. Use oc run instead of
	// docker? Requires an auth'd oc client available to the broker, but that can
	// be baked into the broker's container runtime. This is done as a temporary soln.
	//
	// TODO: Need to figure out the right way to accomplish running metacontainers
	// in remote runtimes longterm.
	////////////////////////////////////////////////////////////////////////////////
	//oc run ansible-service-broker-apb --env "OPENSHIFT_TARGET=10.1.2.2:8443"
	//--env "OPENSHIFT_USER=admin" --env "OPENSHIFT_PASS=derp"
	//--image=apb/ansible-service-broker-ansibleapp --restart=Never --
	//provision -e "dockerhub_user=eriknelson" -e "dockerhub_pass=derp"
	//-e "openshift_target=10.1.2.2:8443" -e "openshift_user=admin" -e "openshift_pass=derp"

	// NOTE: Older approach when docker is easily available to the broker to run
	// metacontainers, i.e., just running on
	//return RunCommand("docker", "run",
	//"-e", fmt.Sprintf("OPENSHIFT_TARGET=%s", clusterConfig.Target),
	//"-e", fmt.Sprintf("OPENSHIFT_USER=%s", clusterConfig.User),
	//"-e", fmt.Sprintf("OPENSHIFT_PASS=%s", clusterConfig.Password),
	//spec.Name, action, "--extra-vars", string(params))

	if !clusterConfig.InCluster {
		err = c.refreshLoginToken(clusterConfig)

		if err != nil {
			c.log.Error("Error occurred while refreshing login token! Aborting apb run.")
			c.log.Error(err.Error())
			return nil, err
		}
		c.log.Notice("Login token successfully refreshed.")
	}

	c.log.Debug("Running OC run...")
	c.log.Debug("clusterConfig:")
	if !clusterConfig.InCluster {
		c.log.Debug("target: [ %s ]", clusterConfig.Target)
		c.log.Debug("user: [ %s ]", clusterConfig.User)
		c.log.Debug("password:[ %s ]", clusterConfig.Password)
	}
	c.log.Debug("name:[ %s ]", spec.Name)
	c.log.Debug("image:[ %s ]", spec.Image)
	c.log.Debug("action:[ %s ]", action)
	c.log.Debug("params:[ %s ]", string(params))

	if clusterConfig.InCluster {
		return RunCommand("oc", "run", fmt.Sprintf("aa-%s", uuid.New()),
			fmt.Sprintf("--image-pull-policy=Always"),
			fmt.Sprintf("--image=%s", spec.Image), "--restart=Never",
			"--", action, "--extra-vars", string(params))
	} else {
		return RunCommand("oc", "run", fmt.Sprintf("aa-%s", uuid.New()),
			"--env", fmt.Sprintf("OPENSHIFT_TARGET=%s", clusterConfig.Target),
			"--env", fmt.Sprintf("OPENSHIFT_USER=%s", clusterConfig.User),
			"--env", fmt.Sprintf("OPENSHIFT_PASS=%s", clusterConfig.Password),
			fmt.Sprintf("--image-pull-policy=Always"),
			fmt.Sprintf("--image=%s", spec.Image), "--restart=Never",
			"--", action, "--extra-vars", string(params))
	}
}

func (c *Client) PullImage(imageName string) error {
	// Under what circumstances does this error out?
	c.dockerClient.PullImage(docker.PullImageOptions{
		Repository:   imageName,
		OutputStream: os.Stdout,
	}, docker.AuthConfiguration{})
	return nil
}

func (c *Client) refreshLoginToken(clusterConfig ClusterConfig) error {
	return OcLogin(c.log,
		"--insecure-skip-tls-verify", clusterConfig.Target,
		"-u", clusterConfig.User,
		"-p", clusterConfig.Password,
	)
}

func OcLogin(log *logging.Logger, args ...string) error {
	log.Debug("Logging into openshift...")
	log.Debug(fmt.Sprintf("Using args: ['%s']", strings.Join(args, "', '")))

	fullArgs := append([]string{"login"}, args...)

	output, err := RunCommand("oc", fullArgs...)

	if err != nil {
		log.Debug(string(output))
		return err
	}

	log.Debug("No error reported after running oc login. Cmd output:")
	log.Debug(string(output))
	return nil
}
