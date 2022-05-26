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
	"sync"

	"golang.org/x/time/rate"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/util/flowcontrol"
)

const (
	DefaultQPS   = rest.DefaultQPS * 3
	DefaultBurst = rest.DefaultBurst * 3
)

var (
	rateLimiters     = map[string]*rateLimiter{}
	rateLimitersLock sync.Mutex

	defaultQPS   = DefaultQPS
	defaultBurst = DefaultBurst
)

func GetDefaultRateLimiter() flowcontrol.RateLimiter {
	return GetRateLimiter("")
}

func GetRateLimiter(name string) flowcontrol.RateLimiter {
	rateLimitersLock.Lock()
	defer rateLimitersLock.Unlock()

	if v, ok := rateLimiters[name]; ok {
		return v
	}

	l := &rateLimiter{
		limiter: rate.NewLimiter(rate.Limit(defaultQPS), defaultBurst),
		clock:   clock{},
		qps:     defaultQPS,
	}

	rateLimiters[name] = l

	return l
}

func GetUnattachedRateLimiter() flowcontrol.RateLimiter {
	return &rateLimiter{
		limiter: rate.NewLimiter(rate.Limit(defaultQPS), defaultBurst),
		clock:   clock{},
		qps:     defaultQPS,
	}
}

func SetDefaultBurst(q int) {
	rateLimitersLock.Lock()
	defer rateLimitersLock.Unlock()

	defaultBurst = q

	for _, v := range rateLimiters {
		v.setBurst(q)
	}
}

func SetDefaultQPS(q float32) {
	rateLimitersLock.Lock()
	defer rateLimitersLock.Unlock()

	defaultQPS = q

	for _, v := range rateLimiters {
		v.setQPS(q)
	}
}
