package sidecar

import (
	pbImplPongV1 "github.com/arangodb/kube-arangodb/integrations/pong/v1"
	pbPongV1 "github.com/arangodb/kube-arangodb/integrations/pong/v1/definition"
	"github.com/arangodb/kube-arangodb/pkg/util/svc"
	"github.com/spf13/cobra"
)

func init() {
	registerer.MustRegister(pbPongV1.Name, func(cmd *cobra.Command) (svc.Handler, error) {
		return pbImplPongV1.New()
	})
}
