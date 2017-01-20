package errors

import "strings"

type Errors []error

func (es Errors) Error() string {
	s := make([]string, len(es))
	for i, e := range es {
		s[i] = e.Error()
	}
	return strings.Join(s, "\n")
}
