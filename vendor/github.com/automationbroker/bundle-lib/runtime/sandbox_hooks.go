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

package runtime

// PreSandboxCreate - The pre sand box creation function will be called
// before the sandbox is created for the bundle. This function should not expect
// to panic and should fail gracefully by bubbling up the error and cleaning up
// after itself.
// Parameters in order of adding to the function.
// string - pod name is also the svc accounts name.
// string - namespace of the transient namespace.
// list of strings - target namespaces.
// string - abp role.
// return error.
type PreSandboxCreate func(string, string, []string, string) error

// AddPostCreateSandbox - Adds a post create sandbox function to the runtime.
// Once the sandbox is created all of the functions that have been added here
// will be executed.
func (p *provider) addPreCreateSandbox(f PreSandboxCreate) {
	p.preSandboxCreate = append(p.preSandboxCreate, f)
}

// PostSandboxCreate - The post sand box creation function will be called
// after the sandbox is created for the APB. This function should not expect
// to panic and should fail gracefully by bubbling up the error and cleaning up
// after itself.
// Parameters in order of adding to the function.
// string - pod name is also the svc accounts name.
// string - namespace of the transient namespace.
// list of strings - target namespaces.
// string - abp role.
// return error.
type PostSandboxCreate func(string, string, []string, string) error

// AddPostCreateSandbox - Adds a post create sandbox function to the runtime.
// Once the sandbox is created all of the functions that have been added here
// will be executed.
func (p *provider) addPostCreateSandbox(f PostSandboxCreate) {
	p.postSandboxCreate = append(p.postSandboxCreate, f)
}

// PreSandboxDestroy - The pre sand box destroy function will be called
// before the sandbox is destroyed. This could mean the namespace is kept around
// if the apb failed and configuration conditions are met. This function should not
// expect to panic and should fail gracefully by bubbling up the error. This
// function should also not delete the namespace or the pod directly. This
// will most likely be used to clean up resources in pre/post create sandbox hooks.
// Parameters:
// string - pod / svc-account name
// string - pod transient namespace
// []string - target namespaces
type PreSandboxDestroy func(string, string, []string) error

func (p *provider) addPreDestroySandbox(f PreSandboxDestroy) {
	p.preSandboxDestroy = append(p.preSandboxDestroy, f)
}

// PostSandboxDestroy - The post sand box destroy function will be called
// after the sandbox is destroyed. This could mean the namespace is kept around
// if the apb failed and configuration conditions are met. This function should not
// expect to panic and should fail gracefully by bubbling up the error. This
// function should also not delete the namespace or the pod directly. This
// will most likely be used to clean up resources in pre/post create sandbox hooks.
// Parameters:
// string - pod / svc-account name
// string - pod transient namespace
// []string - target namespaces
type PostSandboxDestroy func(string, string, []string) error

func (p *provider) addPostDestroySandbox(f PostSandboxDestroy) {
	p.postSandboxDestroy = append(p.postSandboxDestroy, f)
}
