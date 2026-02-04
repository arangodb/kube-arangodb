package sidecar

import (
	"github.com/arangodb/kube-arangodb/pkg/util"
	"github.com/arangodb/kube-arangodb/pkg/util/svc"
	"github.com/spf13/cobra"
)

type registerFunc func(cmd *cobra.Command) (svc.Handler, error)

var registerer = util.NewRegisterer[string, registerFunc]()

func services(cmd *cobra.Command) ([]svc.Handler, error) {
	var r []svc.Handler

	for _, v := range registerer.Items() {
		l := logger.Str("name", v.K)
		if h, err := v.V(cmd); err != nil {
			l.Err(err).Warn("Failed to register service")
			return nil, err
		} else {
			l.Info("Registered service")
			r = append(r, h)
		}
	}

	return r, nil
}
