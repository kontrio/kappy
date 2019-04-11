package kubernetes

import (
	"context"
	"net/http"

	"github.com/apex/log"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

func CreateNamespace(client *kubernetes.ClientSet, namespace string) error {
	namespaceResource := corev1.Namespace{
		ObjectMeta: &metav1.ObjectMeta{
			Name: namespace,
		},
		Spec: corev1.NamespaceSpec{},
	}

	upsertCommand := UpsertCommand{
		Create: func() (err error) {
			_, err = client.CoreV1().Namespaces().Create(&namespaceResource)

			if err == nil {
				log.Infof("Successfully created namespace %s in namespace %s", namespace)
			}
		},
		Update: func() (err error) {
			_, err = client.CoreV1().Namespaces().Update(&namespaceResource)

			if err == nil {
				log.Infof("Successfully updated namespace %s in namespace %s", namespace)
			}
		},
	}

	return upsertCommand.Do()
}
