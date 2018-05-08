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
	"encoding/json"
	"flag"
	"fmt"
	"os"

	apb "github.com/automationbroker/bundle-lib/bundle"
	"github.com/automationbroker/bundle-lib/clients"
	crd "github.com/openshift/ansible-service-broker/pkg/dao/crd"
	etcd "github.com/openshift/ansible-service-broker/pkg/dao/etcd"
	"github.com/sirupsen/logrus"
)

var options struct {
	EtcdHost           string
	EtcdCAFile         string
	EtcdClientCert     string
	EtcdClientKey      string
	EtcdPort           int
	MigrationNamespace string
}

func init() {
	flag.IntVar(&options.EtcdPort, "port", 2379, "user '--port' option to specify the port to connect to etcd")
	flag.StringVar(&options.EtcdHost, "host", "", "host to connect to etcd server")
	flag.StringVar(&options.EtcdCAFile, "ca-file", "", "CA certificate file path to be used to connect over TLS with etcd server")
	flag.StringVar(&options.EtcdClientCert, "client-cert", "", "client cert file path to authenticate to etcd server. If used must also use client-key")
	flag.StringVar(&options.EtcdClientKey, "client-key", "", "client key file path to authenticate to etcd server. If used must also use client-cert")
	flag.StringVar(&options.MigrationNamespace, "namespace", "", "namepsace that the migration should run in")
	flag.Parse()
}

var crdDao *crd.Dao
var etcdDao *etcd.Dao

type extCredsMigration struct {
	ExtCreds  apb.ExtractedCredentials
	ID        string
	Namespace string
}

func main() {
	con := clients.EtcdConfig{
		EtcdHost:       options.EtcdHost,
		EtcdPort:       options.EtcdPort,
		EtcdCaFile:     options.EtcdCAFile,
		EtcdClientKey:  options.EtcdClientKey,
		EtcdClientCert: options.EtcdClientCert,
	}
	logrus.SetLevel(logrus.DebugLevel)
	clients.InitEtcdConfig(con)
	logrus.Infof("etcd configuration: %v", con)
	var err error
	etcdDao, err = etcd.NewDao()
	if err != nil {
		panic(fmt.Sprintf("Unable to connect to etcd - %v", err))
	}
	crdDao, err = crd.NewDao(options.MigrationNamespace)
	if err != nil {
		panic(fmt.Sprintf("Unable to create crd client - %v", err))
	}
	k8scli, err := clients.Kubernetes()
	if err != nil {
		panic(fmt.Sprintf("Unable to get kubernetes client - %v", err))
	}

	//
	// convert specs to bundles
	//
	etcdSpecs, err := etcdDao.BatchGetSpecs("/spec")
	if err != nil && !etcdDao.IsNotFoundError(err) {
		panic(fmt.Sprintf("Unable to get all specs from etcd - %v", err))
	}
	crdSavedSpecs := []*apb.Spec{}
	for _, s := range etcdSpecs {
		err := crdDao.SetSpec(s.ID, s)
		if err != nil {
			logrus.Errorf("unable to save crd spec - %v", err)
			revertCrdSavedSpecs(crdSavedSpecs)
			os.Exit(2)
		}
		crdSavedSpecs = append(crdSavedSpecs, s)
	}

	//
	// convert all the service instances
	//
	siSaved := []*apb.ServiceInstance{}
	siJSONStrs, err := etcdDao.BatchGetRaw("/service_instance")
	if err != nil && !etcdDao.IsNotFoundError(err) {
		revertServiceInstance(siSaved)
		revertCrdSavedSpecs(crdSavedSpecs)
		panic(fmt.Sprintf("Unable to migrate all the service instances - %v", err))
	}
	if siJSONStrs != nil {
		for _, str := range *siJSONStrs {
			si := apb.ServiceInstance{}
			err := json.Unmarshal([]byte(str), &si)
			if err != nil {
				revertServiceInstance(siSaved)
				revertCrdSavedSpecs(crdSavedSpecs)
				panic(fmt.Sprintf("Unable to migrate all the service instances json unmarshal error - %v", err))
			}
			err = crdDao.SetServiceInstance(si.ID.String(), &si)
			if err != nil {
				revertServiceInstance(siSaved)
				revertCrdSavedSpecs(crdSavedSpecs)
				panic(fmt.Sprintf("Unable to migrate all the service instances set service instance - %v", err))
			}
			siSaved = append(siSaved, &si)
		}
	}

	//
	// convert all the service bindings
	//
	biSaved := []*apb.BindInstance{}
	biJSONStrs, err := etcdDao.BatchGetRaw("/bind_instance")
	if err != nil && !etcdDao.IsNotFoundError(err) {
		revertServiceInstance(siSaved)
		revertCrdSavedSpecs(crdSavedSpecs)
		panic(fmt.Sprintf("Unable to migrate all the service instances - %v", err))
	}
	if biJSONStrs != nil {
		for _, str := range *biJSONStrs {
			bi := apb.BindInstance{}
			err := json.Unmarshal([]byte(str), &bi)
			if err != nil {
				revertBindInstance(biSaved)
				revertServiceInstance(siSaved)
				revertCrdSavedSpecs(crdSavedSpecs)
				panic(fmt.Sprintf("Unable to migrate all the binding instances - %v", err))
			}
			err = crdDao.SetBindInstance(bi.ID.String(), &bi)
			if err != nil {
				revertBindInstance(biSaved)
				revertServiceInstance(siSaved)
				revertCrdSavedSpecs(crdSavedSpecs)
				panic(fmt.Sprintf("Unable to migrate all the binding instances - %v", err))
			}
			biSaved = append(biSaved, &bi)
		}
	}

	//
	// convert job states
	//
	jsSaved := []*apb.JobState{}
	for _, si := range siSaved {
		// Convert all the job states
		jobStateNodes, err := etcdDao.BatchGetRaw(fmt.Sprintf("/state/%v/job", si.ID.String()))
		if err != nil && !etcdDao.IsNotFoundError(err) {
			revertBindInstance(biSaved)
			revertServiceInstance(siSaved)
			revertCrdSavedSpecs(crdSavedSpecs)
			panic(fmt.Sprintf("Unable to migrate all the jobs states - %v", err))
		}
		if jobStateNodes != nil {
			for _, jsStr := range *jobStateNodes {
				js := apb.JobState{}
				err := json.Unmarshal([]byte(jsStr), &js)
				if err != nil {
					revertBindInstance(biSaved)
					revertServiceInstance(siSaved)
					revertCrdSavedSpecs(crdSavedSpecs)
					panic(fmt.Sprintf("Unable to migrate all the job states- %v", err))
				}
				_, err = crdDao.SetState(si.ID.String(), js)
				if err != nil {
					revertBindInstance(biSaved)
					revertServiceInstance(siSaved)
					revertCrdSavedSpecs(crdSavedSpecs)
					panic(fmt.Sprintf("Unable to migrate all the job states  - %v", err))
				}
				jsSaved = append(jsSaved, &js)
			}
		}
	}

	//
	// Convert extracted credentials
	//
	extCredsSaved := []extCredsMigration{}
	for _, si := range siSaved {
		labels := map[string]string{"apbAction": "provision", "apbName": si.Spec.FQName}
		extcredsJSON, err := etcdDao.GetRaw(fmt.Sprintf("/extracted_credentials/%v", si.ID))
		if etcdDao.IsNotFoundError(err) {
			continue
		}
		if err != nil {
			revertExtractedCredentials(extCredsSaved)
			revertBindInstance(biSaved)
			revertServiceInstance(siSaved)
			revertCrdSavedSpecs(crdSavedSpecs)
			panic(fmt.Sprintf("Unable to migrate all extracted credentials - %v", err))
		}

		ec := apb.ExtractedCredentials{}
		err = json.Unmarshal([]byte(extcredsJSON), &ec)
		if err != nil {
			revertExtractedCredentials(extCredsSaved)
			revertBindInstance(biSaved)
			revertServiceInstance(siSaved)
			revertCrdSavedSpecs(crdSavedSpecs)
			panic(fmt.Sprintf("Unable to migrate all extracted credentials - %v", err))
		}
		err = k8scli.SaveExtractedCredentialSecret(si.ID.String(), options.MigrationNamespace, ec.Credentials, labels)
		if err != nil {
			revertExtractedCredentials(extCredsSaved)
			revertBindInstance(biSaved)
			revertServiceInstance(siSaved)
			revertCrdSavedSpecs(crdSavedSpecs)
			panic(fmt.Sprintf("Unable to migrate all the binding instances - %v", err))
		}
		extCredsSaved = append(extCredsSaved, extCredsMigration{
			ExtCreds:  ec,
			ID:        si.ID.String(),
			Namespace: options.MigrationNamespace,
		})
	}
	for _, bi := range biSaved {
		labels := map[string]string{"apbAction": "bind"}
		extcredsJSON, err := etcdDao.GetRaw(fmt.Sprintf("/extracted_credentials/%v", bi.ID))
		if err != nil && !etcdDao.IsNotFoundError(err) {
			revertExtractedCredentials(extCredsSaved)
			revertBindInstance(biSaved)
			revertServiceInstance(siSaved)
			revertCrdSavedSpecs(crdSavedSpecs)
			panic(fmt.Sprintf("Unable to migrate all extracted credentials - %v", err))
		} else if etcdDao.IsNotFoundError(err) {
			// We need to get the service instances binding.
			extcredsJSON, err = etcdDao.GetRaw(fmt.Sprintf("/extracted_credentials/%v", bi.ServiceID))
			if err != nil {
				revertExtractedCredentials(extCredsSaved)
				revertBindInstance(biSaved)
				revertServiceInstance(siSaved)
				revertCrdSavedSpecs(crdSavedSpecs)
				panic(fmt.Sprintf("Unable to migrate all extracted credentials - %v", err))
			}

		}

		ec := apb.ExtractedCredentials{}
		err = json.Unmarshal([]byte(extcredsJSON), &ec)
		if err != nil {
			revertExtractedCredentials(extCredsSaved)
			revertBindInstance(biSaved)
			revertServiceInstance(siSaved)
			revertCrdSavedSpecs(crdSavedSpecs)
			panic(fmt.Sprintf("Unable to migrate all extracted credentials - %v", err))
		}
		k8scli, err := clients.Kubernetes()
		if err != nil {
			revertExtractedCredentials(extCredsSaved)
			revertBindInstance(biSaved)
			revertServiceInstance(siSaved)
			revertCrdSavedSpecs(crdSavedSpecs)
			panic(fmt.Sprintf("Unable to migrate all extracted credentials - %v", err))
		}
		err = k8scli.SaveExtractedCredentialSecret(bi.ID.String(), options.MigrationNamespace, ec.Credentials, labels)
		if err != nil {
			revertExtractedCredentials(extCredsSaved)
			revertBindInstance(biSaved)
			revertServiceInstance(siSaved)
			revertCrdSavedSpecs(crdSavedSpecs)
			panic(fmt.Sprintf("Unable to migrate all the binding instances - %v", err))
		}
		extCredsSaved = append(extCredsSaved, extCredsMigration{
			ExtCreds:  ec,
			ID:        bi.ID.String(),
			Namespace: options.MigrationNamespace,
		})
	}
}

func revertCrdSavedSpecs(specs []*apb.Spec) {
	err := crdDao.BatchDeleteSpecs(specs)
	if err != nil {
		panic(fmt.Sprintf("revert failed - %v", err))
	}
	logrus.Infof("reverted saved specs - exiting now - migration failed")
}

func revertBindInstance(bindInstnaces []*apb.BindInstance) {
	for _, bi := range bindInstnaces {
		err := crdDao.DeleteBindInstance(bi.ID.String())
		if err != nil {
			panic(fmt.Sprintf("revert failed - %v", err))
		}
	}
	logrus.Infof("reverted saved binding instances")
}

func revertServiceInstance(siSaved []*apb.ServiceInstance) {
	for _, si := range siSaved {
		err := crdDao.DeleteServiceInstance(si.ID.String())
		if err != nil {
			panic(fmt.Sprintf("revert failed - %v", err))
		}
	}
	logrus.Infof("reverted service instances")
}

func revertExtractedCredentials(extCredsSaved []extCredsMigration) {
	for _, extCred := range extCredsSaved {
		//no DAO to delete saved state. Going to the CRD Client
		k8sclient, err := clients.Kubernetes()
		if err != nil {
			panic(fmt.Sprintf("revert failed - %v", err))
		}
		err = k8sclient.DeleteExtractedCredentialSecret(extCred.ID, extCred.Namespace)
		if err != nil {
			panic(fmt.Sprintf("revert failed - %v", err))
		}
	}
	fmt.Printf("reverted extracted credentials")
}
