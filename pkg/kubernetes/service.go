package kubernetes

import (
	"github.com/apex/log"
	"github.com/kontrio/kappy/pkg/model"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/client-go/kubernetes"
)

func createService(serviceDefinition *model.ServiceDefinition, namespace string) corev1.Service {
	serviceSelector := make(map[string]string)
	serviceSelector["name"] = serviceDefinition.Name

	//TODO: hardcoded
	var port int32 = 80
	targetPort := 3000

	return corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      serviceDefinition.Name,
			Namespace: namespace,
		},
		Spec: corev1.ServiceSpec{
			Selector: serviceSelector,
			Ports: []corev1.ServicePort{
				corev1.ServicePort{
					Port:       port,
					TargetPort: intstr.FromInt(targetPort),
				},
			},
		},
	}
}

func CreateUpdateService(client *kubernetes.Clientset, serviceDefinition *model.ServiceDefinition, namespace string) error {
	service := createService(serviceDefinition, namespace)

	upsertCmd := UpsertCommand{
		Create: func() (err error) {
			_, err = client.CoreV1().Services(namespace).Create(&service)

			if err == nil {
				log.Infof("Created service %s", serviceDefinition.Name)
			}
			return
		},
		Update: func() (err error) {
			_, err = client.CoreV1().Services(namespace).Update(&service)

			if err == nil {
				log.Infof("Updated service %s", serviceDefinition.Name)
			}
			return
		},
	}

	return upsertCmd.Do()
}
