package sidecar

import "github.com/arangodb/kube-arangodb/pkg/logging"

var (
	logger = logging.Global().RegisterAndGetLogger("sidecar", logging.Info)
)
