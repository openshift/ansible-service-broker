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
	"k8s.io/kubernetes/pkg/apis/rbac"
	"k8s.io/kubernetes/pkg/registry/rbac/validation"

	"github.com/automationbroker/bundle-lib/bundle"
	"github.com/automationbroker/bundle-lib/clients"
	"github.com/automationbroker/config"
	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"github.com/openshift/ansible-service-broker/pkg/auth"
	"github.com/openshift/ansible-service-broker/pkg/broker"
	"github.com/openshift/ansible-service-broker/pkg/version"
	"github.com/pborman/uuid"
	log "github.com/sirupsen/logrus"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
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

type handler struct {
	router           mux.Router
	broker           broker.Broker
	brokerConfig     *config.Config
	clusterRoleRules []rbac.PolicyRule
}

// authHandler - does the authentication for the routes
func authHandler(h http.Handler, providers []auth.Provider) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

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

func userInfoHandler(h http.Handler) http.Handler {
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
func NewHandler(b broker.Broker, brokerConfig *config.Config, prefix string,
	providers []auth.Provider, clusterRoleRules []rbac.PolicyRule) http.Handler {
	h := handler{
		router:           *mux.NewRouter(),
		broker:           b,
		brokerConfig:     brokerConfig,
		clusterRoleRules: clusterRoleRules,
	}
	var s *mux.Router
	if prefix == "/" {
		s = &h.router
	} else {
		s = h.router.PathPrefix(prefix).Subrouter()
	}

	s.HandleFunc("/v2/bootstrap", createVarHandler(h.bootstrap)).Methods("POST")
	s.HandleFunc("/v2/catalog", createVarHandler(h.catalog)).Methods("GET")
	s.HandleFunc("/v2/service_instances/{instance_uuid}", createVarHandler(h.getinstance)).Methods("GET")
	s.HandleFunc("/v2/service_instances/{instance_uuid}", createVarHandler(h.provision)).Methods("PUT")
	s.HandleFunc("/v2/service_instances/{instance_uuid}", createVarHandler(h.update)).Methods("PATCH")
	s.HandleFunc("/v2/service_instances/{instance_uuid}", createVarHandler(h.deprovision)).Methods("DELETE")
	s.HandleFunc("/v2/service_instances/{instance_uuid}/service_bindings/{binding_uuid}",
		createVarHandler(h.getbind)).Methods("GET")
	s.HandleFunc("/v2/service_instances/{instance_uuid}/service_bindings/{binding_uuid}",
		createVarHandler(h.bind)).Methods("PUT")
	s.HandleFunc("/v2/service_instances/{instance_uuid}/service_bindings/{binding_uuid}",
		createVarHandler(h.unbind)).Methods("DELETE")
	s.HandleFunc("/v2/service_instances/{instance_uuid}/last_operation",
		createVarHandler(h.lastoperation)).Methods("GET")
	s.HandleFunc("/v2/service_instances/{instance_uuid}/service_bindings/{binding_uuid}/last_operation",
		createVarHandler(h.lastoperation)).Methods("GET")

	if brokerConfig.GetBool("broker.dev_broker") {
		s.HandleFunc("/v2/apb", createVarHandler(h.apbAddSpec)).Methods("POST")
		s.HandleFunc("/v2/apb/{spec_id}", createVarHandler(h.apbRemoveSpec)).Methods("DELETE")
		s.HandleFunc("/v2/apb", createVarHandler(h.apbRemoveSpecs)).Methods("DELETE")
	}

	return handlers.LoggingHandler(os.Stdout, userInfoHandler(authHandler(h, providers)))
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

func (h handler) getinstance(w http.ResponseWriter, r *http.Request, params map[string]string) {
	defer r.Body.Close()
	h.printRequest(r)

	instanceUUID := uuid.Parse(params["instance_uuid"])
	if instanceUUID == nil {
		writeResponse(w, http.StatusBadRequest, broker.ErrorResponse{Description: "invalid instance_uuid"})
		return
	}

	// TODO: typically the methods on the broker return a response this
	// was an old utility method that I'm re-purposing. I think we should
	// make this consistent with the other methods in the broker.
	si, err := h.broker.GetServiceInstance(instanceUUID)
	if err != nil {
		switch err {
		case broker.ErrorNotFound: // return 404
			writeResponse(w, http.StatusNotFound, broker.ErrorResponse{Description: err.Error()})
		default: // return 422
			writeResponse(w, http.StatusUnprocessableEntity, broker.ErrorResponse{Description: err.Error()})
		}
		return
	}

	// planParameterKey is unexported. Using the value here instead of
	// refactoring the world. Besides with the above comment, this code
	// would all live in the broker.go instead of here.
	planID, ok := (*si.Parameters)["_apb_plan_id"].(string)
	if !ok {
		log.Warning("Could not retrieve the current plan name from parameters")
	}

	sir := broker.ServiceInstanceResponse{ServiceID: si.ID.String(), PlanID: planID, Parameters: *si.Parameters}

	writeDefaultResponse(w, http.StatusOK, sir, err)
}

func (h handler) provision(w http.ResponseWriter, r *http.Request, params map[string]string) {
	defer r.Body.Close()
	h.printRequest(r)

	instanceUUID := uuid.Parse(params["instance_uuid"])
	if instanceUUID == nil {
		writeResponse(w, http.StatusBadRequest, broker.ErrorResponse{Description: "invalid instance_uuid"})
		return
	}

	// ignore the error, if async can't be parsed it will be false
	async, _ := strconv.ParseBool(r.FormValue("accepts_incomplete"))

	var req *broker.ProvisionRequest
	err := readRequest(r, &req)

	if err != nil {
		writeResponse(w, http.StatusBadRequest, broker.ErrorResponse{Description: "could not read request: " + err.Error()})
		return
	}

	userInfo, ok := r.Context().Value(UserInfoContext).(broker.UserInfo)
	if !h.brokerConfig.GetBool("broker.auto_escalate") {
		if !ok {
			log.Debugf("unable to retrieve user info from request context")
			// if no user, we should error out with bad request.
			writeResponse(w, http.StatusBadRequest, broker.ErrorResponse{
				Description: "Invalid user info from originating origin header.",
			})
			return
		}

		if ok, status, err := h.validateUser(userInfo, req.Context.Namespace); !ok {
			writeResponse(w, status, broker.ErrorResponse{Description: err.Error()})
			return
		}
	} else {
		log.Debugf("Auto Escalate has been set to true, we are escalating permissions")
	}
	// Ok let's provision this bad boy
	resp, err := h.broker.Provision(instanceUUID, req, async, userInfo)

	if err != nil {
		log.Errorf("provision error %+v", err)
		switch err {
		case broker.ErrorDuplicate:
			writeResponse(w, http.StatusConflict, broker.ProvisionResponse{})
		case broker.ErrorProvisionInProgress:
			writeResponse(w, http.StatusAccepted, resp)
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

	// ignore the error, if async can't be parsed it will be false
	async, _ := strconv.ParseBool(r.FormValue("accepts_incomplete"))

	userInfo, ok := r.Context().Value(UserInfoContext).(broker.UserInfo)
	if !h.brokerConfig.GetBool("broker.auto_escalate") {
		if !ok {
			log.Debugf("unable to retrieve user info from request context")
			// if no user, we should error out with bad request.
			writeResponse(w, http.StatusBadRequest, broker.ErrorResponse{
				Description: "Invalid user info from originating origin header.",
			})
			return
		}

		if ok, status, err := h.validateUser(userInfo, req.Context.Namespace); !ok {
			writeResponse(w, status, broker.ErrorResponse{Description: err.Error()})
			return
		}
	} else {
		log.Debugf("Auto Escalate has been set to true, we are escalating permissions")
	}

	resp, err := h.broker.Update(instanceUUID, req, async, userInfo)

	if err != nil {
		switch err {
		case broker.ErrorUpdateInProgress:
			writeResponse(w, http.StatusAccepted, resp)
		case broker.ErrorNotFound:
			writeResponse(w, http.StatusBadRequest, broker.ErrorResponse{Description: err.Error()})
		case broker.ErrorNoUpdateRequested:
			writeResponse(w, http.StatusOK, resp)
		case broker.ErrorNoUpdateRequested:
			writeResponse(w, http.StatusOK, resp)
		case broker.ErrorPlanNotFound,
			broker.ErrorParameterNotUpdatable,
			broker.ErrorParameterNotFound,
			broker.ErrorParameterUnknownEnum,
			broker.ErrorPlanUpdateNotPossible:
			writeResponse(w, http.StatusBadRequest, broker.ErrorResponse{Description: err.Error()})
		default:
			writeResponse(w, http.StatusInternalServerError, broker.ErrorResponse{Description: err.Error()})
		}
	} else if async {
		writeDefaultResponse(w, http.StatusAccepted, resp, err)
	} else {
		writeDefaultResponse(w, http.StatusOK, resp, err)
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

	// ignore the error, if async can't be parsed it will be false
	async, _ := strconv.ParseBool(r.FormValue("accepts_incomplete"))

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

	nsDeleted, err := isNamespaceDeleted(serviceInstance.Context.Namespace)
	if err != nil {
		writeResponse(w, http.StatusInternalServerError, broker.ErrorResponse{Description: err.Error()})
		return
	}

	userInfo, ok := r.Context().Value(UserInfoContext).(broker.UserInfo)
	if !h.brokerConfig.GetBool("broker.auto_escalate") {
		if !ok {
			log.Debugf("unable to retrieve user info from request context")
			// if no user, we should error out with bad request.
			writeResponse(w, http.StatusBadRequest, broker.ErrorResponse{
				Description: "Invalid user info from originating origin header.",
			})
			return
		}

		if !nsDeleted {
			ok, status, err := h.validateUser(userInfo, serviceInstance.Context.Namespace)
			if !ok {
				writeResponse(w, status, broker.ErrorResponse{Description: err.Error()})
				return
			}
		}
	} else {
		log.Debugf("Auto Escalate has been set to true, we are escalating permissions")
	}

	resp, err := h.broker.Deprovision(serviceInstance, planID, nsDeleted, async, userInfo)

	if err != nil {
		switch err {
		case broker.ErrorNotFound:
			writeResponse(w, http.StatusGone, broker.DeprovisionResponse{})
			return
		case broker.ErrorBindingExists:
			writeResponse(w, http.StatusBadRequest, broker.DeprovisionResponse{})
			return
		case broker.ErrorDeprovisionInProgress:
			writeResponse(w, http.StatusAccepted, resp)
			return
		default:
			writeResponse(w, http.StatusInternalServerError, broker.ErrorResponse{Description: err.Error()})
			return
		}
	} else if async {
		writeDefaultResponse(w, http.StatusAccepted, resp, err)
	} else {
		writeDefaultResponse(w, http.StatusCreated, resp, err)
	}
}

func (h handler) getbind(w http.ResponseWriter, r *http.Request, params map[string]string) {
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

	serviceInstance, err := h.broker.GetServiceInstance(instanceUUID)
	if err != nil {
		switch err {
		case broker.ErrorNotFound:
			writeResponse(w, http.StatusBadRequest, broker.ErrorResponse{Description: err.Error()})
		default:
			writeResponse(w, http.StatusInternalServerError, broker.ErrorResponse{Description: err.Error()})
		}
	}

	resp, err := h.broker.GetBind(serviceInstance, bindingUUID)

	if err != nil {
		switch err {
		case broker.ErrorNotFound:
			writeResponse(w, http.StatusNotFound, broker.ErrorResponse{Description: err.Error()})
		default:
			writeResponse(w, http.StatusBadRequest, broker.ErrorResponse{Description: err.Error()})
		}
		return
	}

	log.Debug("handler: bind found")
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

	// ignore the error, if async can't be parsed it will be false
	async, _ := strconv.ParseBool(r.FormValue("accepts_incomplete"))

	if !async && h.brokerConfig.GetBool("broker.launch_apb_on_bind") {
		log.Warning("launch_apb_on_bind is enabled, but accepts_incomplete is false, binding may fail")
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
		return
	}

	userInfo, ok := r.Context().Value(UserInfoContext).(broker.UserInfo)
	if !h.brokerConfig.GetBool("broker.auto_escalate") {
		if !ok {
			log.Debugf("unable to retrieve user info from request context")
			// if no user, we should error out with bad request.
			writeResponse(w, http.StatusBadRequest, broker.ErrorResponse{
				Description: "Invalid user info from originating origin header.",
			})
			return
		}

		if ok, status, err := h.validateUser(userInfo, serviceInstance.Context.Namespace); !ok {
			writeResponse(w, status, broker.ErrorResponse{Description: err.Error()})
			return
		}
	} else {
		log.Debugf("Auto Escalate has been set to true, we are escalating permissions")
	}

	// process binding request
	resp, ranAsync, err := h.broker.Bind(serviceInstance, bindingUUID, req, async, userInfo)

	if err != nil {
		switch err {
		case broker.ErrorDuplicate:
			writeResponse(w, http.StatusConflict, broker.BindResponse{})
		case broker.ErrorBindingExists:
			writeResponse(w, http.StatusOK, resp)
		case broker.ErrorNotFound:
			writeResponse(w, http.StatusBadRequest, broker.ErrorResponse{Description: err.Error()})
		default:
			writeResponse(w, http.StatusBadRequest, broker.ErrorResponse{Description: err.Error()})
		}
		return
	}
	if ranAsync {
		writeDefaultResponse(w, http.StatusAccepted, resp, err)
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

	// ignore the error, if async can't be parsed it will be false
	async, _ := strconv.ParseBool(r.FormValue("accepts_incomplete"))

	if !async && h.brokerConfig.GetBool("broker.launch_apb_on_bind") {
		log.Warning("launch_apb_on_bind is enabled, but accepts_incomplete is false, unbinding may fail")
	}

	serviceInstance, err := h.broker.GetServiceInstance(instanceUUID)
	if err != nil {
		switch err {
		case broker.ErrorNotFound:
			writeResponse(w, http.StatusGone, nil)
		default:
			writeResponse(w, http.StatusInternalServerError, broker.ErrorResponse{Description: err.Error()})
		}
		return
	}

	bindInstance, err := h.broker.GetBindInstance(bindingUUID)
	if err != nil {
		switch err {
		case broker.ErrorNotFound:
			writeResponse(w, http.StatusGone, nil)
		default:
			writeResponse(w, http.StatusInternalServerError, broker.ErrorResponse{Description: err.Error()})
		}
		return
	}

	nsDeleted, err := isNamespaceDeleted(serviceInstance.Context.Namespace)
	if err != nil {
		writeResponse(w, http.StatusInternalServerError, broker.ErrorResponse{Description: err.Error()})
		return
	}

	userInfo, ok := r.Context().Value(UserInfoContext).(broker.UserInfo)
	if !h.brokerConfig.GetBool("broker.auto_escalate") {
		if !ok {
			log.Debugf("unable to retrieve user info from request context")
			// if no user, we should error out with bad request.
			writeResponse(w, http.StatusBadRequest, broker.ErrorResponse{
				Description: "Invalid user info from originating origin header.",
			})
			return
		}
		if !nsDeleted {
			if ok, status, err := h.validateUser(userInfo, serviceInstance.Context.Namespace); !ok {
				writeResponse(w, status, broker.ErrorResponse{Description: err.Error()})
				return
			}
		}
	} else {
		log.Debugf("Auto Escalate has been set to true, we are escalating permissions")
	}

	resp, ranAsync, err := h.broker.Unbind(serviceInstance, bindInstance, planID, nsDeleted, async, userInfo)

	switch {
	case err == broker.ErrorNotFound: // return 404
		log.Debugf("Binding not found.")
		writeResponse(w, http.StatusNotFound, broker.ErrorResponse{Description: err.Error()})
	case err == broker.ErrorUnbindingInProgress:
		writeResponse(w, http.StatusAccepted, resp)
	case err != nil: // return 500
		log.Errorf("Unknown error: %v", err)
		writeResponse(w, http.StatusInternalServerError, broker.ErrorResponse{Description: err.Error()})
	case ranAsync == true: // return 202
		writeDefaultResponse(w, http.StatusAccepted, resp, err)
	default: // return 200
		writeDefaultResponse(w, http.StatusOK, resp, err)
	}
	return
}

func isNamespaceDeleted(name string) (bool, error) {
	k8scli, err := clients.Kubernetes()
	if err != nil {
		return false, err
	}

	namespace, err := k8scli.Client.CoreV1().Namespaces().Get(name, metav1.GetOptions{})
	if err != nil {
		return false, err
	}

	return namespace == nil || namespace.Status.Phase == v1.NamespaceTerminating, nil
}

func (h handler) lastoperation(w http.ResponseWriter, r *http.Request, params map[string]string) {
	defer r.Body.Close()
	h.printRequest(r)

	instanceUUID := uuid.Parse(params["instance_uuid"])
	if instanceUUID == nil {
		writeResponse(w, http.StatusBadRequest, broker.ErrorResponse{Description: "invalid instance_uuid"})
		return
	}

	// we have a binding job
	if strings.Index(r.URL.Path, "/service_bindings/") > 0 {
		bindingUUID := uuid.Parse(params["binding_uuid"])
		if bindingUUID == nil {
			writeResponse(w, http.StatusBadRequest, broker.ErrorResponse{Description: "invalid binding_uuid"})
			return
		}

		// let's see if the bindInstance exists or not. We don't need the
		// actual instance, just need to know if it is there.
		_, err := h.broker.GetBindInstance(bindingUUID)
		if err != nil {
			switch err {
			case broker.ErrorNotFound:
				writeResponse(w, http.StatusGone, make(map[string]interface{}, 1))
			default:
				writeResponse(w, http.StatusInternalServerError, broker.ErrorResponse{Description: err.Error()})
			}
			return
		}

		//
		// Since we have a binding, let's use the binding id as the instance id
		//
		instanceUUID = bindingUUID
	}

	req := broker.LastOperationRequest{}

	// operation is expected
	if op := r.FormValue("operation"); op != "" {
		req.Operation = op
	} else {
		errmsg := fmt.Sprintf("operation not supplied for a last_operation with instance_uuid [%s]", instanceUUID)
		log.Error(errmsg)
		writeResponse(w, http.StatusBadRequest, broker.ErrorResponse{Description: errmsg})
		return
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
	if err == broker.ErrorNotFound { // return 404
		writeResponse(w, http.StatusGone, broker.ErrorResponse{Description: "Job not found"})
		return
	}

	writeDefaultResponse(w, http.StatusOK, resp, err)
}

// apbAddSpec - Development only route. Will be used by for local developers to add images to the catalog.
func (h handler) apbAddSpec(w http.ResponseWriter, r *http.Request, params map[string]string) {
	log.Debug("handler::apbAddSpec")
	// Read Request for an image name

	// create helper method from MockRegistry
	ansibleBroker, ok := h.broker.(broker.DevBroker)
	if !ok {
		log.Errorf("unable to use broker - %T as ansible service broker", h.broker)
		writeResponse(w, http.StatusInternalServerError, broker.ErrorResponse{Description: "Internal server error"})
		return
	}
	// Decode
	spec64Yaml := r.FormValue("apbSpec")
	if spec64Yaml == "" {
		log.Errorf("Could not find form parameter apbSpec")
		writeResponse(w, http.StatusBadRequest, broker.ErrorResponse{Description: "Could not parameter apbSpec"})
		return
	}
	decodedSpecYaml, err := base64.StdEncoding.DecodeString(spec64Yaml)
	if err != nil {
		log.Errorf("Something went wrong decoding spec from encoded string - %v err -%v", spec64Yaml, err)
		writeResponse(w, http.StatusBadRequest, broker.ErrorResponse{Description: "Invalid parameter encoding"})
		return
	}
	log.Debug("Successfully decoded pushed spec:")
	log.Debugf("%s", decodedSpecYaml)

	var spec bundle.Spec
	if err = yaml.Unmarshal([]byte(decodedSpecYaml), &spec); err != nil {
		log.Errorf("Unable to decode yaml - %v to spec err - %v", decodedSpecYaml, err)
		writeResponse(w, http.StatusBadRequest, broker.ErrorResponse{Description: "Invalid parameter yaml"})
		return
	}
	log.Infof("Assuming pushed APB runtime version [%v]", version.MaxRuntimeVersion)
	spec.Runtime = version.MaxRuntimeVersion

	log.Debug("Unmarshalled into apb.Spec:")
	log.Debugf("%+v", spec)

	resp, err := ansibleBroker.AddSpec(spec)
	if err != nil {
		log.Errorf("An error occurred while trying to add a spec via apb push:")
		log.Errorf("%s", err.Error())
		writeResponse(w, http.StatusInternalServerError,
			broker.ErrorResponse{Description: err.Error()})
		return
	}

	writeDefaultResponse(w, http.StatusOK, resp, err)
}

func (h handler) apbRemoveSpec(w http.ResponseWriter, r *http.Request, params map[string]string) {
	ansibleBroker, ok := h.broker.(broker.DevBroker)
	if !ok {
		log.Errorf("unable to use broker - %T as ansible service broker", h.broker)
		writeResponse(w, http.StatusInternalServerError, broker.ErrorResponse{Description: "Internal server error"})
		return
	}
	specID := params["spec_id"]

	var err error
	if specID != "" {
		err = ansibleBroker.RemoveSpec(specID)
	} else {
		log.Errorf("Unable to find spec id in request")
		writeResponse(w, http.StatusBadRequest, broker.ErrorResponse{Description: "No Spec/service id found."})
		return
	}
	writeDefaultResponse(w, http.StatusNoContent, struct{}{}, err)
}

func (h handler) apbRemoveSpecs(w http.ResponseWriter, r *http.Request, params map[string]string) {
	ansibleBroker, ok := h.broker.(broker.DevBroker)
	if !ok {
		log.Errorf("unable to use broker - %T as ansible service broker", h.broker)
		writeResponse(w, http.StatusInternalServerError, broker.ErrorResponse{Description: "Internal server error"})
		return
	}
	err := ansibleBroker.RemoveSpecs()
	writeDefaultResponse(w, http.StatusNoContent, struct{}{}, err)
}

// printRequest - will print the request with the body.
func (h handler) printRequest(req *http.Request) {
	if h.brokerConfig.GetBool("broker.output_request") {
		b, err := httputil.DumpRequest(req, true)
		if err != nil {
			log.Errorf("unable to dump request to log: %v", err)
		}
		log.Infof("Request: %q", b)
	}
}

// validateUser will use the cached cluster role's rules, and retrieve
// the rules for the user in the namespace to determine if the user's roles
// can cover the  all of the cluster role's rules.
func (h handler) validateUser(userInfo broker.UserInfo, namespace string) (bool, int, error) {
	openshiftClient, err := clients.Openshift()
	if err != nil {
		return false, http.StatusInternalServerError, fmt.Errorf("Unable to connect to the cluster")
	}
	// Retrieving the rules for the user in the namespace.
	s := userInfo.Scopes
	if userInfo.Extra != nil {
		scope, ok := userInfo.Extra["scopes.authorization.openshift.io"]
		switch {
		case ok && userInfo.Scopes != nil:
			log.Infof("Unable to determine correct scope to use. Found both top level scope and scope in extras.")
			return false, http.StatusForbidden, fmt.Errorf("unable to determine correct scope to use")
		case ok:
			s = scope
		}
	}
	prs, err := openshiftClient.SubjectRulesReview(userInfo.Username, userInfo.Groups, s, namespace)
	if err != nil {
		return false, http.StatusInternalServerError, fmt.Errorf("Unable to connect to the cluster")
	}
	if covered, _ := validation.Covers(prs, h.clusterRoleRules); !covered {
		return false, http.StatusForbidden, fmt.Errorf("User does not have sufficient permissions")
	}
	return true, http.StatusOK, nil
}
