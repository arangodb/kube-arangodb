package context

import (
	"github.com/arangodb/go-driver/v2/arangodb"

	pbImplAuthorizationV1Shared "github.com/arangodb/kube-arangodb/integrations/authorization/v1/shared"
	"github.com/arangodb/kube-arangodb/pkg/util"
	"github.com/arangodb/kube-arangodb/pkg/util/cache"
)

const (
	ArangoDBClientCache util.ContextObject[cache.Object[arangodb.Client]] = "arangodb-client-cache-object"

	AuthZClientPlugin util.ContextObject[pbImplAuthorizationV1Shared.Evaluator] = "authz-client-plugin"
)
