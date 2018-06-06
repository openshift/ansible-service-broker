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

package registries

import (
	"regexp"
	"sync"

	log "github.com/sirupsen/logrus"
)

type filterMode uint8

const (
	filterModeBoth filterMode = iota
	filterModeWhite
	filterModeBlack
	filterModeNone
)

type failedRegexp struct {
	regex string
	err   error
}

// Filter - will handle the filtering by using a black list and white list
// of regular expressions.
type Filter struct {
	whitelist         []string
	blacklist         []string
	whiteRegexp       []*regexp.Regexp
	blackRegexp       []*regexp.Regexp
	failedWhiteRegexp []failedRegexp
	failedBlackRegexp []failedRegexp
}

// Init - Initializes Filter, precompiling regex
func (f *Filter) Init() {
	compiled, failed := compileRegexp(f.whitelist)
	f.whiteRegexp = compiled
	f.failedWhiteRegexp = failed
	compiled, failed = compileRegexp(f.blacklist)
	f.blackRegexp = compiled
	f.failedBlackRegexp = failed
}

func compileRegexp(regexStrs []string) ([]*regexp.Regexp, []failedRegexp) {
	regexps := make([]*regexp.Regexp, 0, len(regexStrs))
	failedRegexps := []failedRegexp{}

	for _, str := range regexStrs {
		if str == "" {
			log.Debugf("Ignoring empty whitelist or blacklist regex.")
			continue
		}
		cr, err := regexp.Compile(str)
		if err != nil {
			failedRegexps = append(failedRegexps, failedRegexp{str, err})
			continue
		}

		regexps = append(regexps, cr)
	}

	return regexps, failedRegexps
}

// Run - Executes filter based on white and blacklists
func (f *Filter) Run(totalList []string) ([]string, []string) {
	filterMode := f.getFilterMode()
	if filterMode == filterModeNone {
		return nil, totalList
	}

	whiteMatchSet, blackMatchSet := genMatchSets(
		filterMode, f.whiteRegexp, f.blackRegexp, totalList,
	)

	return applyMatchSets(whiteMatchSet, blackMatchSet, totalList)
}

// FilterMode - FilterMode getter
func (f *Filter) getFilterMode() filterMode {
	if len(f.whiteRegexp) != 0 && len(f.blackRegexp) != 0 {
		return filterModeBoth
	} else if len(f.whiteRegexp) != 0 {
		return filterModeWhite
	} else if len(f.blackRegexp) != 0 {
		return filterModeBlack
	}

	return filterModeNone
}

type matchSetT map[string]bool

func genMatchSets(
	filterMode filterMode,
	whiteRegexp []*regexp.Regexp,
	blackRegexp []*regexp.Regexp,
	totalList []string,
) (matchSetT, matchSetT) {
	var whiteMatchSet, blackMatchSet matchSetT

	if filterMode == filterModeBoth {
		whiteMatchChan := genMatchSet(whiteRegexp, totalList)
		blackMatchChan := genMatchSet(blackRegexp, totalList)
		whiteMatchSet = <-whiteMatchChan
		blackMatchSet = <-blackMatchChan
	} else if filterMode == filterModeWhite {
		whiteMatchChan := genMatchSet(whiteRegexp, totalList)
		whiteMatchSet = <-whiteMatchChan
	} else if filterMode == filterModeBlack {
		blackMatchChan := genMatchSet(blackRegexp, totalList)
		blackMatchSet = <-blackMatchChan
	}

	return whiteMatchSet, blackMatchSet
}

func genMatchSet(
	regexpList []*regexp.Regexp,
	totalList []string,
) <-chan matchSetT {
	matchChunksChan := make(chan []string)

	var wg sync.WaitGroup
	wg.Add(len(regexpList))

	// Produce set of matches for each regex against each item in totalList
	// Run each regex against totalList concurrently
	// Expect one match set per regex, each containing the matches found
	// when running the regex over the totalList
	for _, rx := range regexpList {
		go func(rx *regexp.Regexp) {
			matchChunk := []string{}
			for _, testStr := range totalList {
				if ok := rx.MatchString(testStr); ok {
					matchChunk = append(matchChunk, testStr)
				}
			}
			matchChunksChan <- matchChunk
			wg.Done()
		}(rx)
	}

	// Join workers and close chunks channel when finished
	go func() {
		wg.Wait()
		close(matchChunksChan)
	}()

	return reduceChunks(matchChunksChan)
}

func reduceChunks(chunks <-chan []string) <-chan matchSetT {
	// Fan-in chunks, reduce to single set of matches
	out := make(chan matchSetT)
	go func() {
		matchSet := make(matchSetT)
		// Read and process chunks as they complete until input chan is closed
		for chunk := range chunks {
			for _, match := range chunk {
				matchSet[match] = true
			}
		}
		// Output single reduced matchset and close channel
		out <- matchSet
		close(out)
	}()
	return out
}

// applyMatchSets - Returns validNames and filteredNames
func applyMatchSets(
	whiteMatchSet matchSetT,
	blackMatchSet matchSetT,
	totalList []string,
) ([]string, []string) {
	filteredVals := []string{}
	totalSet := toMatchSetT(totalList)

	if len(whiteMatchSet) == 0 {
		// If nothing is whitelisted, filter everything
		filteredVals = totalList
		totalSet = nil
	} else if len(blackMatchSet) != 0 && len(whiteMatchSet) != 0 {
		// Blacklist matches override white
		for k := range blackMatchSet {
			if _, ok := blackMatchSet[k]; ok {
				delete(whiteMatchSet, k)
			}
		}

		// Only whitelisted vals pass
		for _, k := range totalList {
			if _, ok := whiteMatchSet[k]; !ok {
				delete(totalSet, k)
				filteredVals = append(filteredVals, k)
			}
		}
	} else if len(whiteMatchSet) != 0 {
		// Only whitelisted vals pass
		for _, k := range totalList {
			if _, ok := whiteMatchSet[k]; !ok {
				delete(totalSet, k)
				filteredVals = append(filteredVals, k)
			}
		}
	} else {
		// Filter everything in the blacklist
		for _, k := range totalList {
			if _, ok := blackMatchSet[k]; ok {
				delete(totalSet, k)
				filteredVals = append(filteredVals, k)
			}
		}
	}

	return toSlice(totalSet), filteredVals
}

func toMatchSetT(s []string) matchSetT {
	m := make(matchSetT)
	for _, k := range s {
		m[k] = true
	}
	return m
}

func toSlice(m matchSetT) []string {
	s := make([]string, len(m))
	i := 0
	for k := range m {
		s[i] = k
		i++
	}
	return s
}
