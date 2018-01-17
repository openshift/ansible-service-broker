package user

import (
	kapi "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// Auth system gets identity name and provider
// POST to UserIdentityMapping, get back error or a filled out UserIdentityMapping object

// +genclient
// +genclient:nonNamespaced

// User -
type User struct {
	metav1.TypeMeta
	metav1.ObjectMeta

	FullName string

	Identities []string

	Groups []string
}

// List -
type List struct {
	metav1.TypeMeta
	metav1.ListMeta
	Items []User
}

// +genclient
// +genclient:nonNamespaced

// Identity -
type Identity struct {
	metav1.TypeMeta
	metav1.ObjectMeta

	// ProviderName is the source of identity information
	ProviderName string

	// ProviderUserName uniquely represents this identity in the scope of the provider
	ProviderUserName string

	// User is a reference to the user this identity is associated with
	// Both Name and UID must be set
	User kapi.ObjectReference

	Extra map[string]string
}

// IdentityList -
type IdentityList struct {
	metav1.TypeMeta
	metav1.ListMeta
	Items []Identity
}

// +genclient
// +genclient:nonNamespaced
// +genclient:onlyVerbs=get,create,update,delete

// IdentityMapping -
type IdentityMapping struct {
	metav1.TypeMeta
	metav1.ObjectMeta

	Identity kapi.ObjectReference
	User     kapi.ObjectReference
}

// +genclient
// +genclient:nonNamespaced

// Group represents a referenceable set of Users
type Group struct {
	metav1.TypeMeta
	metav1.ObjectMeta

	Users []string
}

// GroupList -
type GroupList struct {
	metav1.TypeMeta
	metav1.ListMeta
	Items []Group
}
