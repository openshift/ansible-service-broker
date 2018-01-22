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

package handler

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"

	"github.com/openshift/ansible-service-broker/pkg/broker"
)

func readRequest(r *http.Request, obj interface{}) error {
	if r.Header.Get("Content-Type") != "application/json" {
		return errors.New("error: invalid content-type")
	}

	return json.NewDecoder(r.Body).Decode(&obj)
}

func writeResponse(w http.ResponseWriter, code int, obj interface{}) error {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	// return json.NewEncoder(w).Encode(obj)

	// pretty-print for easier debugging
	b, err := json.Marshal(obj)
	if err != nil {
		return err
	}
	i := bytes.Buffer{}
	json.Indent(&i, b, "", "  ")
	i.WriteString("\n")
	_, err = w.Write(i.Bytes())
	return err
}

func writeDefaultResponse(w http.ResponseWriter, code int, resp interface{}, err error) error {
	if err == nil {
		return writeResponse(w, code, resp)
	}

	return writeResponse(w, http.StatusInternalServerError, broker.ErrorResponse{Description: err.Error()})
}
