package awsutil

import (
	"encoding/base64"
	"github.com/apex/log"
	k8s "k8s.io/client-go/tools/clientcmd/api"
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

	authInfo := k8s.AuthInfo{
		Token: authToken,
	}

	certificateAuthorityData, errEncoding := base64.StdEncoding.DecodeString(*eksCluster.CertificateAuthority.Data)

	if errEncoding != nil {
		return nil, errEncoding
	}

	cluster := k8s.Cluster{
		Server:                   eksCluster.Endpoint,
		InsecureSkipTLSVerify:    false,
		CertificateAuthorityData: certificateAuthorityData,
	}

	log.Debugf("Using cluster: %s", cluster.Cluster.Server)

	context := k8s.Context{
		Cluster:  clusterName,
		AuthInfo: "iam-aws-auth",
	}

	clusters := make(map[string]string)
	clusters[clusterName] = &cluster
	
	contexts := map[string]string{
		"main": &context
	}

	return &k8s.Config{
		Clusters:       clusters,
		AuthInfos:      map[string]string{
			"iam-aws-auth": &authInfo,
		},
		Contexts:       contexts,
		CurrentContext: "main",
	}, nil
}
