# Bearer token authentication proposal

## Introduction

The service catalog currently supports basic auth as well as bearer token
authentication. The broker will need to be enhanced to support bearer token
authentication.

## Problem Description
The broker will need to support bearer token authentication.

## <Implementation Details>
Bearer token authentication uses the `Authorization` header followed by `Bearer`
+ 1 space + a base64 encoded token, see [RFC 6750 Section 2.1] [1]

The broker will have a new `BearerAuth` struct that implements the [Provider] [2]
interface.

A new `TokenService` interface that will validate tokens either via Bearer
token, shared key, or against a external service, etc. Basically used to deal with
non username/password authentication.

This new adapter will handle Bearer token and possibly SSL certificate
authentication. The original UserService would handle BasicAuth, OAuth and any
other model where a `UserPrincipal` will be returned.


```golang
type TokenService interface {
   Validate(token string) bool
}
```

There will be specific implementations of the TokenService which will validate
the token. One simple one will simply read the token from the filesystem.
Another one could connect to a service to validate the token.


```golang
// FileTokenService - reads a token from a static file
type FileTokenService struct {
   tokenFile string
}

func (f FileTokenService) Validate(token string) bool {
    // read the file from tokenFile
    // compare the token from the file to the one given
    savedToken, err := ioutil.ReadFile(tokenFile)
    if err != nil {
        return false
    }

    return token == savedToken
}
```

If we have an external service to handle validation something like this would be
used:

```golang
type SomeExternalService struct {
    service SomeAuthService
}

func NewSomeExternalService(service SomeAuthService) SomeExternalService {
    return SomeExternalService{service: service}
}

func (s SomeExternalService) Validate(token string) bool {
    return s.service.Validate(token)

    // this could also be a simple get the value from the
    // service, then compare it. Or much more complex if need be.
    // Basically the idea is the AuthProviders calling Validate
    // don't really care.
}
```

Basic questions about the impact to Broker and APBs:

 - How will the broker's behavior change?

   The broker will now have 2 ways for authentication: basic auth and bearer
   token. This can continue to expand as we see the need.

 - Will this change APBs?

   The bearer token will have no affect on the APBs.

#### Issues

 - What's the best way to configure different auths?

   For example, I want to have a `BasicAuth` that uses the `FileUserService` and a
   `BasicAuth` that uses `DBUserService`, a fictitious service that loads users from a
   database. Today's configuration does not support specifying a service backend
   to a particular `AuthProvider`. Thoughts?


## Work Items
 - Add new `bearer_auth.go` file containing the `BearerAuth` struct and
   associated methods.
 - Add "bearer" to the `createProvider` method
   - update broker configuration
   - update deployment template
 - Add a `TokenService` interface definition to `auth.go`
 - Implement a `FileTokenService` in `file_token_service.go`
 - Investigate hooking up with OpenShift auth server
   - new implementation of the `TokenService`
   - new configuration item in broker config
   - probably new configuration to setup (maybe) the auth server on OpenShift

Other items to consider for consistency sake, but not directly required for this
proposal:

 - Rename `UserServiceAdapter` to `UserService`
 - Rename `FileUserServiceAdapter` to `FileUserService`
 - Move `FileUserServiceAdapter` to `file_user_service.go`

## References
[1]: https://tools.ietf.org/html/rfc6750#section-2.1 "RFC 6750 Section 2.1"
[2]: https://github.com/openshift/ansible-service-broker/blob/bearerauth/pkg/auth/auth.go#L20-L25 "Provider interface"
[3]: https://docs.openshift.com/container-platform/3.6/architecture/additional_concepts/authentication.html#oauth-server-metadata "OAuth Server Metadata"
