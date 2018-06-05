//
// Copyright 2017 ArangoDB GmbH, Cologne, Germany
//
// The Programs (which include both the software and documentation) contain
// proprietary information of ArangoDB GmbH; they are provided under a license
// agreement containing restrictions on use and disclosure and are also
// protected by copyright, patent and other intellectual and industrial
// property laws. Reverse engineering, disassembly or decompilation of the
// Programs, except to the extent required to obtain interoperability with
// other independently created software or as specified by law, is prohibited.
//
// It shall be the licensee's responsibility to take all appropriate fail-safe,
// backup, redundancy, and other measures to ensure the safe use of
// applications if the Programs are used for purposes such as nuclear,
// aviation, mass transit, medical, or other inherently dangerous applications,
// and ArangoDB GmbH disclaims liability for any damages caused by such use of
// the Programs.
//
// This software is the confidential and proprietary information of ArangoDB
// GmbH. You shall not disclose such confidential and proprietary information
// and shall use it only in accordance with the terms of the license agreement
// you entered into with ArangoDB GmbH.
//
// Author Ewout Prangsma
//

package client

import (
	"context"
	"time"

	"github.com/arangodb/arangosync/tasks"
	"github.com/pkg/errors"
)

// API of a sync master/worker
type API interface {
	// Close this client
	Close() error
	// Get the version of the sync master/worker
	Version(ctx context.Context) (VersionInfo, error)
	// Get the role of the sync master/worker
	Role(ctx context.Context) (Role, error)
	// Health performs a quick health check.
	// Returns an error when anything is wrong. If so, check Status.
	Health(ctx context.Context) error
	// Returns the master API (only valid when Role returns master)
	Master() MasterAPI
	// Returns the worker API (only valid when Role returns worker)
	Worker() WorkerAPI

	// Set the ID of the client that is making requests.
	SetClientID(id string)
	// SetShared marks the client as shared.
	// Closing a shared client will not close all idle connections.
	SetShared()
	// SynchronizeMasterEndpoints ensures that the client is using all known master
	// endpoints.
	// Do not use for connections to workers.
	// Returns true when endpoints have changed.
	SynchronizeMasterEndpoints(ctx context.Context) (bool, error)
	// Endpoint returns the currently used endpoint for this client.
	Endpoint() Endpoint
}

const (
	// ClientIDHeaderKey is the name of a request header containing the ID that is
	// making the request.
	ClientIDHeaderKey = "X-ArangoSync-Client-ID"
)

// MasterAPI contains API of sync master
type MasterAPI interface {
	// Gets the current status of synchronization towards the local cluster.
	Status(ctx context.Context) (SyncInfo, error)
	// Configure the master to synchronize the local cluster from a given remote cluster.
	Synchronize(ctx context.Context, input SynchronizationRequest) error
	// Configure the master to stop & completely cancel the current synchronization of the
	// local cluster from a remote cluster.
	// Errors:
	// - RequestTimeoutError when input.WaitTimeout is non-zero and the inactive stage is not reached in time.
	CancelSynchronization(ctx context.Context, input CancelSynchronizationRequest) (CancelSynchronizationResponse, error)
	// Reset a failed shard synchronization.
	ResetShardSynchronization(ctx context.Context, dbName, colName string, shardIndex int) error
	// Update the maximum allowed time between messages in a task channel.
	SetMessageTimeout(ctx context.Context, timeout time.Duration) error
	// Return a list of all known master endpoints of this datacenter.
	// The resulting endpoints are usable from inside and outside the datacenter.
	GetEndpoints(ctx context.Context) (Endpoint, error)
	// Return a list of master endpoints of the leader (syncmaster) of this datacenter.
	// Length of returned list will be 1 or the call will fail because no master is available.
	// In the very rare occasion that the leadership is changing during this call, a list
	// of length 0 can be returned.
	// The resulting endpoint is usable only within the same datacenter.
	GetLeaderEndpoint(ctx context.Context) (Endpoint, error)
	// Return a list of known masters in this datacenter.
	Masters(ctx context.Context) ([]MasterInfo, error)

	InternalMasterAPI
}

// WorkerAPI contains API of sync worker
type WorkerAPI interface {
	InternalWorkerAPI
}

type VersionInfo struct {
	Version string `json:"version"`
	Build   string `json:"build"`
}

// MasterInfo contains information about a single master.
type MasterInfo struct {
	// Unique identifier of the master
	ID string `json:"id"`
	// Internal endpoint of the master
	Endpoint string `json:"endpoint"`
	// Is this master the current leader
	Leader bool `json:"leader"`
}

type RoleInfo struct {
	Role Role `json:"role"`
}

type Role string

const (
	RoleMaster Role = "master"
	RoleWorker Role = "worker"
)

func (r Role) IsMaster() bool { return r == RoleMaster }
func (r Role) IsWorker() bool { return r == RoleWorker }

type ChannelPrefixInfo struct {
	Prefix string `json:"prefix"`
}

// SyncInfo holds the JSON info returned from `GET /_api/sync`
type SyncInfo struct {
	Source         Endpoint           `json:"source"`                   // Endpoint of sync master on remote cluster
	Status         SyncStatus         `json:"status"`                   // Overall status of (incoming) synchronization
	Shards         []ShardSyncInfo    `json:"shards,omitempty"`         // Status of incoming synchronization per shard
	Outgoing       []OutgoingSyncInfo `json:"outgoing,omitempty"`       // Status of outgoing synchronization
	MessageTimeout time.Duration      `json:"messageTimeout,omitempty"` // Maximum time between messages in a task channel
}

// OutgoingSyncInfo holds JSON info returned as part of `GET /_api/sync`
// regarding a specific target for outgoing synchronization data.
type OutgoingSyncInfo struct {
	ID       string          `json:"id"`               // ID of sync master to which data is being send
	Endpoint Endpoint        `json:"endpoint"`         // Endpoint of sync masters to which data is being send
	Status   SyncStatus      `json:"status"`           // Overall status for this outgoing target
	Shards   []ShardSyncInfo `json:"shards,omitempty"` // Status of outgoing synchronization per shard for this target
}

// ShardSyncInfo holds JSON info returned as part of `GET /_api/sync`
// regarding a specific shard.
type ShardSyncInfo struct {
	Database              string        `json:"database"`                 // Database containing the collection - shard
	Collection            string        `json:"collection"`               // Collection containing the shard
	ShardIndex            int           `json:"shardIndex"`               // Index of the shard (0..)
	Status                SyncStatus    `json:"status"`                   // Status of this shard
	StatusMessage         string        `json:"status_message,omitempty"` // Human readable message about the status of this shard
	Delay                 time.Duration `json:"delay,omitempty"`          // Delay between other datacenter and us.
	LastMessage           time.Time     `json:"last_message"`             // Time of last message received by the task handling this shard
	LastDataChange        time.Time     `json:"last_data_change"`         // Time of last message that resulted in a data change, received by the task handling this shard
	LastShardMasterChange time.Time     `json:"last_shard_master_change"` // Time of when we last had a change in the status of the shard master
	ShardMasterKnown      bool          `json:"shard_master_known"`       // Is the shard master known?
}

type SyncStatus string

const (
	// SyncStatusInactive indicates that no synchronization is taking place
	SyncStatusInactive SyncStatus = "inactive"
	// SyncStatusInitializing indicates that synchronization tasks are being setup
	SyncStatusInitializing SyncStatus = "initializing"
	// SyncStatusInitialSync indicates that initial synchronization of collections is ongoing
	SyncStatusInitialSync SyncStatus = "initial-sync"
	// SyncStatusRunning indicates that all collections have been initially synchronized
	// and normal transaction synchronization is active.
	SyncStatusRunning SyncStatus = "running"
	// SyncStatusCancelling indicates that the synchronization process is being cancelled.
	SyncStatusCancelling SyncStatus = "cancelling"
	// SyncStatusFailed indicates that the synchronization process has encountered an unrecoverable failure
	SyncStatusFailed SyncStatus = "failed"
)

var (
	// ValidSyncStatusValues is a list of all possible sync status values.
	ValidSyncStatusValues = []SyncStatus{
		SyncStatusInactive,
		SyncStatusInitializing,
		SyncStatusInitialSync,
		SyncStatusRunning,
		SyncStatusCancelling,
		SyncStatusFailed,
	}
)

// Normalize converts an empty status to inactive.
func (s SyncStatus) Normalize() SyncStatus {
	if s == "" {
		return SyncStatusInactive
	}
	return s
}

// Equals returns true when the other status is equal to the given
// status (both normalized).
func (s SyncStatus) Equals(other SyncStatus) bool {
	return s.Normalize() == other.Normalize()
}

// IsInactiveOrEmpty returns true if the given status equals inactive or is empty.
func (s SyncStatus) IsInactiveOrEmpty() bool {
	return s == SyncStatusInactive || s == ""
}

// IsInitialSyncOrRunning returns true if the given status equals initial-sync or running.
func (s SyncStatus) IsInitialSyncOrRunning() bool {
	return s == SyncStatusInitialSync || s == SyncStatusRunning
}

// IsActive returns true if the given status indicates an active state.
// The is: initializing, initial-sync or running
func (s SyncStatus) IsActive() bool {
	return s == SyncStatusInitializing || s == SyncStatusInitialSync || s == SyncStatusRunning
}

//
// TLSAuthentication contains configuration for using client certificates
// and TLS verification of the server.
type TLSAuthentication = tasks.TLSAuthentication

type SynchronizationRequest struct {
	// Endpoint of sync master of the source cluster
	Source Endpoint `json:"source"`
	// Authentication of the master
	Authentication TLSAuthentication `json:"authentication"`
}

// Clone returns a deep copy of the given request.
func (r SynchronizationRequest) Clone() SynchronizationRequest {
	c := r
	c.Source = r.Source.Clone()
	return c
}

// IsSame returns true if both requests contain the same values.
// The source is considered the same is the intersection of existing & given source is not empty.
// We consider an intersection because:
// - Servers can be down, resulting in a temporary missing endpoint
// - Customer can specify only 1 of all servers
func (r SynchronizationRequest) IsSame(other SynchronizationRequest) bool {
	if r.Source.Intersection(other.Source).IsEmpty() {
		return false
	}
	if r.Authentication.ClientCertificate != other.Authentication.ClientCertificate {
		return false
	}
	if r.Authentication.ClientKey != other.Authentication.ClientKey {
		return false
	}
	if r.Authentication.CACertificate != other.Authentication.CACertificate {
		return false
	}
	return true
}

// Validate checks the values of the given request and returns an error
// in case of improper values.
// Returns nil on success.
func (r SynchronizationRequest) Validate() error {
	if len(r.Source) == 0 {
		return errors.Wrap(BadRequestError, "source missing")
	}
	if err := r.Source.Validate(); err != nil {
		return errors.Wrapf(BadRequestError, "Invalid source: %s", err.Error())
	}
	if r.Authentication.ClientCertificate == "" {
		return errors.Wrap(BadRequestError, "clientCertificate missing")
	}
	if r.Authentication.ClientKey == "" {
		return errors.Wrap(BadRequestError, "clientKey missing")
	}
	if r.Authentication.CACertificate == "" {
		return errors.Wrap(BadRequestError, "caCertificate missing")
	}
	return nil
}

type CancelSynchronizationRequest struct {
	// WaitTimeout is the amount of time the cancel function will wait
	// until the synchronization has reached an `inactive` state.
	// If this value is zero, the cancel function will only switch to the canceling state
	// but not wait until the `inactive` state is reached.
	WaitTimeout time.Duration `json:"wait_timeout,omitempty"`
	// Force is set if you want to end the synchronization even if the source
	// master cannot be reached.
	Force bool `json:"force,omitempty"`
	// ForceTimeout is the amount of time the syncmaster tries to contact
	// the source master to notify it about cancelling the synchronization.
	// This fields is only used when Force is true.
	ForceTimeout time.Duration `json:"force_timeout,omitempty"`
}

type CancelSynchronizationResponse struct {
	// Aborted is set when synchronization has cancelled (state is now inactive)
	// but the source sync master was not notified.
	// This is only possible when the Force flags is set on the request.
	Aborted bool `json:"aborted,omitempty"`
	// Source is the endpoint of sync master on remote cluster that we used
	// to be synchronizing from.
	Source Endpoint `json:"source,omitempty"`
	// ClusterID is the ID of the local synchronization cluster.
	ClusterID string `json:"cluster_id,omitempty"`
}

type SetMessageTimeoutRequest struct {
	MessageTimeout time.Duration `json:"messageTimeout"`
}

type EndpointsResponse struct {
	Endpoints Endpoint `json:"endpoints"`
}

type MastersResponse struct {
	Masters []MasterInfo `json:"masters"`
}
