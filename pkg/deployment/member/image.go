package member

import (
	api "github.com/arangodb/kube-arangodb/pkg/apis/deployment/v1"
)

func GetImageLicense(image *api.ImageInfo) string {
	if image.Enterprise {
		return "enterprise"
	}
	return "community"
}
