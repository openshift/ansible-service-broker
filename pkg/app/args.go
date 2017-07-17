package app

import (
	"fmt"

	flags "github.com/jessevdk/go-flags"
)

// Args - Command line arguments for the ansbile service broker.
type Args struct {
	ConfigFile string `short:"c" long:"config" description:"Config File" default:"/etc/ansible-service-broker/config.yaml"`
	Version    bool   `short:"v" long:"version" description:"Print version information"`
	Insecure   bool   `long:"insecure" description:"Run Ansible Service Broker in insecure mode."`
}

// CreateArgs - Will return the arguments that were passed in to the application
func CreateArgs() (Args, error) {
	args := Args{}

	_, err := flags.Parse(&args)
	if err != nil {
		fmt.Printf("err - %v", err)
		return args, err
	}
	return args, nil
}
