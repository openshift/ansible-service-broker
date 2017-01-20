package ansibleapp

import (
	"fmt"
	"os"
	"os/exec"
	"path"
)

const RUN_SCRIPT_NAME = "hello-ansibleapp.sh"

func ProvisionHelloAnsibleApp() {
	fmt.Println("Provisioning ansible app...")
	runAction("provision")
}

func DeprovisionHelloAnsibleApp() {
	fmt.Println("Deprovisioning ansible app...")
	runAction("deprovision")
}

func runAction(action string) {
	runScript := path.Join(ProjectRoot(), RUN_SCRIPT_NAME)
	fmt.Println("Running:")
	fmt.Println(runScript)
	fmt.Println(action)
	runCommand(runScript, action)
}

func runCommand(bin string, args ...string) {
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
