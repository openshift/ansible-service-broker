package main

import (
	"fmt"
	"net"
	"net/http"

	"github.com/fusor/ansible-service-broker/pkg/broker/ansibleapp"
	"github.com/fusor/ansible-service-broker/pkg/handler"
	"github.com/spf13/pflag"
	"k8s.io/kubernetes/pkg/util/logs"
)

var addr = pflag.String("addr", ":8000", "listen address")
var createProjects = pflag.Bool("create-projects", false, "set true to create a project per service instance")

func main() {
	logs.InitLogs()

	broker, err := ansibleapp.NewBroker("My Test Broker")
	if err != nil {
		panic(err)
	}

	l, err := net.Listen("tcp", *addr)
	if err != nil {
		panic(err)
	}

	fmt.Println("running")

	err = http.Serve(l, handler.NewHandler(broker))
	if err != nil {
		panic(err)
	}
}
