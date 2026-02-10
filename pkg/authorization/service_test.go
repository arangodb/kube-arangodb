package authorization

import (
	"context"
	"fmt"
	goStrings "strings"
	"testing"
	"time"

	"github.com/dchest/uniuri"
	"github.com/stretchr/testify/require"

	"github.com/arangodb/go-driver/v2/arangodb"

	"github.com/arangodb/kube-arangodb/pkg/authorization/client"
	"github.com/arangodb/kube-arangodb/pkg/authorization/service"
	"github.com/arangodb/kube-arangodb/pkg/util"
	"github.com/arangodb/kube-arangodb/pkg/util/arangod/db"
	"github.com/arangodb/kube-arangodb/pkg/util/cache"
	"github.com/arangodb/kube-arangodb/pkg/util/svc"
	"github.com/arangodb/kube-arangodb/pkg/util/tests"
	"github.com/arangodb/kube-arangodb/pkg/util/tests/tgrpc"
)

func Handler(t *testing.T) svc.Handler {
	return NewAuthorizer(db.NewClient(cache.NewObject(tests.TestArangoDBConfig(t).ClientCache())).
		CreateDatabase(fmt.Sprintf("db-%s", goStrings.ToLower(uniuri.NewLen(8))), &arangodb.CreateDatabaseOptions{}).
		CreateCollection("_users", db.StaticProps(arangodb.CreateCollectionPropertiesV2{
			IsSystem: util.NewType(true),
		})).Database())
}

func Server(t *testing.T, ctx context.Context) svc.ServiceStarter {
	local, err := svc.NewService(svc.Configuration{
		Address: "127.0.0.1:0",
		Gateway: &svc.ConfigurationGateway{
			Address: "127.0.0.1:0",
		},
	}, Handler(t))
	require.NoError(t, err)

	return local.Start(ctx)
}

func Client(t *testing.T, ctx context.Context) service.AuthorizationPoolServiceClient {
	start := Server(t, ctx)

	return tgrpc.NewGRPCClient(t, ctx, service.NewAuthorizationPoolServiceClient, start.Address())
}

func Test_Service(t *testing.T) {
	ctx, c := context.WithCancel(t.Context())
	defer c()

	q := Client(t, ctx)

	z := client.NewClient(t.Context(), q)

	require.True(t, z.Wait(time.Second))

	time.Sleep(5 * time.Minute)
}
