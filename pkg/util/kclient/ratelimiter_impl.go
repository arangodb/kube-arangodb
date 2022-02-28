//
// DISCLAIMER
//
// Copyright 2016-2022 ArangoDB GmbH, Cologne, Germany
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//
// Copyright holder is ArangoDB GmbH, Cologne, Germany
//

package kclient

import (
	"context"
	"sync"

	"golang.org/x/time/rate"
	"k8s.io/client-go/util/flowcontrol"
)

var _ flowcontrol.RateLimiter = &rateLimiter{}

type rateLimiter struct {
	lock sync.Mutex

	limiter *rate.Limiter
	clock   clock
	qps     float32
}

func (r *rateLimiter) setBurst(d int) {
	r.lock.Lock()
	defer r.lock.Unlock()

	r.limiter.SetBurst(d)
}

func (r *rateLimiter) setQPS(d float32) {
	r.lock.Lock()
	defer r.lock.Unlock()

	r.qps = d
	r.limiter.SetLimit(rate.Limit(d))
}

func (r *rateLimiter) Accept() {
	r.lock.Lock()
	defer r.lock.Unlock()

	now := r.clock.Now()
	r.clock.Sleep(r.limiter.ReserveN(now, 1).DelayFrom(now))
}

func (r *rateLimiter) Stop() {
}

func (r *rateLimiter) QPS() float32 {
	r.lock.Lock()
	defer r.lock.Unlock()

	return r.qps
}

func (r *rateLimiter) Wait(ctx context.Context) error {
	r.lock.Lock()
	defer r.lock.Unlock()

	return r.limiter.Wait(ctx)
}

func (r *rateLimiter) TryAccept() bool {
	r.lock.Lock()
	defer r.lock.Unlock()

	return r.limiter.AllowN(r.clock.Now(), 1)
}
