package auth

import "net/http"

// AuthProvider - interface for all Auth providers
type AuthProvider interface {
	GetPrincipal(*http.Request) (Principal, error)
}

// Principal - principal user or other identity of some kind with access to the
// broker
type Principal interface {
	GetType() string
	GetName() string
}

type UserServiceAdapter interface {
	FindByLogin(string) (User, error)
	ValidateUser(string, string) bool
}

type User struct {
	Username string
	Password string
}

func (u User) GetType() string {
	return "user"
}

func (u User) GetName() string {
	return u.Username
}

type DefaultUserServiceAdapter struct {
}

func (d DefaultUserServiceAdapter) FindByLogin(string) (User, error) {
	return User{Username: "asb", Password: "password"}, nil
}
func (d DefaultUserServiceAdapter) ValidateUser(username string, password string) bool {
	return false
}
