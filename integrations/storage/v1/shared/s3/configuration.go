package s3

import (
	awsHelper "github.com/arangodb/kube-arangodb/pkg/util/aws"
)

type Configuration struct {
	BucketName string

	Client awsHelper.Config
}
