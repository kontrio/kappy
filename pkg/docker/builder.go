package docker

import (
	"fmt"

	"github.com/apex/log"
	"github.com/kontrio/kappy/pkg/model"
)

func RunBuild(definition *model.BuildDefinition, extraTags []string) error {
	var dockerArgs = []string{"build"}

	dockerArgs = append(dockerArgs, "--file", definition.Dockerfile)

	for _, cacheFrom := range definition.CacheFrom {
		dockerArgs = append(dockerArgs, "--cache-from", cacheFrom)
	}

	for _, tag := range definition.Tags {
		dockerArgs = append(dockerArgs, "--tag", tag)
	}

	for _, tag := range extraTags {
		dockerArgs = append(dockerArgs, "--tag", tag)
	}

	if len(definition.ShmSize) > 0 {
		dockerArgs = append(dockerArgs, "--shm-size", definition.ShmSize)
	}

	for key, arg := range definition.BuildArgs {
		buildArg := fmt.Sprintf("%s=%s", key, arg)
		dockerArgs = append(dockerArgs, "--build-arg", buildArg)
	}

	for _, label := range definition.Labels {
		dockerArgs = append(dockerArgs, "--label", label)
	}

	if len(definition.Target) > 0 {
		dockerArgs = append(dockerArgs, "--target", definition.Target)
	}

	dockerArgs = append(dockerArgs, definition.Context)
	return RunDocker(dockerArgs)
}

func PushImage(tags []string) error {
	var lastError error = nil

	for _, tag := range tags {
		err := RunDocker([]string{"push", tag})

		if err != nil {
			log.Errorf("Failed to push %s - %s", tag, err)
			lastError = err
		}
	}

	return lastError
}
