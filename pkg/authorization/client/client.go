package client

import (
	"context"
	"io"
	"sync"
	"time"

	"golang.org/x/sync/errgroup"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/durationpb"

	pbSharedV1 "github.com/arangodb/kube-arangodb/integrations/shared/v1/definition"
	"github.com/arangodb/kube-arangodb/pkg/authorization/service"
	"github.com/arangodb/kube-arangodb/pkg/authorization/types"
)

func NewClient(ctx context.Context, c service.AuthorizationPoolServiceClient) Client {
	client := &client{
		client: c,
		closed: make(chan struct{}),
	}

	go client.run(ctx)

	return client
}

type Client interface {
	Ready() bool
	Wait(timeout time.Duration) bool
}

type client struct {
	client service.AuthorizationPoolServiceClient

	closed chan struct{}

	policies clientSet[*types.Policy]
	roles    clientSet[*types.Role]
}

func (c *client) Wait(timeout time.Duration) bool {
	timerT := time.NewTimer(timeout)
	defer timerT.Stop()

	tickerT := time.NewTicker(50 * time.Millisecond)
	defer tickerT.Stop()

	for {
		select {
		case <-c.closed:
			return false
		case <-timerT.C:
			return false
		case <-tickerT.C:
			if c.Ready() {
				return true
			}
		}
	}
}

func (c *client) Ready() bool {
	return c.policies.Ready() && c.roles.Ready()
}

type clientSet[T proto.Message] struct {
	lock sync.RWMutex

	items map[string]T

	updated time.Time
}

func (c *clientSet[T]) Ready() bool {
	c.lock.RLock()
	defer c.lock.RUnlock()

	return time.Since(c.updated) < time.Minute
}

func (c *clientSet[T]) Get() map[string]T {
	c.lock.RLock()
	defer c.lock.RUnlock()

	return c.items
}

func (c *clientSet[T]) Set(items map[string]T) {
	cp := make(map[string]T, len(items))
	for k, v := range items {
		cp[k] = v
	}

	c.lock.Lock()
	defer c.lock.Unlock()

	c.items = cp
	c.updated = time.Now()
}

func (c *client) run(ctx context.Context) {
	defer close(c.closed)
	for {
		if err := c.runE(ctx); err != nil {
			logger.Err(err).Warn("Authorization pool client error")
		}

		select {
		case <-ctx.Done():
			return
		default:
			continue
		}
	}
}

func (c *client) runE(ctx context.Context) error {
	g, gctx := errgroup.WithContext(ctx)

	g.Go(func() error {
		return c.runPoliciesE(gctx)
	})

	g.Go(func() error {
		return c.runRolesE(gctx)
	})

	return g.Wait()
}

func (c *client) runPoliciesE(ctx context.Context) error {
	policies := map[string]*types.Policy{}

	var index uint32

	{
		response, err := c.client.GetPolicy(ctx, &pbSharedV1.Empty{})
		if err != nil {
			return err
		}

		for {
			spec, err := response.Recv()
			if err != nil {
				if err == io.EOF {
					break
				}
				return err
			}

			for _, item := range spec.GetItems() {
				policies[item.GetName()] = item.GetItem()
				index = item.GetIndex()
			}
		}
	}

	c.policies.Set(policies)

	logger.Trace("Policies init complete")

	for {
		changes, err := c.client.PoolPolicyChanges(ctx, &service.AuthorizationPoolRequest{
			Start:   index,
			Timeout: durationpb.New(15 * time.Second),
		})
		if err != nil {
			return err
		}

		for {
			spec, err := changes.Recv()
			if err != nil {
				logger.Err(err).Info("Policy changes failed")
				break
			}

			logger.Int("items", len(spec.Items)).Trace("Received policy update")

			for _, item := range spec.GetItems() {
				policies[item.GetName()] = item.GetItem()
				index = item.GetIndex()
			}
		}

		c.policies.Set(policies)
	}
}

func (c *client) runRolesE(ctx context.Context) error {
	roles := map[string]*types.Role{}

	var index uint32

	{
		response, err := c.client.GetRole(ctx, &pbSharedV1.Empty{})
		if err != nil {
			return err
		}

		for {
			spec, err := response.Recv()
			if err != nil {
				if err == io.EOF {
					break
				}
				return err
			}

			for _, item := range spec.GetItems() {
				roles[item.GetName()] = item.GetItem()
				index = item.GetIndex()
			}
		}
	}

	c.roles.Set(roles)

	logger.Trace("Roles init complete")

	for {
		changes, err := c.client.PoolRoleChanges(ctx, &service.AuthorizationPoolRequest{
			Start:   index,
			Timeout: durationpb.New(15 * time.Second),
		})
		if err != nil {
			return err
		}

		for {
			spec, err := changes.Recv()
			if err != nil {
				logger.Err(err).Info("Role changes failed")
				break
			}

			logger.Int("items", len(spec.Items)).Trace("Received roles update")

			for _, item := range spec.GetItems() {
				roles[item.GetName()] = item.GetItem()
				index = item.GetIndex()
			}
		}

		c.roles.Set(roles)
	}
}
