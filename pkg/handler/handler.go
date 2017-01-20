package handler

import (
	"net/http"

	"github.com/fusor/ansible-service-broker/pkg/broker"
	"github.com/gorilla/mux"
	"github.com/pborman/uuid"
	"k8s.io/kubernetes/pkg/api/errors"
)

// TODO: implement asynchronous operations
// TODO: authentication / authorization

type handler struct {
	router mux.Router
	broker broker.Broker
}

func NewHandler(b broker.Broker) http.Handler {
	h := handler{broker: b}

	// TODO: handle X-Broker-API-Version header, currently poorly defined
	root := h.router.Headers("X-Broker-API-Version", "2.9").Subrouter()

	root.HandleFunc("/v2/catalog", h.catalog).Methods("GET")
	root.HandleFunc("/v2/service_instances/{instance_uuid}", h.provision).Methods("PUT")
	root.HandleFunc("/v2/service_instances/{instance_uuid}", h.update).Methods("PATCH")
	root.HandleFunc("/v2/service_instances/{instance_uuid}", h.deprovision).Methods("DELETE")
	root.HandleFunc("/v2/service_instances/{instance_uuid}/service_bindings/{binding_uuid}", h.bind).Methods("PUT")
	root.HandleFunc("/v2/service_instances/{instance_uuid}/service_bindings/{binding_uuid}", h.unbind).Methods("DELETE")

	return h
}

func (h handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	h.router.ServeHTTP(w, r)
}

func (h handler) catalog(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	resp, err := h.broker.Catalog()

	writeDefaultResponse(w, http.StatusOK, resp, err)
}

func (h handler) provision(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	instanceUUID := uuid.Parse(mux.Vars(r)["instance_uuid"])
	if instanceUUID == nil {
		writeResponse(w, http.StatusBadRequest, broker.ErrorResponse{Description: "invalid instance_uuid"})
		return
	}

	var req *broker.ProvisionRequest
	err := readRequest(r, &req)
	if err != nil {
		writeResponse(w, http.StatusBadRequest, broker.ErrorResponse{Description: err.Error()})
		return
	}

	resp, err := h.broker.Provision(instanceUUID, req)

	if errors.IsNotFound(err) || errors.IsInvalid(err) {
		writeResponse(w, http.StatusBadRequest, broker.ErrorResponse{Description: err.Error()})
	} else if errors.IsAlreadyExists(err) {
		writeResponse(w, http.StatusConflict, broker.ProvisionResponse{})
	} else {
		writeDefaultResponse(w, http.StatusCreated, resp, err)
	}
}

func (h handler) update(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	instanceUUID := uuid.Parse(mux.Vars(r)["instance_uuid"])
	if instanceUUID == nil {
		writeResponse(w, http.StatusBadRequest, broker.ErrorResponse{Description: "invalid instance_uuid"})
		return
	}

	var req *broker.UpdateRequest
	if err := readRequest(r, &req); err != nil {
		writeResponse(w, http.StatusBadRequest, broker.ErrorResponse{Description: err.Error()})
		return
	}

	resp, err := h.broker.Update(instanceUUID, req)

	writeDefaultResponse(w, http.StatusOK, resp, err)
}

func (h handler) deprovision(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	instanceUUID := uuid.Parse(mux.Vars(r)["instance_uuid"])
	if instanceUUID == nil {
		writeResponse(w, http.StatusBadRequest, broker.ErrorResponse{Description: "invalid instance_uuid"})
		return
	}

	resp, err := h.broker.Deprovision(instanceUUID)

	if errors.IsNotFound(err) {
		writeResponse(w, http.StatusGone, broker.DeprovisionResponse{})
	} else {
		writeDefaultResponse(w, http.StatusOK, resp, err)
	}
}

func (h handler) bind(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	instanceUUID := uuid.Parse(mux.Vars(r)["instance_uuid"])
	if instanceUUID == nil {
		writeResponse(w, http.StatusBadRequest, broker.ErrorResponse{Description: "invalid instance_uuid"})
		return
	}

	bindingUUID := uuid.Parse(mux.Vars(r)["binding_uuid"])
	if bindingUUID == nil {
		writeResponse(w, http.StatusBadRequest, broker.ErrorResponse{Description: "invalid binding_uuid"})
		return
	}

	var req *broker.BindRequest
	if err := readRequest(r, &req); err != nil {
		writeResponse(w, http.StatusInternalServerError, broker.ErrorResponse{Description: err.Error()})
		return
	}

	resp, err := h.broker.Bind(instanceUUID, bindingUUID, req)

	writeDefaultResponse(w, http.StatusCreated, resp, err)
}

func (h handler) unbind(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	instanceUUID := uuid.Parse(mux.Vars(r)["instance_uuid"])
	if instanceUUID == nil {
		writeResponse(w, http.StatusBadRequest, broker.ErrorResponse{Description: "invalid instance_uuid"})
		return
	}

	bindingUUID := uuid.Parse(mux.Vars(r)["binding_uuid"])
	if bindingUUID == nil {
		writeResponse(w, http.StatusBadRequest, broker.ErrorResponse{Description: "invalid binding_uuid"})
		return
	}

	err := h.broker.Unbind(instanceUUID, bindingUUID)

	if errors.IsNotFound(err) {
		writeResponse(w, http.StatusGone, struct{}{})
	} else {
		writeDefaultResponse(w, http.StatusOK, struct{}{}, err)
	}
	return
}
