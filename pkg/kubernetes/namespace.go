package kubernetes

import (
	"context"
	"net/http"

	"github.com/ericchiang/k8s"
	corev1 "github.com/ericchiang/k8s/apis/core/v1"
	metav1 "github.com/ericchiang/k8s/apis/meta/v1"
)

func CreateNamespace(client *k8s.Client, namespace string) error {
	namespaceResource := corev1.Namespace{
		Metadata: &metav1.ObjectMeta{
			Name: &namespace,
		},
		Spec: &corev1.NamespaceSpec{},
	}

	err := client.Create(context.Background(), &namespaceResource)

	if err != nil {
		if apiErr, ok := err.(*k8s.APIError); ok {
			if apiErr.Code == http.StatusConflict {
				return nil
			} else {
				return err
			}
		}
	}

	return nil
}
