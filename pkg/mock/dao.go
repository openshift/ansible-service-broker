package mock

import (
	"fmt"

	apb "github.com/automationbroker/bundle-lib/bundle"
)

// SubscriberDAO is mock DAO
type SubscriberDAO struct {
	calls     map[string]int
	Errs      map[string]error
	assertErr []error
	AssertOn  map[string]func(...interface{}) error
	Object    map[string]interface{}
}

// SetExtractedCredentials sets extracted credentials
func (mp *SubscriberDAO) SetExtractedCredentials(id string, extCreds *apb.ExtractedCredentials) error {
	assert := mp.AssertOn["SetExtractedCredentials"]
	if nil != assert {
		if err := assert(id, extCreds); err != nil {
			mp.assertErr = append(mp.assertErr, err)
			return err
		}
	}
	mp.calls["SetExtractedCredentials"]++
	return mp.Errs["SetExtractedCredentials"]

}

// SetState sets the JobState
func (mp *SubscriberDAO) SetState(id string, state apb.JobState) (string, error) {
	assert := mp.AssertOn["SetState"]
	if nil != assert {
		if err := assert(id, state); err != nil {
			mp.assertErr = append(mp.assertErr, err)
			return id, err
		}
	}
	mp.calls["SetState"]++
	return id, mp.Errs["SetState"]

}

// DeleteExtractedCredentials deletes extracted credentials
func (mp *SubscriberDAO) DeleteExtractedCredentials(id string) error {
	assert := mp.AssertOn["DeleteExtractedCredentials"]
	if nil != assert {
		if err := assert(id); err != nil {
			mp.assertErr = append(mp.assertErr, err)
			return err
		}
	}
	mp.calls["DeleteExtractedCredentials"]++
	return mp.Errs["DeleteExtractedCredentials"]
}

// DeleteServiceInstance deletes the serviceInstance
func (mp *SubscriberDAO) DeleteServiceInstance(id string) error {
	assert := mp.AssertOn["DeleteServiceInstance"]
	if nil != assert {
		if err := assert(id); err != nil {
			mp.assertErr = append(mp.assertErr, err)
			return err
		}
	}
	mp.calls["DeleteServiceInstance"]++
	return mp.Errs["DeleteServiceInstance"]
}

// GetServiceInstance gets a serviceInstance by id
func (mp *SubscriberDAO) GetServiceInstance(id string) (*apb.ServiceInstance, error) {
	assert := mp.AssertOn["GetServiceInstance"]
	if nil != assert {
		if err := assert(id); err != nil {
			mp.assertErr = append(mp.assertErr, err)
			return nil, err
		}
	}
	mp.calls["GetServiceInstance"]++
	retOb := mp.Object["GetServiceInstance"]
	if nil == retOb {
		return nil, mp.Errs["GetServiceInstance"]
	}
	return retOb.(*apb.ServiceInstance), mp.Errs["GetServiceInstance"]
}

// DeleteBinding mock impl
func (mp *SubscriberDAO) DeleteBinding(bindingInst apb.BindInstance, serviceInst apb.ServiceInstance) error {
	assert := mp.AssertOn["DeleteBinding"]
	if nil != assert {
		if err := assert(bindingInst, serviceInst); err != nil {
			mp.assertErr = append(mp.assertErr, err)
			return err
		}
	}
	mp.calls["DeleteBinding"]++
	return mp.Errs["DeleteBinding"]
}

// SetServiceInstance mock impl
func (mp *SubscriberDAO) SetServiceInstance(id string, serviceInstance *apb.ServiceInstance) error {
	assert := mp.AssertOn["SetServiceInstance"]
	if nil != assert {
		if err := assert(id); err != nil {
			mp.assertErr = append(mp.assertErr, err)
			return err
		}
	}
	mp.calls["SetServiceInstance"]++
	return mp.Errs["SetServiceInstance"]
}

// GetBindInstance mock impl
func (mp *SubscriberDAO) GetBindInstance(id string) (*apb.BindInstance, error) {
	assert := mp.AssertOn["GetBindInstance"]
	if nil != assert {
		if err := assert(id); err != nil {
			mp.assertErr = append(mp.assertErr, err)
			return nil, err
		}
	}
	mp.calls["GetBindInstance"]++
	retOb := mp.Object["GetBindInstance"]
	if nil == retOb {
		return nil, mp.Errs["GetBindInstance"]
	}
	return retOb.(*apb.BindInstance), mp.Errs["GetBindInstance"]
}

// CheckCalls will check the calls made match the expected calls
func (mp *SubscriberDAO) CheckCalls(calls map[string]int) error {
	for k, v := range calls {
		if mp.calls[k] != v {
			return fmt.Errorf("expected %d calls to %s but got %d ", v, k, mp.calls[k])
		}
	}
	return nil
}

// AssertErrors returns any assert errors
func (mp *SubscriberDAO) AssertErrors() []error {
	return mp.assertErr
}

// NewSubscriberDAO returns mock SubscriberDAO
func NewSubscriberDAO() *SubscriberDAO {
	return &SubscriberDAO{
		Errs:     map[string]error{},
		calls:    map[string]int{},
		AssertOn: map[string]func(...interface{}) error{},
		Object:   map[string]interface{}{},
	}
}
