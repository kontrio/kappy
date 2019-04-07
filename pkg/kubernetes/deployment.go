package kubernetes

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/apex/log"
	"github.com/ericchiang/k8s"
	appsv1 "github.com/ericchiang/k8s/apis/apps/v1"
	corev1 "github.com/ericchiang/k8s/apis/core/v1"
	metav1 "github.com/ericchiang/k8s/apis/meta/v1"
	"github.com/kontr/kappy/pkg/model"
)

var _true bool = true
var _false bool = false

func canonicalImageName(imageName string, dockerRegistry string, deployVersion string, hasBuildSection bool) string {
	versionedImage := imageName

	if !hasBuildSection {
		return imageName
	}

	if !strings.Contains(imageName, "@sha256:") {
		versionedImage = fmt.Sprintf("%s:%s", imageName, deployVersion)
	}

	if len(dockerRegistry) == 0 {
		return versionedImage
	}

	return fmt.Sprintf("%s/%s", dockerRegistry, versionedImage)
}

func createDeploymentResource(serviceDef *model.ServiceDefinition, serviceConfig *model.ServiceConfig, namespace, deployVersion, dockerRegistry string) appsv1.Deployment {
	serviceName := serviceDef.Name

	if len(namespace) == 0 {
		namespace = "default"
	}

	containers := []*corev1.Container{}

	for _, container := range serviceDef.Containers {
		containerName := container.Name
		containerConfig := serviceConfig.GetContainerConfigByName(containerName)

		envVars := []*corev1.EnvVar{}

		if containerConfig != nil {
			for key, value := range containerConfig.Env {
				envVars = append(envVars, &corev1.EnvVar{
					Name:  &key,
					Value: &value,
				})
			}
		}

		log.Debugf("Configuring container: %s.%s", serviceName, containerName)

		defaultContainerPort := corev1.ContainerPort{
			ContainerPort: &container.ExposePort,
		}

		imageName := canonicalImageName(container.Image, dockerRegistry, deployVersion, container.Build != nil)

		secretName := fmt.Sprintf("%s-%s-secrets", serviceName, containerName)

		containers = append(containers, &corev1.Container{
			Name:    &container.Name,
			Image:   &imageName,
			Command: container.Command,
			Args:    container.Args,
			Ports:   []*corev1.ContainerPort{&defaultContainerPort},
			Env:     envVars,
			EnvFrom: []*corev1.EnvFromSource{
				&corev1.EnvFromSource{
					SecretRef: &corev1.SecretEnvSource{
						LocalObjectReference: &corev1.LocalObjectReference{
							Name: &secretName,
						},
						Optional: &_true,
					},
				},
			},
		})
	}

	deploymentSpec := appsv1.DeploymentSpec{
		Replicas: &serviceDef.Replicas,
		Selector: &metav1.LabelSelector{
			MatchLabels: map[string]string{
				"name": serviceDef.Name,
			},
		},
		Template: &corev1.PodTemplateSpec{
			Metadata: &metav1.ObjectMeta{
				Labels: map[string]string{
					"name": serviceDef.Name,
				},
			},
			Spec: &corev1.PodSpec{
				Containers: containers,
			},
		},
	}

	deployment := appsv1.Deployment{
		Metadata: &metav1.ObjectMeta{
			Name:      &serviceDef.Name,
			Namespace: &namespace,
		},
		Spec: &deploymentSpec,
	}

	return deployment
}

func CreateUpdateDeployment(client *k8s.Client, serviceDefinition *model.ServiceDefinition, serviceConfig *model.ServiceConfig, namespace, deployVersion, dockerRegistry string) error {
	deployment := createDeploymentResource(serviceDefinition, serviceConfig, namespace, deployVersion, dockerRegistry)
	err := client.Create(context.Background(), &deployment)

	if err != nil {
		if apiErr, ok := err.(*k8s.APIError); ok {
			if apiErr.Code == http.StatusConflict {
				err = client.Update(context.Background(), &deployment)
				if err != nil {
					return err
				}

				log.Infof("Updated deployment %s", serviceDefinition.Name)
			}
		}
		return err
	}

	log.Infof("Created deployment %s", serviceDefinition.Name)
	return nil
}
