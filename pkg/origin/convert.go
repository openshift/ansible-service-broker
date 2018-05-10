package origin

import (
	"github.com/openshift/api/authorization/v1"
	"k8s.io/kubernetes/pkg/apis/rbac"
)

//ConvertAPIPolicyRulesToRBACPolicyRules -  Convert Policy Rules.
func ConvertAPIPolicyRulesToRBACPolicyRules(in []v1.PolicyRule) []rbac.PolicyRule {
	rules := make([]rbac.PolicyRule, 0, len(in))
	for _, rule := range in {
		// We need to split this rule into multiple rules for RBAC
		if isResourceRule(&rule) && isNonResourceRule(&rule) {
			r1 := rbac.PolicyRule{
				Verbs:         rule.Verbs,
				APIGroups:     rule.APIGroups,
				Resources:     rule.Resources,
				ResourceNames: rule.ResourceNames,
			}
			r2 := rbac.PolicyRule{
				Verbs:           rule.Verbs,
				NonResourceURLs: rule.NonResourceURLsSlice,
			}
			rules = append(rules, r1, r2)
		} else {
			r := rbac.PolicyRule{
				APIGroups:       rule.APIGroups,
				Verbs:           rule.Verbs,
				Resources:       rule.Resources,
				ResourceNames:   rule.ResourceNames,
				NonResourceURLs: rule.NonResourceURLsSlice,
			}
			rules = append(rules, r)
		}
	}
	return rules
}

func isResourceRule(rule *v1.PolicyRule) bool {
	return len(rule.APIGroups) > 0 || len(rule.Resources) > 0 || len(rule.ResourceNames) > 0
}

func isNonResourceRule(rule *v1.PolicyRule) bool {
	return len(rule.NonResourceURLsSlice) > 0
}
