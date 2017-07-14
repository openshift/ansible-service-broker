package auth

import (
	"encoding/base64"
	"errors"
	"fmt"
	"net/http"
	"strings"
)

type UserPrincipal struct {
	username string
	// might need a set of permissions etc
}

func (u UserPrincipal) GetType() string {
	return "user"
}

func (u UserPrincipal) GetName() string {
	return u.username
}

type BasicAuth struct {
	usa UserServiceAdapter `json:"usa"`
}

func NewBasicAuth(userSvcAdapter UserServiceAdapter) BasicAuth {
	return BasicAuth{usa: userSvcAdapter}
}

func (b BasicAuth) GetPrincipal(r *http.Request) (Principal, error) {
	var username string
	var password string

	// get Authorization header
	authheader := r.Header.Get("Authorization")
	if strings.HasPrefix(strings.ToUpper(authheader), "BASIC ") {
		// get the encoded part of the header
		decodedheader, err := base64.StdEncoding.DecodeString(authheader[6:])
		if err != nil {
			fmt.Println(err.Error())
		}
		userpass := strings.Split(string(decodedheader), ":")
		username = userpass[0]

		if len(userpass) > 1 {
			password = userpass[1]
		}

		if !b.usa.ValidateUser(username, password) {
			return nil, errors.New("invalid credentials")
		}
	}

	return b.createPrincipal(username)
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
