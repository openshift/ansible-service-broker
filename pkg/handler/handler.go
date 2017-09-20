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

package handler

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httputil"
	"os"
	"strconv"
	"strings"

	yaml "gopkg.in/yaml.v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/kubernetes/pkg/apis/rbac"
	"k8s.io/kubernetes/pkg/registry/rbac/validation"

	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	logging "github.com/op/go-logging"
	"github.com/openshift/ansible-service-broker/pkg/apb"
	"github.com/openshift/ansible-service-broker/pkg/auth"
	"github.com/openshift/ansible-service-broker/pkg/broker"
	"github.com/openshift/ansible-service-broker/pkg/clients"
	"github.com/pborman/uuid"
)

// RequestContextKey - keys that will be used in the request context
type RequestContextKey string

const (
	// OriginatingIdentityHeader is the header for the originating identity
	// or the user to check/impersonate
	OriginatingIdentityHeader = "X-Broker-API-Originating-Identity"
	// UserInfoContext - Broker.UserInfo retrieved from the
	// originating identity header
	UserInfoContext RequestContextKey = "userInfo"
)

// TODO: implement asynchronous operations

type handler struct {
	router           mux.Router
	broker           broker.Broker
	log              *logging.Logger
	brokerConfig     broker.Config
	clusterRoleRules []rbac.PolicyRule
}

// authHandler - does the authentication for the routes
func authHandler(h http.Handler, providers []auth.Provider, log *logging.Logger) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		// TODO: determine what to do with the Principal. We don't really have a
		// context or a session to store it on. Do we need it past this?
		var principalFound error
		for _, provider := range providers {
			principal, err := provider.GetPrincipal(r)
			if principal != nil {
				log.Debug("We found one. HOORAY!")
				// we found our principal, stop looking
				break
			}
			if err != nil {
				principalFound = err
			}
		}
		// if we went through the providers and found no principals. We will
		// have found an error
		if principalFound != nil {
			log.Debug("no principal found")
			writeResponse(w, http.StatusUnauthorized, broker.ErrorResponse{Description: principalFound.Error()})
			return
		}

		h.ServeHTTP(w, r)
	})
}

func userInfoHandler(h http.Handler, log *logging.Logger) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		//Retrieve the UserInfo from request if available.
		userJSONStr := r.Header.Get(OriginatingIdentityHeader)
		if userJSONStr != "" {
			userStr := strings.Split(userJSONStr, " ")
			if len(userStr) != 2 {
				//If we do not understand the user, but something was sent, we should return a 404.
				log.Debugf("Not enough values in header "+
					"for originating origin header - %v", userJSONStr)
				writeResponse(w, http.StatusBadRequest, broker.ErrorResponse{
					Description: "Invalid User Info in Originating Identity Header",
				})
				return
			}
			userInfo := broker.UserInfo{}
			uStr, err := base64.StdEncoding.DecodeString(userStr[1])
			if err != nil {
				//If we do not understand the user, but something was sent, we should return a 404.
				log.Debugf("Unable to decode base64 encoding "+
					"for originating origin header - %v", err)
				writeResponse(w, http.StatusBadRequest, broker.ErrorResponse{
					Description: "Invalid User Info in Originating Identity Header",
				})
				return
			}
			err = json.Unmarshal(uStr, &userInfo)
			if err != nil {
				log.Debugf("Unable to marshal into object "+
					"for originating origin header - %v", err)
				//If we do not understand the user, but something was sent, we should return a 404.
				writeResponse(w, http.StatusBadRequest, broker.ErrorResponse{
					Description: "Invalid User Info in Originating Identity Header",
				})
				return
			}
			r = r.WithContext(context.WithValue(
				r.Context(), UserInfoContext, userInfo),
			)
		} else {
			log.Debugf("Unable to find originating origin header")
		}
		h.ServeHTTP(w, r)
	})
}

// GorillaRouteHandler - gorilla route handler
// making the handler methods more testable by moving the reliance of mux.Vars()
// outside of the handlers themselves
type GorillaRouteHandler func(http.ResponseWriter, *http.Request)

// VarHandler - Variable route handler.
type VarHandler func(http.ResponseWriter, *http.Request, map[string]string)

func createVarHandler(r VarHandler) GorillaRouteHandler {
	return func(writer http.ResponseWriter, request *http.Request) {
		r(writer, request, mux.Vars(request))
	}
}

// NewHandler - Create a new handler by attaching the routes and setting logger and broker.
func NewHandler(b broker.Broker, log *logging.Logger, brokerConfig broker.Config, prefix string,
	providers []auth.Provider, clusterRoleRules []rbac.PolicyRule,
) http.Handler {
	h := handler{
		router:           *mux.NewRouter(),
		broker:           b,
		log:              log,
		brokerConfig:     brokerConfig,
		clusterRoleRules: clusterRoleRules,
	}
	var s *mux.Router
	if prefix == "/" {
		s = &h.router
	} else {
		s = h.router.PathPrefix(prefix).Subrouter()
	}

	// TODO: Reintroduce router restriction based on API version when settled upstream
	// root := h.router.Headers("X-Broker-API-Version", "2.9").Subrouter()

	s.HandleFunc("/v2/bootstrap", createVarHandler(h.bootstrap)).Methods("POST")
	s.HandleFunc("/v2/catalog", createVarHandler(h.catalog)).Methods("GET")
	s.HandleFunc("/v2/service_instances/{instance_uuid}", createVarHandler(h.provision)).Methods("PUT")
	s.HandleFunc("/v2/service_instances/{instance_uuid}", createVarHandler(h.update)).Methods("PATCH")
	s.HandleFunc("/v2/service_instances/{instance_uuid}", createVarHandler(h.deprovision)).Methods("DELETE")
	s.HandleFunc("/v2/service_instances/{instance_uuid}/service_bindings/{binding_uuid}",
		createVarHandler(h.bind)).Methods("PUT")
	s.HandleFunc("/v2/service_instances/{instance_uuid}/service_bindings/{binding_uuid}",
		createVarHandler(h.unbind)).Methods("DELETE")
	s.HandleFunc("/v2/service_instances/{instance_uuid}/last_operation",
		createVarHandler(h.lastoperation)).Methods("GET")

	if h.brokerConfig.DevBroker {
		s.HandleFunc("/apb/spec", createVarHandler(h.apbAddSpec)).Methods("POST")
		s.HandleFunc("/apb/spec/{spec_id}", createVarHandler(h.apbRemoveSpec)).Methods("DELETE")
		s.HandleFunc("/apb/spec", createVarHandler(h.apbRemoveSpecs)).Methods("DELETE")
	}

	return handlers.LoggingHandler(os.Stdout, authHandler(userInfoHandler(h, log), providers, log))
}

func (h handler) bootstrap(w http.ResponseWriter, r *http.Request, params map[string]string) {
	defer r.Body.Close()
	h.printRequest(r)
	resp, err := h.broker.Bootstrap()
	writeDefaultResponse(w, http.StatusOK, resp, err)
}

func (h handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	h.router.ServeHTTP(w, r)
}

func (h handler) catalog(w http.ResponseWriter, r *http.Request, params map[string]string) {
	defer r.Body.Close()
	h.printRequest(r)

	resp, err := h.broker.Catalog()

	writeDefaultResponse(w, http.StatusOK, resp, err)
}

func (h handler) provision(w http.ResponseWriter, r *http.Request, params map[string]string) {
	defer r.Body.Close()
	h.printRequest(r)

	instanceUUID := uuid.Parse(params["instance_uuid"])
	if instanceUUID == nil {
		writeResponse(w, http.StatusBadRequest, broker.ErrorResponse{Description: "invalid instance_uuid"})
		return
	}

	var async bool

	// ignore the error, if async can't be parsed it will be false
	async, _ = strconv.ParseBool(r.FormValue("accepts_incomplete"))

	var req *broker.ProvisionRequest
	err := readRequest(r, &req)

	if err != nil {
		writeResponse(w, http.StatusBadRequest, broker.ErrorResponse{Description: "could not read request: " + err.Error()})
		return
	}

	if !h.brokerConfig.AutoEscalate {
		userInfo, ok := r.Context().Value(UserInfoContext).(broker.UserInfo)
		if !ok {
			h.log.Debugf("unable to retrieve user info from request context")
			// if no user, we should error out with bad request.
			writeResponse(w, http.StatusBadRequest, broker.ErrorResponse{
				Description: "Invalid user info from originating origin header.",
			})
			return
		}

		if ok, status, err := h.validateUser(userInfo.Username, req.Context.Namespace); !ok {
			writeResponse(w, status, broker.ErrorResponse{Description: err.Error()})
			return
		}
	} else {
		h.log.Debugf("Auto Escalate has been set to true, we are escalating permissions")
	}
	// Ok let's provision this bad boy
	resp, err := h.broker.Provision(instanceUUID, req, async)

	if err != nil {
		switch err {
		case broker.ErrorDuplicate:
			writeResponse(w, http.StatusConflict, broker.ProvisionResponse{})
		case broker.ErrorAlreadyProvisioned:
			writeResponse(w, http.StatusOK, resp)
		case broker.ErrorNotFound:
			writeResponse(w, http.StatusBadRequest, broker.ErrorResponse{Description: err.Error()})
		default:
			writeResponse(w, http.StatusBadRequest, broker.ErrorResponse{Description: err.Error()})
		}
	} else if async {
		writeDefaultResponse(w, http.StatusAccepted, resp, err)
	} else {
		writeDefaultResponse(w, http.StatusCreated, resp, err)
	}
}

func (h handler) update(w http.ResponseWriter, r *http.Request, params map[string]string) {
	defer r.Body.Close()
	h.printRequest(r)

	h.log.Debug(params["instance_uuid"])

	instanceUUID := uuid.Parse(params["instance_uuid"])
	if instanceUUID == nil {
		writeResponse(w, http.StatusBadRequest, broker.ErrorResponse{Description: "invalid instance_uuid"})
		return
	}

	var req *broker.UpdateRequest

	if err := readRequest(r, &req); err != nil {
		writeResponse(w, http.StatusBadRequest, broker.ErrorResponse{Description: err.Error()})
		return
	}

	var async bool

	// ignore the error, if async can't be parsed it will be false
	async, _ = strconv.ParseBool(r.FormValue("accepts_incomplete"))

	resp, err := h.broker.Update(instanceUUID, req, async)

	if err != nil {
		switch err {
		case broker.ErrorAlreadyProvisioned:
			writeResponse(w, http.StatusOK, resp)
		case broker.ErrorNotFound:
			writeResponse(w, http.StatusBadRequest, broker.ErrorResponse{Description: err.Error()})
		default:
			writeResponse(w, http.StatusBadRequest, broker.ErrorResponse{Description: err.Error()})
		}
	} else if async {
		writeDefaultResponse(w, http.StatusAccepted, resp, err)
	} else {
		writeDefaultResponse(w, http.StatusCreated, resp, err)
	}
}

func (h handler) deprovision(w http.ResponseWriter, r *http.Request, params map[string]string) {
	defer r.Body.Close()
	h.printRequest(r)

	instanceUUID := uuid.Parse(params["instance_uuid"])
	if instanceUUID == nil {
		writeResponse(w, http.StatusBadRequest, broker.ErrorResponse{Description: "invalid instance_uuid"})
		return
	}

	var async bool
	// ignore the error, if async can't be parsed it will be false
	async, _ = strconv.ParseBool(r.FormValue("accepts_incomplete"))

	planID := r.FormValue("plan_id")
	if planID == "" {
		writeResponse(w, http.StatusBadRequest, broker.ErrorResponse{Description: "deprovision request missing plan_id query parameter"})
	}

	serviceInstance, err := h.broker.GetServiceInstance(instanceUUID)
	if err != nil {
		switch err {
		case broker.ErrorNotFound:
			writeResponse(w, http.StatusGone, broker.DeprovisionResponse{})
			return
		default:
			writeResponse(w, http.StatusInternalServerError, broker.ErrorResponse{Description: err.Error()})
			return
		}
	}

	if !h.brokerConfig.AutoEscalate {
		userInfo, ok := r.Context().Value(UserInfoContext).(broker.UserInfo)
		if !ok {
			h.log.Debugf("unable to retrieve user info from request context")
			// if no user, we should error out with bad request.
			writeResponse(w, http.StatusBadRequest, broker.ErrorResponse{
				Description: "Invalid user info from originating origin header.",
			})
			return
		}

		if ok, status, err := h.validateUser(userInfo.Username, serviceInstance.Context.Namespace); !ok {
			writeResponse(w, status, broker.ErrorResponse{Description: err.Error()})
			return
		}
	} else {
		h.log.Debugf("Auto Escalate has been set to true, we are escalating permissions")
	}

	resp, err := h.broker.Deprovision(serviceInstance, planID, async)

	if err != nil {
		h.log.Debug("err for deprovision - %#v", err)
	}

	switch err {
	case broker.ErrorNotFound:
		writeResponse(w, http.StatusGone, broker.DeprovisionResponse{})
		return
	case broker.ErrorBindingExists:
		writeResponse(w, http.StatusBadRequest, broker.DeprovisionResponse{})
		return
	case broker.ErrorDeprovisionInProgress:
		writeResponse(w, http.StatusAccepted, broker.DeprovisionResponse{})
		return
	}

	writeDefaultResponse(w, http.StatusOK, resp, err)
}

func (h handler) bind(w http.ResponseWriter, r *http.Request, params map[string]string) {
	defer r.Body.Close()
	h.printRequest(r)

	// validate input uuids
	instanceUUID := uuid.Parse(params["instance_uuid"])
	if instanceUUID == nil {
		writeResponse(w, http.StatusBadRequest, broker.ErrorResponse{Description: "invalid instance_uuid"})
		return
	}

	bindingUUID := uuid.Parse(params["binding_uuid"])
	if bindingUUID == nil {
		writeResponse(w, http.StatusBadRequest, broker.ErrorResponse{Description: "invalid binding_uuid"})
		return
	}

	var req *broker.BindRequest
	if err := readRequest(r, &req); err != nil {
		writeResponse(w, http.StatusInternalServerError, broker.ErrorResponse{Description: err.Error()})
		return
	}

	serviceInstance, err := h.broker.GetServiceInstance(instanceUUID)
	if err != nil {
		switch err {
		case broker.ErrorNotFound:
			writeResponse(w, http.StatusBadRequest, broker.ErrorResponse{Description: err.Error()})
		default:
			writeResponse(w, http.StatusInternalServerError, broker.ErrorResponse{Description: err.Error()})
		}
	}

	if !h.brokerConfig.AutoEscalate {
		userInfo, ok := r.Context().Value(UserInfoContext).(broker.UserInfo)
		if !ok {
			h.log.Debugf("unable to retrieve user info from request context")
			// if no user, we should error out with bad request.
			writeResponse(w, http.StatusBadRequest, broker.ErrorResponse{
				Description: "Invalid user info from originating origin header.",
			})
			return
		}

		if ok, status, err := h.validateUser(userInfo.Username, serviceInstance.Context.Namespace); !ok {
			writeResponse(w, status, broker.ErrorResponse{Description: err.Error()})
			return
		}
	} else {
		h.log.Debugf("Auto Escalate has been set to true, we are escalating permissions")
	}

	// process binding request
	resp, err := h.broker.Bind(serviceInstance, bindingUUID, req)

	if err != nil {
		switch err {
		case broker.ErrorDuplicate:
			writeResponse(w, http.StatusConflict, broker.BindResponse{})
		case broker.ErrorAlreadyProvisioned:
			writeResponse(w, http.StatusOK, resp)
		case broker.ErrorNotFound:
			writeResponse(w, http.StatusBadRequest, broker.ErrorResponse{Description: err.Error()})
		default:
			writeResponse(w, http.StatusBadRequest, broker.ErrorResponse{Description: err.Error()})
		}
	} else {
		writeDefaultResponse(w, http.StatusCreated, resp, err)
	}
}

func (h handler) unbind(w http.ResponseWriter, r *http.Request, params map[string]string) {
	defer r.Body.Close()
	h.printRequest(r)

	instanceUUID := uuid.Parse(params["instance_uuid"])
	if instanceUUID == nil {
		writeResponse(w, http.StatusBadRequest, broker.ErrorResponse{Description: "invalid instance_uuid"})
		return
	}

	bindingUUID := uuid.Parse(params["binding_uuid"])
	if bindingUUID == nil {
		writeResponse(w, http.StatusBadRequest, broker.ErrorResponse{Description: "invalid binding_uuid"})
		return
	}
	planID := r.FormValue("plan_id")
	if planID == "" {
		writeResponse(w, http.StatusBadRequest, broker.ErrorResponse{Description: "unbind request missing plan_id query parameter"})
	}

	serviceInstance, err := h.broker.GetServiceInstance(instanceUUID)
	if err != nil {
		writeResponse(w, http.StatusGone, nil)
		return
	}

	if !h.brokerConfig.AutoEscalate {
		userInfo, ok := r.Context().Value(UserInfoContext).(broker.UserInfo)
		if !ok {
			h.log.Debugf("unable to retrieve user info from request context")
			// if no user, we should error out with bad request.
			writeResponse(w, http.StatusBadRequest, broker.ErrorResponse{
				Description: "Invalid user info from originating origin header.",
			})
			return
		}
		if ok, status, err := h.validateUser(userInfo.Username, serviceInstance.Context.Namespace); !ok {
			writeResponse(w, status, broker.ErrorResponse{Description: err.Error()})
			return
		}
	} else {
		h.log.Debugf("Auto Escalate has been set to true, we are escalating permissions")
	}

	resp, err := h.broker.Unbind(serviceInstance, bindingUUID, planID)

	if errors.IsNotFound(err) {
		writeResponse(w, http.StatusGone, resp)
	} else {
		writeDefaultResponse(w, http.StatusOK, resp, err)
	}
	return
}

func (h handler) lastoperation(w http.ResponseWriter, r *http.Request, params map[string]string) {
	defer r.Body.Close()
	h.printRequest(r)

	instanceUUID := uuid.Parse(params["instance_uuid"])
	if instanceUUID == nil {
		writeResponse(w, http.StatusBadRequest, broker.ErrorResponse{Description: "invalid instance_uuid"})
		return
	}

	req := broker.LastOperationRequest{}

	// operation is rqeuired
	if op := r.FormValue("operation"); op != "" {
		req.Operation = op
	} else {
		h.log.Warning(fmt.Sprintf("operation not supplied, relying solely on the instance_uuid [%s]", instanceUUID))
	}

	// service_id is optional
	if serviceID := r.FormValue("service_id"); serviceID != "" {
		req.ServiceID = serviceID
	}

	// plan_id is optional
	if planID := r.FormValue("plan_id"); planID != "" {
		req.PlanID = planID
	}

	resp, err := h.broker.LastOperation(instanceUUID, &req)

	writeDefaultResponse(w, http.StatusOK, resp, err)
}

// apbAddSpec - Development only route. Will be used by for local developers to add images to the catalog.
func (h handler) apbAddSpec(w http.ResponseWriter, r *http.Request, params map[string]string) {
	h.log.Debug("handler::apbAddSpec")
	// Read Request for an image name

	// create helper method from MockRegistry
	ansibleBroker, ok := h.broker.(broker.DevBroker)
	if !ok {
		h.log.Errorf("unable to use broker - %T as ansible service broker", h.broker)
		writeResponse(w, http.StatusInternalServerError, broker.ErrorResponse{Description: "Internal server error"})
		return
	}
	// Decode
	spec64Yaml := r.FormValue("apbSpec")
	if spec64Yaml == "" {
		h.log.Errorf("Could not find form parameter apbSpec")
		writeResponse(w, http.StatusBadRequest, broker.ErrorResponse{Description: "Could not parameter apbSpec"})
		return
	}
	decodedSpecYaml, err := base64.StdEncoding.DecodeString(spec64Yaml)
	if err != nil {
		h.log.Errorf("Something went wrong decoding spec from encoded string - %v err -%v", spec64Yaml, err)
		writeResponse(w, http.StatusBadRequest, broker.ErrorResponse{Description: "Invalid parameter encoding"})
		return
	}
	h.log.Debug("Successfully decoded pushed spec:")
	h.log.Debugf("%s", decodedSpecYaml)

	var spec apb.Spec
	if err = yaml.Unmarshal([]byte(decodedSpecYaml), &spec); err != nil {
		h.log.Errorf("Unable to decode yaml - %v to spec err - %v", decodedSpecYaml, err)
		writeResponse(w, http.StatusBadRequest, broker.ErrorResponse{Description: "Invalid parameter yaml"})
		return
	}

	h.log.Debug("Unmarshalled into apb.Spec:")
	h.log.Debugf("%+v", spec)

	resp, err := ansibleBroker.AddSpec(spec)
	if err != nil {
		h.log.Errorf("An error occurred while trying to add a spec via apb push:")
		h.log.Errorf("%s", err.Error())
		writeResponse(w, http.StatusInternalServerError,
			broker.ErrorResponse{Description: err.Error()})
		return
	}

	writeDefaultResponse(w, http.StatusOK, resp, err)
}

func (h handler) apbRemoveSpec(w http.ResponseWriter, r *http.Request, params map[string]string) {
	ansibleBroker, ok := h.broker.(broker.DevBroker)
	if !ok {
		h.log.Errorf("unable to use broker - %T as ansible service broker", h.broker)
		writeResponse(w, http.StatusInternalServerError, broker.ErrorResponse{Description: "Internal server error"})
		return
	}
	specID := params["spec_id"]

	var err error
	if specID != "" {
		err = ansibleBroker.RemoveSpec(specID)
	} else {
		h.log.Errorf("Unable to find spec id in request")
		writeResponse(w, http.StatusBadRequest, broker.ErrorResponse{Description: "No Spec/service id found."})
		return
	}
	writeDefaultResponse(w, http.StatusNoContent, struct{}{}, err)
}

func (h handler) apbRemoveSpecs(w http.ResponseWriter, r *http.Request, params map[string]string) {
	ansibleBroker, ok := h.broker.(broker.DevBroker)
	if !ok {
		h.log.Errorf("unable to use broker - %T as ansible service broker", h.broker)
		writeResponse(w, http.StatusInternalServerError, broker.ErrorResponse{Description: "Internal server error"})
		return
	}
	err := ansibleBroker.RemoveSpecs()
	writeDefaultResponse(w, http.StatusNoContent, struct{}{}, err)
}

// printRequest - will print the request with the body.
func (h handler) printRequest(req *http.Request) {
	if h.brokerConfig.OutputRequest {
		b, err := httputil.DumpRequest(req, true)
		if err != nil {
			h.log.Errorf("unable to dump request to log: %v", err)
		}
		h.log.Infof("Request: %q", b)
	}
}

// validateUser will use the cached cluster role's rules, and retrieve
// the rules for the user in the namespace to determine if the user's roles
// can cover the  all of the cluster role's rules.
func (h handler) validateUser(userName, namespace string) (bool, int, error) {
	openshiftClient, err := clients.Openshift(h.log)
	if err != nil {
		return false, http.StatusInternalServerError, fmt.Errorf("Unable to connect to the cluster")
	}
	// Retrieving the rules for the user in the namespace.
	prs, err := openshiftClient.SubjectRulesReview(userName, namespace, h.log)
	if err != nil {
		return false, http.StatusInternalServerError, fmt.Errorf("Unable to connect to the cluster")
	}
	if covered, _ := validation.Covers(prs, h.clusterRoleRules); !covered {
		return false, http.StatusBadRequest, fmt.Errorf("User does not have sufficient permissions")
	}
	return true, http.StatusOK, nil
}
