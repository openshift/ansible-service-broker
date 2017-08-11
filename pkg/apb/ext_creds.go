package apb

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/openshift/ansible-service-broker/pkg/runtime"

	logging "github.com/op/go-logging"
)

// ExtractCredentials - Extract credentials from pod in a certain namespace.
func ExtractCredentials(
	podname string, namespace string, log *logging.Logger,
) (*ExtractedCredentials, error) {
	log.Debug("Calling monitorOutput on " + podname)
	bindOutput, err := monitorOutput(namespace, podname, log)
	if bindOutput == nil {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	return buildExtractedCredentials(bindOutput)
}

func monitorOutput(namespace string, podname string, log *logging.Logger) ([]byte, error) {
	// TODO: Error handling here
	// It would also be nice to gather the script output that exec runs
	// instead of only getting the credentials

	for r := 1; r <= apbWatchRetries; r++ {
		// err will be the return code from the exec command
		// Use the error code to deterine the state
		failedToExec := errors.New("exit status 1")
		credsNotAvailable := errors.New("exit status 2")

		output, err := runtime.RunCommand("oc", "exec", podname, gatherCredentialsCMD, "--namespace="+namespace)
		podCompleted := strings.Contains(string(output), "current phase is Succeeded") ||
			strings.Contains(string(output), "cannot exec into a container in a completed pod")

		if err == nil {
			log.Notice("[%s] Bind credentials found", podname)
			return output, nil
		} else if podCompleted && err.Error() == failedToExec.Error() {
			log.Notice("[%s] APB completed", podname)
			return nil, nil
		} else if err.Error() == failedToExec.Error() {
			log.Info(string(output))
			log.Warning("[%s] Retry attempt %d: Failed to exec into the container", podname, r)
		} else if err.Error() == credsNotAvailable.Error() {
			log.Info(string(output))
			log.Warning("[%s] Retry attempt %d: Bind credentials not availble yet", podname, r)
		} else {
			log.Info(string(output))
			log.Warning("[%s] Retry attempt %d: Failed to exec into the container", podname, r)
		}

		log.Warning("[%s] Retry attempt %d: exec into %s failed", podname, r, podname)
		time.Sleep(time.Duration(apbWatchInterval) * time.Second)
	}

	timeout := fmt.Sprintf("[%s] ExecTimeout: Failed to gather bind credentials after %d retries", podname, apbWatchRetries)
	return nil, errors.New(timeout)
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

func decodeOutput(output []byte) (map[string]interface{}, error) {
	str := string(output)

	decodedjson, err := base64.StdEncoding.DecodeString(str)
	if err != nil {
		return nil, err
	}

	decoded := make(map[string]interface{})
	json.Unmarshal(decodedjson, &decoded)
	return decoded, nil
}
