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

package bundle

import (
	"github.com/coreos/go-semver/semver"
	log "github.com/sirupsen/logrus"
)

// These constants describe the minimum and maximum
// accepted APB spec versions. They are used to filter
// acceptable APBs.

// MinSpecVersion constant to describe minimum supported spec version
const MinSpecVersion = "1.0.0"

// MaxSpecVersion constant to describe maximum supported spec version
const MaxSpecVersion = "1.0.0"

// These constants describe the minimum and maximum
// accepted APB runtime versions. They are used to filter
// acceptable APBs.

// MinRuntimeVersion constant to describe minimum supported runtime version
const MinRuntimeVersion = 1

// MaxRuntimeVersion constant to describe maximum supported runtime version
const MaxRuntimeVersion = 2

// The minimum/maximum Bundle spec semantic versions
var minSpecSemver = semver.New("1.0.0")
var maxSpecSemver = semver.New("1.0.0")

// ValidateVersion - Ensure the Bundle Spec Version and Bundle Runtime Version
// are within bounds
func (s *Spec) ValidateVersion() bool {
	return s.checkVersion() && s.checkRuntime()
}

func (s *Spec) checkVersion() bool {
	specSemver, err := semver.NewVersion(s.Version)
	if err != nil {
		if s.Version == "1.0" {
			log.Debugf("Spec [%v] version (%v) not semver compatible", s.FQName, s.Version)
			return true
		}
		return false
	}

	if specSemver.Compare(*minSpecSemver) < 0 {
		log.Errorf("Spec version (%v) is less than the minimum version %v", s.Version, MinSpecVersion)
		return false
	}

	if specSemver.Compare(*maxSpecSemver) > 0 {
		log.Errorf("Spec version (%v) is greater than the maximum version %v", s.Version, MaxSpecVersion)
		return false
	}

	return true
}

func (s *Spec) checkRuntime() bool {
	return s.Runtime >= MinRuntimeVersion && s.Runtime <= MaxRuntimeVersion
}
