package awsutil

import (
	"encoding/base64"
	"github.com/apex/log"
	"github.com/kontrio/kappy/pkg/k8sauth"
	"k8s.io/client-go/rest"
)

func GetKubeConfig(clusterName string) (*rest.Config, error) {
	log.Infof("Getting EKS cluster information for cluster: '%s'", clusterName)
	eksCluster, errAws := GetClusterInfo(clusterName)

	if errAws != nil {
		log.Debugf("Failed to get EKS cluster information for cluster %s - %s", clusterName, errAws)
		return nil, errAws
	}

	authToken, errAuth := GetIamAuthToken(clusterName)

	if errAuth != nil {
		return nil, errAuth
	}

	certificateAuthorityData, errEncoding := base64.StdEncoding.DecodeString(*eksCluster.CertificateAuthority.Data)

	if errEncoding != nil {
		return nil, errEncoding
	}

	config := k8sauth.ConfigFromTokenAuth(k8sauth.TokenAuth{
		Host:   *eksCluster.Endpoint,
		Token:  authToken,
		CAData: certificateAuthorityData,
	})

	return config, nil
}
