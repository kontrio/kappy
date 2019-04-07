package awsutil

import (
	"encoding/base64"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ecr"
	"github.com/kontr/kappy/pkg/docker"
	"os"
	"strings"
	"time"
)

type ecrAuth struct {
	Token         string
	User          string
	Pass          string
	ProxyEndpoint string
	ExpiresAt     time.Time
}

func IsEcrRegistry(registry string) bool {
	return strings.Contains(registry, "dkr.ecr")
}

func getRegion() string {
	region, isSet := os.LookupEnv("AWS_REGION")

	if isSet {
		return region
	}

	// We should get these also from the Kappy config?
	return "eu-west-1"
}

func DoEcrLogin(registry *string) error {
	awsSession := session.New()
	ecrSvc := ecr.New(awsSession, aws.NewConfig().WithMaxRetries(2).WithRegion(getRegion()))

	registryId := string(*registry)[:12]

	params := &ecr.GetAuthorizationTokenInput{
		RegistryIds: []*string{&registryId},
	}

	resp, errAws := ecrSvc.GetAuthorizationToken(params)

	if errAws != nil {
		return errAws
	}

	authTokens := make([]ecrAuth, len(resp.AuthorizationData))

	for i, auth := range resp.AuthorizationData {

		// extract base64 token
		data, errB64 := base64.StdEncoding.DecodeString(*auth.AuthorizationToken)

		if errB64 != nil {
			return errB64
		}

		token := strings.SplitN(string(data), ":", 2)

		authTokens[i] = ecrAuth{
			Token:         *auth.AuthorizationToken,
			User:          token[0],
			Pass:          token[1],
			ProxyEndpoint: *(auth.ProxyEndpoint),
			ExpiresAt:     *(auth.ExpiresAt),
		}
	}

	for _, authToken := range authTokens {
		dockerErr := docker.RunDocker([]string{"login", "-u", authToken.User, "-p", authToken.Pass, authToken.ProxyEndpoint})

		if dockerErr != nil {
			return dockerErr
		}
	}
	return nil
}
