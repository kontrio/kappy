package kubernetes

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/apex/log"
	"github.com/kontrio/kappy/pkg/kstrings"
	"github.com/kontrio/kappy/pkg/model"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

var _true bool = true
var _false bool = false

func canonicalImageName(imageName string, dockerRegistry string, deployVersion string, namespace string, hasBuildSection bool, hasConfigBuildSection bool) string {
	versionedImage := imageName

	if !hasBuildSection && !hasConfigBuildSection {
		return imageName
	}

	if !strings.Contains(imageName, "@sha256:") {
		if hasConfigBuildSection {
			versionedImage = fmt.Sprintf("%s:%s-%s", imageName, deployVersion, namespace)
		} else {
			versionedImage = fmt.Sprintf("%s:%s", imageName, deployVersion)
		}
	}

	if kstrings.IsEmpty(&dockerRegistry) {
		return versionedImage
	}

	return fmt.Sprintf("%s/%s", dockerRegistry, versionedImage)
}

func createDeploymentResource(serviceDef *model.ServiceDefinition, serviceConfig *model.ServiceConfig, namespace, deployVersion, dockerRegistry string) appsv1.Deployment {
	serviceName := serviceDef.Name

	if len(namespace) == 0 {
		namespace = "default"
	}

	containers := []corev1.Container{}

	for _, container := range serviceDef.Containers {
		containerName := container.Name
		containerConfig := serviceConfig.GetContainerConfigByName(containerName)

		envVars := []corev1.EnvVar{}

		hasContainerConfig := containerConfig != nil
		hasContainerConfigBuildSection := hasContainerConfig && containerConfig.Build != nil

		explicitPort := false

		if hasContainerConfig {
			for key, value := range containerConfig.Env {
				if key == "PORT" {
					explicitPort = true
				}

				envVars = append(envVars, corev1.EnvVar{
					Name:  key,
					Value: value,
				})
			}
		}

		log.Debugf("Configuring container: %s.%s", serviceName, containerName)

		defaultContainerPort := corev1.ContainerPort{
			ContainerPort: container.ExposePort,
		}

		if !explicitPort {
			envVars = append(envVars, corev1.EnvVar{
				Name:  "PORT",
				Value: strconv.Itoa(int(container.ExposePort)),
			})
		}

		imageName := canonicalImageName(container.Image, dockerRegistry, deployVersion, namespace, container.Build != nil, hasContainerConfigBuildSection)

		secretName := fmt.Sprintf("%s-%s-secrets", serviceName, containerName)

		containers = append(containers, corev1.Container{
			Name:    container.Name,
			Image:   imageName,
			Command: container.Command,
			Args:    container.Args,
			Ports:   []corev1.ContainerPort{defaultContainerPort},
			Env:     envVars,
			EnvFrom: []corev1.EnvFromSource{
				corev1.EnvFromSource{
					SecretRef: &corev1.SecretEnvSource{
						LocalObjectReference: corev1.LocalObjectReference{
							Name: secretName,
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
		Template: corev1.PodTemplateSpec{
			ObjectMeta: metav1.ObjectMeta{
				Name:      serviceDef.Name,
				Namespace: namespace,
				Labels: map[string]string{
					"name":          serviceDef.Name,
					"kappy.managed": serviceDef.Name,
				},
			},
			Spec: corev1.PodSpec{
				Containers: containers,
			},
		},
	}

	deployment := appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      serviceDef.Name,
			Namespace: namespace,
			Labels: map[string]string{
				"name":          serviceDef.Name,
				"kappy.managed": serviceDef.Name,
			},
		},
		Spec: deploymentSpec,
	}

	return deployment
}

func UpsertDeployment(client *kubernetes.Clientset, serviceDefinition *model.ServiceDefinition, serviceConfig *model.ServiceConfig, namespace, deployVersion, dockerRegistry string) error {
	deployment := createDeploymentResource(serviceDefinition, serviceConfig, namespace, deployVersion, dockerRegistry)

	upsert := UpsertCommand{
		Create: func() (err error) {
			_, err = client.AppsV1().Deployments(namespace).Create(&deployment)
			if err == nil {
				log.Infof("Created deployment %s", serviceDefinition.Name)
			}

			return
		},
		Update: func() (err error) {
			_, err = client.AppsV1().Deployments(namespace).Update(&deployment)
			if err == nil {
				log.Infof("Updated deployment %s", serviceDefinition.Name)
			}
			return
		},
	}

	return upsert.Do()
}
