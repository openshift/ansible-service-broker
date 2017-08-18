package apb

import (
	"sync"

	meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	logging "github.com/op/go-logging"
	"github.com/openshift/ansible-service-broker/pkg/clients"
)

// SecretsConfig - Entry for a secret config block in broker config
type SecretsConfig struct {
	Title   string `yaml:"title"`
	ApbName string `yaml:"apb_name"`
	Secret  string `yaml:"secret"`
}

// Validate - Ensures that the secrets config is valid (ie, all strings are
// non-empty
func (c SecretsConfig) Validate() bool {
	for _, str := range []string{c.Title, c.ApbName, c.Secret} {
		if str == "" {
			return false
		}
	}
	return true
}

// AssociationRule - A rule to associate apbs with a secrets
type AssociationRule struct {
	apbName string
	secret  string
}

type secretsCache struct {
	mapping map[string][]string
	rwSync  sync.RWMutex
	rules   []AssociationRule
	config  []SecretsConfig
	log     *logging.Logger
}

var secrets secretsCache

// GetSecrets - Returns a list of secrets to be attached to a specified spec
func GetSecrets(spec *Spec) []string {
	secrets.rwSync.RLock()
	defer secrets.rwSync.RUnlock()
	return secrets.mapping[spec.FQName]
}

// AddSecrets - Uses the AssociationRules generated from config to link specs to
// secrets and add them to the global secrets cache
func AddSecrets(specs []*Spec) {
	for _, spec := range specs {
		AddSecretsFor(spec)
	}
}

// AddSecretsFor - Uses AssociationRules for a given spec to link the spec to
// secrets and add them to the global secrets cache
func AddSecretsFor(spec *Spec) {
	secrets.rwSync.Lock()
	defer secrets.rwSync.Unlock()

	for _, rule := range secrets.rules {
		if match(spec, rule) {
			addSecret(spec, rule)
		}
	}
}

func addSecret(spec *Spec, rule AssociationRule) {
	secrets.mapping[spec.FQName] = append(secrets.mapping[spec.FQName], rule.secret)
}

func match(spec *Spec, rule AssociationRule) bool {
	return spec.FQName == rule.apbName
}

// InitializeSecretsCache - Generates AssociationRules from config and
// initializes the global secrets cache
func InitializeSecretsCache(config []SecretsConfig, log *logging.Logger) {
	rules := []AssociationRule{}
	for _, cfg := range config {
		rules = append(rules, AssociationRule{cfg.ApbName, cfg.Secret})
	}
	secrets = secretsCache{
		mapping: make(map[string][]string),
		rwSync:  sync.RWMutex{},
		log:     log,
		rules:   rules,
		config:  config,
	}
}

// FilterSecrets - Filters all parameters masked by a secret out of the given
// specs
func FilterSecrets(inSpecs []*Spec) ([]*Spec, error) {
	for _, spec := range inSpecs {
		secrets.log.Debugf("Filtering spec %v", spec.FQName)
		for _, secret := range GetSecrets(spec) {
			secretKeys, err := getSecretKeys(secret)
			if err != nil {
				return nil, err
			}
			secrets.log.Debugf("Found secret with name %v", secret)
			spec.Plans = filterPlans(spec.Plans, secretKeys)
		}
	}
	return inSpecs, nil
}

func filterPlans(inPlans []Plan, secretKeys []string) []Plan {
	newPlans := []Plan{}
	for _, plan := range inPlans {
		plan.Parameters = filterParameters(plan.Parameters, secretKeys)
		newPlans = append(newPlans, plan)
	}
	return newPlans
}

func filterParameters(inParams []ParameterDescriptor, secretKeys []string) []ParameterDescriptor {
	newParams := []ParameterDescriptor{}

	for _, param := range inParams {
		if !paramInSecret(param, secretKeys) {
			newParams = append(newParams, param)
		}

	}
	return newParams
}

func paramInSecret(param ParameterDescriptor, secretKeys []string) bool {
	for _, key := range secretKeys {
		if key == param.Name {
			secrets.log.Debugf("Param %v matched", param.Name, key)
			return true
		}
	}
	return false
}

func getSecretKeys(secretName string) ([]string, error) {
	k8scli, err := clients.Kubernetes(secrets.log)
	if err != nil {
		return nil, err
	}

	secretData, err := k8scli.CoreV1().Secrets("ansible-service-broker").Get(secretName, meta_v1.GetOptions{})
	if err != nil {
		return nil, err
	}

	ret := []string{}
	for key := range secretData.Data {
		ret = append(ret, key)
	}
	return ret, nil
}
