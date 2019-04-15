package docker

import (
	"fmt"
	"io"
	"os"
	"os/exec"

	"github.com/kr/pty"
)

func RunDocker(args []string, env map[string]string) error {
	dockerCmd := exec.Command("docker", args...)

	dockerCmd.Env = keyValueStrings(env)

	outFile, err := pty.Start(dockerCmd)

	if err != nil {
		return err
	}

	io.Copy(os.Stdout, outFile)
	return nil
}

func keyValueStrings(env map[string]string) []string {
	keyVals := []string{}
	for key, value := range env {
		keyVals = append(keyVals, fmt.Sprintf("%s=%s", key, value))
	}

	return keyVals
}
