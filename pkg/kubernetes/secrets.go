package kubernetes

import (
	"context"
	"net/http"

	"github.com/apex/log"
	"github.com/ericchiang/k8s"

	corev1 "github.com/ericchiang/k8s/apis/core/v1"
	metav1 "github.com/ericchiang/k8s/apis/meta/v1"
)

func CreateSecret(client *k8s.Client, secretRef, namespace string, secrets map[string]string, labels map[string]string) error {
	// Only alias to default for logging purposes, sending the empty string in the protobuf
	// defaults to default anyway and is more efficient
	ns := namespace
	if len(namespace) == 0 {
		ns = "default"
	}

	log.Infof("Creating secrets '%s' in namespace: '%s'", secretRef, ns)

	secret := corev1.Secret{
		Metadata: &metav1.ObjectMeta{
			Name:      &secretRef,
			Namespace: &namespace,
			Labels:    labels,
		},
		StringData: secrets,
	}

	return CreateUpdateK8sResource(client, context.Background(), &secret)
}

func CreateUpdateK8sResource(client *k8s.Client, ctx context.Context, resource *corev1.Secret) error {
	err := client.Create(ctx, resource)
	if err != nil {
		if apiErr, ok := err.(*k8s.APIError); ok {
			if apiErr.Code == http.StatusConflict {
				err = client.Update(ctx, resource)
				if err != nil {
					return err
				}
			}
		}
		return err
	}
	return nil
}
