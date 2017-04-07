package apb

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"regexp"
	"strings"
	"time"

	logging "github.com/op/go-logging"
)

var _log *logging.Logger

var ContainerCreatingError = "status: ContainerCreating, still waiting to start"
var TimeoutFreq = 6    // Seconds
var TotalTimeout = 900 // 15min

// HACK ALERT!
// A lot of the current approach to extracting credentials and monitoring
// output is *very* experimental and error prone. Entire approach is going
// to be thrown out and redone asap.

func extractCredentials(
	output []byte, log *logging.Logger,
) (*ExtractedCredentials, error) {
	_log = log
	log.Info("{%s}", string(output))

	log.Debug("Calling getPodName")
	podname, _ := getPodName(output, log)
	log.Debug("Calling monitorOutput on " + podname)
	credOut, _ := monitorOutput(podname)
	log.Debug("oc log output: %s", string(credOut))

	var creds *ExtractedCredentials
	var err error
	creds, err = buildExtractedCredentials(credOut)

	if err.Error() == ContainerCreatingError {
		// HACK: this is HORRIBLE. but there is definitely a time between a bind
		// and when the container is up.
		totalRetries := TotalTimeout / TimeoutFreq
		retries := 1
		for {
			if retries == totalRetries {
				errstr := "TIMED OUT WAITING FOR CONTAINER TO COME UP"
				log.Error(errstr)
				return nil, errors.New(errstr)
			}

			//time.Sleep(TimeoutFreq * time.Second)
			time.Sleep(time.Duration(TimeoutFreq) * time.Second)
			log.Info("Container not up yet, retrying %d of %d on pod %s", retries, totalRetries, podname)
			credOut, _ = monitorOutput(podname)
			log.Debug("oc log output: \n%s", string(credOut))
			creds, err = buildExtractedCredentials(credOut)

			if err == nil {
				if creds != nil {
					log.Debug("Pod reporting finished and returned Credentials")
				} else {
					log.Debug("Pod reporting finished and DID NOT return Credentials")
				}
				break
			}

			if err.Error() == ContainerCreatingError {
				// Container is still not up, retry
				retries++
				continue
			}

			// unexpected error
			return nil, err
		}
	} else {
		// unexpected error
		return nil, err
	}

	return creds, err
}

// HACK: this really is a crappy way of getting output
func getPodName(output []byte, log *logging.Logger) (string, error) {
	r, err := regexp.Compile(`^pod[ \"]*(.*?)[ \"]*created`)
	if err != nil {
		return "", err
	}

	podname := r.FindStringSubmatch(string(output))

	if log != nil {
		log.Debug("%v", podname)
		log.Debug("%d", len(podname)-1)
	}

	return podname[len(podname)-1], nil
}

func monitorOutput(podname string) ([]byte, error) {
	return runCommand("oc", "logs", "-f", podname)
}

func buildExtractedCredentials(output []byte) (*ExtractedCredentials, error) {
	if strings.Contains(string(output), "ContainerCreating") {
		// Still waiting for container to come up
		_log.Debug("buildExtractedCredentials::Still waiting for container to come up!!")
		return nil, errors.New(ContainerCreatingError)
	}

	result, err := decodeOutput(output)
	if err != nil {
		return nil, err
	}

	creds := make(map[string]interface{})
	for k, v := range result {
		creds[k] = v
	}

	return &ExtractedCredentials{Credentials: creds}, nil
}

func decodeOutput(output []byte) (map[string]string, error) {
	// Possible return states
	// 1) nil, nil -> No credentials found, no errors occurred. Valid.
	// 2) creds, nil -> Credentials found, no errors occurred. Valid.
	// 3) nil, err -> Credentials found, no errors occurred. Error state.
	str := string(output)
	startIdx := strings.Index(str, "<BIND_CREDENTIALS>")
	startOffset := startIdx + len("<BIND_CREDENTIALS>")
	endIdx := strings.Index(str, "</BIND_CREDENTIALS>")

	if startIdx < 0 || endIdx < 0 {
		startIdx = strings.Index(str, "<BIND_ERROR>")
		startOffset := startIdx + len("<BIND_ERROR>")
		endIdx := strings.Index(str, "</BIND_ERROR>")
		if startIdx > -1 && endIdx > -1 {
			// Case 3, error reported
			_log.Debug("Case 3, found error")
			return nil, errors.New(str[startOffset:endIdx])
		}

		// Case 1, no creds found, no errors occurred
		_log.Debug("No creds found, no errors occurred")
		return nil, nil
	}

	_log.Debug("Attempting decode")
	decodedjson, err := base64.StdEncoding.DecodeString(str[startOffset:endIdx])
	if err != nil {
		return nil, err
	}

	_log.Debug("Raw string %s", decodedjson)
	decoded := make(map[string]string)
	json.Unmarshal(decodedjson, &decoded)
	// Case 2, creds successfully found and decoded
	_log.Debug("Unmarshaled decoded %+v", decoded)
	return decoded, nil
}
