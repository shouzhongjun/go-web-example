package pythonrunner

import (
	"github.com/duke-git/lancet/v2/system"
	"os"
	"os/exec"
)

func RunPythonScript(scriptPath string, args ...string) (string, error) {
	file, err := os.OpenFile(scriptPath, os.O_RDONLY, 0666)
	if err != nil {
		return "", err
	}
	stdout, stderr, err := system.ExecCommand("python3", func(cmd *exec.Cmd) {
		cmd.Stdin = file
		cmd.Args = append(cmd.Args, args...)

	})
	if err != nil {
		return "", err
	}
	if len(stderr) > 0 {
		return "", err
	}
	return stdout, nil
}
