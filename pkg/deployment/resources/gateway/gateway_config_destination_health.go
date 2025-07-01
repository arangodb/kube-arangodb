package gateway

import (
	"github.com/arangodb/kube-arangodb/pkg/apis/shared"
	"github.com/arangodb/kube-arangodb/pkg/util"
	pbEnvoyCoreV3 "github.com/envoyproxy/go-control-plane/envoy/config/core/v3"
	"google.golang.org/protobuf/types/known/durationpb"
	"time"
)

type ConfigDestinationHealthChecks []ConfigDestinationHealthCheck

func (c ConfigDestinationHealthChecks) Validate() error {
	return shared.ValidateInterfaceList(c)
}

func (c ConfigDestinationHealthChecks) Render() []*pbEnvoyCoreV3.HealthCheck {
	ret := make([]*pbEnvoyCoreV3.HealthCheck, len(c))
	for id := range c {
		ret[id] = c[id].Render()
	}
	return ret
}

type ConfigDestinationHealthCheck struct {
	Timeout *time.Duration `json:"timeout,omitempty"`

	Interval *time.Duration `json:"interval,omitempty"`
}

func (c ConfigDestinationHealthCheck) Validate() error {
	return nil
}

func (c ConfigDestinationHealthCheck) Render() *pbEnvoyCoreV3.HealthCheck {
	return &pbEnvoyCoreV3.HealthCheck{
		Timeout:  durationpb.New(util.OptionalType(c.Timeout, time.Second)),
		Interval: durationpb.New(util.OptionalType(c.Interval, time.Second)),

		HealthChecker: &pbEnvoyCoreV3.HealthCheck_TcpHealthCheck_{
			TcpHealthCheck: &pbEnvoyCoreV3.HealthCheck_TcpHealthCheck{},
		},
	}
}
