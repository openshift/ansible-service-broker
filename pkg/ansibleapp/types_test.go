package ansibleapp

import (
	"encoding/json"
	"fmt"
	ft "github.com/fusor/ansible-service-broker/pkg/fusortest"
	"reflect"
	"testing"
)

const SpecId = "ab094014-b740-495e-b178-946d5aa97ebf"
const SpecName = "fusor/etherpad-ansibleapp"
const SpecBindable = false
const SpecAsync = "optional"
const SpecDescription = "A note taking webapp"

var SpecJSON = fmt.Sprintf(`
{
  "id": "%s",
	"description": "%s",
  "name": "%s",
  "bindable": %t,
  "async": "%s"
}
`, SpecId, SpecDescription, SpecName, SpecBindable, SpecAsync)

func TestSpecLoadJSON(t *testing.T) {
	s := Spec{}
	err := LoadJSON(SpecJSON, &s)
	if err != nil {
		panic(err)
	}

	ft.AssertEqual(t, s.Id, SpecId)
	ft.AssertEqual(t, s.Description, SpecDescription)
	ft.AssertEqual(t, s.Name, SpecName)
	ft.AssertEqual(t, s.Bindable, SpecBindable)
	ft.AssertEqual(t, s.Async, SpecAsync)
}

func TestSpecDumpJSON(t *testing.T) {
	s := Spec{
		Id:          SpecId,
		Description: SpecDescription,
		Name:        SpecName,
		Bindable:    SpecBindable,
		Async:       SpecAsync,
	}

	var knownMap interface{}
	var subjectMap interface{}

	raw, err := DumpJSON(&s)
	if err != nil {
		panic(err)
	}
	json.Unmarshal([]byte(SpecJSON), &knownMap)
	json.Unmarshal([]byte(raw), &subjectMap)

	ft.AssertTrue(t, reflect.DeepEqual(knownMap, subjectMap))
}
