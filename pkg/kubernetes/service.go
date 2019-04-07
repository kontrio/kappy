package kubernetes

import (
	"context"
	"net/http"

	"github.com/ericchiang/k8s"
	corev1 "github.com/ericchiang/k8s/apis/core/v1"
	metav1 "github.com/ericchiang/k8s/apis/meta/v1"
	"github.com/ericchiang/k8s/util/intstr"
	"github.com/kontrio/kappy/pkg/model"
)

func createService(serviceDefinition *model.ServiceDefinition, namespace string) corev1.Service {
	serviceSelector := make(map[string]string)
	serviceSelector["name"] = serviceDefinition.Name
	serviceSelector["namespace"] = namespace

	//TODO: hardcoded
	var port int32 = 80
	var targetPort int32 = 3000

	return corev1.Service{
		Metadata: &metav1.ObjectMeta{
			Name:      &serviceDefinition.Name,
			Namespace: &namespace,
		},
		Spec: &corev1.ServiceSpec{
			Selector: serviceSelector,
			Ports: []*corev1.ServicePort{
				&corev1.ServicePort{
					Port: &port,
					TargetPort: &intstr.IntOrString{
						IntVal: &targetPort,
					},
				},
			},
		},
	}
}

func CreateUpdateService(client *k8s.Client, serviceDefinition *model.ServiceDefinition, namespace string) error {
	resource := createService(serviceDefinition, namespace)
	err := client.Create(context.Background(), &resource)
	if err != nil {
		if apiErr, ok := err.(*k8s.APIError); ok {
			if apiErr.Code == http.StatusConflict {
				err = client.Update(context.Background(), &resource)
				if err != nil {
					return err
				}
			}
		}
		return err
	}
	return nil
}
