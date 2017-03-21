package fusortest

import (
	"bytes"
	"encoding/json"
	"fmt"
	"regexp"
	"testing"
)

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

func AssertEqual(t *testing.T, a interface{}, b interface{}, message ...string) {
	assert(t, a, b, message, func(a interface{}, b interface{}) bool {
		return a == b
	})
}

func AssertNotEqual(t *testing.T, a interface{}, b interface{}, message ...string) {
	assert(t, a, b, message, func(a interface{}, b interface{}) bool {
		return a != b
	})
}

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

func AssertNotNil(t *testing.T, a interface{}, message ...string) {
	if a != nil {
		return
	}

	var msg string
	if len(message) != 0 {
		msg = message[0]
	}

	if len(message) == 0 {
		msg = fmt.Sprintf("%v is nil!", a)
	}
	t.Fatal(msg)
}

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

func StripNewline(input string) string {
	re := regexp.MustCompile("\\n")
	return re.ReplaceAllString(input, "")
}

func MinifyJSON(input string) string {
	var mm interface{}
	json.Unmarshal([]byte(input), &mm)
	fmt.Println("before")
	fmt.Printf("%v", mm)
	fmt.Println("after")
	dat, _ := json.Marshal(mm)
	return string(dat)
}
