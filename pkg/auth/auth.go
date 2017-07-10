package auth

import "net/http"

// AuthProvider - interface for all Auth providers
type AuthProvider interface {
	getPrincipal(*http.Request) Principal
}

// Principal - principal user or other identity of some kind with access to the
// broker
type Principal interface {
	GetType() string
	GetName() string
}

type UserServiceAdapter interface {
	FindByLogin(string) User
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

func (d DefaultUserServiceAdapter) FindByLogin(string) User {
	return User{Username: "asb", Password: "password"}
}
func (d DefaultUserServiceAdapter) ValidateUser(username string, password string) bool {
	return false
}
