package ansibleapp

import (
	"encoding/json"
)

type SpecManifest map[string]*Spec

type Spec struct {
	Id          string `json:"id"`
	Name        string `json:"name"`
	Bindable    bool   `json:"bindable"`
	Description string `json:"description"`

	// required, optional, unsupported
	Async string `json:"async"`
}

func (s *Spec) LoadJSON(payload string) *Spec {
	json.Unmarshal([]byte(payload), s)
	return s
}

func (s *Spec) DumpJSON() string {
	payload, err := json.Marshal(s)
	if err != nil {
		panic(err)
	}

	return string(payload)
}

func NewSpecManifest(specs []*Spec) SpecManifest {
	manifest := make(map[string]*Spec)
	for _, spec := range specs {
		manifest[spec.Id] = spec
	}
	return manifest
}
