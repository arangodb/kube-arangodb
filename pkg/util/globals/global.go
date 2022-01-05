//
// DISCLAIMER
//
// Copyright 2016-2021 ArangoDB GmbH, Cologne, Germany
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

package globals

import "time"

const (
	DefaultKubernetesTimeout     = 2 * time.Second
	DefaultArangoDTimeout        = time.Second * 10
	DefaultReconciliationTimeout = time.Minute

	DefaultKubernetesRequestBatchSize = 256

	DefaultBackupConcurrentUploads = 4
)

var globalObj = &globals{
	timeouts: &globalTimeouts{
		requests:       NewTimeout(DefaultKubernetesTimeout),
		arangod:        NewTimeout(DefaultArangoDTimeout),
		reconciliation: NewTimeout(DefaultReconciliationTimeout),
	},
	kubernetes: &globalKubernetes{
		requestBatchSize: NewInt64(DefaultKubernetesRequestBatchSize),
	},
	backup: &globalBackup{
		concurrentUploads: NewInt(DefaultBackupConcurrentUploads),
	},
}

func GetGlobals() Globals {
	return globalObj
}

func GetGlobalTimeouts() GlobalTimeouts {
	return globalObj.timeouts
}

type Globals interface {
	Timeouts() GlobalTimeouts
	Kubernetes() GlobalKubernetes
	Backup() GlobalBackup
}

type globals struct {
	timeouts   *globalTimeouts
	kubernetes *globalKubernetes
	backup     *globalBackup
}

func (g globals) Backup() GlobalBackup {
	return g.backup
}

func (g globals) Kubernetes() GlobalKubernetes {
	return g.kubernetes
}

func (g globals) Timeouts() GlobalTimeouts {
	return g.timeouts
}

type GlobalKubernetes interface {
	RequestBatchSize() Int64
}

type globalKubernetes struct {
	requestBatchSize Int64
}

func (g *globalKubernetes) RequestBatchSize() Int64 {
	return g.requestBatchSize
}

type GlobalBackup interface {
	ConcurrentUploads() Int
}

type globalBackup struct {
	concurrentUploads Int
}

func (g *globalBackup) ConcurrentUploads() Int {
	return g.concurrentUploads
}

type GlobalTimeouts interface {
	Reconciliation() Timeout

	Kubernetes() Timeout
	ArangoD() Timeout
}

type globalTimeouts struct {
	requests, arangod, reconciliation Timeout
}

func (g *globalTimeouts) Reconciliation() Timeout {
	return g.reconciliation
}

func (g *globalTimeouts) ArangoD() Timeout {
	return g.arangod
}

func (g *globalTimeouts) Kubernetes() Timeout {
	return g.requests
}
