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

package bundle

import (
	"fmt"

	"github.com/automationbroker/bundle-lib/runtime"
	log "github.com/sirupsen/logrus"
)

const (
	// GatherCredentialsCommand - Command used when execing for bind credentials
	// moving this constant here because eventually Extracting creds will
	// need to be moved to runtime. Therefore keeping all of this together
	// makes sense
	GatherCredentialsCommand = "broker-bind-creds"
)

var (
	// ErrExtractedCredentialsNotFound - Extracted Credentials are not found.
	ErrExtractedCredentialsNotFound = fmt.Errorf("credentials not found")
)

// RecoverExtractCredentials - Recover extracted credentials.
func RecoverExtractCredentials(podname, ns, fqname, id string, method JobMethod, targets []string, rt int) error {
	defer runtime.Provider.DestroySandbox(podname, ns, targets, clusterConfig.Namespace, clusterConfig.KeepNamespace, clusterConfig.KeepNamespaceOnError)
	credBytes, err := runtime.Provider.ExtractCredentials(podname, ns, rt)
	if err != nil {
		log.Errorf("bundle unable to extract credentials - %v", err)
		return err
	}
	creds, err := buildExtractedCredentials(credBytes)
	if err != nil {
		log.Errorf("bundle unable to build extracted credentials - %v", err)
		return err
	}
	labels := map[string]string{"apbAction": string(method), "apbName": fqname}
	err = runtime.Provider.CreateExtractedCredential(id, clusterConfig.Namespace, creds.Credentials, labels)
	if err != nil {
		log.Errorf("Bundle unable to save extracted credentials - %v", err)
		return err
	}
	return nil
}

// GetExtractedCredentials - Will get the extracted credentials for a caller of the APB package.
func GetExtractedCredentials(id string) (*ExtractedCredentials, error) {
	creds, err := runtime.Provider.GetExtractedCredential(id, clusterConfig.Namespace)
	if err != nil {
		switch {
		case err == runtime.ErrCredentialsNotFound:
			log.Debugf("extracted credential secret not found - %v", id)
			return nil, ErrExtractedCredentialsNotFound
		default:
			log.Errorf("unable to get the extracted credential secret - %v", err)
			return nil, err
		}
	}
	return &ExtractedCredentials{Credentials: creds}, nil
}

// DeleteExtractedCredentials - Will delete the extracted credentials for a caller of the APB package.
// Please use this method with caution.
func DeleteExtractedCredentials(id string) error {
	return runtime.Provider.DeleteExtractedCredential(id, clusterConfig.Namespace)
}

// SetExtractedCredentials - Will set new credentials for an id.
// Please use this method with caution.
func SetExtractedCredentials(id string, creds *ExtractedCredentials) error {
	return runtime.Provider.CreateExtractedCredential(id, clusterConfig.Namespace, creds.Credentials, nil)
}
