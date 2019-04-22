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
var serviceToBuild string

var buildCmd = &cobra.Command{
	Use:   "build [stackname] [servicename]",
	Short: "Build an application or a set of applications and push to docker repositories",
	Args: func(cmd *cobra.Command, args []string) (err error) {
		config, err = pkg.LoadConfig(&KappyFile)

		if err != nil {
			return
		}

		if len(args) < 1 {
			return fmt.Errorf("Requires [stackname] argument")
		}

		if config.GetStackByName(args[0]) == nil {
			return fmt.Errorf("Stack '%s' is not defined in the .kappy configuration", args[0])
		}

		serviceToBuild = ""
		if len(args) > 1 {
			serviceToBuild = args[1]
		}

		return
	},
	Run: func(cmd *cobra.Command, args []string) {
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

		buildRecords := getBuildableImagesForStack(config, version, serviceToBuild, stackName)

		for _, buildRecord := range buildRecords {
			err := docker.RunBuild(buildRecord.buildDef, buildRecord.extraTags)

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

		log.Infof("Use this version with \"kappy deploy %s --version %s\" to deploy this build", stackName, version)
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

func getBuildableImagesForStack(config *model.Config, version, serviceToBuild, stackName string) []buildRecord {
	imagesToBuild := []buildRecord{}

	stackDefinition := config.GetStackByName(stackName)

	if stackDefinition == nil {
		log.Warnf("Could not find a stack configured %s", stackName)
		return imagesToBuild
	}

	for serviceName, definition := range config.Services {

		if !kstrings.IsEmpty(&serviceToBuild) && serviceToBuild != serviceName {
			continue
		}

		for _, container := range definition.Containers {
			imageName := container.Image

			hasContainerConfig := false

			serviceConfig := stackDefinition.GetServiceConfig(serviceName)

			containerConfig := &model.ContainerConfig{}
			if serviceConfig != nil {
				containerConfig = serviceConfig.GetContainerConfigByName(container.Name)
				hasContainerConfig = containerConfig != nil
			}

			if ((hasContainerConfig && containerConfig.Build != nil) || container.Build != nil) && !kstrings.IsEmpty(&imageName) {
				buildDefinition := container.Build
				if hasContainerConfig {
					buildDefinition = model.MergeBuildDefinitions(container.Build, containerConfig.Build)
				}

				extraTags := []string{}

				var versionImageTag string

				if hasContainerConfig && containerConfig.Build != nil {
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
