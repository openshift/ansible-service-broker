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

package auth

import (
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"path"
	"strings"

	logging "github.com/op/go-logging"
	"github.com/openshift/ansible-service-broker/pkg/config"
)

// Config - Configuration for authentication
type Config struct {
	Type    string `yaml:"type"`
	Enabled bool   `yaml:"enabled"`
}

// Provider - an auth provider is an adapter that provides the principal
// object required for authentication. This can be a User, a System, or some
// other entity.
type Provider interface {
	GetPrincipal(*http.Request) (Principal, error)
}

// Principal - principal user or other identity of some kind with access to the
// broker.
type Principal interface {
	GetType() string
	GetName() string
	// TODO: add roles?
}

// UserServiceAdapter - is the interface for a service that stores Users. It can
// be anything you want: file, database, whatever as long as you can search and
// validate them.
type UserServiceAdapter interface {
	FindByLogin(string) (User, error)
	ValidateUser(string, string) bool
}

// User - a User from the service adapter
type User struct {
	Username string
	Password string
}

// GetType - returns the type of Principal. This is a user principal.
func (u User) GetType() string {
	return "user"
}

// GetName - returns the name of the User Principal.
func (u User) GetName() string {
	return u.Username
}

// FileUserServiceAdapter - a file based UserServiceAdapter which seeds its
// users from a file.
type FileUserServiceAdapter struct {
	filedir string
	log     *logging.Logger
	userdb  map[string]User
}

func (d *FileUserServiceAdapter) buildDB() error {
	userfile := path.Join(d.filedir, "username")
	passfile := path.Join(d.filedir, "password")
	username, uerr := ioutil.ReadFile(userfile)
	if uerr != nil {
		d.log.Error("Error reading username. %v", uerr.Error())
		return uerr
	}
	password, perr := ioutil.ReadFile(passfile)
	if perr != nil {
		d.log.Error("Error reading password. %v", perr.Error())
		return perr
	}

	// userdb is probably overkill, but if we ever want to allow multiple users,
	// it'll come in handy.
	d.userdb = make(map[string]User)
	// since it's also the key
	unamestr := string(username)
	d.userdb[unamestr] = User{Username: unamestr, Password: string(password)}

	return nil
}

// FindByLogin - given a login name, this will return the associated User or
// an error
func (d FileUserServiceAdapter) FindByLogin(login string) (User, error) {

	// TODO: add some error checking
	if user, ok := d.userdb[login]; ok {
		return user, nil
	}

	return User{}, errors.New("user not found")
}

// ValidateUser - returns true if the given user's credentials match a user
// in our backend storage. Returns fals otherwise.
func (d FileUserServiceAdapter) ValidateUser(username string, password string) bool {
	user, err := d.FindByLogin(username)
	if err != nil {
		d.log.Debug("user not found, returning false")
		return false
	}

	if user.Username == username && user.Password == password {
		d.log.Debug("user found, returning true")
		return true
	}

	return false
}

// NewFileUserServiceAdapter - constructor for the FUSA
func NewFileUserServiceAdapter(dir string, log *logging.Logger) (*FileUserServiceAdapter, error) {
	if dir == "" {
		return nil, fmt.Errorf("directory is empty, defaulting to %s", dir)
	}

	fusa := FileUserServiceAdapter{filedir: dir, log: log}
	err := fusa.buildDB()
	if err != nil {
		log.Error("we had a problem building the DB for FileUserServiceAdapter. ", err)
		return nil, err
	}
	return &fusa, nil
}

// GetProviders - returns the list of configured providers
func GetProviders(authConfig *config.Config, log *logging.Logger) []Provider {
	providers := make([]Provider, 0, len(authConfig.GetSubConfig("broker.auth").ToMap()))

	for t := range authConfig.GetSubConfig("broker.auth").ToMap() {
		fmt.Printf("%v - here", t)
		if authConfig.GetBool(fmt.Sprintf("broker.auth.%v.enabled", t)) {
			fmt.Printf("here")
			provider, err := createProvider(t, log)
			if err != nil {
				log.Warning("Unable to create provider for %v. %v", t, err)
				continue
			}
			providers = append(providers, provider)
		}
	}

	return providers
}

func createProvider(providerType string, log *logging.Logger) (Provider, error) {
	switch strings.ToLower(providerType) {
	case "basic":
		log.Info("Configured for basic auth")
		usa, err := GetUserServiceAdapter(log)
		if err != nil {
			return nil, err
		}
		return NewBasicAuth(usa, log), nil
	// add case "oauth":
	default:
		panic("Unknown auth provider")
	}
}

// GetUserServiceAdapter returns the configured UserServiceAdapter
func GetUserServiceAdapter(log *logging.Logger) (UserServiceAdapter, error) {
	// TODO: really need to figure out a better way to define what
	// should be returned.
	return NewFileUserServiceAdapter("/var/run/asb-auth", log)
}
