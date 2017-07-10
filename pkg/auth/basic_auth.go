package auth

import "net/http"

type BasicAuth struct {
	usa UserServiceAdapter `json:"usa"`
}

func NewBasicAuth(userSvcAdapter UserServiceAdapter) BasicAuth {
	return BasicAuth{usa: userSvcAdapter}
}

func (b BasicAuth) GetPrincipal(r *http.Request) Principal {
	// get Authorization header
	authHdr := r.Header.Get("Authorization")
	//	if (authHdr.Get
	return nil
}
