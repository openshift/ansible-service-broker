package apb

import (
	"sync"

	meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	logging "github.com/op/go-logging"
	"github.com/openshift/ansible-service-broker/pkg/clients"
)

type secret string

// secrets:
// - title: All database apbs
// 	whitelist:
// 	- *db-apb
// 	secrets:
// 	- default_db_credentials
// - title: All amazon apbs
// 	whitelist:
// 	- amazon-*
// 	- aws-*
// 	secrets:
// 	- default_db_credentials
// 	- aws_production_credentials

type SecretsConfig struct {
	Title   string
	ApbName string
	Secret  string
	// Whitelist []string
	// Secrets   []string
}

func (c SecretsConfig) Validate() bool {
	// strs := append(c.Whitelist, append(c.Secrets, c.Title)...)

	// for _, str := range strs {
	for _, str := range []string{c.Title, c.ApbName, c.Secret} {
		if str == "" {
			return false
		}
	}
	return true
}

type AssociationRule struct {
	apbName string
	secret  string
	// apbFilter registries.Filter
	// planFilter registries.Filter
	// secretKeys       []string
}

type secretsCache struct {
	mapping map[string][]string
	rwSync  sync.RWMutex
	rules   []AssociationRule
	config  []SecretsConfig
	log     *logging.Logger
}

var secrets secretsCache

func GetSecrets(spec *Spec) []string {
	secrets.rwSync.RLock()
	defer secrets.rwSync.RUnlock()
	return secrets.mapping[spec.FQName]
}

func AddSecrets(specs []*Spec) {
	for _, spec := range specs {
		AddSecretsFor(spec)
	}
}

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

func NewSecrets(config []SecretsConfig, log *logging.Logger) {
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

func FilterSecrets(inSpecs []*Spec) ([]*Spec, error) {
	for _, spec := range inSpecs {
		for _, secret := range GetSecrets(spec) {
			secretKeys, err := getSecretKeys(secret)
			if err != nil {
				return nil, err
			}
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

// func FilterParameters(

// func FilterParameters(inSpecs []*Spec) ([]*Spec, string) {
// 	var wg sync.WaitGroup
// 	wg.Add(len(inSpecs))

// 	type resultT struct {
// 		spec       *Spec
// 		failReason string
// 	}

// 	results := make(chan resultT)
// 	for _, inSpec := range inSpecs {
// 		go func(spec *Spec) {
// 			defer wg.Done()
// 			updatedSpec, failReason := s.filterSecrets(spec)
// 			results <- resultT{updatedSpec, failReason}
// 		}(inSpec)
// 	}

// 	go func() {
// 		wg.Wait()
// 		close(results)
// 	}()

// 	filtered := make([]*Spec, 0, len(inSpecs))
// 	for _, spec := range len(inSpecs) {
// 		filtered = append(filtered, spec)
// 	}

// 	return filtered, nil

// }
