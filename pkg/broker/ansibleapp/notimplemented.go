package ansibleapp

import (
	"net/http"

	"k8s.io/kubernetes/pkg/api/errors"
	"k8s.io/kubernetes/pkg/api/unversioned"
)

var notImplemented = &errors.StatusError{
	ErrStatus: unversioned.Status{
		Code:    http.StatusNotImplemented,
		Message: http.StatusText(http.StatusNotImplemented),
	},
}
