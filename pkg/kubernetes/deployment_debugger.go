package kubernetes

import (
	"context"
	"fmt"
	"github.com/apex/log"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/client-go/kubernetes"
)

func GetFailingPods(ctx context.Context, client *kubernetes.Clientset, namespace, deploymentName string) error {
	labelSet := labels.Set{
		"name": deploymentName,
	}
	list, err := client.CoreV1().Pods(namespace).List(metav1.ListOptions{LabelSelector: labelSet.AsSelector().String()})

	if err != nil {
		return fmt.Errorf("Failed to list pods: %s", err)
	}

	failedPods := map[string]*corev1.Pod{}
	unknownPods := map[string]*corev1.Pod{}

	for _, pod := range list.Items {
		podName := pod.ObjectMeta.Name
		log.Infof("Pod: %s Status: %s", podName, pod.Status.Phase)
		switch pod.Status.Phase {
		case corev1.PodRunning, corev1.PodSucceeded:
			continue
		case corev1.PodPending:
			unknownPods[podName] = &pod
		case corev1.PodFailed:
			failedPods[podName] = &pod
		default:
			unknownPods[podName] = &pod
		}
	}

	if len(failedPods) > 0 {
		// Debug
		log.Errorf("Failing pods: %d", len(failedPods))
	}

	return nil
}

func extractErrorsFromPodStatus(pod *corev1.Pod) error {
	if pod.Status.Phase == corev1.PodRunning || pod.Status.Phase == corev1.PodSucceeded {
		return nil
	}

	for _, containerStatus := range pod.Status.ContainerStatuses {
		containerReady := containerStatus.Ready

		if containerReady == true {
			continue
		}

	}

	return nil
}
