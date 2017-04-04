package apb

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"regexp"
	"strings"
	"time"

	logging "github.com/op/go-logging"
)

// TODO: Figure out the right way to allow apb to log
// It's passed in here, but that's a hard coupling point to
// github.com/op/go-logging, which is used all over the broker
// Maybe apb defines its own interface and accepts that optionally
// Little looser, but still not great
func Bind(
	instance *ServiceInstance,
	parameters *Parameters,
	clusterConfig ClusterConfig, log *logging.Logger,
) (*BindData, error) {
	log.Notice("============================================================")
	log.Notice("                       BINDING                              ")
	log.Notice("============================================================")
	log.Notice(fmt.Sprintf("Parameters: %v", parameters))
	log.Notice("============================================================")

	var client *Client
	var err error

	if client, err = NewClient(log); err != nil {
		return nil, err
	}

	if err = client.PullImage(instance.Spec.Name); err != nil {
		return nil, err
	}

	output, err := client.RunImage("bind", clusterConfig, instance.Spec, parameters)

	if err != nil {
		log.Error("Problem running image", err)
		return nil, err
	}

	log.Info("{" + string(output) + "}")

	log.Debug("Calling getPodName")
	podname, _ := getPodName(output, log)
	log.Debug("Calling monitorOutput on " + podname)

	out := make(chan []byte)
	go func() {
		monitorOutput(podname, out, log)
	}()

	var bindout []byte
	for msg := range out {
		log.Info(string(msg))
		bindout = msg
	}

	return buildBindData(bindout)
}

// HACK: this really is a crappy way of getting output
func getPodName(output []byte, log *logging.Logger) (string, error) {
	r, err := regexp.Compile(`^pod[ \"]*(.*?)[ \"]*created`)
	if err != nil {
		return "", err
	}

	podname := r.FindStringSubmatch(string(output))

	if log != nil {
		log.Debug(fmt.Sprintf("%v", podname))
		log.Debug(fmt.Sprintf("%d", (len(podname) - 1)))
	}

	return podname[len(podname)-1], nil
}

func monitorOutput(podname string, mon chan []byte, log *logging.Logger) {
	// TODO: Error handling here
	// It would also be nice to gather the script output that exec runs
	// instead of only getting the credentials
	retries := 5
	for r := 1; r <= retries; r++ {
		output, _ := runCommand("oc", "exec", podname, "broker-bind-creds")

		if strings.Contains(string(output), "BIND_CREDENTIALS") {
			mon <- output
			close(mon)
			return
		}

		time.Sleep(time.Duration(r) * time.Second)
		log.Warning("Retry attempt %d: exec into %s failed", r, podname)
	}
	t := fmt.Sprintf("ExecTimeout: Failed to gather bind credentials after %d retries", retries)
	mon <- []byte(t)
	close(mon)
	return
}

func buildBindData(output []byte) (*BindData, error) {
	// parse the output
	result, err := decodeOutput(output)
	if err != nil {
		return nil, err
	}

	fu := make(map[string]interface{})
	for k, v := range result {
		fu[k] = v
	}

	return &BindData{Credentials: fu}, nil
}

func decodeOutput(output []byte) (map[string]string, error) {
	str := string(output)
	startIdx := strings.Index(str, "<BIND_CREDENTIALS>")
	startOffset := startIdx + len("<BIND_CREDENTIALS>")
	endIdx := strings.Index(str, "</BIND_CREDENTIALS>")

	if startIdx < 0 || endIdx < 0 {
		// look for BIND_ERROR
		startIdx = strings.Index(str, "<BIND_ERROR>")
		startOffset := startIdx + len("<BIND_ERROR>")
		endIdx := strings.Index(str, "</BIND_ERROR>")
		if startIdx > -1 && endIdx > -1 {
			return nil, errors.New(str[startOffset:endIdx])
		}
		return nil, errors.New("Unable to parse output")
	}

	decodedjson, err := base64.StdEncoding.DecodeString(str[startOffset:endIdx])
	if err != nil {
		return nil, err
	}

	decoded := make(map[string]string)
	json.Unmarshal(decodedjson, &decoded)
	return decoded, nil
}
