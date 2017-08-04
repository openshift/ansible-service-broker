package auth

import (
	"errors"
	"net/http"

	logging "github.com/op/go-logging"
)

// UserPrincipal - represents a User as a Principal to the auth system.
type UserPrincipal struct {
	username string
	// might need a set of permissions etc
}

// GetType - returns "user" indicating it is a UserPrincipal
func (u UserPrincipal) GetType() string {
	return "user"
}

// GetName - returns user's name
func (u UserPrincipal) GetName() string {
	return u.username
}

// BasicAuth - Performs an HTTP Basic Auth validation.
type BasicAuth struct {
	usa UserServiceAdapter
	log *logging.Logger
}

// NewBasicAuth - constructs a BasicAuth instance.
func NewBasicAuth(userSvcAdapter UserServiceAdapter, log *logging.Logger) BasicAuth {
	return BasicAuth{usa: userSvcAdapter, log: log}
}

// GetPrincipal - returns the User Principal that matches the credentials in the
// Authorization header.
func (b BasicAuth) GetPrincipal(r *http.Request) (Principal, error) {
	if username, password, ok := r.BasicAuth(); ok {
		if !b.usa.ValidateUser(username, password) {
			return nil, errors.New("invalid credentials")
		}
		return b.createPrincipal(username)
	}

	return nil, errors.New("invalid credentials, corrupt header")
}

func (b BasicAuth) createPrincipal(username string) (Principal, error) {
	// don't care about the user right now, just trying to see if it
	// exists. In the future we might want to check its permissions etc.
	_, err := b.usa.FindByLogin(username)
	if err != nil {
		return nil, err
	}
	return UserPrincipal{username: username}, nil
}
