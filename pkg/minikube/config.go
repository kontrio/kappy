package minikube

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net"
	"path"

	"github.com/apex/log"
	"github.com/kontrio/kappy/pkg/k8sauth"
	"github.com/kontrio/kappy/pkg/kstrings"
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

	ipAddress := config.Driver.IpAddress

	// Supports the "none" minikube driver
	if kstrings.IsEmpty(&config.Driver.IpAddress) {
		machineIP, err := machineIP()
		if err != nil {
			log.Errorf("Could not determine the machines IP address from the network interfaces: %s", err)
		} else {
			ipAddress = machineIP
		}
	}

	url := fmt.Sprintf("https://%s:8443", ipAddress)
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

	log.Debugf("FULL minikube config: \n%s", string(configData))

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

func machineIP() (string, error) {
	ifaces, err := net.Interfaces()
	if err != nil {
		return "", err
	}
	for _, iface := range ifaces {
		if iface.Flags&net.FlagUp == 0 {
			continue // interface down
		}
		if iface.Flags&net.FlagLoopback != 0 {
			continue // loopback interface
		}
		addrs, err := iface.Addrs()
		if err != nil {
			return "", err
		}
		for _, addr := range addrs {
			var ip net.IP
			switch v := addr.(type) {
			case *net.IPNet:
				ip = v.IP
			case *net.IPAddr:
				ip = v.IP
			}
			if ip == nil || ip.IsLoopback() {
				continue
			}
			ip = ip.To4()
			if ip == nil {
				continue // not an ipv4 address
			}
			return ip.String(), nil
		}
	}
	return "", fmt.Errorf("Could not determine local IP address")
}
