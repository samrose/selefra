package utils

import (
	"bytes"
	"os/exec"
)

func RunCommand(command string, args ...string) (stdout string, stderr string, err error) {

	stdOutBuff := bytes.Buffer{}
	stdErrBuff := bytes.Buffer{}

	cmd := exec.Command(command, args...)
	cmd.Stdout = &stdOutBuff
	cmd.Stderr = &stdErrBuff

	err = cmd.Run()

	stdout = stdOutBuff.String()
	stderr = stdErrBuff.String()
	//
	//if err != nil {
	//	diagnostics.AddErrorMsg("Run command %s error, error msg = %s, stdout = %s, stderr = %s", command, err.Error(), stdout, stderr)
	//}
	return
}
