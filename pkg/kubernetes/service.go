package kubernetes

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/apex/log"
	"github.com/kontrio/kappy/pkg/model"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/client-go/kubernetes"
)

func createService(serviceDefinition *model.ServiceDefinition, namespace string) (*corev1.Service, error) {
	serviceSelector := make(map[string]string)
	serviceSelector["name"] = serviceDefinition.Name

	servicePorts := []corev1.ServicePort{}

	for _, mapping := range serviceDefinition.ServicePorts {
		mapping := strings.Split(mapping, ":")
		if len(mapping) != 2 {
			return nil, fmt.Errorf("Port mapping for service: %s is invalid '%s', expected `serviceport:containerport`", serviceDefinition.Name, mapping)
		}

		servicePort, errParseServicePort := strconv.ParseInt(mapping[0], 10, 32)

		if errParseServicePort != nil {
			return nil, fmt.Errorf("Could not parse service port: %s", errParseServicePort)
		}

		containerPort, errParseContainerPort := strconv.ParseInt(mapping[1], 10, 32)

		if errParseContainerPort != nil {
			return nil, fmt.Errorf("Could not parse container port: %s", errParseServicePort)
		}

		servicePorts = append(servicePorts, corev1.ServicePort{
			Port:       int32(servicePort),
			TargetPort: intstr.FromInt(int(containerPort)),
		})
	}

	if len(servicePorts) == 0 {
		//TODO: hardcoded defaults legacy
		servicePorts = append(servicePorts, corev1.ServicePort{
			Port:       80,
			TargetPort: intstr.FromInt(3000),
		})
	}

	return &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      serviceDefinition.Name,
			Namespace: namespace,
		},
		Spec: corev1.ServiceSpec{
			Selector: serviceSelector,
			Ports:    servicePorts,
		},
	}, nil
}

func CreateUpdateService(client *kubernetes.Clientset, serviceDefinition *model.ServiceDefinition, namespace string) error {
	service, errValid := createService(serviceDefinition, namespace)

	if errValid != nil {
		return errValid
	}

	upsertCmd := UpsertCommand{
		Create: func() (err error) {
			_, err = client.CoreV1().Services(namespace).Create(service)

			if err == nil {
				log.Infof("Created service %s", serviceDefinition.Name)
			} else {
				// We need to now get the current version from the Kube API in order to apply an update to that version
				result, errGet := getCurrentService(client, serviceDefinition.Name, namespace)

				if errGet != nil {
					return errGet
				}

				service.ObjectMeta.ResourceVersion = result.ObjectMeta.ResourceVersion
				service.Spec.ClusterIP = result.Spec.ClusterIP
			}

			return
		},
		Update: func() (err error) {

			_, err = client.CoreV1().Services(namespace).Update(service)

			if err == nil {
				log.Infof("Updated service %s", serviceDefinition.Name)
			}
			return
		},
	}

	return upsertCmd.Do()
}

func getCurrentService(client *kubernetes.Clientset, name, namespace string) (result *corev1.Service, err error) {
	result, err = client.CoreV1().Services(namespace).Get(name, metav1.GetOptions{ResourceVersion: ""})

	if err != nil {
		log.Errorf("Failed to get service %s in namespace %s: %s", name, namespace, err)
	}
	return
}
