package ansibleapp

import (
	"encoding/json"
	"fmt"
	ft "github.com/fusor/ansible-service-broker/pkg/fusortest"
	"reflect"
	"testing"
)

const SPEC_ID = "ab094014-b740-495e-b178-946d5aa97ebf"
const SPEC_NAME = "fusor/etherpad-ansibleapp"
const SPEC_BINDABLE = false
const SPEC_ASYNC = "optional"

var SPEC_JSON = fmt.Sprintf(`
{
  "id": "%s",
  "name": "%s",
  "bindable": %t,
  "async": "%s"
}
`, SPEC_ID, SPEC_NAME, SPEC_BINDABLE, SPEC_ASYNC)

func TestSpecLoadJSON(t *testing.T) {
	s := Spec{}
	s.LoadJSON(SPEC_JSON)

	ft.AssertEqual(t, s.Id, SPEC_ID)
	ft.AssertEqual(t, s.Name, SPEC_NAME)
	ft.AssertEqual(t, s.Bindable, SPEC_BINDABLE)
	ft.AssertEqual(t, s.Async, SPEC_ASYNC)

}

func TestSpecDumpJSON(t *testing.T) {
	s := Spec{
		Id:       SPEC_ID,
		Name:     SPEC_NAME,
		Bindable: SPEC_BINDABLE,
		Async:    SPEC_ASYNC,
	}

	var knownMap interface{}
	var subjectMap interface{}

	json.Unmarshal([]byte(SPEC_JSON), &knownMap)
	json.Unmarshal([]byte(s.DumpJSON()), &subjectMap)

	ft.AssertTrue(t, reflect.DeepEqual(knownMap, subjectMap))
}
