package auth

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
)

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

type FileUserServiceAdapter struct {
	userfile string
	userdb   map[string]User
}

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

func (d FileUserServiceAdapter) FindByLogin(login string) (User, error) {

	fmt.Println(login)
	user := d.userdb[login]

	fmt.Println("username: " + user.Username)
	return user, nil
}

func (d FileUserServiceAdapter) ValidateUser(username string, password string) bool {
	return false
}
