package cmd

import (
	"fmt"
	"os"

	"github.com/apex/log"
	"github.com/kontrio/kappy/pkg"
	"github.com/kontrio/kappy/pkg/awsutil"
	"github.com/kontrio/kappy/pkg/docker"
	"github.com/kontrio/kappy/pkg/git"
	"github.com/kontrio/kappy/pkg/kstrings"
	"github.com/kontrio/kappy/pkg/model"
	"github.com/spf13/cobra"
)

var ShouldPush bool = false

var buildCmd = &cobra.Command{
	Use:   "build [stackname]",
	Short: "Build an application or a set of applications and push to docker repositories",
	Args:  cobra.ArbitraryArgs,
	Run: func(cmd *cobra.Command, args []string) {
		log.Debug("Loading config file")
		config, err := pkg.LoadConfig()

		if err != nil {
			log.Errorf("Failed to load config file %s", err)
			os.Exit(1)
		}

		if len(args) == 0 {
			log.Errorf("Expected stackname argument")
			os.Exit(1)
			return
		}

		stackName := args[0]

		log.Debugf("Using docker_registry: %s", config.DockerRegistry)

		if ShouldPush && awsutil.IsEcrRegistry(config.DockerRegistry) {
			log.Debugf("Logging into AWS ECR registry..")
			errEcrLogin := awsutil.DoEcrLogin(&config.DockerRegistry)

			if errEcrLogin != nil {
				log.Errorf("Failed to login to ECR registry: %s - %s", config.DockerRegistry, errEcrLogin)
				os.Exit(1)
				return
			}
		}

		version, errVersion := git.GetVersionInfo()

		if errVersion != nil {
			log.Errorf("Failed to find a 'git' commit to tag builds with: %s", errVersion)
			os.Exit(1)
			return
		}

		buildRecords := getBuildableImagesForStack(config, version, stackName)

		for _, buildRecord := range buildRecords {
			err = docker.RunBuild(buildRecord.buildDef, buildRecord.extraTags)

			if err != nil {
				log.Errorf("Failed to build image: %s - %s", buildRecord.imageName, err)
				os.Exit(1)
			}

			if ShouldPush {
				errPush := docker.PushImage(buildRecord.extraTags)

				if errPush != nil {
					log.Errorf("Failed to push image %s - %s", buildRecord.imageName, errPush)
				}
			}
		}

		log.Infof("Use this version with \"kappy deploy [stackname] --version %s\" to deploy this build", version)
		log.Infof("Versioned: '%s'", version)
	},
}

func initBuildCmd() {
	buildCmd.Flags().BoolVarP(&ShouldPush, "push", "p", false, "push built images immediately")
}

type buildRecord struct {
	buildDef  *model.BuildDefinition
	extraTags []string
	imageName string
}

func getBuildableImagesForStack(config *model.Config, version, stackName string) []buildRecord {
	imagesToBuild := []buildRecord{}

	stackDefinition := config.GetStackByName(stackName)

	if stackDefinition == nil {
		log.Warnf("Could not find a stack configured %s", stackName)
		return imagesToBuild
	}

	for serviceName, definition := range config.Services {
		for _, container := range definition.Containers {
			containerConfig := stackDefinition.GetServiceConfig(serviceName).GetContainerConfigByName(container.Name)
			imageName := container.Image

			if (containerConfig.Build != nil || container.Build != nil) && !kstrings.IsEmpty(&imageName) {
				buildDefinition := model.MergeBuildDefinitions(container.Build, containerConfig.Build)
				extraTags := []string{}

				var versionImageTag string

				if containerConfig.Build != nil {
					versionImageTag = fmt.Sprintf("%s-%s", version, stackName)
				} else {
					versionImageTag = fmt.Sprintf("%s", version)
				}

				// If a specific build config exists for a stack, then we append the stackname to the built image
				extraTags = append(extraTags, fmt.Sprintf("%s/%s:%s", config.DockerRegistry, imageName, versionImageTag))

				for _, tag := range container.Build.Tags {
					extraTags = append(extraTags, fmt.Sprintf("%s/%s", config.DockerRegistry, tag))
				}

				imagesToBuild = append(imagesToBuild, buildRecord{
					imageName: imageName,
					buildDef:  buildDefinition,
					extraTags: extraTags,
				})
			}
		}
	}
	return imagesToBuild
}

func getBuildableImages(config *model.Config, version string) []buildRecord {
	imagesToBuild := []buildRecord{}

	for _, definition := range config.Services {
		for _, container := range definition.Containers {
			imageName := container.Image

			if container.Build != nil && len(imageName) > 0 {
				extraTags := []string{}

				extraTags = append(extraTags, fmt.Sprintf("%s/%s:%s", config.DockerRegistry, imageName, version))

				for _, tag := range container.Build.Tags {
					extraTags = append(extraTags, fmt.Sprintf("%s/%s", config.DockerRegistry, tag))
				}

				imagesToBuild = append(imagesToBuild, buildRecord{
					imageName: imageName,
					buildDef:  container.Build,
					extraTags: extraTags,
				})
			}
		}
	}

	return imagesToBuild
}
