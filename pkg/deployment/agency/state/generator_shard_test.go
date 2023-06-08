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

type ShardGeneratorInterface interface {
	WithPlan(servers ...Server) ShardGeneratorInterface
	WithCurrent(servers ...Server) ShardGeneratorInterface
	Add() CollectionGeneratorInterface
}

type shardGenerator struct {
	col collectionGenerator

	plan    Servers
	current Servers
}

func (s shardGenerator) WithPlan(servers ...Server) ShardGeneratorInterface {
	s.plan = servers
	return s
}

func (s shardGenerator) WithCurrent(servers ...Server) ShardGeneratorInterface {
	s.current = servers
	return s
}

func (s shardGenerator) Add() CollectionGeneratorInterface {
	c := s.col

	if c.shards == nil {
		c.shards = map[int]shardGenerator{}
	}

	c.shards[id()] = s

	return c
}
