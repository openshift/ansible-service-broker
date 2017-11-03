//
// Copyright (c) 2017 Red Hat, Inc.
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
// Red Hat trademarks are not licensed under Apache License, Version 2.
// No permission is granted to use or replicate Red Hat trademarks that
// are incorporated in this software or its documentation.
//

package adapters

import (
	logging "github.com/op/go-logging"
	"github.com/openshift/ansible-service-broker/pkg/apb"
	"github.com/openshift/ansible-service-broker/pkg/clients"
	yaml "gopkg.in/yaml.v1"
	meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"strings"
)

const localOpenShiftName = "openshift-registry"

// LocalOpenShiftAdapter - Docker Hub Adapter
type LocalOpenShiftAdapter struct {
	Config Configuration
	Log    *logging.Logger
}

// RegistryName - Retrieve the registry name
func (r LocalOpenShiftAdapter) RegistryName() string {
	return localOpenShiftName
}

// GetImageNames - retrieve the images
func (r LocalOpenShiftAdapter) GetImageNames() ([]string, error) {
	r.Log.Debug("LocalOpenShiftAdapter::GetImageNames")
	r.Log.Debug("BundleSpecLabel: %s", BundleSpecLabel)

	openshiftClient, err := clients.Openshift(r.Log)
	if err != nil {
		r.Log.Errorf("Failed to instantiate OpenShift client")
		return nil, err
	}

	images, err := openshiftClient.ListRegistryImages(r.Log)
	if err != nil {
		r.Log.Errorf("Failed to load registry images")
		return nil, err
	}

	return images, nil
}

// FetchSpecs - retrieve the spec for the image names.
func (r LocalOpenShiftAdapter) FetchSpecs(imageNames []string) ([]*apb.Spec, error) {
	r.Log.Debug("LocalOpenShiftAdapter::FetchSpecs")
	specList := []*apb.Spec{}
	registryIP, err := r.getServiceIP("docker-registry", "default")
	if err != nil {
		r.Log.Errorf("Failed get docker-registry service information.")
		return nil, err
	}

	openshiftClient, err := clients.Openshift(r.Log)
	if err != nil {
		r.Log.Errorf("Failed to instantiate OpenShift client.")
		return nil, err
	}

	fqImages, err := openshiftClient.ConvertRegistryImagesToSpecs(r.Log, imageNames)
	if err != nil {
		r.Log.Errorf("Failed to load registry images")
		return nil, err
	}

	for _, image := range fqImages {
		spec, err := r.loadSpec(image.DecodedSpec)
		if err != nil {
			r.Log.Errorf("Failed to load image spec")
			continue
		}
		if strings.HasPrefix(image.Name, registryIP) {
			// Image has proper registry IP prefix
			spec.Image = image.Name
			namespace := strings.Split(image.Name, "/")[1]
			for _, ns := range r.Config.Namespaces {
				if ns == namespace {
					r.Log.Debugf("Image [%v] is in configured namespace [%v]. Adding to SpecList.", image.Name, ns)
					specList = append(specList, spec)
				}
			}
		} else {
			r.Log.Debugf("Image does not have proper registry IP prefix. Something went wrong.")
		}
	}

	return specList, nil
}

func (r LocalOpenShiftAdapter) loadSpec(yamlSpec []byte) (*apb.Spec, error) {
	r.Log.Debug("LocalOpenShiftAdapter::LoadSpec")
	spec := &apb.Spec{}

	err := yaml.Unmarshal(yamlSpec, spec)
	if err != nil {
		r.Log.Errorf("Something went wrong loading decoded spec yaml, %s", err)
		return nil, err
	}
	return spec, nil
}

func (r LocalOpenShiftAdapter) getServiceIP(service string, namespace string) (string, error) {
	k8scli, err := clients.Kubernetes(r.Log)
	if err != nil {
		return "", err
	}

	serviceData, err := k8scli.CoreV1().Services(namespace).Get(service, meta_v1.GetOptions{})
	if err != nil {
		r.Log.Warningf("Unable to load service '%s' from namespace '%s'", service, namespace)
		return "", err
	}
	r.Log.Debugf("Found service with name %v", service)

	return serviceData.Spec.ClusterIP, nil
}
