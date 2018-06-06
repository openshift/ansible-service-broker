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
	"strings"
	"time"

	log "github.com/sirupsen/logrus"

	"github.com/automationbroker/bundle-lib/clients"
	"k8s.io/apimachinery/pkg/util/wait"
)

type openshift struct{}

func (o openshift) getRuntime() string {
	return "openshift"
}

func (o openshift) shouldJoinNetworks() (bool, PostSandboxCreate, PostSandboxDestroy) {
	ocli, err := clients.Openshift()
	if err != nil {
		log.Errorf("unable to get openshift client - %v", err)
		// Defaulting if anything goes wrong to not join the networks.
		return false, nil, nil
	}
	pluginName, err := ocli.GetClusterNetworkPlugin()
	log.Debugf("plugin for the network - %v", pluginName)
	if err != nil {
		// The plugins could not be defined. ex: when using oc cluster up
		// or a pure k8s cluster. Therefore making this a notice.
		log.Infof("unable to retrieve the network plugin, defaulting to not joining networks - %v", err)
		// Defaulting to not join the networks.
		return false, nil, nil
	}

	// Case insensitive check here because want to prepare if things change.
	if strings.ToLower(pluginName) == "redhat/openshift-ovs-multitenant" {
		log.Debugf("stating that the pluginname is multitenant - %v", pluginName)
		return true, addPodNetworks, isolatePodNetworks
	}
	return false, nil, nil
}

func addPodNetworks(pod, ns string, targetNS []string, apbRole string) error {
	log.Debugf("adding pod networks together namespace: %v, target namespaces: %v", ns, targetNS)
	// Check to make sure that we have a target namespace.
	if len(targetNS) < 1 {
		return fmt.Errorf("Can not find target namespace to add to its network")
	}
	o, err := clients.Openshift()
	if err != nil {
		return err
	}
	// Get corresponding NetNamespace for given namespace
	netns, err := o.GetNetNamespace(ns)
	if err != nil {
		return err
	}
	_, err = o.JoinNamespacesNetworks(netns, targetNS[0])
	if err != nil {
		log.Errorf("Unable to join netns: %v to nsTarget: %v", netns.Name, targetNS[0])
		return err
	}

	//  wait for some time, to determine if the change was applied correctly.
	backoff := wait.Backoff{
		Steps:    15,
		Duration: 500 * time.Millisecond,
		Factor:   1.1,
	}
	return wait.ExponentialBackoff(backoff, func() (bool, error) {
		return didAnnotationUpdate("join", netns.NetName)
	})
}

func isolatePodNetworks(pod, ns string, targetNS []string) error {
	log.Debugf("adding pod networks together namespace: %v, target namespaces: %v", ns, targetNS)
	// Check to make sure that we have a target namespace.
	if len(targetNS) < 1 {
		return fmt.Errorf("Can not find target namespace to add to its network")
	}
	o, err := clients.Openshift()
	if err != nil {
		return err
	}
	// Get corresponding NetNamespace for given namespace
	netns, err := o.GetNetNamespace(ns)
	if err != nil {
		return err
	}
	_, err = o.IsolateNamespacesNetworks(netns, targetNS[0])
	if err != nil {
		log.Errorf("Unable to join netns: %v to nsTarget: %v", netns.Name, targetNS[0])
		return err
	}

	//  wait for some time, to determine if the change was applied correctly.
	backoff := wait.Backoff{
		Steps:    15,
		Duration: 500 * time.Millisecond,
		Factor:   1.1,
	}
	return wait.ExponentialBackoff(backoff, func() (bool, error) {
		return didAnnotationUpdate("join", netns.NetName)
	})
}

func didAnnotationUpdate(action, name string) (bool, error) {
	o, err := clients.Openshift()
	if err != nil {
		return false, err
	}
	updatedNetNs, err := o.GetNetNamespace(name)
	if err != nil {
		return false, err
	}
	args := strings.Split(updatedNetNs.Annotations[clients.ChangePodNetworkAnnotation], ":")
	if args[0] == "" {
		return true, nil
	}
	switch action {
	case "isolate":
		if args[0] == "isolate" {
			return true, nil
		}
	case "join":
		if args[0] == "join" {
			return true, nil
		}
	}

	// The format of the annotation is "<action>:<namespace to join to/remove from>"
	// The only actions that we will ever see is join or isolate.
	// This means the annotation was not found to be updated, therefore this errored, and nothing changed.
	// Pod network change not applied yet
	return false, nil
}
