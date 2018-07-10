//
// Copyright (c) 2018 Red Hat, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//    http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//

package adapters

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	b64 "encoding/base64"

	"github.com/automationbroker/bundle-lib/bundle"
	"github.com/automationbroker/bundle-lib/clients"
	v1image "github.com/openshift/api/image/v1"
	log "github.com/sirupsen/logrus"
	yaml "gopkg.in/yaml.v1"
	meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var (
	errRuntimeNotFound = errors.New("runtime not found")
)

type containerConfig struct {
	Labels imageLabel `json:"Labels"`
}

type imageMetadata struct {
	ContainerConfig containerConfig `json:"ContainerConfig"`
}

const localOpenShiftName = "openshift-registry"

// LocalOpenShiftAdapter - Docker Hub Adapter
type LocalOpenShiftAdapter struct {
	Config Configuration
}

// RegistryName - Retrieve the registry name
func (r LocalOpenShiftAdapter) RegistryName() string {
	return localOpenShiftName
}

// GetImageNames - retrieve the images
func (r LocalOpenShiftAdapter) GetImageNames() ([]string, error) {
	log.Debug("LocalOpenShiftAdapter::GetImageNames")
	log.Debugf("BundleSpecLabel: %s", BundleSpecLabel)

	openshiftClient, err := clients.Openshift()
	if err != nil {
		log.Errorf("Failed to instantiate OpenShift client")
		return nil, err
	}

	imageClient := openshiftClient.Image()
	if r.Config.Namespaces == nil {
		log.Debug("Didn't find any namespace in configuration, assuming `openshift`.")
		r.Config.Namespaces = append(r.Config.Namespaces, "openshift")
	}
	imageList := []string{}
	for _, ns := range r.Config.Namespaces {
		is, err := imageClient.ImageStreams(ns).List(meta_v1.ListOptions{})
		if err != nil {
			log.Errorf("Failed to get list of imagestreams for namespace [%v]: %v", ns, err)
			continue
		}
		for _, i := range is.Items {
			imageList = append(imageList, fmt.Sprintf("%v/%v", ns, i.Name))
		}
	}
	return imageList, nil
}

// FetchSpecs - retrieve the spec for the image names.
func (r LocalOpenShiftAdapter) FetchSpecs(imageNames []string) ([]*bundle.Spec, error) {
	log.Debug("LocalOpenShiftAdapter::FetchSpecs")
	specList := []*bundle.Spec{}

	openshiftClient, err := clients.Openshift()
	if err != nil {
		log.Errorf("Failed to instantiate OpenShift client.")
		return nil, err
	}
	imageClient := openshiftClient.Image()

	if r.Config.Tag == "" {
		log.Debug("No tag specified in config, assuming `latest`")
		r.Config.Tag = "latest"
	}

	for _, image := range imageNames {
		fullName := strings.Split(image, "/")
		if len(fullName) < 2 {
			log.Errorf("Image name [%v] not in expected format, skipping.", image)
			continue
		}
		ns := fullName[0]
		iName := fullName[1]
		imTag, err := imageClient.ImageStreamTags(ns).Get(fmt.Sprintf("%v:%v", iName, r.Config.Tag), meta_v1.GetOptions{})
		if err != nil {
			log.Errorf("Failed to get image for imagestream [%v]: %v", image, err)
			continue
		}
		spec, err := r.loadSpec(imTag.Image)
		if err != nil {
			log.Errorf("Failed to load spec for [%v]: %v", image, err)
			continue
		}
		specList = append(specList, spec)
	}

	return specList, nil
}

func (r LocalOpenShiftAdapter) loadSpec(image v1image.Image) (*bundle.Spec, error) {
	log.Debug("LocalOpenShiftAdapter::LoadSpec")
	b, err := image.DockerImageMetadata.MarshalJSON()
	if err != nil {
		log.Errorf("unable to get json docker image metadata: %v", err)
		return nil, err
	}
	i := imageMetadata{}
	err = json.Unmarshal(b, &i)
	if err != nil {
		log.Errorf("unable to get unmarshal json docker image metadata: %v", err)
		return nil, err
	}
	spec := &bundle.Spec{}

	decodedSpecYaml, err := b64.StdEncoding.DecodeString(i.ContainerConfig.Labels.Spec)
	if err != nil {
		log.Errorf("Failed to decode spec: %v", err)
		return nil, err
	}
	err = yaml.Unmarshal([]byte(decodedSpecYaml), spec)
	if err != nil {
		log.Errorf("Something went wrong loading decoded spec yaml, %s", err)
		return nil, err
	}
	spec.Runtime, err = getAPBRuntimeVersion(i.ContainerConfig.Labels.Runtime)
	if err != nil {
		log.Errorf("Failed to parse image runtime version")
		return nil, errRuntimeNotFound
	}
	spec.Image = strings.Split(image.DockerImageReference, "@")[0]
	return spec, nil
}

func (r LocalOpenShiftAdapter) getServiceIP(service string, namespace string) (string, error) {
	k8s, err := clients.Kubernetes()
	if err != nil {
		return "", err
	}

	serviceData, err := k8s.Client.CoreV1().Services(namespace).Get(service, meta_v1.GetOptions{})
	if err != nil {
		log.Warningf("Unable to load service '%s' from namespace '%s'", service, namespace)
		return "", err
	}
	log.Debugf("Found service with name %v", service)

	return serviceData.Spec.ClusterIP, nil
}
