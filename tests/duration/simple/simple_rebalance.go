//
// DISCLAIMER
//
// Copyright 2020 ArangoDB GmbH, Cologne, Germany
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
// Author Ewout Prangsma
//

package simple

// rebalanceShards attempts to rebalance shards over the existing servers.
// The operation is expected to succeed.
func (t *simpleTest) rebalanceShards() error {
	/*opts := struct{}{}
	operationTimeout, retryTimeout := t.OperationTimeout, t.RetryTimeout
	t.log.Info().Msgf("Rebalancing shards...")
	if _, err := t.client.Post("/_admin/cluster/rebalanceShards", nil, nil, opts, "", nil, []int{202}, []int{400, 403, 503}, operationTimeout, retryTimeout); err != nil {
		// This is a failure
		t.rebalanceShardsCounter.failed++
		t.reportFailure(test.NewFailure("Failed to rebalance shards: %v", err))
		return maskAny(err)
	}
	t.rebalanceShardsCounter.succeeded++
	t.log.Info().Msgf("Rebalancing shards succeeded")*/
	return nil
}
