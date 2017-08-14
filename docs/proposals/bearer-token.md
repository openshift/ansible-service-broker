# Bearer token authentication proposal

## Introduction

The service catalog currently supports basic auth as well as bearer token
authentication. The broker will need to be enhanced to support bearer token
authentication.

## Problem Description
The broker will need to support bearer token authentication.

## <Implementation Details>
Bearer token authentication uses the `Authorization` header followed by `Bearer`
+ 1 space + a base64 encoded token [] [1]

The body of the proposal will be filled with details about the feature you're
trying to land.

Some good questions to answer in detail:
 - How will this improve the broker?
 - How will the broker's behavior change?
 - Will this change APBs?

## Work Items
A list of items that you plan to implement. You don't have to follow it
exactly during implementation, but it's good to compact all the details
about the proposal into a series of steps that anyone can follow.

example:
 - Add a new pkg SpeedUpBindings
 - Build SpeedUpBindings functions so bindings are faster
 - Integrate SpeedUpBindings into the Binding workflow
 - Fix tests so they use SpeedUpBindings

## References
[1] [RFC 6750 Section 2.1](https://tools.ietf.org/html/rfc6750#section-2.1)
