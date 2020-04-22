package tls

import (
	"context"
	"github.com/arangodb/go-driver"
	"net/http"
)

type Details struct {

}

func NewClient(c driver.Connection) Client {
	return &client{
		c:c,
	}
}

type Client interface {

}

type client struct {
	c driver.Connection
}

func (c * client) GetTLS(ctx context.Context) (Details, error) {
	c.c.NewRequest(http.MethodGet, )
}