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

package main

import (
	"flag"
	"fmt"
	"net/http"
	"strings"

	crd "github.com/openshift/ansible-service-broker/pkg/dao/crd"

	"github.com/sirupsen/logrus"
)

var options struct {
	BrokerNamespace string
	Port            int
}

func init() {
	flag.IntVar(&options.Port, "port", 1337, "port that the dashboard-redirector should listen on")
	flag.StringVar(&options.BrokerNamespace, "namespace", "ansible-service-broker", "namespace that the broker resides in")
	flag.Parse()
}

var crdDao *crd.Dao

var ServiceInstanceID = "XXX"

func main() {
	var err error

	crdDao, err = crd.NewDao(options.BrokerNamespace)
	if err != nil {
		panic(fmt.Sprintf("Unable to create crd client - %v", err))
	}

	logrus.Info("Trying to load batch specs as a sanity check...")
	specs, err := crdDao.BatchGetSpecs("")
	if err != nil {
		panic(fmt.Sprintf("Sanity check failed! -> %v", err))
	} else {
		logrus.Info("Sanity check passed! Loaded specs: %v", specs)
	}

	http.HandleFunc("/", redirect)
	portStr := fmt.Sprintf(":%d", options.Port)

	logrus.Infof("Dashboard redirector listening on port [%s]", portStr)
	err = http.ListenAndServe(portStr, nil)
	if err != nil {
		logrus.Fatal("ListenAndServe: ", err)
	}
}

func redirect(w http.ResponseWriter, r *http.Request) {
	var errMsg string
	logrus.Info("Checking for form")

	id := r.FormValue("id")
	if id == "" {
		errMsg := "Did not find service instance id as a query param!"
		logrus.Error(errMsg)
		http.Error(w, errMsg, http.StatusBadRequest)
		return
	} else {
		logrus.Infof("Got request for service instance %s, looking up dashboard_url", id)
	}

	si, err := crdDao.GetServiceInstance(id)
	if err != nil {
		errMsg = fmt.Sprintf("Something went wrong trying to load service instance [%s] -> %s", id, err)
		logrus.Errorf(errMsg, id, err.Error())
		http.Error(w, errMsg, http.StatusInternalServerError)
		return
	}

	logrus.Info("Successfully loaded SI: %+v", si)

	if si.DashboardURL == "" {
		errMsg = fmt.Sprintf("No DashboardURL set for requested instance! %d", id)
		logrus.Infof("%s, returning 404", errMsg)
		http.Error(w, errMsg, http.StatusNotFound)
		return
	}

	logrus.Infof("DashboardURL found: %s, 301 redirecting", si.DashboardURL)
	var redirectURL string
	if !strings.HasPrefix("http", si.DashboardURL) {
		redirectURL = fmt.Sprintf("http://%s", si.DashboardURL)
	} else {
		redirectURL = si.DashboardURL
	}

	http.Redirect(w, r, redirectURL, 301)
}
