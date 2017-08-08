package secrets

import (
	meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	logging "github.com/op/go-logging"
	"github.com/openshift/ansible-service-broker/pkg/apb"
	"github.com/openshift/ansible-service-broker/pkg/clients"
)

type secret string

type Config struct {
	Title   string
	ApbName string
	Secret  string
}

func (c Config) Validate() bool {
	for _, str := range []string{c.Title, c.ApbName, c.Secret} {
		if str == "" {
			return false
		}
	}
	return true
}

type Secret struct {
	apbName string
	secret  string
}

type Secrets struct {
	config  []Config
	secrets []Secret
	log     *logging.Logger
}

func NewSecrets(config []Config, log *logging.Logger) Secrets {
	sekrets := []Secret{}
	for _, cfg := range config {
		sekrets = append(sekrets, Secret{cfg.ApbName, cfg.Secret})
	}
	return Secrets{config, sekrets, log}
}

func (s Secrets) Filter(inSpecs []*apb.Spec) ([]*apb.Spec, error) {

	for _, spec := range inSpecs {

		for _, secret := range s.secrets {
			if spec.FQName == secret.apbName {
				keys, err := getKeys(secret.secret, s.log)
				if err != nil {
					return nil, err
				}

				newPlans := []apb.Plan{}
				for _, plan := range spec.Plans {
					newParams := []apb.ParameterDescriptor{}

					for _, param := range plan.Parameters {
						found := false

						for _, key := range keys {
							if key == param.Name {
								found = true
							}
						}

						if !found {
							newParams = append(newParams, param)
						}

					}
					plan.Parameters = newParams
					newPlans = append(newPlans, plan)
				}

				spec.Plans = newPlans
			}
		}
	}
	return inSpecs, nil
}

func getKeys(secretName string, log *logging.Logger) ([]string, error) {
	k8scli, err := clients.Kubernetes(log)
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
