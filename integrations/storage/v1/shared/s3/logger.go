package s3

import "github.com/arangodb/kube-arangodb/pkg/logging"

var logger = logging.Global().RegisterAndGetLogger("integration-storage-v1-s3", logging.Info)
