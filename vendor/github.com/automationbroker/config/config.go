//
// Copyright (c) 2018 Red Hat, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//    http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//
// Red Hat trademarks are not licensed under Apache License, Version 2.
// No permission is granted to use or replicate Red Hat trademarks that
// are incorporated in this software or its documentation.
//

package config

import (
	"fmt"
	"io/ioutil"
	"strings"
	"sync"

	log "github.com/sirupsen/logrus"
	yaml "gopkg.in/yaml.v1"
)

// Config - The base config for the pieces of the applcation
type Config struct {
	config map[string]interface{}
	mutex  sync.RWMutex
}

// CreateConfig - Read config file and create the Config struct
func CreateConfig(configFile string) (*Config, error) {
	var err error
	// Load struct
	var dat []byte
	if dat, err = ioutil.ReadFile(configFile); err != nil {
		return &Config{config: map[string]interface{}{}, mutex: sync.RWMutex{}}, err
	}
	c := map[string]interface{}{}

	if err = yaml.Unmarshal(dat, &c); err != nil {
		return &Config{config: map[string]interface{}{}, mutex: sync.RWMutex{}}, err
	}
	return &Config{config: c, mutex: sync.RWMutex{}}, nil
}

// NewConfigFromMap - Create a new config object from the underlying map
func NewConfigFromMap(m map[string]interface{}) *Config {
	return &Config{config: m, mutex: sync.RWMutex{}}
}

// Empty - Determines if the configuration is empty
func (c *Config) Empty() bool {
	return len(c.config) == 0
}

// GetString - Retrieve the configuration value as a string.
func (c *Config) GetString(key string) string {
	c.mutex.RLock()
	subMap := createNewMap(c.config)
	//Can unlock the map for reading after we copy the map.
	c.mutex.RUnlock()
	keys := strings.Split(key, ".")
	val := retrieveValueFromKeys(keys, subMap)
	if v, ok := val.(string); ok {
		return v
	}
	log.Debugf("Unable to get %v from config", key)
	return ""
}

// GetSliceOfStrings - Retrieve the configuration value that is a slice of strings.
func (c *Config) GetSliceOfStrings(key string) []string {
	c.mutex.RLock()
	subMap := createNewMap(c.config)
	//Can unlock the map for reading after we copy the map.
	c.mutex.RUnlock()
	keys := strings.Split(key, ".")
	val := retrieveValueFromKeys(keys, subMap)
	var s []string
	if v, ok := val.([]interface{}); ok {
		for _, str := range v {
			if st, ok := str.(string); ok {
				s = append(s, st)
			} else {
				return nil
			}
		}
		return s
	}
	log.Debugf("Unable to get %v from config", key)
	return nil
}

// GetInt - Retrieve the configuration value as a int.
func (c *Config) GetInt(key string) int {
	c.mutex.RLock()
	subMap := createNewMap(c.config)
	//Can unlock the map for reading after we copy the map.
	c.mutex.RUnlock()
	keys := strings.Split(key, ".")
	val := retrieveValueFromKeys(keys, subMap)
	if v, ok := val.(int); ok {
		return v
	}
	log.Debugf("Unable to get %v from config", key)
	return 0
}

// GetBool - Retrieve the configuration value as a bool.
func (c *Config) GetBool(key string) bool {
	c.mutex.RLock()
	subMap := createNewMap(c.config)
	//Can unlock the map for reading after we copy the map.
	c.mutex.RUnlock()
	keys := strings.Split(key, ".")
	val := retrieveValueFromKeys(keys, subMap)
	if v, ok := val.(bool); ok {
		return v
	}
	log.Debugf("Unable to get %v from config", key)
	return false
}

// GetFloat64 - Retrieve the configuration value as a float64
func (c *Config) GetFloat64(key string) float64 {
	c.mutex.RLock()
	subMap := createNewMap(c.config)
	//Can unlock the map for reading after we copy the map.
	c.mutex.RUnlock()
	keys := strings.Split(key, ".")
	val := retrieveValueFromKeys(keys, subMap)
	if v, ok := val.(float64); ok {
		return v
	}
	log.Debugf("Unable to get %v from config", key)
	return float64(0)
}

// GetFloat32 - Retrieve the configuration value as a float32
func (c *Config) GetFloat32(key string) float32 {
	c.mutex.RLock()
	subMap := createNewMap(c.config)
	//Can unlock the map for reading after we copy the map.
	c.mutex.RUnlock()
	keys := strings.Split(key, ".")
	val := retrieveValueFromKeys(keys, subMap)
	if v, ok := val.(float64); ok {
		return float32(v)
	}
	log.Debugf("Unable to get %v from config", key)
	return float32(0)
}

// GetSubConfig - Retrieve the sub map
func (c *Config) GetSubConfig(key string) *Config {
	c.mutex.RLock()
	subMap := createNewMap(c.config)
	//Can unlock the map for reading after we copy the map.
	c.mutex.RUnlock()
	keys := strings.Split(key, ".")
	val := retrieveValueFromKeys(keys, subMap)
	switch val.(type) {
	case *Config:
		return val.(*Config)
	default:
		log.Debugf("Unable to get %v from config", key)
		return &Config{config: map[string]interface{}{}, mutex: sync.RWMutex{}}
	}
}

// GetSubConfigArray - Retrieve an array of sub configs
// Example
// - name: key
//   - key: value
//     key1: value1
//   - key: newvalue
//     key1: newvalue1
// [Config{key: value, key1: value1}, Config{key: newvalue, key1: newvalue1}]
func (c *Config) GetSubConfigArray(key string) []*Config {
	c.mutex.RLock()
	subMap := createNewMap(c.config)
	keys := strings.Split(key, ".")
	val := retrieveValueFromKeys(keys, subMap)
	switch val.(type) {
	case []interface{}:
		return createSubConfigForArray(val.([]interface{}))
	default:
		log.Debugf("Unable to get %v from config", key)
		return []*Config{}
	}
}

func createSubConfigForArray(vals []interface{}) []*Config {
	configs := []*Config{}
	for _, value := range vals {
		if v, ok := value.(map[string]interface{}); ok {
			configs = append(configs, &Config{mutex: sync.RWMutex{}, config: v})
		} else if v, ok := value.(map[interface{}]interface{}); ok {
			// If no name field don't fill up the map
			s, err := createStringMap(v)
			if err != nil {
				return configs
			}
			configs = append(configs, &Config{mutex: sync.RWMutex{}, config: s})
		} else {
			//We don't know what to do if they are not key value pairs.
			return configs
		}
	}
	return configs
}

// ToMap - Retrieve a copy of the undlying map for the config.
func (c *Config) ToMap() map[string]interface{} {
	c.mutex.RLock()
	subMap := createNewMap(c.config)
	//Can unlock the map for reading after we copy the map.
	c.mutex.RUnlock()
	return subMap
}

//createNewMap - Need to create a new map to work with.
func createNewMap(config map[string]interface{}) map[string]interface{} {
	newMap := map[string]interface{}{}
	for k, v := range config {
		newMap[k] = v
	}
	return newMap
}

func retrieveValueFromKeys(keys []string, subMap map[string]interface{}) interface{} {
	var val interface{}
	for i, key := range keys {
		val = subMap[key]
		switch val.(type) {
		case map[string]interface{}:
			//If the final value then we should return the a new
			// configuration with the submap as the config
			subMap = val.(map[string]interface{})
			if i == len(keys)-1 {
				return &Config{mutex: sync.RWMutex{}, config: val.(map[string]interface{})}
			}
		case map[interface{}]interface{}:
			var err error
			subMap, err = createStringMap(val.(map[interface{}]interface{}))
			if err != nil {
				return &Config{mutex: sync.RWMutex{}, config: map[string]interface{}{}}
			}
			// We do not know what to do if the key is not a string. Error here.
			if i == len(keys)-1 {
				return &Config{mutex: sync.RWMutex{}, config: subMap}
			}
		case []interface{}:
			//If array of values we will attempt to make a map using the a
			// name field on the underlying object.
			//TODO: we should eventually add this back to the original map,
			if i == len(keys)-1 {
				return val
			}
			// then we could check for existence of this sub map.
			subMap = createSubMapFromArray(val.([]interface{}))
			if len(subMap) > 0 && i == len(keys)-1 {
				return &Config{config: subMap, mutex: sync.RWMutex{}}
			}
		default:
			if i == len(keys)-1 {
				return val
			}
			//If invalid key with no value, return an empty configuration.
			return &Config{config: map[string]interface{}{}, mutex: sync.RWMutex{}}
		}
	}
	return val
}

// createSubMapFromArray - create the submap from the array
// checks if the name field is in the underlying map and will use as the key
func createSubMapFromArray(val []interface{}) map[string]interface{} {
	subMap := map[string]interface{}{}
	for _, value := range val {
		if v, ok := value.(map[string]interface{}); ok {
			// If no name field don't fill up the map
			if name, ok := v["name"]; ok {
				if n, ok := name.(string); ok {
					subMap[n] = value
				}
			}
		} else if v, ok := value.(map[interface{}]interface{}); ok {
			// If no name field don't fill up the map
			s, err := createStringMap(v)
			if err != nil {
				return subMap
			}
			if name, ok := s["name"]; ok {
				if n, ok := name.(string); ok {
					subMap[n] = value
				}
			} else if t, ok := s["type"]; ok {
				if tp, ok := t.(string); ok {
					subMap[tp] = value
				}
			}
		} else {
			//We don't know what to do if they are not key value pairs.
			return subMap
		}
	}
	return subMap
}

func createStringMap(val map[interface{}]interface{}) (map[string]interface{}, error) {
	subMap := map[string]interface{}{}
	for key, v := range val {
		if k, ok := key.(string); ok {
			subMap[k] = v
		} else {
			return subMap, fmt.Errorf("Unable to understand non string key")
		}
	}
	return subMap, nil
}
