package kubernetes

import (
	"errors"
	"fmt"
	"strings"

	"github.com/kontrio/kappy/pkg/awsutil"
	"github.com/kontrio/kappy/pkg/minikube"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

func GetConfig(clusterName string) (*rest.Config, error) {

	parts := strings.Split(clusterName, ":")

	if len(parts) != 2 {
		return nil, errors.New("Could not get cluster name must be in the format: 'type:name'")
	}

	clusterType := parts[0]
	name := parts[1]

	switch clusterType {
	case "eks":
		return awsutil.GetKubeConfig(name)
	case "minikube":
		return minikube.GetKubeConfig(name)
	}

	return nil, errors.New(fmt.Sprintf("Unknown cluster type: %s (%s)", clusterType, clusterName))
}

func CreateClient(clusterName string) (*kubernetes.Clientset, error) {
	config, errConfig := GetConfig(clusterName)

	if errConfig != nil {
		return nil, errConfig
	}

	return kubernetes.NewForConfig(config)
}
