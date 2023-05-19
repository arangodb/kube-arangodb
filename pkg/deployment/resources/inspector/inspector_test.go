//
// DISCLAIMER
//
// Copyright 2016-2023 ArangoDB GmbH, Cologne, Germany
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

package inspector

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil/inspector"
	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil/inspector/refresh"
	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil/inspector/throttle"
	"github.com/arangodb/kube-arangodb/pkg/util/kclient"
)

type loaderTestDefinition struct {
	tg  func(t throttle.Components) throttle.Throttle
	get func(i inspector.Inspector) refresh.Inspector
}

var loaderTestDefinitions = map[string]loaderTestDefinition{
	"Secret": {
		tg: func(t throttle.Components) throttle.Throttle {
			return t.Secret()
		},
		get: func(i inspector.Inspector) refresh.Inspector {
			return i.Secret()
		},
	},
	"Service": {
		tg: func(t throttle.Components) throttle.Throttle {
			return t.Service()
		},
		get: func(i inspector.Inspector) refresh.Inspector {
			return i.Service()
		},
	},
	"ServiceAccount": {
		tg: func(t throttle.Components) throttle.Throttle {
			return t.ServiceAccount()
		},
		get: func(i inspector.Inspector) refresh.Inspector {
			return i.ServiceAccount()
		},
	},
	"Node": {
		tg: func(t throttle.Components) throttle.Throttle {
			return t.Node()
		},
		get: func(i inspector.Inspector) refresh.Inspector {
			return i.Node()
		},
	},
	"PersistentVolume": {
		tg: func(t throttle.Components) throttle.Throttle {
			return t.PersistentVolume()
		},
		get: func(i inspector.Inspector) refresh.Inspector {
			return i.PersistentVolume()
		},
	},
	"Pod": {
		tg: func(t throttle.Components) throttle.Throttle {
			return t.Pod()
		},
		get: func(i inspector.Inspector) refresh.Inspector {
			return i.Pod()
		},
	},
	"PodDisruptionBudget": {
		tg: func(t throttle.Components) throttle.Throttle {
			return t.PodDisruptionBudget()
		},
		get: func(i inspector.Inspector) refresh.Inspector {
			return i.PodDisruptionBudget()
		},
	},
	"ServiceMonitor": {
		tg: func(t throttle.Components) throttle.Throttle {
			return t.ServiceMonitor()
		},
		get: func(i inspector.Inspector) refresh.Inspector {
			return i.ServiceMonitor()
		},
	},
	"ArangoMember": {
		tg: func(t throttle.Components) throttle.Throttle {
			return t.ArangoMember()
		},
		get: func(i inspector.Inspector) refresh.Inspector {
			return i.ArangoMember()
		},
	},
	"ArangoTask": {
		tg: func(t throttle.Components) throttle.Throttle {
			return t.ArangoTask()
		},
		get: func(i inspector.Inspector) refresh.Inspector {
			return i.ArangoTask()
		},
	},
	"ArangoClusterSynchronizations": {
		tg: func(t throttle.Components) throttle.Throttle {
			return t.ArangoClusterSynchronization()
		},
		get: func(i inspector.Inspector) refresh.Inspector {
			return i.ArangoClusterSynchronization()
		},
	},
}

func getAllTypes() []string {
	r := make([]string, 0, len(loaderTestDefinitions))

	for k := range loaderTestDefinitions {
		r = append(r, k)
	}

	return r
}

func Test_Inspector_RefreshMatrix(t *testing.T) {
	c := kclient.NewFakeClient()

	tc := throttle.NewThrottleComponents(time.Hour, time.Hour, time.Hour, time.Hour, time.Hour, time.Hour, time.Hour, time.Hour, time.Hour, time.Hour, time.Hour, time.Hour, time.Hour)

	i := NewInspector(tc, c, "test", "test")

	require.NoError(t, i.Refresh(context.Background()))

	combineAllTypes(func(changed, not []string) {
		{
			times := getTimes(i)
			time.Sleep(time.Millisecond)

			for _, k := range changed {
				require.NoError(t, loaderTestDefinitions[k].get(i).Refresh(context.Background()))
			}

			ntimes := getTimes(i)

			for _, k := range changed {
				require.NotEqual(t, times[k], ntimes[k])
			}

			for _, k := range not {
				require.Equal(t, times[k], ntimes[k])
			}

			for _, k := range changed {
				loaderTestDefinitions[k].tg(i.GetThrottles()).Invalidate()
			}

			for _, k := range changed {
				require.NoError(t, loaderTestDefinitions[k].get(i).Refresh(context.Background()))
			}

			ntimes = getTimes(i)

			for _, k := range changed {
				require.NotEqual(t, times[k], ntimes[k])
			}

			for _, k := range not {
				require.Equal(t, times[k], ntimes[k])
			}
		}

		{
			times := getTimes(i)
			time.Sleep(time.Millisecond)

			require.NoError(t, i.Refresh(context.Background()))

			ntimes := getTimes(i)
			for k, v := range times {
				require.Equal(t, v, ntimes[k])
			}

			for _, k := range changed {
				loaderTestDefinitions[k].tg(i.GetThrottles()).Invalidate()
			}

			require.NoError(t, i.Refresh(context.Background()))

			ntimes = getTimes(i)

			for _, k := range changed {
				require.NotEqual(t, times[k], ntimes[k])
			}

			for _, k := range not {
				require.Equal(t, times[k], ntimes[k])
			}
		}
	})
}

func combineAllTypes(f func(changed, not []string)) {
	t := getAllTypes()
	cmb := make([]bool, len(t))

	cmbc := make([]bool, len(t))

	getAllCombinations(cmb, func() {
		copy(cmbc, cmb)
		z := 0
		for i := 0; i < len(cmb); i++ {
			if !cmbc[i] {
				for j := len(cmb) - 1; j > i; j-- {
					if cmbc[j] {
						t[i], t[j] = t[j], t[i]
						cmbc[i], cmbc[j] = cmbc[j], cmbc[i]
						break
					}
				}
			}
		}
		for i := 0; i < len(cmb); i++ {
			if cmb[i] {
				z++
			}
		}
		f(t[0:z], t[z:])
		copy(cmbc, cmb)
		for i := 0; i < len(cmb); i++ {
			if !cmbc[i] {
				for j := len(cmb) - 1; j > i; j-- {
					if cmbc[j] {
						t[i], t[j] = t[j], t[i]
						cmbc[i], cmbc[j] = cmbc[j], cmbc[i]
						break
					}
				}
			}
		}
	})
}

func getAllCombinations(cmb []bool, f func()) {
	for {
		f()

		if !bumpCombination(cmb, 0) {
			return
		}
	}
}

func bumpCombination(cmd []bool, index int) bool {
	if index >= len(cmd) {
		return false
	}

	if cmd[index] {
		cmd[index] = false
		return bumpCombination(cmd, index+1)
	}

	cmd[index] = true
	return true
}

func getTimes(i inspector.Inspector) map[string]time.Time {
	r := map[string]time.Time{}

	for k, v := range loaderTestDefinitions {
		r[k] = v.get(i).LastRefresh()
	}

	return r
}

func Test_Inspector_Load(t *testing.T) {
	c := kclient.NewFakeClient()

	i := NewInspector(throttle.NewAlwaysThrottleComponents(), c, "test", "test")

	require.NoError(t, i.Refresh(context.Background()))
}

func Test_Inspector_Invalidate(t *testing.T) {
	c := kclient.NewFakeClient()

	tc := throttle.NewThrottleComponents(time.Hour, time.Hour, time.Hour, time.Hour, time.Hour, time.Hour, time.Hour, time.Hour, time.Hour, time.Hour, time.Hour, time.Hour, time.Hour)

	i := NewInspector(tc, c, "test", "test")

	require.NoError(t, i.Refresh(context.Background()))

	for n, q := range loaderTestDefinitions {
		t.Run(n, func(t *testing.T) {
			t.Run("Specific", func(t *testing.T) {
				times := getTimes(i)
				time.Sleep(20 * time.Millisecond)

				t.Run("Refresh", func(t *testing.T) {
					require.NoError(t, q.get(i).Refresh(context.Background()))
				})

				t.Run("Ensure time changed", func(t *testing.T) {
					ntimes := getTimes(i)
					for k, v := range times {
						if k == n {
							require.NotEqual(t, v, ntimes[k])
						} else {
							require.Equal(t, v, ntimes[k])
						}
					}
				})

				t.Run("Ensure time changed", func(t *testing.T) {
					ntimes := getTimes(i)
					for k, v := range times {
						if k == n {
							require.NotEqual(t, v, ntimes[k])
						} else {
							require.Equal(t, v, ntimes[k])
						}
					}
				})
			})
			t.Run("All", func(t *testing.T) {
				times := getTimes(i)
				time.Sleep(20 * time.Millisecond)

				t.Run("Refresh", func(t *testing.T) {
					require.NoError(t, i.Refresh(context.Background()))
				})

				t.Run("Ensure time did not change", func(t *testing.T) {
					ntimes := getTimes(i)
					for k, v := range times {
						require.Equal(t, v, ntimes[k])
					}
				})

				t.Run("Invalidate", func(t *testing.T) {
					q.tg(i.GetThrottles()).Invalidate()
				})

				t.Run("Refresh after invalidate", func(t *testing.T) {
					require.NoError(t, i.Refresh(context.Background()))
				})

				t.Run("Ensure time changed", func(t *testing.T) {
					ntimes := getTimes(i)
					for k, v := range times {
						if k == n {
							require.NotEqual(t, v, ntimes[k])
						} else {
							require.Equal(t, v, ntimes[k])
						}
					}
				})
			})
		})
	}
}
