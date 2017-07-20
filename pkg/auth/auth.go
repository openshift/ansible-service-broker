package auth

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
)

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
	userfile string
	userdb   map[string]User
}

// NewFileUserServiceAdapter - constructor for the FUSA
func NewFileUserServiceAdapter(filename string) FileUserServiceAdapter {
	fusa := FileUserServiceAdapter{userfile: filename}
	fusa.buildDb()
	return fusa
}

func (d *FileUserServiceAdapter) buildDb() {
	content, err := ioutil.ReadFile(d.userfile)
	if err != nil {
		fmt.Println(err.Error())
	}
	userinfo := strings.Split(string(content), "\n")
	d.userdb = make(map[string]User)
	d.userdb[userinfo[0]] = User{Username: userinfo[0], Password: userinfo[1]}
}

// FindByLogin - given a login name, this will return the associated User or an
// error
func (d FileUserServiceAdapter) FindByLogin(login string) (User, error) {

	// TODO: add some error checking
	user := d.userdb[login]

	return user, nil
}

// ValidateUser - returns true if the given user's credentials match a user in
// our backend storage. Returns fals otherwise.
func (d FileUserServiceAdapter) ValidateUser(username string, password string) bool {
	user, err := d.FindByLogin(username)
	if err != nil {
		return false
	}

	if user.Username == username && user.Password == password {
		return true
	}

	return false
}

// Handler - does the authentication for the routes
func Handler(h http.Handler, providers []Provider) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// TODO: loop through the providers

		// TODO: determine what to do with the Principal. We don't really have a
		// context or a session to store it on. Do we need it past this?
		_, err := providers[0].GetPrincipal(r)
		if err != nil {
			http.Error(w, err.Error(), http.StatusUnauthorized)
			return
		}
		h.ServeHTTP(w, r)
	})
}
