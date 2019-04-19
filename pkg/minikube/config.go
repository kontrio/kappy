package minikube

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"path"

	"github.com/apex/log"
	"github.com/kontrio/kappy/pkg/k8sauth"
	"github.com/mitchellh/go-homedir"
	"k8s.io/client-go/rest"
)

func GetKubeConfig(clusterName string) (*rest.Config, error) {
	log.Infof("Getting Minikube cluster information for cluster: '%s'", clusterName)
	log.Warnf("Minikube support is experimental, proceed with low expectations!")

	config, err := readMiniKubeMachineConfig(clusterName)

	if err != nil {
		return nil, err
	}

	log.Debugf("Reading clientcert: %s", config.HostOptions.AuthOptions.ClientCertPath)
	clientCert, errCrt := ioutil.ReadFile(config.HostOptions.AuthOptions.ClientCertPath)
	if errCrt != nil {
		return nil, errCrt
	}

	log.Debugf("Reading clientcert: %s", config.HostOptions.AuthOptions.ClientKeyPath)
	clientKey, errKey := ioutil.ReadFile(config.HostOptions.AuthOptions.ClientKeyPath)
	if errKey != nil {
		return nil, errKey
	}

	log.Debugf("Reading clientcert: %s", config.HostOptions.AuthOptions.CaCertPath)
	caCert, errCa := ioutil.ReadFile(config.HostOptions.AuthOptions.CaCertPath)
	if errCa != nil {
		return nil, errCa
	}

	url := fmt.Sprintf("https://%s:8443", config.Driver.IpAddress)
	log.Debugf("Using kube api server %s", url)

	return k8sauth.ConfigFromTLSClientAuth(k8sauth.TLSClientAuth{
		KeyData:  clientKey,
		CertData: clientCert,
		CAData:   caCert,

		Host: url,
	}), nil
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

	// HACK: For some reason the config cert pem is wrong
	crtPath := path.Join(homeDir, ".minikube", "client.crt")
	config.HostOptions.AuthOptions.ClientCertPath = crtPath

	keyPath := path.Join(homeDir, ".minikube", "client.key")
	config.HostOptions.AuthOptions.ClientKeyPath = keyPath

	caPath := path.Join(homeDir, ".minikube", "ca.crt")
	config.HostOptions.AuthOptions.CaCertPath = caPath

	log.Debugf("%#v", config)

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
