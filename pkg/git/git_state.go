package git

import (
	"os"
	"os/exec"
	"strings"
)

var ciEnvVars = []string{
	"CI_COMMIT_SHA", // gitlab
	"TRAVIS_COMMIT", // travis
}

func GetVersionInfo() (string, error) {

	for _, envVar := range ciEnvVars {
		fromEnv, isSet := os.LookupEnv(envVar)
		if isSet {
			return fromEnv[:8], nil
		}
	}

	return getGitCommitSha()
}

func getGitCommitSha() (string, error) {
	output, err := exec.Command("git", "rev-parse", "--short", "HEAD").Output()

	if err != nil {
		return "", err
	}

	return strings.TrimSpace(string(output)), err
}
