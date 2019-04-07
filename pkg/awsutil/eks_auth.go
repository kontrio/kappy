package awsutil

import (
	"os/exec"
	"strings"

	"github.com/apex/log"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/eks"
)

func GetIamAuthToken(clusterName string) (string, error) {
	output, err := exec.Command("k8s-aws-authenticator", "token", "--cluster-id", clusterName, "--token-only").Output()
	return strings.TrimSpace(string(output)), err
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
