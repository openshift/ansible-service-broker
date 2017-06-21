package clients

import (
	docker "github.com/fsouza/go-dockerclient"
	logging "github.com/op/go-logging"
)

func Docker(log *logging.Logger) (*docker.Client, error) {
	dockerClient, err := docker.NewClient(DockerSocket)
	if err != nil {
		log.Error("Could not load docker client")
		return nil, err
	}
	return dockerClient, nil
}
