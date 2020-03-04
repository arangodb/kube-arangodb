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

import (
	"context"
	"fmt"
	"io"
	"math/rand"
	"os"
	"sort"
	"sync"
	"sync/atomic"
	"time"

	driver "github.com/arangodb/go-driver"
	"github.com/pkg/errors"
	"github.com/rs/zerolog"

	"github.com/arangodb/kube-arangodb/tests/duration/test"
)

type SimpleConfig struct {
	MaxDocuments   int
	MaxCollections int
}

const (
	initialDocumentCount = 999
)

type simpleTest struct {
	SimpleConfig
	activeMutex                         sync.Mutex
	logPath                             string
	reportDir                           string
	log                                 zerolog.Logger
	listener                            test.TestListener
	stop                                chan struct{}
	active                              bool
	pauseRequested                      bool
	paused                              bool
	client                              driver.Client
	db                                  driver.Database
	failures                            int
	actions                             int
	collections                         map[string]*collection
	collectionsMutex                    sync.Mutex
	lastCollectionIndex                 int32
	readExistingCounter                 counter
	readExistingWrongRevisionCounter    counter
	readNonExistingCounter              counter
	createCounter                       counter
	createCollectionCounter             counter
	removeExistingCollectionCounter     counter
	updateExistingCounter               counter
	updateExistingWrongRevisionCounter  counter
	updateNonExistingCounter            counter
	replaceExistingCounter              counter
	replaceExistingWrongRevisionCounter counter
	replaceNonExistingCounter           counter
	deleteExistingCounter               counter
	deleteExistingWrongRevisionCounter  counter
	deleteNonExistingCounter            counter
	importCounter                       counter
	queryCreateCursorCounter            counter
	queryNextBatchCounter               counter
	queryNextBatchNewCoordinatorCounter counter
	queryLongRunningCounter             counter
	rebalanceShardsCounter              counter
	queryUpdateCounter                  counter
	queryUpdateLongRunningCounter       counter
}

type counter struct {
	succeeded int
	failed    int
}

type collection struct {
	name         string
	existingDocs map[string]UserDocument
}

// NewSimpleTest creates a simple test
func NewSimpleTest(log zerolog.Logger, reportDir string, config SimpleConfig) test.TestScript {
	return &simpleTest{
		SimpleConfig: config,
		reportDir:    reportDir,
		log:          log,
		collections:  make(map[string]*collection),
	}
}

// Name returns the name of the script
func (t *simpleTest) Name() string {
	return "simple"
}

// Start triggers the test script to start.
// It should spwan actions in a go routine.
func (t *simpleTest) Start(client driver.Client, listener test.TestListener) error {
	t.activeMutex.Lock()
	defer t.activeMutex.Unlock()

	if t.active {
		// No restart unless needed
		return nil
	}

	t.listener = listener
	t.client = client
	ctx := context.Background()
	db, err := client.Database(ctx, "_system")
	if err != nil {
		return maskAny(err)
	}
	t.db = db

	// Cleanup of old data
	for i := 1; i <= t.MaxCollections; i++ {
		col, err := db.Collection(ctx, t.getCollectionName(i))
		if err == nil {
			if err := col.Remove(ctx); err != nil {
				return errors.Wrapf(err, "Failed to remove collection %s", col.Name())
			}
		} else if !driver.IsNotFound(err) {
			return maskAny(err)
		}
	}

	t.active = true
	go t.testLoop()
	return nil
}

// Stop any running test. This should not return until tests are actually stopped.
func (t *simpleTest) Stop() error {
	t.activeMutex.Lock()
	defer t.activeMutex.Unlock()

	if !t.active {
		// No active, nothing to stop
		return nil
	}

	stop := make(chan struct{})
	t.stop = stop
	<-stop
	return nil
}

// Interrupt the tests, but be prepared to continue.
func (t *simpleTest) Pause() error {
	t.pauseRequested = true
	return nil
}

// Resume running the tests, where Pause interrupted it.
func (t *simpleTest) Resume() error {
	t.pauseRequested = false
	return nil
}

// Status returns the current status of the test
func (t *simpleTest) Status() test.TestStatus {
	cc := func(name string, c counter) test.Counter {
		return test.Counter{
			Name:      name,
			Succeeded: c.succeeded,
			Failed:    c.failed,
		}
	}

	status := test.TestStatus{
		Active:   t.active && !t.paused,
		Pausing:  t.pauseRequested && t.paused,
		Failures: t.failures,
		Actions:  t.actions,
		Counters: []test.Counter{
			cc("#collections created", t.createCollectionCounter),
			cc("#collections removed", t.removeExistingCollectionCounter),
			cc("#documents created", t.createCounter),
			cc("#existing documents read", t.readExistingCounter),
			cc("#existing documents updated", t.updateExistingCounter),
			cc("#existing documents replaced", t.replaceExistingCounter),
			cc("#existing documents removed", t.deleteExistingCounter),
			cc("#existing documents wrong revision read", t.readExistingWrongRevisionCounter),
			cc("#existing documents wrong revision updated", t.updateExistingWrongRevisionCounter),
			cc("#existing documents wrong revision replaced", t.replaceExistingWrongRevisionCounter),
			cc("#existing documents wrong revision removed", t.deleteExistingWrongRevisionCounter),
			cc("#non-existing documents read", t.readNonExistingCounter),
			cc("#non-existing documents updated", t.updateNonExistingCounter),
			cc("#non-existing documents replaced", t.replaceNonExistingCounter),
			cc("#non-existing documents removed", t.deleteNonExistingCounter),
			cc("#import operations", t.importCounter),
			cc("#create AQL cursor operations", t.queryCreateCursorCounter),
			cc("#fetch next AQL cursor batch operations", t.queryNextBatchCounter),
			cc("#fetch next AQL cursor batch after coordinator change operations", t.queryNextBatchNewCoordinatorCounter),
			cc("#long running AQL query operations", t.queryLongRunningCounter),
			cc("#rebalance shards operations", t.rebalanceShardsCounter),
			cc("#update AQL query operations", t.queryUpdateCounter),
			cc("#long running update AQL query operations", t.queryUpdateLongRunningCounter),
		},
	}

	t.collectionsMutex.Lock()
	for _, c := range t.collections {
		status.Messages = append(status.Messages,
			fmt.Sprintf("Current #documents in %s: %d", c.name, len(c.existingDocs)),
		)
	}
	t.collectionsMutex.Unlock()

	return status
}

// CollectLogs copies all logging info to the given writer.
func (t *simpleTest) CollectLogs(w io.Writer) error {
	if logPath := t.logPath; logPath == "" {
		// Nothing to log yet
		return nil
	} else {
		rd, err := os.Open(logPath)
		if err != nil {
			return maskAny(err)
		}
		defer rd.Close()
		if _, err := io.Copy(w, rd); err != nil {
			return maskAny(err)
		}
		return nil
	}
}

func (t *simpleTest) shouldStop() bool {
	// Should we stop?
	if stop := t.stop; stop != nil {
		stop <- struct{}{}
		return true
	}
	return false
}

type UserDocument struct {
	Key   string `json:"_key"`
	rev   string // Note that we do not export this field!
	Value int    `json:"value"`
	Name  string `json:"name"`
	Odd   bool   `json:"odd"`
}

// Equals returns true when the value fields of `d` and `other` are the equal.
func (d UserDocument) Equals(other UserDocument) bool {
	return d.Value == other.Value &&
		d.Name == other.Name &&
		d.Odd == other.Odd
}

func (t *simpleTest) reportFailure(f test.Failure) {
	t.failures++
	t.listener.ReportFailure(f)
}

func (t *simpleTest) testLoop() {
	t.active = true
	t.actions = 0
	defer func() { t.active = false }()

	if err := t.createAndInitCollection(); err != nil {
		t.log.Error().Msgf("Failed to create&init first collection: %v. Giving up", err)
		return
	}

	var plan []int
	planIndex := 0
	for {
		// Should we stop
		if t.shouldStop() {
			return
		}
		if t.pauseRequested {
			t.paused = true
			time.Sleep(time.Second * 2)
			continue
		}
		t.paused = false
		t.actions++
		if plan == nil || planIndex >= len(plan) {
			plan = createTestPlan(20) // Update when more tests are added
			planIndex = 0
		}

		switch plan[planIndex] {
		case 0:
			// Create collection with initial data
			if len(t.collections) < t.MaxCollections && rand.Intn(100)%2 == 0 {
				if err := t.createAndInitCollection(); err != nil {
					t.log.Error().Msgf("Failed to create&init collection: %v", err)
				}
			}
			planIndex++

		case 1:
			// Remove an existing collection
			if len(t.collections) > 1 && rand.Intn(100)%2 == 0 {
				c := t.selectRandomCollection()
				if err := t.removeExistingCollection(c); err != nil {
					t.log.Error().Msgf("Failed to remove existing collection: %#v", err)
				}
			}
			planIndex++

		case 2:
			// Create a random document
			if len(t.collections) > 0 {
				c := t.selectRandomCollection()
				if len(c.existingDocs) < t.MaxDocuments {
					userDoc := UserDocument{
						Key:   c.createNewKey(true),
						Value: rand.Int(),
						Name:  fmt.Sprintf("User %d", time.Now().Nanosecond()),
						Odd:   time.Now().Nanosecond()%2 == 1,
					}
					if rev, err := t.createDocument(c, userDoc, userDoc.Key); err != nil {
						t.log.Error().Msgf("Failed to create document: %#v", err)
					} else {
						userDoc.rev = rev
						c.existingDocs[userDoc.Key] = userDoc

						// Now try to read it, it must exist
						//t.client.SetCoordinator("")
						if _, err := t.readExistingDocument(c, userDoc.Key, rev, false, false); err != nil {
							t.log.Error().Msgf("Failed to read just-created document '%s': %#v", userDoc.Key, err)
						}
					}
				}
			}
			planIndex++

		case 3:
			// Read a random existing document
			if len(t.collections) > 0 {
				c := t.selectRandomCollection()
				if len(c.existingDocs) > 0 {
					randomKey, rev := c.selectRandomKey()
					if _, err := t.readExistingDocument(c, randomKey, rev, false, false); err != nil {
						t.log.Error().Msgf("Failed to read existing document '%s': %#v", randomKey, err)
					}
				}
			}
			planIndex++

		case 4:
			// Read a random existing document but with wrong revision
			planIndex++

		case 5:
			// Read a random non-existing document
			if len(t.collections) > 0 {
				c := t.selectRandomCollection()
				randomKey := c.createNewKey(false)
				if err := t.readNonExistingDocument(c.name, randomKey); err != nil {
					t.log.Error().Msgf("Failed to read non-existing document '%s': %#v", randomKey, err)
				}
			}
			planIndex++

		case 6:
			// Remove a random existing document
			if len(t.collections) > 0 {
				c := t.selectRandomCollection()
				if len(c.existingDocs) > 0 {
					randomKey, rev := c.selectRandomKey()
					if err := t.removeExistingDocument(c.name, randomKey, rev); err != nil {
						t.log.Error().Msgf("Failed to remove existing document '%s': %#v", randomKey, err)
					} else {
						// Remove succeeded, key should no longer exist
						c.removeExistingKey(randomKey)

						// Now try to read it, it should not exist
						//t.client.SetCoordinator("")
						if err := t.readNonExistingDocument(c.name, randomKey); err != nil {
							t.log.Error().Msgf("Failed to read just-removed document '%s': %#v", randomKey, err)
						}
					}
				}
			}
			planIndex++

		case 7:
			// Remove a random existing document but with wrong revision
			planIndex++

		case 8:
			// Remove a random non-existing document
			if len(t.collections) > 0 {
				c := t.selectRandomCollection()
				randomKey := c.createNewKey(false)
				if err := t.removeNonExistingDocument(c.name, randomKey); err != nil {
					t.log.Error().Msgf("Failed to remove non-existing document '%s': %#v", randomKey, err)
				}
			}
			planIndex++

		case 9:
			// Update a random existing document
			if len(t.collections) > 0 {
				c := t.selectRandomCollection()
				if len(c.existingDocs) > 0 {
					randomKey, rev := c.selectRandomKey()
					if newRev, err := t.updateExistingDocument(c, randomKey, rev); err != nil {
						t.log.Error().Msgf("Failed to update existing document '%s': %#v", randomKey, err)
					} else {
						// Updated succeeded, now try to read it, it should exist and be updated
						//t.client.SetCoordinator("")
						if _, err := t.readExistingDocument(c, randomKey, newRev, false, false); err != nil {
							t.log.Error().Msgf("Failed to read just-updated document '%s': %#v", randomKey, err)
						}
					}
				}
			}
			planIndex++

		case 10:
			// Update a random existing document but with wrong revision
			planIndex++

		case 11:
			// Update a random non-existing document
			if len(t.collections) > 0 {
				c := t.selectRandomCollection()
				randomKey := c.createNewKey(false)
				if err := t.updateNonExistingDocument(c.name, randomKey); err != nil {
					t.log.Error().Msgf("Failed to update non-existing document '%s': %#v", randomKey, err)
				}
			}
			planIndex++

		case 12:
			// Replace a random existing document
			if len(t.collections) > 0 {
				c := t.selectRandomCollection()
				if len(c.existingDocs) > 0 {
					randomKey, rev := c.selectRandomKey()
					if newRev, err := t.replaceExistingDocument(c, randomKey, rev); err != nil {
						t.log.Error().Msgf("Failed to replace existing document '%s': %#v", randomKey, err)
					} else {
						// Replace succeeded, now try to read it, it should exist and be replaced
						//t.client.SetCoordinator("")
						if _, err := t.readExistingDocument(c, randomKey, newRev, false, false); err != nil {
							t.log.Error().Msgf("Failed to read just-replaced document '%s': %#v", randomKey, err)
						}
					}
				}
			}
			planIndex++

		case 13:
			// Replace a random existing document but with wrong revision
			planIndex++

		case 14:
			// Replace a random non-existing document
			if len(t.collections) > 0 {
				c := t.selectRandomCollection()
				randomKey := c.createNewKey(false)
				if err := t.replaceNonExistingDocument(c.name, randomKey); err != nil {
					t.log.Error().Msgf("Failed to replace non-existing document '%s': %#v", randomKey, err)
				}
			}
			planIndex++

		case 15:
			// Query documents
			planIndex++

		case 16:
			// Query documents (long running)
			if len(t.collections) > 0 {
				c := t.selectRandomCollection()
				if err := t.queryDocumentsLongRunning(c); err != nil {
					t.log.Error().Msgf("Failed to query (long running) documents: %#v", err)
				}
			}
			planIndex++

		case 17:
			// Rebalance shards
			if err := t.rebalanceShards(); err != nil {
				t.log.Error().Msgf("Failed to rebalance shards: %#v", err)
			}
			planIndex++

		case 18:
			// AQL update query
			if len(t.collections) > 0 {
				c := t.selectRandomCollection()
				if len(c.existingDocs) > 0 {
					randomKey, _ := c.selectRandomKey()
					if newRev, err := t.queryUpdateDocuments(c, randomKey); err != nil {
						t.log.Error().Msgf("Failed to update document using AQL query: %#v", err)
					} else {
						// Updated succeeded, now try to read it (anywhere), it should exist and be updated
						//t.client.SetCoordinator("")
						if _, err := t.readExistingDocument(c, randomKey, newRev, false, false); err != nil {
							t.log.Error().Msgf("Failed to read just-updated document '%s': %#v", randomKey, err)
						}
					}
				}
			}
			planIndex++

		case 19:
			// Long running AQL update query
			if len(t.collections) > 0 {
				c := t.selectRandomCollection()
				if len(c.existingDocs) > 0 {
					randomKey, _ := c.selectRandomKey()
					if newRev, err := t.queryUpdateDocumentsLongRunning(c, randomKey); err != nil {
						t.log.Error().Msgf("Failed to update document using long running AQL query: %#v", err)
					} else {
						// Updated succeeded, now try to read it (anywhere), it should exist and be updated
						//t.client.SetCoordinator("")
						if _, err := t.readExistingDocument(c, randomKey, newRev, false, false); err != nil {
							t.log.Error().Msgf("Failed to read just-updated document '%s': %#v", randomKey, err)
						}
					}
				}
			}
			planIndex++
		}
		time.Sleep(time.Second * 2)
	}
}

// createTestPlan creates an int-array of 'steps' long with all values from 0..steps-1 in random order.
func createTestPlan(steps int) []int {
	plan := make([]int, steps)
	for i := 0; i < steps; i++ {
		plan[i] = i
	}
	test.Shuffle(sort.IntSlice(plan))
	return plan
}

// createNewCollectionName returns a new (unique) collection name
func (t *simpleTest) createNewCollectionName() string {
	index := atomic.AddInt32(&t.lastCollectionIndex, 1)
	return t.getCollectionName(int(index))
}

// getCollectionName returns a collection name with given index
func (t *simpleTest) getCollectionName(index int) string {
	return fmt.Sprintf("simple_user_%d", index)
}

func (t *simpleTest) selectRandomCollection() *collection {
	index := rand.Intn(len(t.collections))
	for _, c := range t.collections {
		if index == 0 {
			return c
		}
		index--
	}
	return nil // This should never be reached when len(t.collections) > 0
}

func (t *simpleTest) registerCollection(c *collection) {
	t.collectionsMutex.Lock()
	defer t.collectionsMutex.Unlock()
	t.collections[c.name] = c
}

func (t *simpleTest) unregisterCollection(c *collection) {
	t.collectionsMutex.Lock()
	defer t.collectionsMutex.Unlock()
	delete(t.collections, c.name)
}

func (t *simpleTest) createAndInitCollection() error {
	c := &collection{
		name:         t.createNewCollectionName(),
		existingDocs: make(map[string]UserDocument),
	}
	if err := t.createCollection(c, 9, 2); err != nil {
		t.reportFailure(test.NewFailure("Creating collection '%s' failed: %v", c.name, err))
		return maskAny(err)
	}
	t.registerCollection(c)
	t.createCollectionCounter.succeeded++
	t.actions++

	// Import documents
	if err := t.importDocuments(c); err != nil {
		t.reportFailure(test.NewFailure("Failed to import documents: %#v", err))
	}
	t.actions++

	// Check imported documents
	for k := range c.existingDocs {
		if t.shouldStop() || t.pauseRequested {
			return nil
		}
		if _, err := t.readExistingDocument(c, k, "", true, false); err != nil {
			t.reportFailure(test.NewFailure("Failed to read existing document '%s': %#v", k, err))
		}
		t.actions++
	}

	// Create sample users
	for i := 0; i < initialDocumentCount; i++ {
		if t.shouldStop() || t.pauseRequested {
			return nil
		}
		userDoc := UserDocument{
			Key:   fmt.Sprintf("doc%05d", i),
			Value: i,
			Name:  fmt.Sprintf("User %d", i),
			Odd:   i%2 == 1,
		}
		if rev, err := t.createDocument(c, userDoc, userDoc.Key); err != nil {
			t.reportFailure(test.NewFailure("Failed to create document: %#v", err))
		} else {
			userDoc.rev = rev
			c.existingDocs[userDoc.Key] = userDoc
		}
		t.actions++
	}
	return nil
}

func (c *collection) createNewKey(record bool) string {
	for {
		key := fmt.Sprintf("newkey%07d", rand.Int31n(100*1000))
		_, found := c.existingDocs[key]
		if !found {
			if record {
				c.existingDocs[key] = UserDocument{}
			}
			return key
		}
	}
}

func (c *collection) removeExistingKey(key string) {
	delete(c.existingDocs, key)
}

func (c *collection) selectRandomKey() (string, string) {
	index := rand.Intn(len(c.existingDocs))
	for k, v := range c.existingDocs {
		if index == 0 {
			return k, v.rev
		}
		index--
	}
	return "", "" // This should never be reached when len(t.existingDocs) > 0
}

func (c *collection) selectWrongRevision(key string) (string, bool) {
	correctRev := c.existingDocs[key].rev
	for _, v := range c.existingDocs {
		if v.rev != correctRev && v.rev != "" {
			return v.rev, true
		}
	}
	return "", false // This should never be reached when len(t.existingDocs) > 1
}
