package v1

import (
	pbImplStorageV1SharedS3 "github.com/arangodb/kube-arangodb/integrations/storage/v1/shared/s3"
	"github.com/arangodb/kube-arangodb/pkg/util/errors"
	"github.com/arangodb/kube-arangodb/pkg/util/svc"
)

func New(cfg Configuration) (svc.Handler, error) {
	switch cfg.Type {
	case ConfigurationTypeS3:

		impl, err := pbImplStorageV1SharedS3.NewS3Impl(cfg.S3)
		if err != nil {
			return nil, errors.Wrap(err, "failed to create S3 service server")
		}

		return impl, nil
	default:
		return nil, errors.New("currently only 's3' storage type is supported")
	}
}
