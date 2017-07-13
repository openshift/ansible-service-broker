package handler

import (
	"encoding/base64"
	"fmt"
	"net/http"
	"net/http/httputil"
	"os"
	"strconv"

	yaml "gopkg.in/yaml.v1"
	"k8s.io/apimachinery/pkg/api/errors"

	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	logging "github.com/op/go-logging"
	"github.com/openshift/ansible-service-broker/pkg/apb"
	"github.com/openshift/ansible-service-broker/pkg/broker"
	"github.com/pborman/uuid"
)

// TODO: implement asynchronous operations
// TODO: authentication / authorization

type handler struct {
	router       mux.Router
	broker       broker.Broker
	log          *logging.Logger
	brokerConfig broker.Config
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
func NewHandler(b broker.Broker, log *logging.Logger, brokerConfig broker.Config) http.Handler {
	h := handler{
		router:       *mux.NewRouter(),
		broker:       b,
		log:          log,
		brokerConfig: brokerConfig,
	}

	// TODO: Reintroduce router restriction based on API version when settled upstream
	//root := h.router.Headers("X-Broker-API-Version", "2.9").Subrouter()

	h.router.HandleFunc("/v2/bootstrap", createVarHandler(h.bootstrap)).Methods("POST")
	h.router.HandleFunc("/v2/catalog", createVarHandler(h.catalog)).Methods("GET")
	h.router.HandleFunc("/v2/service_instances/{instance_uuid}", createVarHandler(h.provision)).Methods("PUT")
	h.router.HandleFunc("/v2/service_instances/{instance_uuid}", createVarHandler(h.update)).Methods("PATCH")
	h.router.HandleFunc("/v2/service_instances/{instance_uuid}", createVarHandler(h.deprovision)).Methods("DELETE")
	h.router.HandleFunc("/v2/service_instances/{instance_uuid}/service_bindings/{binding_uuid}",
		createVarHandler(h.bind)).Methods("PUT")
	h.router.HandleFunc("/v2/service_instances/{instance_uuid}/service_bindings/{binding_uuid}",
		createVarHandler(h.unbind)).Methods("DELETE")
	h.router.HandleFunc("/v2/service_instances/{instance_uuid}/last_operation",
		createVarHandler(h.lastoperation)).Methods("GET")

	if h.brokerConfig.DevBroker {
		h.router.HandleFunc("/apb/spec", createVarHandler(h.apbAddSpec)).Methods("POST")
		h.router.HandleFunc("/apb/spec/{spec_id}", createVarHandler(h.apbRemoveSpec)).Methods("DELETE")
		h.router.HandleFunc("/apb/spec", createVarHandler(h.apbRemoveSpecs)).Methods("DELETE")
	}

	return handlers.LoggingHandler(os.Stdout, h)
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
	queryparams := r.URL.Query()

	if val, ok := queryparams["accepts_incomplete"]; ok {
		// ignore the error, if async can't be parsed it will be false
		async, _ = strconv.ParseBool(val[0])
	}

	var req *broker.ProvisionRequest
	err := readRequest(r, &req)

	if err != nil {
		writeResponse(w, http.StatusBadRequest, broker.ErrorResponse{Description: "could not read request: " + err.Error()})
		return
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

	resp, err := h.broker.Update(instanceUUID, req)

	writeDefaultResponse(w, http.StatusOK, resp, err)
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
	queryparams := r.URL.Query()
	if val, ok := queryparams["accepts_incomplete"]; ok {
		// ignore the error, if async can't be parsed it will be false
		async, _ = strconv.ParseBool(val[0])
	}

	resp, err := h.broker.Deprovision(instanceUUID, async)

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

	// process binding request
	resp, err := h.broker.Bind(instanceUUID, bindingUUID, req)

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

	resp, err := h.broker.Unbind(instanceUUID, bindingUUID)

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

	queryparams := r.URL.Query()

	// operation is rqeuired
	if val, ok := queryparams["operation"]; ok {
		req.Operation = val[0]
	} else {
		h.log.Warning(fmt.Sprintf("operation not supplied, relying solely on the instance_uuid [%s]", instanceUUID))
	}

	// service_id is optional
	if val, ok := queryparams["service_id"]; ok {
		req.ServiceID = val[0]
	}

	// plan_id is optional
	if val, ok := queryparams["plan_id"]; ok {
		req.PlanID = uuid.Parse(val[0])
	}

	resp, err := h.broker.LastOperation(instanceUUID, &req)

	writeDefaultResponse(w, http.StatusOK, resp, err)
}

// apbAddSpec - Development only route. Will be used by for local developers to add images to the catalog.
func (h handler) apbAddSpec(w http.ResponseWriter, r *http.Request, params map[string]string) {
	//Read Request for an image name

	// create helper method from MockRegistry
	ansibleBroker, ok := h.broker.(broker.DevBroker)
	if !ok {
		h.log.Errorf("unable to use broker - %T as ansible service broker", h.broker)
		writeResponse(w, http.StatusInternalServerError, broker.ErrorResponse{Description: "Internal server error"})
		return
	}
	//Decode
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
	var spec apb.Spec
	if err = yaml.Unmarshal([]byte(decodedSpecYaml), &spec); err != nil {
		h.log.Errorf("Unable to decode yaml - %v to spec err - %v", decodedSpecYaml, err)
		writeResponse(w, http.StatusBadRequest, broker.ErrorResponse{Description: "Invalid parameter yaml"})
		return
	}
	resp, err := ansibleBroker.AddSpec(spec)
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

//printRequest - will print the request with the body.
func (h handler) printRequest(req *http.Request) {
	if h.brokerConfig.OutputRequest {
		b, err := httputil.DumpRequest(req, true)
		if err != nil {
			h.log.Errorf("unable to dump request to log: %v", err)
		}
		h.log.Infof("Request: %q", b)
	}
}
