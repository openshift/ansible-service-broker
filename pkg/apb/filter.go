package apb

import (
	"errors"
	"io/ioutil"
	"regexp"
	"sync"

	yaml "gopkg.in/yaml.v2"
)

type filterMode uint8

const (
	filterModeBoth filterMode = iota
	filterModeWhite
	filterModeBlack
	filterModeNone
)

type Filter struct {
	whitelistFile string
	blacklistFile string
	whitelist     []string
	blacklist     []string
}

// NewFilter - Creates a new ApbFilter
func NewFilter(whitelistFile string, blacklistFile string) Filter {
	return Filter{
		whitelistFile: whitelistFile,
		blacklistFile: blacklistFile,
		whitelist:     []string{},
		blacklist:     []string{},
	}
}

// Init - Initializes a filter, reading in whitelist and blacklist files
// Will error out if files do not exist, or neither files are provided
func (f *Filter) Init() error {
	if f.whitelistFile == "" && f.blacklistFile == "" {
		return errors.New("No whitelist or blacklist file specified. Cannot init Filter")
	}

	if f.whitelistFile != "" {
		contents, err := ioutil.ReadFile(f.whitelistFile)
		if err != nil {
			return err
		}
		if yaml.Unmarshal(contents, &f.whitelist); err != nil {
			return err
		}
	}

	if f.blacklistFile != "" {
		contents, err := ioutil.ReadFile(f.blacklistFile)
		if err != nil {
			return err
		}
		if yaml.Unmarshal(contents, &f.blacklist); err != nil {
			return err
		}
	}

	return nil
}

// Run - Executes filter based on white and blacklists
func (f *Filter) Run(totalList []string) ([]string, []string) {
	filterMode := f.getFilterMode()
	if filterMode == filterModeNone {
		return totalList, nil
	}

	whiteMatchSet, blackMatchSet := genMatchSets(
		filterMode, f.whitelist, f.blacklist, totalList,
	)

	return applyMatchSets(whiteMatchSet, blackMatchSet, totalList)
}

// FilterMode - FilterMode getter
func (f *Filter) getFilterMode() filterMode {
	if len(f.whitelist) != 0 && len(f.blacklist) != 0 {
		return filterModeBoth
	} else if len(f.whitelist) != 0 {
		return filterModeWhite
	} else if len(f.blacklist) != 0 {
		return filterModeBlack
	}

	return filterModeNone
}

type matchSetT map[string]bool

func genMatchSets(
	filterMode filterMode,
	whitelist []string,
	blacklist []string,
	totalList []string,
) (matchSetT, matchSetT) {
	var whiteMatchSet, blackMatchSet matchSetT

	if filterMode == filterModeBoth {
		whiteMatchChan := genMatchSet(whitelist, totalList)
		blackMatchChan := genMatchSet(blacklist, totalList)
		whiteMatchSet = <-whiteMatchChan
		blackMatchSet = <-blackMatchChan
	} else if filterMode == filterModeWhite {
		whiteMatchChan := genMatchSet(whitelist, totalList)
		whiteMatchSet = <-whiteMatchChan
	} else if filterMode == filterModeBlack {
		blackMatchChan := genMatchSet(blacklist, totalList)
		blackMatchSet = <-blackMatchChan
	}

	return whiteMatchSet, blackMatchSet
}

func genMatchSet(
	regexList []string,
	totalList []string,
) <-chan matchSetT {
	matchChunksChan := make(chan []string)

	var wg sync.WaitGroup
	wg.Add(len(regexList))

	// Produce set of matches for each regex against each item in totalList
	// Run each regex against totalList concurrently
	// Expect one match set per regex, each containing the matches found
	// when running the regex over the totalList
	for _, regex := range regexList {
		go func(regex string) {
			matchChunk := []string{}
			for _, testStr := range totalList {
				if ok, _ := regexp.MatchString(regex, testStr); ok {
					matchChunk = append(matchChunk, testStr)
				}
			}
			matchChunksChan <- matchChunk
			wg.Done()
		}(regex)
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
		// Read and process chunks as they complete until input chan is cloesd
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

	if len(whiteMatchSet) != 0 && len(blackMatchSet) != 0 {
		// Blacklist matches override white
		for k, _ := range blackMatchSet {
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
	for k, _ := range m {
		s[i] = k
		i++
	}
	return s
}
