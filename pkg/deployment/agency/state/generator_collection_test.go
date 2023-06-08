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

package state

type CollectionGeneratorInterface interface {
	WithWriteConcern(wc int) CollectionGeneratorInterface
	WithReplicationFactor(rf int) CollectionGeneratorInterface
	WithShard() ShardGeneratorInterface
	Add() DatabaseGeneratorInterface
}

type collectionGenerator struct {
	db  databaseGenerator
	col string

	wc     *int
	rf     *ReplicationFactor
	shards map[int]shardGenerator
}

func (c collectionGenerator) Add() DatabaseGeneratorInterface {
	d := c.db
	if d.collections == nil {
		d.collections = map[string]collectionGenerator{}
	}

	d.collections[c.col] = c

	return d
}

func (c collectionGenerator) WithShard() ShardGeneratorInterface {
	return shardGenerator{
		col: c,
	}
}

func (c collectionGenerator) WithWriteConcern(wc int) CollectionGeneratorInterface {
	c.wc = &wc
	return c
}

func (c collectionGenerator) WithReplicationFactor(rf int) CollectionGeneratorInterface {
	c.rf = (*ReplicationFactor)(&rf)
	return c
}
