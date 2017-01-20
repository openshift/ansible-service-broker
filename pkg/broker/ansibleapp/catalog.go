package ansibleapp

import (
	"fmt"
	"github.com/fusor/ansible-service-broker/pkg/broker"
	"github.com/pborman/uuid"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"path"
)

type AnsibleAppManifest struct {
	Ansibleapps []AnsibleApp
}

type AnsibleApp struct {
	Name        string
	Description string
	Uuid        string
}

func loadManifest() *AnsibleAppManifest {
	// TODO: Could bundle this in with gobundle, hardcoding for now
	manifestFilePath := path.Join(ProjectRoot(), "etc", "ansibleapp-manifest.yaml")
	fmt.Printf("Reading manifest file %s\n", manifestFilePath)

	manifestFile, err := ioutil.ReadFile(manifestFilePath)
	if err != nil {
		panic(err)
	}

	var manifest AnsibleAppManifest
	err = yaml.Unmarshal(manifestFile, &manifest)
	if err != nil {
		panic(err)
	}

	return &manifest
}

func servicesFromManifest(manifest *AnsibleAppManifest) []broker.Service {
	services := []broker.Service{}

	for _, app := range manifest.Ansibleapps {
		services = append(services, broker.Service{
			Name:        app.Name,
			ID:          uuid.Parse(app.Uuid),
			Description: app.Description,
			Plans:       plans, // TODO? Same question Jim had
		})
	}

	return services
}

func (b Broker) Catalog() (*broker.CatalogResponse, error) {
	services := servicesFromManifest(loadManifest())
	return &broker.CatalogResponse{Services: services}, nil
}
