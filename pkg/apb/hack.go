package apb

import (
	"bytes"
	"os/exec"
	"syscall"
)

var HardcodedClusterConfig = ClusterConfig{
	//Target:   "cap.example.com:8443",
	Target:   "10.1.2.2:8443",
	User:     "admin",
	Password: "admin",
}

// HACK: really need a better way to do docker run
func RunCommand(cmd string, args ...string) ([]byte, error) {
	output, err := exec.Command(cmd, args...).CombinedOutput()
	return output, err
}

func RunCommandWithExitCode(name string, args ...string) (stdout string, stderr string, exitCode int) {
	var outbuf, errbuf bytes.Buffer
	defaultFailedCode := 1
	cmd := exec.Command(name, args...)
	cmd.Stdout = &outbuf
	cmd.Stderr = &errbuf

	err := cmd.Run()
	stdout = outbuf.String()
	stderr = errbuf.String()

	if err != nil {
		if exitError, ok := err.(*exec.ExitError); ok {
			ws := exitError.Sys().(syscall.WaitStatus)
			exitCode = ws.ExitStatus()
		} else {
			exitCode = defaultFailedCode
			if stderr == "" {
				stderr = err.Error()
			}
		}
	} else {
		ws := cmd.ProcessState.Sys().(syscall.WaitStatus)
		exitCode = ws.ExitStatus()
	}
	return
}
