package kubernetes

import (
	"github.com/apex/log"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

func CreateNSResource(namespace string) (*corev1.Namespace, error) {
	return &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: namespace,
		},
		Spec: corev1.NamespaceSpec{},
	}, nil
}

func CreateNamespace(client *kubernetes.Clientset, namespace string) error {
	namespaceResource, _ := CreateNSResource(namespace)

	upsertCommand := UpsertCommand{
		Create: func() (err error) {
			_, err = client.CoreV1().Namespaces().Create(namespaceResource)

			if err == nil {
				log.Infof("Successfully created namespace %s", namespace)
			}
			return
		},
		Update: func() (err error) {
			_, err = client.CoreV1().Namespaces().Update(namespaceResource)

			if err == nil {
				log.Infof("Successfully updated namespace %s", namespace)
			}
			return
		},
	}

	return upsertCommand.Do()
}
