package dao

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/coreos/etcd/client"
	"github.com/coreos/etcd/version"
	logging "github.com/op/go-logging"
	"github.com/openshift/ansible-service-broker/pkg/apb"
	"github.com/openshift/ansible-service-broker/pkg/clients"
	"github.com/pborman/uuid"
)

type Dao struct {
	config clients.EtcdConfig
	log    *logging.Logger
	client client.Client
	kapi   client.KeysAPI // Used to interact with kvp API over HTTP
}

// NewDao - Create a new Dao object
func NewDao(config clients.EtcdConfig, log *logging.Logger) *Dao {
	dao := Dao{
		config: config,
		log:    log,
	}

	dao.client = clients.Clients.EtcdClient
	dao.kapi = client.NewKeysAPI(dao.client)
	return &dao
}

// SetRaw - Allows the setting of the value json string to the key in the kvp API.
func (d *Dao) SetRaw(key string, val string) error {
	_, err := d.kapi.Set(context.Background(), key, val /*opts*/, nil)
	return err
}

func (d *Dao) GetEtcdVersion(config clients.EtcdConfig) (string, string, error) {
	// The next etcd release (1.4) will have client.GetVersion()
	// We'll use this to test our etcd connection for now
	resp, err := http.Get("http://" + config.EtcdHost + ":" + config.EtcdPort + "/version")
	if err != nil {
		return "", "", err
	}

	defer resp.Body.Close()
	body, _ := ioutil.ReadAll(resp.Body)

	switch resp.StatusCode {
	case http.StatusOK:
		var vresp version.Versions
		if err := json.Unmarshal(body, &vresp); err != nil {
			return "", "", err
		}
		return vresp.Server, vresp.Cluster, nil
	default:
		var connectErr error
		if err := json.Unmarshal(body, &connectErr); err != nil {
			return "", "", err
		}
		return "", "", connectErr
	}
}

// GetRaw - gets a specific json string for a key from the kvp API.
func (d *Dao) GetRaw(key string) (string, error) {
	res, err := d.kapi.Get(context.Background(), key /*opts*/, nil)
	if err != nil {
		return "", err
	}

	val := res.Node.Value
	return val, nil
}

// BatchGetRaw - Get multiple  types as individual json strings
// TODO: Streaming interface? Going to need to optimize all this for
// a full-load catalog response of 10k
// This is more likely to be paged given current proposal
// In which case, we need paged Batch gets
// 2 steps?
// GET /spec/manifest [/*ordered ids*/]
// BatchGet(offset, count)?
func (d *Dao) BatchGetRaw(dir string) (*[]string, error) {
	d.log.Debug("Dao::BatchGetRaw")

	var res *client.Response
	var err error

	opts := &client.GetOptions{Recursive: true}
	if res, err = d.kapi.Get(context.Background(), dir, opts); err != nil {
		return nil, err
	}

	specNodes := res.Node.Nodes
	specCount := len(specNodes)

	d.log.Debug("Successfully loaded [ %d ] objects from etcd dir [ %s ]", specCount, dir)

	payloads := make([]string, specCount)
	for i, node := range specNodes {
		payloads[i] = node.Value
	}

	return &payloads, nil
}

// GetSpec - Retrieve the spec for the kvp API.
func (d *Dao) GetSpec(id string) (*apb.Spec, error) {
	spec := &apb.Spec{}
	if err := d.getObject(specKey(id), spec); err != nil {
		return nil, err
	}
	return spec, nil
}

// SetSpec - set spec for an id in the kvp API.
func (d *Dao) SetSpec(id string, spec *apb.Spec) error {
	return d.setObject(specKey(id), spec)
}

// DeleteSpec - Delete the spec for a given spec id.
func (d *Dao) DeleteSpec(specID string) error {
	d.log.Debug(fmt.Sprintf("Dao::DeleteSpec-> [ %s ]", specID))
	_, err := d.kapi.Delete(context.Background(), specKey(specID), nil)
	return err
}

// BatchSetSpecs - set specs based on SpecManifest in the kvp API.
func (d *Dao) BatchSetSpecs(specs apb.SpecManifest) error {
	// TODO: Is there no batch insert in the etcd api?
	for id, spec := range specs {
		err := d.SetSpec(id, spec)
		if err != nil {
			return err
		}
	}

	return nil
}

// BatchGetSpecs - Retrieve all the specs for dir.
func (d *Dao) BatchGetSpecs(dir string) ([]*apb.Spec, error) {
	payloads, err := d.BatchGetRaw(dir)

	if client.IsKeyNotFound(err) {
		return []*apb.Spec{}, nil
	} else if err != nil {
		return []*apb.Spec{}, err
	}

	specs := make([]*apb.Spec, len(*payloads))
	for i, payload := range *payloads {
		spec := &apb.Spec{}
		apb.LoadJSON(payload, spec)
		specs[i] = spec
		d.log.Debug("Batch idx [ %d ] -> [ %s ]", i, spec.Id)
	}

	return specs, nil
}

// BatchDeleteSpecs - set specs based on SpecManifest in the kvp API.
func (d *Dao) BatchDeleteSpecs(specs []*apb.Spec) error {
	// TODO: Is there no batch insert in the etcd api?
	for _, spec := range specs {
		err := d.DeleteSpec(spec.Id)
		if err != nil {
			return err
		}
	}
	return nil
}

// FindByState - Retrieve all the jobs that match state
func (d *Dao) FindJobStateByState(state apb.State) ([]apb.RecoverStatus, error) {
	d.log.Debug("Dao::FindByState")

	var res *client.Response
	var err error

	opts := &client.GetOptions{Recursive: true}
	if res, err = d.kapi.Get(context.Background(), "/state/", opts); err != nil {
		return nil, err
	}

	stateNodes := res.Node.Nodes
	stateCount := len(stateNodes)

	d.log.Debug("Successfully loaded [ %d ] jobstate objects from etcd dir [ /state/ ]", stateCount)

	var recoverstatus []apb.RecoverStatus
	for _, node := range stateNodes {
		k := fmt.Sprintf("%s/job", node.Key)

		status := apb.RecoverStatus{InstanceId: uuid.Parse(stateKeyId(node.Key))}
		jobstate := apb.JobState{}
		nodes, e := d.kapi.Get(context.Background(), k, opts)
		if e != nil {
			// if key is invalid do we keep it?
			d.log.Warning(
				fmt.Sprintf("Error processing jobstate record, moving on to next. %v", e.Error()))
			continue
		}

		for _, n := range nodes.Node.Nodes {
			apb.LoadJSON(n.Value, &jobstate)
			if jobstate.State == state {
				d.log.Debug(fmt.Sprintf(
					"Found! jobstate [%v] matched given state: [%v].", jobstate, state))
				status.State = jobstate
				recoverstatus = append(recoverstatus, status)
			} else {
				// we could probably remove this once we're happy with how this
				// works.
				d.log.Debug(fmt.Sprintf(
					"Skipping, jobstate [%v] did not match given state: [%v].", jobstate, state))
			}
		}
	}

	return recoverstatus, nil
}

// GetServiceInstance - Retrieve specific service instance from the kvp API.
func (d *Dao) GetServiceInstance(id string) (*apb.ServiceInstance, error) {
	spec := &apb.ServiceInstance{}
	if err := d.getObject(serviceInstanceKey(id), spec); err != nil {
		return nil, err
	}
	return spec, nil
}

// SetServiceInstance - Set service instance for an id in the kvp API.
func (d *Dao) SetServiceInstance(id string, serviceInstance *apb.ServiceInstance) error {
	return d.setObject(serviceInstanceKey(id), serviceInstance)
}

// DeleteServiceInstance - Delete the service instance for an service instance id.
func (d *Dao) DeleteServiceInstance(id string) error {
	d.log.Debug(fmt.Sprintf("Dao::DeleteServiceInstance -> [ %s ]", id))
	_, err := d.kapi.Delete(context.Background(), serviceInstanceKey(id), nil)
	return err
}

// GetBindInstance - Retrieve a specific bind instance from the kvp API
func (d *Dao) GetBindInstance(id string) (*apb.BindInstance, error) {
	spec := &apb.BindInstance{}
	if err := d.getObject(bindInstanceKey(id), spec); err != nil {
		return nil, err
	}
	return spec, nil
}

// SetBindInstance - Set the bind instance for id in the kvp API.
func (d *Dao) SetBindInstance(id string, bindInstance *apb.BindInstance) error {
	return d.setObject(bindInstanceKey(id), bindInstance)
}

// DeleteBindInstance - Delete the binding instance for an id in the kvp API.
func (d *Dao) DeleteBindInstance(id string) error {
	d.log.Debug(fmt.Sprintf("Dao::DeleteBindInstance -> [ %s ]", id))
	_, err := d.kapi.Delete(context.Background(), bindInstanceKey(id), nil)
	return err
}

// GetExtractedCredentials - Get the extracted credentials for an id in the kvp API.
func (d *Dao) GetExtractedCredentials(id string) (*apb.ExtractedCredentials, error) {
	extractedCredentials := &apb.ExtractedCredentials{}
	if err := d.getObject(extractedCredentialsKey(id), extractedCredentials); err != nil {
		return nil, err
	}
	return extractedCredentials, nil
}

// SetExtractedCredentials - Set the extracted credentials for an id in the kvp API.
func (d *Dao) SetExtractedCredentials(id string, extractedCredentials *apb.ExtractedCredentials) error {
	return d.setObject(extractedCredentialsKey(id), extractedCredentials)
}

// DeleteExtractedCredentials - delete the extracted credentials for an id in the kvp API.
func (d *Dao) DeleteExtractedCredentials(id string) error {
	_, err := d.kapi.Delete(context.Background(), extractedCredentialsKey(id), nil)
	return err
}

// SetState - Set the Job State in the kvp API for id.
func (d *Dao) SetState(id string, state apb.JobState) error {
	return d.setObject(stateKey(id, state.Token), state)
}

// GetState - Retrieve a job state from the kvp API for an ID and Token.
func (d *Dao) GetState(id string, token string) (apb.JobState, error) {
	state := apb.JobState{}
	if err := d.getObject(stateKey(id, token), &state); err != nil {
		return apb.JobState{State: apb.StateFailed}, err
	}
	return state, nil
}

func (d *Dao) getObject(key string, data interface{}) error {
	raw, err := d.GetRaw(key)
	if err != nil {
		return err
	}
	apb.LoadJSON(raw, data)
	return nil
}

func (d *Dao) setObject(key string, data interface{}) error {
	payload, err := apb.DumpJSON(data)
	if err != nil {
		return err
	}
	return d.SetRaw(key, payload)
}

////////////////////////////////////////////////////////////
// Key generators
////////////////////////////////////////////////////////////

func stateKey(id string, jobid string) string {
	return fmt.Sprintf("/state/%s/job/%s", id, jobid)
}

func stateKeyId(key string) string {
	s := strings.TrimPrefix(key, "/state/")
	s = strings.TrimSuffix(s, "/job")
	return s
}

func extractedCredentialsKey(id string) string {
	return fmt.Sprintf("/extracted_credentials/%s", id)
}

func specKey(id string) string {
	return fmt.Sprintf("/spec/%s", id)
}

func serviceInstanceKey(id string) string {
	return fmt.Sprintf("/service_instance/%s", id)
}

func bindInstanceKey(id string) string {
	return fmt.Sprintf("/bind_instance/%s", id)
}
