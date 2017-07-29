package auth

import (
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"path"
	"strings"

	logging "github.com/op/go-logging"
)

// ConfigEntry - Configuration for authentication
type ConfigEntry struct {
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

// Handler - does the authentication for the routes
func Handler(h http.Handler, providers []Provider, log *logging.Logger) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		// TODO: determine what to do with the Principal. We don't really have a
		// context or a session to store it on. Do we need it past this?
		var principalFound error
		for _, provider := range providers {
			principal, err := provider.GetPrincipal(r)
			if principal != nil {
				log.Debug("We found one. HOORAY!")
				// we found our principal, stop looking
				break
			}
			if err != nil {
				principalFound = err
			}
		}
		// if we went through the providers and found no principals. We will
		// have found an error
		if principalFound != nil {
			log.Debug("no principal found")
			http.Error(w, principalFound.Error(), http.StatusUnauthorized)
			return
		}

		h.ServeHTTP(w, r)
	})
}

// GetProviders - returns the list of configured providers
func GetProviders(entries []ConfigEntry, log *logging.Logger) []Provider {
	providers := make([]Provider, 0, len(entries))

	for _, cfg := range entries {
		if cfg.Enabled {
			provider, err := createProvider(cfg.Type, log)
			if err != nil {
				log.Warning("Unable to create provider for %v. %v", cfg.Type, err)
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
