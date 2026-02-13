package client

import "github.com/arangodb/kube-arangodb/pkg/logging"

var logger = logging.Global().RegisterAndGetLogger("authz-pool-client", logging.Info)
