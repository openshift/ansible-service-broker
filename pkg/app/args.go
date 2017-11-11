//
// Copyright (c) 2017 Red Hat, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//    http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//

package app

import (
	"fmt"

	flags "github.com/jessevdk/go-flags"
)

// Args - Command line arguments for the ansbile service broker.
type Args struct {
	ConfigFile string `short:"c" long:"config" description:"Config File" default:"/etc/ansible-service-broker/config.yaml"`
	Version    bool   `short:"v" long:"version" description:"Print version information"`
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
