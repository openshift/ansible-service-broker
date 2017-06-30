package fusortest

import (
	"bytes"
	"encoding/json"
	"fmt"
	"regexp"
	"testing"
)

// Condition - the function to use for testing.
type Condition func(a interface{}, b interface{}) bool

func assert(t *testing.T, a interface{}, b interface{}, message []string, test Condition) {
	var msg string
	if len(message) != 0 {
		msg = message[0]
	}

	if !test(a, b) {
		fail(t, a, b, msg)
	}
}

func fail(t *testing.T, a interface{}, b interface{}, message string) {
	if len(message) == 0 {
		message = fmt.Sprintf("%v != %v", a, b)
	}
	t.Fatal(message)
}

// AssertEqual - Assert that inputs are equivalent.
func AssertEqual(t *testing.T, a interface{}, b interface{}, message ...string) {
	assert(t, a, b, message, func(a interface{}, b interface{}) bool {
		return a == b
	})
}

// AssertNotEqual - Assert that inputs are not equivalent
func AssertNotEqual(t *testing.T, a interface{}, b interface{}, message ...string) {
	assert(t, a, b, message, func(a interface{}, b interface{}) bool {
		return a != b
	})
}

// AssertTrue - Assert that input is true
func AssertTrue(t *testing.T, a interface{}, message ...string) {
	if a == true {
		return
	}

	var msg string
	if len(message) != 0 {
		msg = message[0]
	}

	if len(message) == 0 {
		msg = fmt.Sprintf("%v is not true!", a)
	}
	t.Fatal(msg)
}

// AssertFalse - Assert that input is false
func AssertFalse(t *testing.T, a interface{}, message ...string) {
	if a == false {
		return
	}

	var msg string
	if len(message) != 0 {
		msg = message[0]
	}

	if len(message) == 0 {
		msg = fmt.Sprintf("%v is not true!", a)
	}
	t.Fatal(msg)
}

// AssertNotNil - Assert that input is not nil
func AssertNotNil(t *testing.T, a interface{}, message ...string) {
	if a != nil {
		return
	}

	var msg string
	if len(message) != 0 {
		msg = message[0]
	}

	if len(message) == 0 {
		msg = fmt.Sprintf("%v is nil! %s", a)
	}
	t.Fatal(msg)
}

// AssertNil - Assert input is nil
func AssertNil(t *testing.T, a interface{}, message ...string) {
	if a == nil {
		return
	}

	var msg string
	if len(message) != 0 {
		msg = message[0]
	}

	if len(message) == 0 {
		msg = fmt.Sprintf("%v is not nil! %s", a)
	}
	t.Fatal(msg)
}

// AssertError - Assert that input is an error response
func AssertError(t *testing.T, body *bytes.Buffer, msg string) {
	var errResp = make(map[string]string)

	if body == nil {
		t.Fatal("invalid response body")
	}

	json.Unmarshal(body.Bytes(), &errResp)
	if errResp["description"] != msg {
		t.Log(errResp["description"])
		t.Fatal("error message does not match")
	}
}

// AssertState - Assert that state contianed in the body is of a certain state.
func AssertState(t *testing.T, body *bytes.Buffer, state string) {
	var resp = make(map[string]string)

	if body == nil {
		t.Fatal("invalid response body")
	}

	json.Unmarshal(body.Bytes(), &resp)
	if resp["state"] != state {
		t.Log(resp["state"])
		t.Fatal("state does not match")
	}
}

// AssertOperation -  Assert that the operation contained in the body is of a certain operation.
func AssertOperation(t *testing.T, body *bytes.Buffer, op string) {
	var resp = make(map[string]string)

	if body == nil {
		t.Fatal("invalid response body")
	}

	json.Unmarshal(body.Bytes(), &resp)
	if resp["operation"] != op {
		t.Log(resp["operation"])
		t.Fatal("state does not match")
	}
}

// StripNewline - String all new lines from string.
func StripNewline(input string) string {
	re := regexp.MustCompile("\\n")
	return re.ReplaceAllString(input, "")
}

// MinifyJSON - Minify the json outputted.
func MinifyJSON(input string) string {
	var mm interface{}
	json.Unmarshal([]byte(input), &mm)
	fmt.Println("before")
	fmt.Printf("%v", mm)
	fmt.Println("after")
	dat, _ := json.Marshal(mm)
	return string(dat)
}
