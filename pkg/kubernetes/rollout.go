package kubernetes

import (
	"context"
	"fmt"
	"github.com/apex/log"

	appsv1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/kubernetes"
	"k8s.io/kubernetes/pkg/kubectl/scheme"
)

func WatchDeployment(ctx context.Context, client *kubernetes.Clientset, namespace, deploymentName string) error {

	fieldSelector := fields.OneTermEqualSelector("metadata.name", deploymentName).String()

	watcher, err := client.AppsV1().Deployments(namespace).Watch(metav1.ListOptions{FieldSelector: fieldSelector})

	if err != nil {
		return err
	}

	defer watcher.Stop()

	for event := range watcher.ResultChan() {

		switch event.Type {
		case watch.Added, watch.Modified:
			deployment := &appsv1.Deployment{}
			err := scheme.Scheme.Convert(event.Object, deployment, nil)

			if err != nil {
				return fmt.Errorf("Failed to convert %T to %T: %v", event.Object, deployment, err)
			}

			status, done, errStatus := getRolloutStatus(deployment)

			if errStatus != nil {
				return errStatus
			}

			log.Infof("%s", status)

			if done {
				return nil
			}
		case watch.Deleted:
			return fmt.Errorf("deployment was deleted")
		default:
			return fmt.Errorf("Unexpected watch event for deployment from Kube API: %#v", event)
		}
	}

	return nil
}

const _deploymentProgressing = "Progressing"
const _deploymentReplicaFailure = "ReplicaFailure"
const _deploymentAvailable = "Available"
const _timedoutReason = "ProgressDeadlineExceeded"

func getRolloutStatus(deployment *appsv1.Deployment) (string, bool, error) {
	status := deployment.Status
	condition := getDeploymentCondition(&status, appsv1.DeploymentProgressing)
	name := deployment.ObjectMeta.Name

	if condition != nil && condition.Reason == _timedoutReason {
		return "", false, fmt.Errorf("deployment %q exceeded its progress deadline", name)
	}

	if status.UpdatedReplicas < *deployment.Spec.Replicas {
		return fmt.Sprintf("%q: %d/%d replicas updated", name, status.UpdatedReplicas, *deployment.Spec.Replicas), false, nil
	}

	if status.Replicas > status.UpdatedReplicas {
		return fmt.Sprintf("%q: %d/%d old replicas terminating", name, status.Replicas, status.UpdatedReplicas), false, nil
	}

	if status.AvailableReplicas < status.UpdatedReplicas {
		return fmt.Sprintf("%q: %d/%d updated replicas are available", name, status.AvailableReplicas, status.UpdatedReplicas), false, nil
	}

	return fmt.Sprintf("%q: %d/%d successfully rolled out", name, status.AvailableReplicas, status.UpdatedReplicas), true, nil
}

func getDeploymentCondition(deploymentStatus *appsv1.DeploymentStatus, conditionStatus appsv1.DeploymentConditionType) *appsv1.DeploymentCondition {
	for _, condition := range deploymentStatus.Conditions {
		log.Debugf("Condition type: if %s == %s", condition.Type, conditionStatus)
		if condition.Type == conditionStatus {
			return &condition
		}
	}
	return nil
}
