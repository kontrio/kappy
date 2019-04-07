package kubernetes

import (
	"context"
	"fmt"
	"net/http"

	"github.com/ericchiang/k8s"
	"github.com/ericchiang/k8s/apis/extensions/v1beta1"
	metav1 "github.com/ericchiang/k8s/apis/meta/v1"
	"github.com/ericchiang/k8s/util/intstr"
	"github.com/kontr/kappy/pkg/model"
)

func createIngress(serviceConfig *model.ServiceConfig, serviceName, namespace string) v1beta1.Ingress {

	rules := []*v1beta1.IngressRule{}
	defaultPath := "/"
	var defaultPort int32 = 80
	certSecretName := fmt.Sprintf("%s-cert", serviceName)

	for _, ingress := range serviceConfig.Ingress {
		rules = append(rules, &v1beta1.IngressRule{
			Host: &ingress,
			IngressRuleValue: &v1beta1.IngressRuleValue{
				Http: &v1beta1.HTTPIngressRuleValue{
					Paths: []*v1beta1.HTTPIngressPath{
						&v1beta1.HTTPIngressPath{
							Path: &defaultPath,
							Backend: &v1beta1.IngressBackend{
								ServiceName: &serviceName,
								ServicePort: &intstr.IntOrString{
									IntVal: &defaultPort,
								},
							},
						},
					},
				},
			},
		})
	}

	return v1beta1.Ingress{
		Metadata: &metav1.ObjectMeta{
			Name:      &serviceName,
			Namespace: &namespace,
			Annotations: map[string]string{
				"certmanager.k8s.io/cluster-issuer":                   "cert-issuer",
				"traefik.ingress.kubernetes.io/frontend-entry-points": "https",
			},
		},
		Spec: &v1beta1.IngressSpec{
			Rules: rules,
			Tls: []*v1beta1.IngressTLS{
				&v1beta1.IngressTLS{
					Hosts:      serviceConfig.Ingress,
					SecretName: &certSecretName,
				},
			},
		},
	}
}

func CreateUpdateIngress(client *k8s.Client, serviceConfig *model.ServiceConfig, serviceName, namespace string) error {
	resource := createIngress(serviceConfig, serviceName, namespace)
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
