package awsutil

import (
	"github.com/apex/log"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/eks"
	awstoken "github.com/kubernetes-sigs/aws-iam-authenticator/pkg/token"
)

func GetIamAuthToken(clusterName string) (string, error) {
	generator, err := awstoken.NewGenerator()
	if err != nil {
		return "", err
	}

	token, errGetToken := generator.Get(clusterName)
	if errGetToken != nil {
		return "", errGetToken
	}

	return token, nil
}

func GetClusterInfo(name string) (*eks.Cluster, error) {
	awsSession := session.New()
	eksService := eks.New(awsSession, aws.NewConfig().WithMaxRetries(2).WithRegion(getRegion()))

	log.Debugf("Getting cluster info for cluster named: %s", name)

	output, err := eksService.DescribeCluster(&eks.DescribeClusterInput{
		Name: &name,
	})

	if err != nil {
		return nil, err
	}

	return output.Cluster, nil
}
