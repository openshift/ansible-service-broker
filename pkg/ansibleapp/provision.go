package ansibleapp

import (
	"fmt"
	"github.com/op/go-logging"
)

// TODO: Figure out the right way to allow ansibleapp to log
// It's passed in here, but that's a hard coupling point to
// github.com/op/go-logging, which is used all over the broker
// Maybe ansibleapp defines its own interface and accepts that optionally
// Little looser, but still not great
func Provision(spec *Spec, parameters *Parameters, log *logging.Logger) error {
	log.Notice("============================================================")
	log.Notice("                       PROVISIONING                         ")
	log.Notice("============================================================")
	log.Notice(fmt.Sprintf("Spec.Id: %s", spec.Id))
	log.Notice(fmt.Sprintf("Spec.Name: %s", spec.Name))
	log.Notice(fmt.Sprintf("Spec.Description: %s", spec.Description))
	log.Notice(fmt.Sprintf("Parameters: %v", parameters))
	log.Notice("============================================================")

	return nil
}
