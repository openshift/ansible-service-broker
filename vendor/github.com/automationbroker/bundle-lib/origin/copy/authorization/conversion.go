package authorization

import (
	"fmt"

	api "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/conversion"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/kubernetes/pkg/apis/rbac"

	"github.com/automationbroker/bundle-lib/origin/copy/user/validation"
)

// reconcileProtectAnnotation is the name of an annotation which prevents reconciliation if set to "true"
// can't use this const in pkg/oc/admin/policy because of import cycle
const reconcileProtectAnnotation = "openshift.io/reconcile-protect"

func addConversionFuncs(scheme *runtime.Scheme) error {
	if err := scheme.AddConversionFuncs(
		ConvertAuthorizationClusterRoleToRBACClusterRole,
		ConvertAuthorizationRoleToRBACRole,
		ConvertAuthorizationClusterRoleBindingToRBACClusterRoleBinding,
		ConvertAuthorizationRoleBindingToRBACRoleBinding,
		ConvertRBACClusterRoleToAuthorizationClusterRole,
		ConvertRBACRoleToAuthorizationRole,
		ConvertRBACClusterRoleBindingToAuthorizationClusterRoleBinding,
		ConvertRBACRoleBindingToAuthorizationRoleBinding,
	); err != nil { // If one of the conversion functions is malformed, detect it immediately.
		return err
	}
	return nil
}

// ConvertAuthorizationClusterRoleToRBACClusterRole - convert to cluster role
func ConvertAuthorizationClusterRoleToRBACClusterRole(in *ClusterRole, out *rbac.ClusterRole, _ conversion.Scope) error {
	out.ObjectMeta = in.ObjectMeta
	out.Annotations = convertAuthorizationAnnotationsToRBACAnnotations(in.Annotations)
	out.Rules = ConvertAPIPolicyRulesToRBACPolicyRules(in.Rules)
	return nil
}

// ConvertAuthorizationRoleToRBACRole - convert role to rbac role
func ConvertAuthorizationRoleToRBACRole(in *Role, out *rbac.Role, _ conversion.Scope) error {
	out.ObjectMeta = in.ObjectMeta
	out.Rules = ConvertAPIPolicyRulesToRBACPolicyRules(in.Rules)
	return nil
}

// ConvertAuthorizationClusterRoleBindingToRBACClusterRoleBinding -
func ConvertAuthorizationClusterRoleBindingToRBACClusterRoleBinding(in *ClusterRoleBinding, out *rbac.ClusterRoleBinding, _ conversion.Scope) error {
	if len(in.RoleRef.Namespace) != 0 {
		return fmt.Errorf("invalid origin cluster role binding %s: attempts to reference role in namespace %q instead of cluster scope", in.Name, in.RoleRef.Namespace)
	}
	var err error
	if out.Subjects, err = convertAPISubjectsToRBACSubjects(in.Subjects); err != nil {
		return err
	}
	out.RoleRef = convertAPIRoleRefToRBACRoleRef(&in.RoleRef)
	out.ObjectMeta = in.ObjectMeta
	return nil
}

// ConvertAuthorizationRoleBindingToRBACRoleBinding -
func ConvertAuthorizationRoleBindingToRBACRoleBinding(in *RoleBinding, out *rbac.RoleBinding, _ conversion.Scope) error {
	if len(in.RoleRef.Namespace) != 0 && in.RoleRef.Namespace != in.Namespace {
		return fmt.Errorf("invalid origin role binding %s: attempts to reference role in namespace %q instead of current namespace %q", in.Name, in.RoleRef.Namespace, in.Namespace)
	}
	var err error
	if out.Subjects, err = convertAPISubjectsToRBACSubjects(in.Subjects); err != nil {
		return err
	}
	out.RoleRef = convertAPIRoleRefToRBACRoleRef(&in.RoleRef)
	out.ObjectMeta = in.ObjectMeta
	return nil
}

//ConvertAPIPolicyRulesToRBACPolicyRules -  Convert Policy Rules.
func ConvertAPIPolicyRulesToRBACPolicyRules(in []PolicyRule) []rbac.PolicyRule {
	rules := make([]rbac.PolicyRule, 0, len(in))
	for _, rule := range in {
		// Origin's authorizer's RuleMatches func ignores rules that have AttributeRestrictions.
		// Since we know this rule will never be respected in Origin, we do not preserve it during conversion.
		if rule.AttributeRestrictions != nil {
			continue
		}
		// We need to split this rule into multiple rules for RBAC
		if isResourceRule(&rule) && isNonResourceRule(&rule) {
			r1 := rbac.PolicyRule{
				Verbs:         rule.Verbs.List(),
				APIGroups:     rule.APIGroups,
				Resources:     rule.Resources.List(),
				ResourceNames: rule.ResourceNames.List(),
			}
			r2 := rbac.PolicyRule{
				Verbs:           rule.Verbs.List(),
				NonResourceURLs: rule.NonResourceURLs.List(),
			}
			rules = append(rules, r1, r2)
		} else {
			r := rbac.PolicyRule{
				APIGroups:       rule.APIGroups,
				Verbs:           rule.Verbs.List(),
				Resources:       rule.Resources.List(),
				ResourceNames:   rule.ResourceNames.List(),
				NonResourceURLs: rule.NonResourceURLs.List(),
			}
			rules = append(rules, r)
		}
	}
	return rules
}

func isResourceRule(rule *PolicyRule) bool {
	return len(rule.APIGroups) > 0 || len(rule.Resources) > 0 || len(rule.ResourceNames) > 0
}

func isNonResourceRule(rule *PolicyRule) bool {
	return len(rule.NonResourceURLs) > 0
}

func convertAPISubjectsToRBACSubjects(in []api.ObjectReference) ([]rbac.Subject, error) {
	subjects := make([]rbac.Subject, 0, len(in))
	for _, subject := range in {
		s := rbac.Subject{
			Name: subject.Name,
		}

		switch subject.Kind {
		case ServiceAccountKind:
			s.Kind = rbac.ServiceAccountKind
			s.Namespace = subject.Namespace
		case UserKind, SystemUserKind:
			s.APIGroup = rbac.GroupName
			s.Kind = rbac.UserKind
		case GroupKind, SystemGroupKind:
			s.APIGroup = rbac.GroupName
			s.Kind = rbac.GroupKind
		default:
			return nil, fmt.Errorf("invalid kind for origin subject: %q", subject.Kind)
		}

		subjects = append(subjects, s)
	}
	return subjects, nil
}

func convertAPIRoleRefToRBACRoleRef(in *api.ObjectReference) rbac.RoleRef {
	return rbac.RoleRef{
		APIGroup: rbac.GroupName,
		Kind:     getRBACRoleRefKind(in.Namespace),
		Name:     in.Name,
	}
}

// Infers the scope of the kind based on the presence of the namespace
func getRBACRoleRefKind(namespace string) string {
	kind := "ClusterRole"
	if len(namespace) != 0 {
		kind = "Role"
	}
	return kind
}

// ConvertRBACClusterRoleToAuthorizationClusterRole - convert rbac cluster role to cluster role
func ConvertRBACClusterRoleToAuthorizationClusterRole(in *rbac.ClusterRole, out *ClusterRole, _ conversion.Scope) error {
	out.ObjectMeta = in.ObjectMeta
	out.Annotations = convertRBACAnnotationsToAuthorizationAnnotations(in.Annotations)
	out.Rules = ConvertRBACPolicyRulesToAuthorizationPolicyRules(in.Rules)
	return nil
}

// ConvertRBACRoleToAuthorizationRole -
func ConvertRBACRoleToAuthorizationRole(in *rbac.Role, out *Role, _ conversion.Scope) error {
	out.ObjectMeta = in.ObjectMeta
	out.Rules = ConvertRBACPolicyRulesToAuthorizationPolicyRules(in.Rules)
	return nil
}

// ConvertRBACClusterRoleBindingToAuthorizationClusterRoleBinding -
func ConvertRBACClusterRoleBindingToAuthorizationClusterRoleBinding(in *rbac.ClusterRoleBinding, out *ClusterRoleBinding, _ conversion.Scope) error {
	var err error
	if out.Subjects, err = convertRBACSubjectsToAuthorizationSubjects(in.Subjects); err != nil {
		return err
	}
	if out.RoleRef, err = convertRBACRoleRefToAuthorizationRoleRef(&in.RoleRef, ""); err != nil {
		return err
	}
	out.ObjectMeta = in.ObjectMeta
	return nil
}

// ConvertRBACRoleBindingToAuthorizationRoleBinding -
func ConvertRBACRoleBindingToAuthorizationRoleBinding(in *rbac.RoleBinding, out *RoleBinding, _ conversion.Scope) error {
	var err error
	if out.Subjects, err = convertRBACSubjectsToAuthorizationSubjects(in.Subjects); err != nil {
		return err
	}
	if out.RoleRef, err = convertRBACRoleRefToAuthorizationRoleRef(&in.RoleRef, in.Namespace); err != nil {
		return err
	}
	out.ObjectMeta = in.ObjectMeta
	return nil
}

func convertRBACSubjectsToAuthorizationSubjects(in []rbac.Subject) ([]api.ObjectReference, error) {
	subjects := make([]api.ObjectReference, 0, len(in))
	for _, subject := range in {
		s := api.ObjectReference{
			Name: subject.Name,
		}

		switch subject.Kind {
		case rbac.ServiceAccountKind:
			s.Kind = ServiceAccountKind
			s.Namespace = subject.Namespace
		case rbac.UserKind:
			s.Kind = determineUserKind(subject.Name, validation.ValidateUserName)
		case rbac.GroupKind:
			s.Kind = determineGroupKind(subject.Name, validation.ValidateGroupName)
		default:
			return nil, fmt.Errorf("invalid kind for rbac subject: %q", subject.Kind)
		}

		subjects = append(subjects, s)
	}
	return subjects, nil
}

// rbac.RoleRef has no namespace field since that can be inferred from the kind of referenced role.
// The Origin role ref (api.ObjectReference) requires its namespace value to match the binding's namespace
// for a binding to a role.  For a binding to a cluster role, the namespace value must be the empty string.
// Thus we have to explicitly provide the namespace value as a parameter and use it based on the role's kind.
func convertRBACRoleRefToAuthorizationRoleRef(in *rbac.RoleRef, namespace string) (api.ObjectReference, error) {
	switch in.Kind {
	case "ClusterRole":
		return api.ObjectReference{Name: in.Name}, nil
	case "Role":
		return api.ObjectReference{Name: in.Name, Namespace: namespace}, nil
	default:
		return api.ObjectReference{}, fmt.Errorf("invalid kind %q for rbac role ref %q", in.Kind, in.Name)
	}
}

// ConvertRBACPolicyRulesToAuthorizationPolicyRules -
func ConvertRBACPolicyRulesToAuthorizationPolicyRules(in []rbac.PolicyRule) []PolicyRule {
	rules := make([]PolicyRule, 0, len(in))
	for _, rule := range in {
		r := PolicyRule{
			APIGroups:       rule.APIGroups,
			Verbs:           sets.NewString(rule.Verbs...),
			Resources:       sets.NewString(rule.Resources...),
			ResourceNames:   sets.NewString(rule.ResourceNames...),
			NonResourceURLs: sets.NewString(rule.NonResourceURLs...),
		}
		rules = append(rules, r)
	}
	return rules
}

func copyMapExcept(in map[string]string, except string) map[string]string {
	out := map[string]string{}
	for k, v := range in {
		if k != except {
			out[k] = v
		}
	}
	return out
}

var stringBool = sets.NewString("true", "false")

func convertAuthorizationAnnotationsToRBACAnnotations(in map[string]string) map[string]string {
	if value, ok := in[reconcileProtectAnnotation]; ok && stringBool.Has(value) {
		out := copyMapExcept(in, reconcileProtectAnnotation)
		if value == "true" {
			out[rbac.AutoUpdateAnnotationKey] = "false"
		} else {
			out[rbac.AutoUpdateAnnotationKey] = "true"
		}
		return out
	}
	return in
}

func convertRBACAnnotationsToAuthorizationAnnotations(in map[string]string) map[string]string {
	if value, ok := in[rbac.AutoUpdateAnnotationKey]; ok && stringBool.Has(value) {
		out := copyMapExcept(in, rbac.AutoUpdateAnnotationKey)
		if value == "true" {
			out[reconcileProtectAnnotation] = "false"
		} else {
			out[reconcileProtectAnnotation] = "true"
		}
		return out
	}
	return in
}
