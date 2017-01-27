package ansibleapp

import (
	"encoding/json"
	"github.com/pborman/uuid"
)

type Parameters map[string]interface{}
type SpecManifest map[string]*Spec

type Spec struct {
	Id          string `json:"id"`
	Name        string `json:"name"`
	Bindable    bool   `json:"bindable"`
	Description string `json:"description"`

	// required, optional, unsupported
	Async string `json:"async"`
}

func LoadJSON(payload string, obj interface{}) error {
	err := json.Unmarshal([]byte(payload), obj)
	if err != nil {
		return err
	}

	return nil
}

func DumpJSON(obj interface{}) (string, error) {
	payload, err := json.Marshal(obj)
	if err != nil {
		return "", err
	}

	return string(payload), nil
}

func NewSpecManifest(specs []*Spec) SpecManifest {
	manifest := make(map[string]*Spec)
	for _, spec := range specs {
		manifest[spec.Id] = spec
	}
	return manifest
}

type ServiceInstance struct {
	Id         uuid.UUID   `json:"id"`
	Spec       *Spec       `json:"spec"`
	Parameters *Parameters `json:"parameters"`
}
