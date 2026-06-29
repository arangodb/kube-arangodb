//
// DISCLAIMER
//
// Copyright 2016-2026 ArangoDB GmbH, Cologne, Germany
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

package transaction

import (
	"context"
	"fmt"
	goHttp "net/http"

	adbDriverV2 "github.com/arangodb/go-driver/v2/arangodb"
	adbDriverV2Shared "github.com/arangodb/go-driver/v2/arangodb/shared"

	"github.com/arangodb/kube-arangodb/pkg/util/errors"
)

type Options struct {
	Transient bool
}
type Transaction struct {
	keys       []KeyChanger
	conditions Conditions
	clientID   string
	options    Options
}

func NewTransaction(clientID string, options Options) (Transaction, error) {
	if clientID == "" {
		return Transaction{}, errors.New("clientID can not be empty")
	}
	return Transaction{
		clientID: clientID,
		options:  options,
	}, nil
}
func (k *Transaction) AddConditionByFullKey(fullKey string, condition KeyConditioner) error {
	if k.conditions == nil {
		k.conditions = make(Conditions)
	}
	if _, ok := k.conditions[fullKey]; ok {
		// For the time being one key can have only one condition.
		// It is a limitation in the agency.
		return fmt.Errorf("too many conditions")
	}
	k.conditions[fullKey] = condition
	return nil
}
func (t *Transaction) AddCondition(key Key, condition KeyConditioner) error {
	fullKey := CreateFullKey(key)
	return t.AddConditionByFullKey(fullKey, condition)
}
func (t *Transaction) AddKey(key KeyChanger) {
	t.keys = append(t.keys, key)
}
func (t *Transaction) GetType() string {
	if t.options.Transient {
		return "transient"
	}
	return "write"
}

var ErrPrecondition = errors.New("precondition failed")

func WriteTransaction(ctx context.Context, cli adbDriverV2.Client, transaction Transaction) (int64, error) {
	keysToChange := make(map[string]any)
	for _, v := range transaction.keys {
		keysToChange[v.GetKey()] = agencyUpdate{
			Operation: string(v.GetOperation()),
			New:       v.GetNew(),
			Val:       v.GetVal(),
		}
	}
	conditions := make(map[string]any)
	if transaction.conditions != nil {
		for key, condition := range transaction.conditions {
			conditions[key] = map[string]any{
				condition.GetName(): condition.GetValue(),
			}
		}
	}
	at := make(agencyTransaction, 0, 3)
	at = append(at, keysToChange)
	at = append(at, conditions)
	if len(transaction.clientID) > 0 {
		at = append(at, transaction.clientID)
	}
	var result agencyResult
	resp, err := cli.Post(ctx, &result, []agencyTransaction{at}, "_api", "agency", transaction.GetType())
	if err != nil {
		if adbDriverV2Shared.IsPreconditionFailed(err) {
			return 0, ErrPrecondition
		}
		return 0, err
	}
	switch resp.Code() {
	case goHttp.StatusOK, goHttp.StatusAccepted, goHttp.StatusCreated:
	// Do nothing.
	case goHttp.StatusPreconditionFailed:
		return 0, ErrPrecondition
	default:
		return 0, errors.Errorf("Unexpected response code %d", resp.Code())
	}
	if len(result.Results) != 1 {
		return 0, errors.Errorf("Unexpected results length %d, but expected 1", len(result.Results))
	}
	transactionID := result.Results[0]
	if transactionID == 0 {
		return 0, ErrPrecondition
	}
	return transactionID, nil
}

type agencyTransaction []any
type agencyUpdate struct {
	Operation string `json:"op,omitempty"`
	New       any    `json:"new,omitempty"`
	Val       any    `json:"val,omitempty"`
}
type agencyResult struct {
	Results []int64 `json:"results"`
}
