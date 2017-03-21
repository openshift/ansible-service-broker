package apb

import (
	"encoding/json"
	"fmt"
	ft "github.com/fusor/ansible-service-broker/pkg/fusortest"
	"reflect"
	"testing"
)

const SpecId = "ab094014-b740-495e-b178-946d5aa97ebf"
const SpecName = "fusor/etherpad-apb"
const SpecBindable = false
const SpecAsync = "optional"
const SpecDescription = "A note taking webapp"
const SpecParameters = `
	[
		{"name": "hostport", "description": "The host TCP port as the external endpoint", "default": 9001, "type": "foo", "required": true},
		{"name": "db_user", "description": "Database User", "default": "db_user", "type": "", "required": true},
		{"name": "db_pass", "description": "Database Password", "default": "db_pass", "type": "", "required": true},
		{"name": "db_name", "description": "Database Name", "default": "db_name", "type": "", "required": true},
		{"name": "db_host", "description": "Database service hostname/ip", "default": "mariadb", "type": "", "required": true},
		{"name": "db_port", "description": "Database service port", "default": 3306, "type": "", "required": true}
	]
`

var expectedSpecParameters = []*ParameterDescriptor{
	&ParameterDescriptor{Name: "hostport", Description: "The host TCP port as the external endpoint", Default: float64(9001), Type: "foo", Required: true},
	&ParameterDescriptor{Name: "db_user", Description: "Database User", Default: "db_user", Type: "", Required: true},
	&ParameterDescriptor{Name: "db_pass", Description: "Database Password", Default: "db_pass", Type: "", Required: true},
	&ParameterDescriptor{Name: "db_name", Description: "Database Name", Default: "db_name", Type: "", Required: true},
	&ParameterDescriptor{Name: "db_host", Description: "Database service hostname/ip", Default: "mariadb", Type: "", Required: true},
	&ParameterDescriptor{Name: "db_port", Description: "Database service port", Default: float64(3306), Type: "", Required: true},
}

var SpecJSON = fmt.Sprintf(`
{
	"id": "%s",
	"description": "%s",
	"name": "%s",
	"bindable": %t,
	"async": "%s",
	"parameters": %s
}
`, SpecId, SpecDescription, SpecName, SpecBindable, SpecAsync, SpecParameters)

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
	ft.AssertTrue(t, reflect.DeepEqual(s.Parameters, expectedSpecParameters))

}

func TestSpecDumpJSON(t *testing.T) {
	s := Spec{
		Id:          SpecId,
		Description: SpecDescription,
		Name:        SpecName,
		Bindable:    SpecBindable,
		Async:       SpecAsync,
		Parameters:  expectedSpecParameters,
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
