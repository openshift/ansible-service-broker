package apb

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"

	docker "github.com/fsouza/go-dockerclient"
	logging "github.com/op/go-logging"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	restclient "k8s.io/client-go/rest"

	"github.com/pborman/uuid"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
	"k8s.io/kubernetes/pkg/api/v1"
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
	context *Context,
	p *Parameters,
) (string, error) {
	// HACK: We're expecting to run containers via go APIs rather than cli cmds
	// TODO: Expecting parameters to be passed here in the future as well

	extraVars, err := createExtraVars(context, p)

	if err != nil {
		return "", err
	}

	if !clusterConfig.InCluster {
		err = c.RefreshLoginToken(clusterConfig)

		if err != nil {
			c.log.Error("Error occurred while refreshing login token! Aborting apb run.")
			c.log.Error(err.Error())
			return "", err
		}
		c.log.Notice("Login token successfully refreshed.")
	}

	c.log.Debug("clusterConfig:")
	if !clusterConfig.InCluster {
		c.log.Debug("target: [ %s ]", clusterConfig.Target)
		c.log.Debug("user: [ %s ]", clusterConfig.User)
	}
	c.log.Debug("name:[ %s ]", spec.Name)
	c.log.Debug("image:[ %s ]", spec.Image)
	c.log.Debug("action:[ %s ]", action)

	// It's a critical error if a Namespace is not provided to the
	// broker because its required to know where to execute the pods and
	// sandbox them based on that Namespace. Should fail fast and loud,
	// with controlled error handling.
	if context.Namespace == "" {
		errStr := "Namespace not found within request context. Cannot perform requested " + action
		c.log.Error(errStr)
		return "", errors.New(errStr)
	}

	ns := context.Namespace
	apbId := fmt.Sprintf("apb-%s", uuid.New())

	sam := NewServiceAccountManager(c.log)
	serviceAccountName, err := sam.CreateApbSandbox(ns, apbId)

	if err != nil {
		c.log.Error(err.Error())
		return apbId, err
	}

	pod := &v1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name: apbId,
		},
		Spec: v1.PodSpec{
			Containers: []v1.Container{
				{
					Name:  "apb",
					Image: spec.Image,
					Args: []string{
						action,
						"--extra-vars",
						extraVars,
					},
					ImagePullPolicy: v1.PullAlways,
				},
			},
			RestartPolicy:      v1.RestartPolicyNever,
			ServiceAccountName: serviceAccountName,
		},
	}

	c.log.Notice(fmt.Sprintf("Creating pod %q in the %s namespace", pod.Name, ns))
	_, err = c.ClusterClient.CoreV1().Pods(ns).Create(pod)

	return apbId, err
}

func (c *Client) PullImage(imageName string) error {
	// Under what circumstances does this error out?
	c.dockerClient.PullImage(docker.PullImageOptions{
		Repository:   imageName,
		OutputStream: os.Stdout,
	}, docker.AuthConfiguration{})
	return nil
}

func (c *Client) RefreshLoginToken(clusterConfig ClusterConfig) error {
	return OcLogin(c.log,
		"--insecure-skip-tls-verify", clusterConfig.Target,
		"-u", clusterConfig.User,
		"-p", clusterConfig.Password,
	)
}

// TODO(fabianvf): This function is also called from broker/broker.go
// We should probably move this logic out of the client to a more
// generic location.
func OcLogin(log *logging.Logger, args ...string) error {
	log.Debug("Logging into openshift...")

	fullArgs := append([]string{"login"}, args...)

	output, err := RunCommand("oc", fullArgs...)

	if err != nil {
		log.Debug(string(output))
		return err
	}
	return nil
}

// TODO: Instead of putting namespace directly as a parameter, we should create a dictionary
// of apb_metadata and put context and other variables in it so we don't pollute the user
// parameter space.
func createExtraVars(context *Context, parameters *Parameters) (string, error) {
	var paramsCopy Parameters
	if parameters != nil && *parameters != nil {
		paramsCopy = *parameters
	} else {
		paramsCopy = make(Parameters)
	}

	if context != nil {
		paramsCopy["namespace"] = context.Namespace
	}
	extraVars, err := json.Marshal(paramsCopy)
	return string(extraVars), err
}
