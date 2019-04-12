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
