package app

import (
	"errors"
	"fmt"

	flags "github.com/jessevdk/go-flags"
)

type Args struct {
	ConfigFile string `short:"c" long:"config" description:"Config File" default:"/etc/ansible-service-broker/config.yaml"`
	Version    bool   `short:"v" long:"version" description:"Print version information"`
}

func CreateArgs() (Args, error) {
	args := Args{}

	_, err := flags.Parse(&args)
	if err != nil {
		return args, err
	}

	err = validateArgs(&args)
	if err != nil {
		return args, err
	}

	return args, nil
}

func validateArgs(args *Args) error {
	var err error
	if args.ConfigFile == "" {
		err = errors.New("must provide a config file location with -c, or --config")
	}

	return err
}

func ArgsUsage() {
	// TODO
	fmt.Println("USAGE: To be implemented...")
}
