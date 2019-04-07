package awsutil

import (
	"encoding/base64"
	"github.com/apex/log"
	"github.com/ericchiang/k8s"
)

func GetKubeConfig(clusterName string) (*k8s.Config, error) {
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

	authInfo := k8s.NamedAuthInfo{
		Name: "iam-aws-auth",
		AuthInfo: k8s.AuthInfo{
			Token: authToken,
		},
	}

	certificateAuthorityData, errEncoding := base64.StdEncoding.DecodeString(*eksCluster.CertificateAuthority.Data)

	if errEncoding != nil {
		return nil, errEncoding
	}

	cluster := k8s.NamedCluster{
		Name: clusterName,
		Cluster: k8s.Cluster{
			Server:                   *eksCluster.Endpoint,
			InsecureSkipTLSVerify:    false,
			CertificateAuthorityData: certificateAuthorityData,
		},
	}

	log.Debugf("Using cluster: %s", cluster.Cluster.Server)

	context := k8s.NamedContext{
		Name: "main",
		Context: k8s.Context{
			Cluster:  clusterName,
			AuthInfo: "iam-aws-auth",
		},
	}

	return &k8s.Config{
		Clusters:       []k8s.NamedCluster{cluster},
		AuthInfos:      []k8s.NamedAuthInfo{authInfo},
		Contexts:       []k8s.NamedContext{context},
		CurrentContext: "main",
	}, nil
}
