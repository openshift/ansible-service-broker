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

A new `UserServiceAdapter` will need to be implemented to know how
to read the token.

Some good questions to answer in detail:

 - How will the broker's behavior change?

   The broker will now have 2 ways for authentication: basic auth and bearer
   token.

 - Will this change APBs?
   The bearer token will have no affect on the APBs.

## Work Items
 - Add new `bearer_auth.go` file containing the `BearerAuth` struct and
   associated methods.
 - Add "bearer" to the `createProvider` method
 - 

## References
[1]: https://tools.ietf.org/html/rfc6750#section-2.1 "RFC 6750 Section 2.1"
[2]: https://github.com/openshift/ansible-service-broker/blob/bearerauth/pkg/auth/auth.go#L20-L25 "Provider interface"


BearerAuth has a TokenServiceAdapter

ValidateToken(token)


Change UserServiceAdapter interface to
type UserServiceAdapter interface {
   FindByLogin(string) (User, error)
   Validate(string, string) bool
}

type TokenServiceAdapter interface {
   Validate(string) bool
}

Could make the validator use a variadic function:

type ServiceAdapterValidator interface {
    Validate(values ...string) bool
}
