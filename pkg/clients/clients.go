package clients

import (
	"sync"
)

type clientResult struct {
	client interface{}
	err    error
}

var instances struct {
	Etcd       clientResult
	Kubernetes clientResult
}

var once struct {
	Etcd       sync.Once
	Kubernetes sync.Once
}
