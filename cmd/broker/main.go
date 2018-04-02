//
// Copyright (c) 2018 Red Hat, Inc.
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

package main

import (
	"fmt"
	"os"

	"github.com/automationbroker/bundle-lib/registries"
	flags "github.com/jessevdk/go-flags"
	"github.com/openshift/ansible-service-broker/pkg/app"
	"github.com/openshift/ansible-service-broker/pkg/version"
)

func main() {

	var args app.Args
	var err error

	// Writing directly to stderr because log has not been bootstrapped
	if args, err = app.CreateArgs(); err != nil {
		if flagsErr, ok := err.(*flags.Error); ok && flagsErr.Type == flags.ErrHelp {
			os.Exit(0)
		} else {
			os.Exit(1)
		}
	}

	if args.Version {
		fmt.Println(version.Version)
		os.Exit(0)
	}

	// To add your custom registries, define an entry in this array.
	regs := []registries.Registry{}

	// CreateApp passing in the args and registries
	app := app.CreateApp(args, regs)
	app.Start()
}
