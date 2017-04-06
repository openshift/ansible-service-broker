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

// HACK ALERT!
// A lot of the current approach to extracting credentials and monitoring
// output is *very* experimental and error prone. Entire approach is going
// to be thrown out and redone asap.

func extractCredentials(
	output []byte, log *logging.Logger,
) (*ExtractedCredentials, error) {
	log.Info("{%s}", string(output))

	log.Debug("Calling getPodName")
	podname, _ := getPodName(output, log)
	log.Debug("Calling monitorOutput on " + podname)
	bindout, _ := monitorOutput(podname)
	log.Info(string(bindout))

	var creds *ExtractedCredentials
	var err error
	creds, err = buildExtractedCredentials(bindout)

	if err != nil {
		// HACK: this is HORRIBLE. but there is definitely a time between a bind
		// and when the container is up.
		time.Sleep(5 * time.Second)
		log.Warning("Trying to monitor the output again on " + podname)
		bindout, _ = monitorOutput(podname)
		log.Info(string(bindout))
		creds, err = buildExtractedCredentials(bindout)
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
		return nil, errors.New("container creating, still waiting to start.")
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

		// Case 1, no creds found, no errors occurred
		return nil, nil
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
