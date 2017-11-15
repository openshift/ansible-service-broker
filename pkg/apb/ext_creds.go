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

package apb

import (
	"encoding/base64"
	"encoding/json"
	"fmt"

	"github.com/openshift/ansible-service-broker/pkg/clients"

	logging "github.com/op/go-logging"
)

// ExtractCredentials - Extract credentials from pod in a certain namespace.
func ExtractCredentials(
	podname string,
	namespace string,
	log *logging.Logger,
) (*ExtractedCredentials, error) {

	k8scli, err := clients.Kubernetes()
	if err != nil {
		return nil, fmt.Errorf("Unable to retrive kubernetes client - %v", err)
	}

	secretData, err := k8scli.GetSecretData(podname, namespace)
	if err != nil {
		return nil, fmt.Errorf("Unable to retrieve secret [ %v ] - %v", podname, err)
	}

	return buildExtractedCredentials(secretData["fields"])
}

func buildExtractedCredentials(output []byte) (*ExtractedCredentials, error) {
	result, err := decodeOutput(output)
	if err != nil {
		return nil, err
	}

	creds := make(map[string]interface{})
	for k, v := range result {
		creds[k] = v
	}

	return &ExtractedCredentials{Credentials: creds}, nil
}

func decodeOutput(output []byte) (map[string]interface{}, error) {
	str := string(output)

	decodedjson, err := base64.StdEncoding.DecodeString(str)
	if err != nil {
		return nil, err
	}

	decoded := make(map[string]interface{})
	json.Unmarshal(decodedjson, &decoded)
	return decoded, nil
}
