package apb

import "os/exec"

var HardcodedClusterConfig = ClusterConfig{
	//Target:   "cap.example.com:8443",
	Target:   "10.1.2.2:8443",
	User:     "admin",
	Password: "admin",
}

// HACK: really need a better way to do docker run
func runCommand(cmd string, args ...string) ([]byte, error) {
	output, err := exec.Command(cmd, args...).CombinedOutput()
	return output, err
}
