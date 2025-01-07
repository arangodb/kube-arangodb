package s3

import (
	"testing"

	"github.com/arangodb/kube-arangodb/pkg/util/tests"
)

func getClient(t *testing.T) Configuration {
	var scfg Configuration

	scfg.Client = tests.GetAWSClientConfig(t)
	scfg.BucketName = tests.GetAWSS3Bucket(t)

	return scfg
}
