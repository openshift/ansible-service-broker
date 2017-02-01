package broker

import (
	"net/http"

	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/apis/meta/v1"
)

var notImplemented = &errors.StatusError{
	ErrStatus: v1.Status{
		Code:    http.StatusNotImplemented,
		Message: http.StatusText(http.StatusNotImplemented),
	},
}
