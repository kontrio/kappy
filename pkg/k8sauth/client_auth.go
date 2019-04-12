package k8sauth

import (
	"k8s.io/client-go/rest"
)

const kappyUserAgent = "kappy/0.0.x alpha"

type TokenAuth struct {
	Token  string
	CAData []byte

	Host string
}

type TLSClientAuth struct {
	KeyData  []byte
	CertData []byte
	CAData   []byte

	Host string
}

func ConfigFromTLSClientAuth(options TLSClientAuth) *rest.Config {
	return &rest.Config{
		Host: options.Host,
		TLSClientConfig: rest.TLSClientConfig{
			Insecure: false,
			KeyData:  options.KeyData,
			CertData: options.CertData,
			CAData:   options.CAData,
		},
	}
}

func ConfigFromTokenAuth(tokenAuth TokenAuth) *rest.Config {
	return &rest.Config{
		Host:        tokenAuth.Host,
		BearerToken: tokenAuth.Token,
		TLSClientConfig: rest.TLSClientConfig{
			Insecure: false,
			CAData:   tokenAuth.CAData,
		},
	}
}

func WithKappyUserAgent(config *rest.Config) *rest.Config {
	return rest.AddUserAgent(config, kappyUserAgent)
}
