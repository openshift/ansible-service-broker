package apb

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	logging "github.com/op/go-logging"
)

var stillWaitingError = "status: still waiting to start"
var timeoutFreq = 6    // Seconds
var totalTimeout = 900 // 15min

// ExtractCredentials - Extract credentials from pod in a certain namespace.
func ExtractCredentials(
	podname string, namespace string, log *logging.Logger,
) (*ExtractedCredentials, error) {
	credOut := make(chan []byte)

	log.Debug("Calling monitorOutput on " + podname)
	go monitorOutput(podname, credOut, log)

	var bindOutput []byte
	for msg := range credOut {
		bindOutput = msg
	}

	return buildExtractedCredentials(bindOutput)
}

func monitorOutput(podname string, mon chan []byte, log *logging.Logger) {
	// TODO: Error handling here
	// It would also be nice to gather the script output that exec runs
	// instead of only getting the credentials
	retries := 20

	for r := 1; r <= retries; r++ {
		output, _ := RunCommand("oc", "exec", podname, "broker-bind-creds")

		stillWaiting := strings.Contains(string(output), "ContainerCreating") ||
			strings.Contains(string(output), "NotFound") ||
			strings.Contains(string(output), "container not found")
		if stillWaiting {
			log.Warning("Retry attempt %d: Waiting for container to start", r)
		} else if strings.Contains(string(output), "BIND_CREDENTIALS") {
			mon <- output
			close(mon)
			log.Notice("Bind credentials found")
			return
		} else {
			log.Warning("Retry attempt %d: exec into %s failed", r, podname)
		}

		log.Debug(string(output))
		time.Sleep(time.Duration(r) * time.Second)

	}
	t := fmt.Sprintf("ExecTimeout: Failed to gather bind credentials after %d retries", retries)
	mon <- []byte(t)
	close(mon)
	return
}

func buildExtractedCredentials(output []byte) (*ExtractedCredentials, error) {
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
