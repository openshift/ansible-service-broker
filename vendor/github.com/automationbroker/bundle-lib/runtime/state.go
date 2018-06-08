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

package runtime

import (
	"fmt"

	"github.com/automationbroker/bundle-lib/clients"
	log "github.com/sirupsen/logrus"
	kerror "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	defaultNamespace     = "ansible-service-broker"
	defaultMountLocation = "/etc/apb/state"
)

// State handles the state for service bundles
type state struct {
	// this is the namespace where the master configs will be stored
	nsTarget string
	// mountLocation is where in the pod the state will be mounted
	mountLocation string
}

// StateManager defines an interface for managing state created by service bundles
type StateManager interface {
	CopyState(fromName, toName, fromNS, toNS string) error
	DeleteState(name string) error
	StateIsPresent(name string) (bool, error)
	MasterName(instanceID string) string
	MasterNamespace() string
	MountLocation() string
}

// CopyState copies the state configmap from one namespace to another
func (s state) CopyState(fromName, toName, fromNS, toNS string) error {
	log.Debugf("state: copying state from namespace %s to ns %s from name %s to name %s", fromNS, toNS, fromName, toName)
	k8s, err := clients.Kubernetes()
	if err != nil {
		return err
	}
	fromClient := k8s.Client.CoreV1().ConfigMaps(fromNS)
	toClient := k8s.Client.CoreV1().ConfigMaps(toNS)
	fromMap, err := fromClient.Get(fromName, metav1.GetOptions{})
	if err != nil {
		if kerror.IsNotFound(err) {
			log.Debug("no state configmap found to copy")
			// can't copy if there is nothing to copy
			return nil
		}
		return err
	}
	toMap, err := toClient.Get(toName, metav1.GetOptions{})
	if err != nil {
		if kerror.IsNotFound(err) {
			fromMap.Namespace = toNS
			fromMap.Name = toName
			fromMap.ResourceVersion = ""
			if _, err := toClient.Create(fromMap); err != nil {
				return err
			}
			return nil
		}
		return err
	}
	for k, v := range fromMap.Data {
		toMap.Data[k] = v
	}
	if _, err := toClient.Update(toMap); err != nil {
		return err
	}
	return nil
}

// MasterName provides a consistent name for the state object in the master namespace
func (s state) MasterName(id string) string {
	return fmt.Sprintf("%s-state", id)
}

// StateIsPresent checks to see is there an object carrying state for ServiceBundle
func (s state) StateIsPresent(stateName string) (bool, error) {
	k8s, err := clients.Kubernetes()
	if err != nil {
		return false, err
	}
	if _, err := k8s.Client.CoreV1().ConfigMaps(s.nsTarget).Get(stateName, metav1.GetOptions{}); err != nil {
		fmt.Println("client returned err ", err)
		if kerror.IsNotFound(err) {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

// DeleteState will remove the state object from the broker namespace
func (s state) DeleteState(name string) error {
	log.Debugf("state: deleting master state %s in ns %s", name, s.nsTarget)
	k8s, err := clients.Kubernetes()
	if err != nil {
		return err
	}
	if err := k8s.Client.CoreV1().ConfigMaps(s.nsTarget).Delete(name, &metav1.DeleteOptions{}); err != nil {
		if kerror.IsNotFound(err) {
			log.Debugf("state: no state configmap found. Nothing to delete")
			return nil
		}
		return err
	}
	return nil
}

// MasterNamespace returns the name of the namespace where the master state is stored
func (s state) MasterNamespace() string {
	return s.nsTarget
}

// MountLocation returns the location where the state will be mounted
func (s state) MountLocation() string {
	return s.mountLocation
}
