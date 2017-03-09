package ansibleapp

import (
	b64 "encoding/base64"
	"encoding/json"
	"github.com/op/go-logging"
	"strings"
)

// TODO: Figure out the right way to allow ansibleapp to log
// It's passed in here, but that's a hard coupling point to
// github.com/op/go-logging, which is used all over the broker
// Maybe ansibleapp defines its own interface and accepts that optionally
// Little looser, but still not great
func Bind(
	parameters *Parameters,
	clusterConfig ClusterConfig, log *logging.Logger,
) (*BindData, error) {
	log.Notice("============================================================")
	log.Notice("                       BINDING                              ")
	log.Notice("============================================================")
	/*
		log.Notice(fmt.Sprintf("Spec.Id: %s", spec.Id))
		log.Notice(fmt.Sprintf("Spec.Name: %s", spec.Name))
		log.Notice(fmt.Sprintf("Spec.Description: %s", spec.Description))
		log.Notice(fmt.Sprintf("Parameters: %v", parameters))
		log.Notice("============================================================")

		var client *Client
		var err error

		if client, err = NewClient(log); err != nil {
			return err
		}

		if err = client.PullImage(spec.Name); err != nil {
			return err
		}

		// HACK: Cluster config needs to come in from the broker. For now, hardcode it
		output, err := client.RunImage("bind", clusterConfig, spec, parameters)

		if err != nil {
			log.Error("Problem running image")
			return err
		}

		log.Info(string(output))
	*/

	/*
		we're going to have to parse the output from the run command to create the
		BindData.
	*/
	output := []byte(`
Login failed (401 Unauthorized)

PLAY [all] *********************************************************************

TASK [setup] *******************************************************************
ok: [localhost]

TASK [Bind] ********************************************************************
changed: [localhost]

TASK [debug] *******************************************************************
ok: [localhost] => {
    "msg": "<BIND_CREDENTIALS>eyJkYiI6ICJmdXNvcl9ndWVzdGJvb2tfZGIiLCAidXNlciI6ICJkdWRlcl90d28iLCAicGFzcyI6ICJkb2c4dHdvIn0=</BIND_CREDENTIALS>"
}

PLAY RECAP *********************************************************************
localhost                  : ok=3    changed=1    unreachable=0    failed=0   
`)
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
