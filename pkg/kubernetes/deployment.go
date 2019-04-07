package kubernetes

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/apex/log"
	"github.com/ericchiang/k8s"
	appsv1 "github.com/ericchiang/k8s/apis/apps/v1"
	corev1 "github.com/ericchiang/k8s/apis/core/v1"
	metav1 "github.com/ericchiang/k8s/apis/meta/v1"
	"github.com/kontrio/kappy/pkg/model"
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
			Labels: map[string]string{
				"name": serviceDef.Name,
			},
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

func WatchDeploymentRollout(client *k8s.Client, serviceName, namespace string) error {
	deployment := appsv1.Deployment{}

	labelSelector := new(k8s.LabelSelector)
	labelSelector.Eq("name", serviceName)

	ctx, cancel := context.WithTimeout(context.Background(), 120*time.Second)
	defer cancel()

	watcher, err := client.Watch(ctx, namespace, &deployment, labelSelector.Selector())
	if err != nil {
		return err
	}

	defer watcher.Close()

	for {
		deployment := new(appsv1.Deployment)
		eventType, errWatch := watcher.Next(deployment)

		log.Debugf("Watched: %s", eventType)

		if errWatch != nil {
			return errWatch
		}

		// Check failures to enter debug mode
		status := deployment.GetStatus()
		condition := getDeploymentCondition(status, _deploymentReplicaFailure)

		if condition != nil {
			log.Errorf("replica failure: %s", condition.GetMessage())
		}

		message, ok, errStatus := getRolloutStatus(deployment)

		log.Infof(message)

		if errStatus != nil {
			return errStatus
		}

		if ok {
			return nil
		}
	}
	return nil
}

var _deploymentProgressing string = "Progressing"
var _deploymentReplicaFailure string = "ReplicaFailure"
var _deploymentAvailable string = "Available"

var _timedoutReason string = "ProgressDeadlineExceeded"

func getRolloutStatus(deployment *appsv1.Deployment) (string, bool, error) {
	status := deployment.GetStatus()
	condition := getDeploymentCondition(status, _deploymentProgressing)
	name := *deployment.Metadata.Name

	if condition != nil && condition.GetReason() == _timedoutReason {
		return "", false, fmt.Errorf("deployment %q exceeded its progress deadline", name)
	}

	if status.GetUpdatedReplicas() < deployment.GetSpec().GetReplicas() {
		return fmt.Sprintf("Waiting for deployment %q rollout to finish: %d of %d new replicas have been updated...", name, status.GetUpdatedReplicas(), deployment.GetSpec().GetReplicas()), false, nil
	}

	if status.GetReplicas() > status.GetUpdatedReplicas() {
		return fmt.Sprintf("Waitinf for deployment %q rollout to finish: waiting for %d old replicas to terminate...", name, status.GetReplicas()-status.GetUpdatedReplicas()), false, nil
	}

	if status.GetAvailableReplicas() < status.GetUpdatedReplicas() {
		return fmt.Sprintf("Waiting for deployment %q rollout to finish: %d of %d updated replicas are available...", name, status.GetAvailableReplicas(), status.GetUpdatedReplicas()), false, nil
	}

	return fmt.Sprintf("deployment %q successfully rolled out", name), true, nil
}

func getDeploymentCondition(deploymentStatus *appsv1.DeploymentStatus, conditionStatus string) *appsv1.DeploymentCondition {
	for _, condition := range deploymentStatus.GetConditions() {
		log.Debugf("Condition type: if %s == %s", *condition.Type, conditionStatus)
		if *condition.Type == conditionStatus {
			return condition
		}
	}
	return nil
}
