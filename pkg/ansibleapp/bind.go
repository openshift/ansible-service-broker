package ansibleapp

import (
	b64 "encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"regexp"
	"strings"
	"time"
)

// TODO: Figure out the right way to allow ansibleapp to log
// It's passed in here, but that's a hard coupling point to
// github.com/op/go-logging, which is used all over the broker
// Maybe ansibleapp defines its own interface and accepts that optionally
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
	bindout, _ := monitorOutput(podname)
	log.Info(string(bindout))

	// HACK ALERT!
	bd, err := buildBindData(bindout)
	if err != nil {
		// HACK: this is HORRIBLE. but there is definitely a time between a bind
		// and when the container is up.
		time.Sleep(5 * time.Second)
		log.Warning("Trying to monitor the output again on " + podname)
		bindout, _ = monitorOutput(podname)
		log.Info(string(bindout))
		return buildBindData(bindout)
	}

	return bd, err
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

func monitorOutput(podname string) ([]byte, error) {
	return runCommand("oc", "logs", "-f", podname)
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

	decodedjson, err := b64.StdEncoding.DecodeString(str[startOffset:endIdx])
	if err != nil {
		return nil, err
	}

	decoded := make(map[string]string)
	json.Unmarshal(decodedjson, &decoded)
	return decoded, nil
}
