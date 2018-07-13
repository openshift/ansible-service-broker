package authorization

// Authorizer - authorizes users via the bundle runtime.
type Authorizer interface {
	Authorize(AuthorizeUser, string) (Decision, error)
}

// AuthorizeUser - an interface for a user object.
type AuthorizeUser interface {
	Username() string
}

// Decision - The outcome of the authorization check
type Decision string

const (
	// DecisionAllowed - The authorizer has determined the action is allowed.
	DecisionAllowed Decision = "allowed"
	// DecisionDeny - The authorizer has determined the action is not allowed.
	DecisionDeny Decision = "deny"
	// DecisionNoOpinion - The authorizer has no opinion,
	// this should mean the action is not allowed.
	DecisionNoOpinion = "no opinion"
)
