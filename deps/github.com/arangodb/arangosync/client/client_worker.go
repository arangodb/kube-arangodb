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
	"net/url"
	"path"
	"time"
)

// StartTask is called by the master to instruct the worker
// to run a task with given instructions.
func (c *client) StartTask(ctx context.Context, data StartTaskRequest) error {
	url := c.createURLs("/_api/task", nil)

	req, err := c.newRequests("POST", url, data)
	if err != nil {
		return maskAny(err)
	}
	if err := c.do(ctx, req, nil); err != nil {
		return maskAny(err)
	}

	return nil
}

// StopTask is called by the master to instruct the worker
// to stop all work on the given task.
func (c *client) StopTask(ctx context.Context, taskID string) error {
	url := c.createURLs(path.Join("/_api/task", taskID), nil)

	req, err := c.newRequests("DELETE", url, nil)
	if err != nil {
		return maskAny(err)
	}
	if err := c.do(ctx, req, nil); err != nil {
		return maskAny(err)
	}

	return nil
}

// SetDirectMQTopicToken configures the token used to access messages of a given channel.
func (c *client) SetDirectMQTopicToken(ctx context.Context, channelName, token string, tokenTTL time.Duration) error {
	url := c.createURLs(path.Join("/_api/mq/direct/channel", url.PathEscape(channelName), "token"), nil)

	data := SetDirectMQTopicTokenRequest{
		Token:    token,
		TokenTTL: tokenTTL,
	}
	req, err := c.newRequests("POST", url, data)
	if err != nil {
		return maskAny(err)
	}
	if err := c.do(ctx, req, nil); err != nil {
		return maskAny(err)
	}

	return nil
}

// GetDirectMQMessages return messages for a given MQ channel.
func (c *client) GetDirectMQMessages(ctx context.Context, channelName string) ([]DirectMQMessage, error) {
	url := c.createURLs(path.Join("/_api/mq/direct/channel", url.PathEscape(channelName), "messages"), nil)

	var result GetDirectMQMessagesResponse
	req, err := c.newRequests("GET", url, nil)
	if err != nil {
		return nil, maskAny(err)
	}
	if err := c.do(ctx, req, &result); err != nil {
		return nil, maskAny(err)
	}

	return result.Messages, nil
}

// CommitDirectMQMessage removes all messages from the given channel up to an including the given offset.
func (c *client) CommitDirectMQMessage(ctx context.Context, channelName string, offset int64) error {
	url := c.createURLs(path.Join("/_api/mq/direct/channel", url.PathEscape(channelName), "commit"), nil)

	data := CommitDirectMQMessageRequest{
		Offset: offset,
	}
	req, err := c.newRequests("POST", url, data)
	if err != nil {
		return maskAny(err)
	}
	if err := c.do(ctx, req, nil); err != nil {
		return maskAny(err)
	}

	return nil
}
