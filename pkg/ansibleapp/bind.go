package ansibleapp

import (
	b64 "encoding/base64"
	"encoding/json"
	"fmt"
	"github.com/op/go-logging"
	"strings"
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

	log.Info(string(output))

	return buildBindData(output)
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

	decodedjson, err := b64.StdEncoding.DecodeString(str[startOffset:endIdx])
	if err != nil {
		return nil, err
	}

	decoded := make(map[string]string)
	json.Unmarshal(decodedjson, &decoded)
	return decoded, nil
}
