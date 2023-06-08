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

import (
	"fmt"
	"strings"
	"testing"

	"github.com/dchest/uniuri"
	"github.com/stretchr/testify/require"

	"github.com/arangodb/kube-arangodb/pkg/util"
)

func NewDatabaseRandomGenerator() DatabaseGeneratorInterface {
	return NewDatabaseGenerator(fmt.Sprintf("d%s", strings.ToLower(uniuri.NewLen(16))))
}

func NewDatabaseGenerator(name string) DatabaseGeneratorInterface {
	return databaseGenerator{
		db: name,
	}
}

type DatabaseGeneratorInterface interface {
	Collection(name string) CollectionGeneratorInterface
	RandomCollection() CollectionGeneratorInterface
	Add() Generator
}

type databaseGenerator struct {
	db string

	collections map[string]collectionGenerator
}

func (d databaseGenerator) RandomCollection() CollectionGeneratorInterface {
	return d.Collection(fmt.Sprintf("c%s", strings.ToLower(uniuri.NewLen(16))))
}

func (d databaseGenerator) Collection(name string) CollectionGeneratorInterface {
	return collectionGenerator{
		db:  d,
		col: name,
	}
}

func (d databaseGenerator) Add() Generator {
	return func(t *testing.T, s *State) {
		if s.Plan.Collections == nil {
			s.Plan.Collections = PlanCollections{}
		}

		if s.Current.Collections == nil {
			s.Current.Collections = CurrentCollections{}
		}

		_, ok := s.Plan.Collections[d.db]
		require.False(t, ok)

		_, ok = s.Current.Collections[d.db]
		require.False(t, ok)

		plan := PlanDBCollections{}
		current := CurrentDBCollections{}

		for col, colDet := range d.collections {
			planShards := Shards{}
			currentShards := CurrentDBCollection{}

			for shard, shardDet := range colDet.shards {
				n := fmt.Sprintf("s%d", shard)

				planShards[n] = shardDet.plan
				currentShards[n] = CurrentDBShard{Servers: shardDet.current}
			}

			planCol := PlanCollection{
				Name:              util.NewType[string](col),
				Shards:            planShards,
				WriteConcern:      colDet.wc,
				ReplicationFactor: colDet.rf,
			}

			plan[col] = planCol
			current[col] = currentShards
		}

		s.Plan.Collections[d.db] = plan
		s.Current.Collections[d.db] = current
	}
}
