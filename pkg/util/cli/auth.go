package cli

import (
	"github.com/arangodb/kube-arangodb/pkg/util"
	"github.com/arangodb/kube-arangodb/pkg/util/cache"
	"github.com/arangodb/kube-arangodb/pkg/util/token"
)

func NewServiceAuth(prefix string, mods ...util.Mod[ServiceConfigurationInput]) ServiceAuth {
var
}

type ServiceAuthConfig struct {
	Enabled bool
	Path    string
}

type ServiceAuth interface {
	FlagRegisterer

	Loader() cache.ObjectFetcher[token.Secret]
}

type serviceAuth struct {
	enabled Flag[bool]

	path string
}
