package clients

import (
	docker "github.com/fsouza/go-dockerclient"
	logging "github.com/op/go-logging"
)

func NewDocker(log *logging.Logger) error {
	dockerClient, err := docker.NewClient(DockerSocket)
	if err != nil {
		log.Error("Could not load docker client")
		return err
	}

	Clients.DockerClient = dockerClient
	return nil
}
