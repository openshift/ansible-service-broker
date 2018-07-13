package k8s

import (
	"fmt"

	"github.com/automationbroker/bundle-lib/authorization"
	"github.com/automationbroker/bundle-lib/clients"
	authv1 "k8s.io/api/authentication/v1"
	authorizationv1 "k8s.io/api/authorization/v1"
	"k8s.io/client-go/kubernetes/typed/authorization/v1"
)

// NewAuthorizer - Create a new authorizer client.
func NewAuthorizer(group, resource, verb string) (authorization.Authorizer, error) {
	k, err := clients.Kubernetes()
	if err != nil {
		return nil, fmt.Errorf("Unable to connect to the cluster")
	}
	return k8sAuthorization{
		resource: authorizationv1.ResourceAttributes{
			Group:    group,
			Resource: resource,
			Verb:     verb,
		},
		client: k.Client.AuthorizationV1().SubjectAccessReviews(),
	}, nil

}

// AuthorizationUser - A user to be used by the k8s authorizer.
type AuthorizationUser struct {
	authv1.UserInfo
}

// Username - return the username.
func (u AuthorizationUser) Username() string {
	return u.UserInfo.Username
}

type k8sAuthorization struct {
	resource authorizationv1.ResourceAttributes
	client   v1.SubjectAccessReviewInterface
}

func (a k8sAuthorization) Authorize(user authorization.AuthorizeUser, location string) (authorization.Decision, error) {
	u, ok := user.(*AuthorizationUser)
	if !ok {
		return authorization.DecisionDeny, fmt.Errorf("unknown user structure")
	}
	r := &a.resource
	r.Namespace = location
	sar := &authorizationv1.SubjectAccessReview{
		Spec: authorizationv1.SubjectAccessReviewSpec{
			User: u.UserInfo.Username,
			UID:  u.UserInfo.UID,
			//Extra:  userInfo.Extra,
			Groups:             u.UserInfo.Groups,
			ResourceAttributes: r,
		},
	}
	sar, err := a.client.Create(sar)
	if err != nil {
		return authorization.DecisionDeny, err
	}
	switch {
	case sar.Status.Denied && sar.Status.Allowed:
		return authorization.DecisionDeny, fmt.Errorf("review has both denied and allowed the request. defaulting to closed")
	case sar.Status.Denied:
		return authorization.DecisionDeny, nil
	case sar.Status.Allowed:
		return authorization.DecisionAllowed, nil
	default:
		return authorization.DecisionNoOpinion, nil
	}
}
