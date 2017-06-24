package clients

import (
	"errors"
	docker "github.com/fsouza/go-dockerclient"
	logging "github.com/op/go-logging"
)

func Docker(log *logging.Logger) (*docker.Client, error) {
	errMsg := "Something went wrong initializing Docker client!"
	once.Docker.Do(func() {
		client, err := newDocker(log)
		if err != nil {
			log.Error(errMsg)
			log.Error(err.Error())
			instances.Docker = clientResult{nil, err}
		}
		instances.Docker = clientResult{client, nil}
	})

	err := instances.Docker.err
	if err != nil {
		log.Error(errMsg)
		log.Error(err.Error())
		return nil, err
	}

	if client, ok := instances.Docker.client.(*docker.Client); ok {
		return client, nil
	} else {
		return nil, errors.New(errMsg)
	}
}

func newDocker(log *logging.Logger) (*docker.Client, error) {
	return docker.NewClient(DockerSocket)
}
