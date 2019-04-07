package docker

import (
	"io"
	"os"
	"os/exec"

	"github.com/kr/pty"
)

func RunDocker(args []string) error {
	dockerCmd := exec.Command("docker", args...)

	outFile, err := pty.Start(dockerCmd)

	if err != nil {
		return err
	}

	io.Copy(os.Stdout, outFile)
	return nil
}
