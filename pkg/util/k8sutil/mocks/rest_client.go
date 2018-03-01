package mocks

import "k8s.io/client-go/rest"

type restClient struct {
}

func (c *restClient) RESTClient() rest.Interface {
	panic("not support")
}
