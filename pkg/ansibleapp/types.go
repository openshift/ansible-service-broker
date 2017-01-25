package ansibleapp

import (
	"encoding/json"
)

type Spec struct {
	Id       string `json:"id"`
	Name     string `json:"name"`
	Bindable bool   `json:"bindable"`

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
