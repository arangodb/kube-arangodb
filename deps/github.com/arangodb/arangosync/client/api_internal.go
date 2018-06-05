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
	"encoding/json"
	"fmt"
	"math"
	"time"

	"github.com/arangodb/arangosync/tasks"
)

// InternalMasterAPI contains the internal API of the sync master.
type InternalMasterAPI interface {
	// Worker -> Master

	// Load configuration data from the master
	ConfigureWorker(ctx context.Context, endpoint string) (WorkerConfiguration, error)
	// Return all registered workers
	RegisteredWorkers(ctx context.Context) ([]WorkerRegistration, error)
	// Return info about a specific worker
	RegisteredWorker(ctx context.Context, id string) (WorkerRegistration, error)
	// Register (or update registration of) a worker
	RegisterWorker(ctx context.Context, endpoint, token, hostID string) (WorkerRegistrationResponse, error)
	// Remove the registration of a worker
	UnregisterWorker(ctx context.Context, id string) error
	// Get info about a specific task
	Task(ctx context.Context, id string) (TaskInfo, error)
	// Get all known tasks
	Tasks(ctx context.Context) ([]TaskInfo, error)
	// Get all known tasks for a given channel
	TasksByChannel(ctx context.Context, channelName string) ([]TaskInfo, error)
	// Notify the master that a task with given ID has completed.
	TaskCompleted(ctx context.Context, taskID string, info TaskCompletedRequest) error
	// Create tasks to start synchronization of a shard in the given db+col.
	SynchronizeShard(ctx context.Context, dbName, colName string, shardIndex int) error
	// Stop tasks to synchronize a shard in the given db+col.
	CancelSynchronizeShard(ctx context.Context, dbName, colName string, shardIndex int) error
	// Report status of the synchronization of a shard back to the master.
	SynchronizeShardStatus(ctx context.Context, entries []SynchronizationShardStatusRequestEntry) error
	// IsChannelRelevant checks if a MQ channel is still relevant
	IsChannelRelevant(ctx context.Context, channelName string) (bool, error)

	// Worker & Master -> Master
	// GetDirectMQTopicEndpoint returns an endpoint that the caller can use to fetch direct MQ messages
	// from.
	// This method requires a directMQ token or client cert for authentication.
	GetDirectMQTopicEndpoint(ctx context.Context, channelName string) (DirectMQTopicEndpoint, error)
	// RenewDirectMQToken renews a given direct MQ token.
	// This method requires a directMQ token for authentication.
	RenewDirectMQToken(ctx context.Context, token string) (DirectMQToken, error)
	// CloneDirectMQToken creates a clone of a given direct MQ token.
	// When the given token is revoked, the newly cloned token is also revoked.
	// This method requires a directMQ token for authentication.
	CloneDirectMQToken(ctx context.Context, token string) (DirectMQToken, error)
	// Add entire direct MQ API
	InternalDirectMQAPI

	// Master -> Master

	// Start a task that sends inventory data to a receiving remote cluster.
	OutgoingSynchronization(ctx context.Context, input OutgoingSynchronizationRequest) (OutgoingSynchronizationResponse, error)
	// Cancel sending synchronization data to the remote cluster with given ID.
	CancelOutgoingSynchronization(ctx context.Context, remoteID string) error
	// Create tasks to send synchronization data of a shard in the given db+col to a remote cluster.
	OutgoingSynchronizeShard(ctx context.Context, remoteID, dbName, colName string, shardIndex int, input OutgoingSynchronizeShardRequest) error
	// Stop tasks to send synchronization data of a shard in the given db+col to a remote cluster.
	CancelOutgoingSynchronizeShard(ctx context.Context, remoteID, dbName, colName string, shardIndex int) error
	// Report status of the synchronization of a shard back to the master.
	OutgoingSynchronizeShardStatus(ctx context.Context, entries []SynchronizationShardStatusRequestEntry) error
	// Reset a failed shard synchronization.
	OutgoingResetShardSynchronization(ctx context.Context, remoteID, dbName, colName string, shardIndex int, newControlChannel, newDataChannel string) error

	// Get a prefix for names of channels that contain message
	// going to this master.
	ChannelPrefix(ctx context.Context) (string, error)
	// Get the local message queue configuration.
	GetMessageQueueConfig(ctx context.Context) (MessageQueueConfig, error)
}

// InternalWorkerAPI contains the internal API of the sync worker.
type InternalWorkerAPI interface {
	// StartTask is called by the master to instruct the worker
	// to run a task with given instructions.
	StartTask(ctx context.Context, data StartTaskRequest) error
	// StopTask is called by the master to instruct the worker
	// to stop all work on the given task.
	StopTask(ctx context.Context, taskID string) error
	// SetDirectMQTopicToken configures the token used to access messages of a given channel.
	SetDirectMQTopicToken(ctx context.Context, channelName, token string, tokenTTL time.Duration) error
	// Add entire direct MQ API
	InternalDirectMQAPI
}

// InternalDirectMQAPI contains the internal API of the sync master/worker wrt direct MQ messages.
type InternalDirectMQAPI interface {
	// GetDirectMQMessages return messages for a given MQ channel.
	GetDirectMQMessages(ctx context.Context, channelName string) ([]DirectMQMessage, error)
	// CommitDirectMQMessage removes all messages from the given channel up to an including the given offset.
	CommitDirectMQMessage(ctx context.Context, channelName string, offset int64) error
}

// MessageQueueConfig contains all deployment configuration info for the local MQ.
type MessageQueueConfig struct {
	Type           string            `json:"type"`
	Endpoints      []string          `json:"endpoints"`
	Authentication TLSAuthentication `json:"authentication"`
}

// Clone returns a deep copy of the given config
func (c MessageQueueConfig) Clone() MessageQueueConfig {
	result := c
	result.Endpoints = append([]string{}, c.Endpoints...)
	return result
}

// ConfigureWorkerRequest is the JSON body for the ConfigureWorker request.
type ConfigureWorkerRequest struct {
	Endpoint string `json:"endpoint"` // Endpoint of the worker
}

// WorkerConfiguration contains configuration data passed from
// the master to the worker.
type WorkerConfiguration struct {
	Cluster struct {
		Endpoints       []string `json:"endpoints"`
		JWTSecret       string   `json:"jwtSecret,omitempty"`
		MaxDocumentSize int      `json:"maxDocumentSize,omitempty"`
		// Minimum replication factor of new/modified collections
		MinReplicationFactor int `json:"min-replication-factor,omitempty"`
		// Maximum replication factor of new/modified collections
		MaxReplicationFactor int `json:"max-replication-factor,omitempty"`
	} `json:"cluster"`
	HTTPServer struct {
		Certificate string `json:"certificate"`
		Key         string `json:"key"`
	} `json:"httpServer"`
	MessageQueue struct {
		MessageQueueConfig // MQ configuration of local MQ
	} `json:"mq"`
}

// SetDefaults fills empty values with defaults
func (c *WorkerConfiguration) SetDefaults() {
	if c.Cluster.MinReplicationFactor <= 0 {
		c.Cluster.MinReplicationFactor = 1
	}
	if c.Cluster.MaxReplicationFactor <= 0 {
		c.Cluster.MaxReplicationFactor = math.MaxInt32
	}
}

// Validate the given configuration.
// Return an error on validation errors, nil when all ok.
func (c WorkerConfiguration) Validate() error {
	if c.Cluster.MinReplicationFactor < 1 {
		return maskAny(fmt.Errorf("MinReplicationFactor must be >= 1"))
	}
	if c.Cluster.MaxReplicationFactor < 1 {
		return maskAny(fmt.Errorf("MaxReplicationFactor must be >= 1"))
	}
	if c.Cluster.MaxReplicationFactor < c.Cluster.MinReplicationFactor {
		return maskAny(fmt.Errorf("MaxReplicationFactor must be >= MinReplicationFactor"))
	}
	return nil
}

type WorkerRegistrations struct {
	Workers []WorkerRegistration `json:"workers"`
}

type WorkerRegistration struct {
	// ID of the worker assigned to it by the master
	ID string `json:"id"`
	// Endpoint of the worker
	Endpoint string `json:"endpoint"`
	// Expiration time of the last registration of the worker
	ExpiresAt time.Time `json:"expiresAt"`
	// ID of the worker when communicating with ArangoDB servers.
	ServerID int64 `json:"serverID"`
	// IF of the host the worker process is running on
	HostID string `json:"host,omitempty"`
}

// Validate the given registration.
// Return nil if ok, error otherwise.
func (wr WorkerRegistration) Validate() error {
	if wr.ID == "" {
		return maskAny(fmt.Errorf("ID empty"))
	}
	if wr.Endpoint == "" {
		return maskAny(fmt.Errorf("Endpoint empty"))
	}
	if wr.ServerID == 0 {
		return maskAny(fmt.Errorf("ServerID == 0"))
	}
	return nil
}

// IsExpired returns true when the given worker is expired.
func (wr WorkerRegistration) IsExpired() bool {
	return time.Now().After(wr.ExpiresAt)
}

type WorkerRegistrationRequest struct {
	Endpoint string `json:"endpoint"`
	Token    string `json:"token,omitempty"`
	HostID   string `json:host,omitempty"`
}

type WorkerRegistrationResponse struct {
	WorkerRegistration
	// Maximum time between message in a task channel.
	MessageTimeout time.Duration `json:"messageTimeout,omitempty"`
}

type StartTaskRequest struct {
	ID string `json:"id"`
	tasks.TaskData
	// MQ configuration of the remote cluster
	RemoteMessageQueueConfig MessageQueueConfig `json:"remote-mq-config"`
}

// OutgoingSynchronizationRequest holds the master->master request
// data for configuring an outgoing inventory stream.
type OutgoingSynchronizationRequest struct {
	// ID of remote cluster
	ID string `json:"id"`
	// Endpoints of sync masters of the remote (target) cluster
	Target   Endpoint `json:"target"`
	Channels struct {
		// Name of MQ topic to send inventory data to.
		Inventory string `json:"inventory"`
	} `json:"channels"`
	// MQ configuration of the remote (target) cluster
	MessageQueueConfig MessageQueueConfig `json:"mq-config"`
}

// Clone returns a deep copy of the given request.
func (r OutgoingSynchronizationRequest) Clone() OutgoingSynchronizationRequest {
	c := r
	c.Target = r.Target.Clone()
	c.MessageQueueConfig = r.MessageQueueConfig.Clone()
	return c
}

// OutgoingSynchronizationResponse holds the answer to an
// master->master request for configuring an outgoing synchronization.
type OutgoingSynchronizationResponse struct {
	// MQ configuration of the remote (source) cluster
	MessageQueueConfig MessageQueueConfig `json:"mq-config"`
}

// OutgoingSynchronizeShardRequest holds the master->master request
// data for configuring an outgoing shard synchronization stream.
type OutgoingSynchronizeShardRequest struct {
	Channels struct {
		// Name of MQ topic to receive control messages on.
		Control string `json:"control"`
		// Name of MQ topic to send data messages to.
		Data string `json:"data"`
	} `json:"channels"`
}

// SynchronizationShardStatusRequest is the request body of a (Outgoing)SynchronizationShardStatus request.
type SynchronizationShardStatusRequest struct {
	Entries []SynchronizationShardStatusRequestEntry `json:"entries"`
}

// SynchronizationShardStatusRequestEntry is a single entry in a SynchronizationShardStatusRequest
type SynchronizationShardStatusRequestEntry struct {
	RemoteID   string                     `json:"remoteID"`
	Database   string                     `json:"database"`
	Collection string                     `json:"collection"`
	ShardIndex int                        `json:"shardIndex"`
	Status     SynchronizationShardStatus `json:"status"`
}

type SynchronizationShardStatus struct {
	// Current status
	Status SyncStatus `json:"status"`
	// Human readable status message
	StatusMessage string `json:"status_message,omitempty"`
	// Delay between us and other data center.
	Delay time.Duration `json:"delay"`
	// Time of last message received by the task handling this shard
	LastMessage time.Time `json:"last_message"`
	// Time of last message that resulted in a data change, received by the task handling this shard
	LastDataChange time.Time `json:"last_data_change"`
	// Time of when we last had a change in the status of the shard master
	LastShardMasterChange time.Time `json:"last_shard_master_change"`
	// Is the shard master known?
	ShardMasterKnown bool `json:"shard_master_known"`
}

// IsSame returns true when the Status & StatusMessage of both statuses
// are equal and the Delay is very close.
func (s SynchronizationShardStatus) IsSame(other SynchronizationShardStatus) bool {
	if s.Status != other.Status || s.StatusMessage != other.StatusMessage ||
		s.LastMessage != other.LastMessage || s.LastDataChange != other.LastDataChange ||
		s.LastShardMasterChange != other.LastShardMasterChange || s.ShardMasterKnown != other.ShardMasterKnown {
		return false
	}
	return !IsSignificantDelayDiff(s.Delay, other.Delay)
}

// TaskCompletedRequest holds the info for a TaskCompleted request.
type TaskCompletedRequest struct {
	Error bool `json:"error,omitempty"`
}

// TaskAssignment contains information of the assignment of a
// task to a worker.
// It is serialized as JSON into the agency.
type TaskAssignment struct {
	// ID of worker the task is assigned to
	WorkerID string `json:"worker_id"`
	// When the assignment was made
	CreatedAt time.Time `json:"created_at"`
	// How many assignments have been made
	Counter int `json:"counter,omitempty"`
}

// TaskInfo contains all information known about a task.
type TaskInfo struct {
	ID         string         `json:"id"`
	Task       tasks.TaskData `json:"task"`
	Assignment TaskAssignment `json:"assignment"`
}

// IsAssigned returns true when the task in given info is assigned to a
// worker, false otherwise.
func (i TaskInfo) IsAssigned() bool {
	return i.Assignment.WorkerID != ""
}

// NeedsCleanup returns true when the entry is subject to cleanup.
func (i TaskInfo) NeedsCleanup() bool {
	return i.Assignment.Counter > 0 && !i.Task.Persistent
}

// TasksResponse is the JSON response for MasterAPI.Tasks method.
type TasksResponse struct {
	Tasks []TaskInfo `json:"tasks,omitempty"`
}

// IsSignificantDelayDiff returns true if there is a significant difference
// between the given delays.
func IsSignificantDelayDiff(d1, d2 time.Duration) bool {
	if d2 == 0 {
		return d1 != 0
	}
	x := float64(d1) / float64(d2)
	return x < 0.9 || x > 1.1
}

// IsChannelRelevantResponse is the JSON response for a MasterAPI.IsChannelRelevant call
type IsChannelRelevantResponse struct {
	IsRelevant bool `json:"isRelevant"`
}

// StatusAPI describes the API provided to task workers used to send status updates to the master.
type StatusAPI interface {
	// SendIncomingStatus queues a given incoming synchronization status entry for sending.
	SendIncomingStatus(entry SynchronizationShardStatusRequestEntry)
	// SendOutgoingStatus queues a given outgoing synchronization status entry for sending.
	SendOutgoingStatus(entry SynchronizationShardStatusRequestEntry)
}

// DirectMQToken provides a token with its TTL
type DirectMQToken struct {
	// Token used to authenticate with the server.
	Token string `json:"token"`
	// How long the token will be valid.
	// Afterwards a new token has to be fetched.
	TokenTTL time.Duration `json:"token-ttl"`
}

// DirectMQTokenRequest is the JSON request body for Renew/Clone direct MQ token request.
type DirectMQTokenRequest struct {
	// Token used to authenticate with the server.
	Token string `json:"token"`
}

// DirectMQTopicEndpoint provides information about an endpoint for Direct MQ messages.
type DirectMQTopicEndpoint struct {
	// Endpoint of the server that can provide messages for a specific topic.
	Endpoint Endpoint `json:"endpoint"`
	// CA certificate used to sign the TLS connection of the server.
	// This is used for verifying the server.
	CACertificate string `json:"caCertificate"`
	// Token used to authenticate with the server.
	Token string `json:"token"`
	// How long the token will be valid.
	// Afterwards a new token has to be fetched.
	TokenTTL time.Duration `json:"token-ttl"`
}

// SetDirectMQTopicTokenRequest is the JSON request body for SetDirectMQTopicToken request.
type SetDirectMQTopicTokenRequest struct {
	// Token used to authenticate with the server.
	Token string `json:"token"`
	// How long the token will be valid.
	// Afterwards a new token has to be fetched.
	TokenTTL time.Duration `json:"token-ttl"`
}

// DirectMQMessage is a direct MQ message.
type DirectMQMessage struct {
	Offset  int64           `json:"offset"`
	Message json.RawMessage `json:"message"`
}

// GetDirectMQMessagesResponse is the JSON body for GetDirectMQMessages response.
type GetDirectMQMessagesResponse struct {
	Messages []DirectMQMessage `json:"messages,omitempty"`
}

// CommitDirectMQMessageRequest is the JSON request body for CommitDirectMQMessage request.
type CommitDirectMQMessageRequest struct {
	Offset int64 `json:"offset"`
}
