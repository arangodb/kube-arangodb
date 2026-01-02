package tls

import (
	"context"
	goHttp "net/http"

	"github.com/arangodb/kube-arangodb/pkg/util"
)

func WithTLSConfigFetcherGen(gen func() TLSConfigFetcher) util.ModEP1[goHttp.Server, context.Context] {
	return WithTLSConfigFetcher(gen())
}

func WithTLSConfigFetcher(fetcher TLSConfigFetcher) util.ModEP1[goHttp.Server, context.Context] {
	return func(in *goHttp.Server, p1 context.Context) error {
		v, err := fetcher.Eval(p1)
		if err != nil {
			return err
		}

		in.TLSConfig = v

		return nil
	}
}
