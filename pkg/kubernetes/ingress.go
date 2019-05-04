package kubernetes

import (
	"fmt"

	"github.com/apex/log"
	"github.com/kontrio/kappy/pkg/model"
	netv1beta1 "k8s.io/api/extensions/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/client-go/kubernetes"
)

func CreateIngressResource(serviceConfig *model.ServiceConfig, serviceName, namespace string) (*netv1beta1.Ingress, error) {

	rules := []netv1beta1.IngressRule{}
	defaultPath := "/"
	defaultPort := 80
	certSecretName := fmt.Sprintf("%s-cert", serviceName)

	for _, ingress := range serviceConfig.Ingress {
		rules = append(rules, netv1beta1.IngressRule{
			Host: ingress,
			IngressRuleValue: netv1beta1.IngressRuleValue{
				HTTP: &netv1beta1.HTTPIngressRuleValue{
					Paths: []netv1beta1.HTTPIngressPath{
						netv1beta1.HTTPIngressPath{
							Path: defaultPath,
							Backend: netv1beta1.IngressBackend{
								ServiceName: serviceName,
								ServicePort: intstr.FromInt(defaultPort),
							},
						},
					},
				},
			},
		})
	}

	return &netv1beta1.Ingress{
		TypeMeta: metav1.TypeMeta{
			Kind: "Ingress",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      serviceName,
			Namespace: namespace,
			Annotations: map[string]string{
				"certmanager.k8s.io/cluster-issuer":                   "cert-issuer",
				"traefik.ingress.kubernetes.io/frontend-entry-points": "https",
			},
			Labels: map[string]string{
				"name":          serviceName,
				"kappy.managed": serviceName,
			},
		},
		Spec: netv1beta1.IngressSpec{
			Rules: rules,
			TLS: []netv1beta1.IngressTLS{
				netv1beta1.IngressTLS{
					Hosts:      serviceConfig.Ingress,
					SecretName: certSecretName,
				},
			},
		},
	}, nil
}

func CreateUpdateIngress(client *kubernetes.Clientset, serviceConfig *model.ServiceConfig, serviceName, namespace string) error {
	resource, _ := CreateIngressResource(serviceConfig, serviceName, namespace)

	upsertCommand := UpsertCommand{
		Create: func() (err error) {
			_, err = client.ExtensionsV1beta1().Ingresses(namespace).Create(resource)

			if err == nil {
				log.Infof("Successfully created ingress %s in namespace %s", serviceName, namespace)
			}
			return
		},
		Update: func() (err error) {
			_, err = client.ExtensionsV1beta1().Ingresses(namespace).Update(resource)

			if err == nil {
				log.Infof("Successfully updated ingress %s in namespace %s", serviceName, namespace)
			}
			return
		},
	}

	return upsertCommand.Do()
}
