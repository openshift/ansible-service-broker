package runtime

import (
	"errors"

	"github.com/automationbroker/bundle-lib/clients"
	log "github.com/sirupsen/logrus"
)

var (
	//ErrCredentialsNotFound - Credentials not found.
	ErrCredentialsNotFound = errors.New("extracted credentials were not found")
)

// ExtractedCredential - Interface to define CRUD operations for
// how to manage extracted credentials
type ExtractedCredential interface {
	// CreateExtractedCredentials - takes id, action, namespace, and credentials will save them.
	CreateExtractedCredential(string, string, map[string]interface{}, map[string]string) error
	// UpdateExtractedCredentials - takes id, action, namespace, and credentials will update them.
	UpdateExtractedCredential(string, string, map[string]interface{}, map[string]string) error
	// GetExtractedCredential - takes id, namespace will get credentials.
	GetExtractedCredential(string, string) (map[string]interface{}, error)
	// DeleteExtractedCredentials - takes id, namesapce and deletes the credentials.
	DeleteExtractedCredential(string, string) error
}

type defaultExtractedCredential struct{}

func (d defaultExtractedCredential) CreateExtractedCredential(ID, ns string,
	extCreds map[string]interface{}, labels map[string]string) error {

	k8scli, err := clients.Kubernetes()
	if err != nil {
		log.Errorf("Unable to get kubernetes client - %v", err)
		return err
	}
	err = k8scli.SaveExtractedCredentialSecret(ID, ns, extCreds, labels)
	if err != nil {
		log.Errorf("unable to save extracted credentials - %v", err)
		return err
	}
	return nil
}

func (d defaultExtractedCredential) UpdateExtractedCredential(ID, ns string,
	extCreds map[string]interface{}, labels map[string]string) error {

	k8scli, err := clients.Kubernetes()
	if err != nil {
		log.Errorf("Unable to get kubernetes client - %v", err)
		return err
	}
	err = k8scli.UpdateExtractedCredentialSecret(ID, ns, extCreds, labels)
	if err != nil {
		log.Errorf("unable to update extracted credentials - %v", err)
		return err
	}
	return nil
}

func (d defaultExtractedCredential) GetExtractedCredential(ID, ns string) (map[string]interface{}, error) {
	k8scli, err := clients.Kubernetes()
	if err != nil {
		log.Errorf("Unable to get kubernetes client - %v", err)
		return nil, err
	}
	creds, err := k8scli.GetExtractedCredentialSecretData(ID, ns)
	if err != nil {
		switch {
		case err == clients.ErrCredentialsNotFound:
			log.Debugf("credentials not found id: %v, namespace: %v", ID, ns)
			return nil, ErrCredentialsNotFound
		default:
			log.Errorf("unable to get extracted credentials - %v", err)
			return nil, err
		}
	}
	return creds, nil
}

func (d defaultExtractedCredential) DeleteExtractedCredential(ID, ns string) error {
	k8scli, err := clients.Kubernetes()
	if err != nil {
		log.Errorf("Unable to get kubernetes client - %v", err)
		return err
	}
	err = k8scli.DeleteExtractedCredentialSecret(ID, ns)
	if err != nil {
		log.Errorf("unable to get extracted credentials - %v", err)
		return err
	}
	return nil
}
