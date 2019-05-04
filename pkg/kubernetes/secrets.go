package kubernetes

import (
	"github.com/apex/log"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

func CreateSecretResource(secretRef, namespace string, secrets map[string]string, labels map[string]string) (*corev1.Secret, error) {
	// Alias to default only for logging purposes.
	ns := namespace
	if len(namespace) == 0 {
		ns = corev1.NamespaceDefault
	}

	log.Infof("Creating secrets '%s' in namespace: '%s'", secretRef, ns)

	encodedSecrets := convertToBytes(secrets)

	return &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      secretRef,
			Namespace: namespace,
			Labels:    labels,
		},
		Data: encodedSecrets,
	}, nil

}

func CreateSecret(client *kubernetes.Clientset, secretRef, namespace string, secrets map[string]string, labels map[string]string) error {
	resource, _ := CreateSecretResource(secretRef, namespace, secrets, labels)
	upsertCmd := UpsertCommand{
		Create: func() (err error) {
			_, err = client.CoreV1().Secrets(namespace).Create(resource)

			if err == nil {
				log.Infof("Successfully created secret %s in namespace %s", secretRef, namespace)
			}
			return
		},
		Update: func() (err error) {
			_, err = client.CoreV1().Secrets(namespace).Update(resource)

			if err == nil {
				log.Infof("Successfully updated secret %s in namespace %s", secretRef, namespace)
			}
			return
		},
	}

	return upsertCmd.Do()
}
