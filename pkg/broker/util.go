package broker

import (
	"fmt"
	"github.com/fusor/ansible-service-broker/pkg/ansibleapp"
	"github.com/pborman/uuid"
	"os"
	"os/exec"
	"path"
)

func RunCommand(bin string, args ...string) {
	cmd := exec.Command(bin, args...) //.CombinedOutput()
	cmd.Stdin = os.Stdin
	cmd.Stdin = os.Stdout
	cmd.Stdin = os.Stderr
	err := cmd.Run()

	if err != nil {
		fmt.Println(err)
	}

	return
}

func ProjectRoot() string {
	gopath := os.Getenv("GOPATH")
	rootPath := path.Join(gopath, "src", "github.com", "fusor",
		"ansible-service-broker")
	return rootPath
}

// TODO: This is going to have to be expanded much more to support things like
// parameters (will need to get passed through as metadata
func SpecToService(spec *ansibleapp.Spec) Service {
	parameterDescriptors := make(map[string]interface{})
	parameterDescriptors["parameters"] = spec.Parameters

	return Service{
		ID:          uuid.Parse(spec.Id),
		Name:        spec.Name,
		Description: spec.Description,
		Bindable:    spec.Bindable,
		Plans:       plans, // HACK; it's still unclear how plans are relevant to us
		Metadata:    parameterDescriptors,
	}
}
