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

var StillWaitingError = "status: still waiting to start"
var TimeoutFreq = 6    // Seconds
var TotalTimeout = 900 // 15min

// HACK ALERT!
// A lot of the current approach to extracting credentials and monitoring
// output is *very* experimental and error prone. Entire approach is going
// to be thrown out and redone asap.

func extractCredentials(
	podname string, log *logging.Logger,
) (*ExtractedCredentials, error) {
	log.Debug("Calling monitorOutput on " + podname)
	credOut, _ := monitorOutput(podname)
	log.Debug("oc log output: %s", string(credOut))

	var creds *ExtractedCredentials
	var err error
	creds, err = buildExtractedCredentials(credOut)

	if err != nil {
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
			} else if err.Error() == StillWaitingError {
				// Known error code that's received when we're either waiting for
				// ContainerCreating, or for the pod resource to be created.
				// These are expected states, and we'll wait until the pod is up.
				log.Debug(err.Error())
			} else {
				log.Notice("WARNING: Unexpected output from apb pod")
				log.Notice("Will keep retrying, but it's possible something has gone wrong.")
				log.Notice(err.Error())
			}

			retries++
		}
	}

	return creds, err
}

func monitorOutput(podname string) ([]byte, error) {
	return RunCommand("oc", "logs", "-f", podname)
}

func buildExtractedCredentials(output []byte) (*ExtractedCredentials, error) {
	stillWaiting := strings.Contains(string(output), "ContainerCreating") ||
		strings.Contains(string(output), "NotFound")

	if stillWaiting {
		// Still waiting for container to come up
		return nil, errors.New(StillWaitingError)
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
			return nil, errors.New(str[startOffset:endIdx])
		}

		if strings.Contains(str, "image can't be pulled") {
			return nil, errors.New("image can't be pulled")
		} else if strings.Contains(str, "FAILED! =>") {
			return nil, errors.New("provision failed, INSERT MESSAGE HERE")
		} else {
			// Case 1, no creds found, no errors occurred
			return nil, nil
		}
	}

	decodedjson, err := base64.StdEncoding.DecodeString(str[startOffset:endIdx])
	if err != nil {
		return nil, err
	}

	decoded := make(map[string]string)
	json.Unmarshal(decodedjson, &decoded)
	// Case 2, creds successfully found and decoded
	return decoded, nil
}
