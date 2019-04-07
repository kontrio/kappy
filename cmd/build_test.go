package cmd

import (
	"github.com/kontrio/kappy/pkg"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestGetBuildableImagesFromConfig(t *testing.T) {
	configOneBuildableImage := pkg.Config{
		DockerRegistry: "test.example.com",
		Services: map[string]pkg.ServiceDefinition{
			"api_a": pkg.ServiceDefinition{
				Replicas: 1,
				Containers: []pkg.ContainerDefinition{
					pkg.ContainerDefinition{
						Name: "default",
						Build: &pkg.BuildDefinition{
							Context:    ".",
							Dockerfile: "Dockerfile",
						},
						Image: "my/imagename",
					},
				},
			},
		},
	}

	buildRecords := getBuildableImages(&configOneBuildableImage)

	assert.Equal(t, 1, len(buildRecords))
	assert.Equal(t, ".", buildRecords[0].buildDef.Context)
	assert.Equal(t, "my/imagename", buildRecords[0].imageName)
	assert.Equal(t, "test.example.com", buildRecords[0].dockerRegistry)
}
