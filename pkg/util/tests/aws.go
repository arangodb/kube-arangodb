package tests

import (
	"testing"

	"github.com/arangodb/kube-arangodb/pkg/util"
	awsHelper "github.com/arangodb/kube-arangodb/pkg/util/aws"
)

const (
	TestAwsProfile util.EnvironmentVariable = "TEST_AWS_PROFILE"
	TestAwsRole    util.EnvironmentVariable = "TEST_AWS_ROLE"
	TestAWSBucket  util.EnvironmentVariable = "TEST_AWS_BUCKET"
)

func GetAWSS3Bucket(t *testing.T) string {
	b, ok := TestAWSBucket.Lookup()
	if !ok {
		t.Skipf("Bucket does not exists")
	}

	return b
}

func GetAWSClientConfig(t *testing.T) awsHelper.Config {
	v, ok := TestAwsProfile.Lookup()
	if !ok {
		t.Skipf("Client does not exists")
	}

	var c awsHelper.Config
	c.Region = "eu-central-1"

	c.Provider.Config = awsHelper.ProviderConfig{
		Profile: v,
	}
	c.Provider.Type = awsHelper.ProviderTypeConfig

	r, ok := TestAwsRole.Lookup()
	if ok {
		c.Provider.Impersonate = awsHelper.ProviderImpersonate{
			Impersonate: true,
			Role:        r,
			Name:        "Test",
		}
	}

	return c
}
