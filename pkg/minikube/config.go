package minikube

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"path"

	"github.com/apex/log"
	"github.com/ericchiang/k8s"
	"github.com/mitchellh/go-homedir"
)

func GetKubeConfig(clusterName string) (*k8s.Config, error) {
	log.Infof("Getting Minikube cluster information for cluster: '%s'", clusterName)
	log.Warnf("Minikube support is experimental, proceed with low expectations!")

	config, err := readMiniKubeMachineConfig(clusterName)

	if err != nil {
		return nil, err
	}

	clientCert, errCrt := ioutil.ReadFile(config.HostOptions.AuthOptions.ClientCertPath)
	if errCrt != nil {
		return nil, errCrt
	}

	clientKey, errKey := ioutil.ReadFile(config.HostOptions.AuthOptions.ClientKeyPath)
	if errKey != nil {
		return nil, errKey
	}

	caCertPath, errCa := ioutil.ReadFile(config.HostOptions.AuthOptions.CaCertPath)
	if errCa != nil {
		return nil, errCa
	}

	log.Debugf("Using client cert: %s key: %s cacert: %s", clientCert, clientKey, caCertPath)

	authInfo := k8s.NamedAuthInfo{
		Name: "minikube-auth",
		AuthInfo: k8s.AuthInfo{
			ClientCertificateData: clientCert,
			ClientKeyData:         clientKey,
		},
	}

	url := fmt.Sprintf("https://%s:8443", config.Driver.IpAddress)
	log.Debugf("Using kube api server %s", url)

	cluster := k8s.NamedCluster{
		Name: clusterName,
		Cluster: k8s.Cluster{
			Server:                   url,
			CertificateAuthorityData: caCertPath,
		},
	}

	context := k8s.NamedContext{
		Name: "main",
		Context: k8s.Context{
			Cluster:  clusterName,
			AuthInfo: "minikube-auth",
		},
	}

	return &k8s.Config{
		Clusters:       []k8s.NamedCluster{cluster},
		AuthInfos:      []k8s.NamedAuthInfo{authInfo},
		Contexts:       []k8s.NamedContext{context},
		CurrentContext: "main",
	}, nil
}

func readMiniKubeMachineConfig(clusterName string) (*minikubeConfig, error) {
	homeDir, err := homedir.Dir()

	if err != nil {
		return nil, err
	}

	configFile := path.Join(homeDir, ".minikube", "machines", clusterName, "config.json")
	configData, errRead := ioutil.ReadFile(configFile)

	if errRead != nil {
		log.Debugf("Failed to read minikube config file %s", errRead)
		return nil, errRead
	}

	config := minikubeConfig{}
	errMarshal := json.Unmarshal(configData, &config)

	if errMarshal != nil {
		log.Debugf("Failed to parse minikube config file %s", errMarshal)
		return nil, errMarshal
	}

	if config.Version != 3 {
		log.Warnf("Minikube config version is '%d', this was built for version '%d', please report any errors with this version", config.Version, 3)
	}

	return &config, nil

}

type minikubeConfig struct {
	Version     int32       `json:"Version"`
	HostOptions hostOptions `json:"HostOptions"`
	Driver      driver      `json:"Driver"`
	Name        string      `json:"name"`
}

type driver struct {
	IpAddress string `json:"IPAddress"`
}

type hostOptions struct {
	AuthOptions authOptions `json:"AuthOptions"`
}

type authOptions struct {
	CertDir          string `json:"CertDir"`
	CaCertPath       string `json:"CaCertPath"`
	CaPrivateKeyPath string `json:"CaPrivateKeyPath"`
	ServerCertPath   string `json:"ServerCertPath"`
	ServerKeyPath    string `json:"ServerKeyPath"`
	ClientKeyPath    string `json:"ClientKeyPath"`
	ClientCertPath   string `json:"ClientCertPath"`
}
