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
	"crypto/sha1"
	"fmt"
	"strings"
	"sync"

	certificates "github.com/arangodb-helper/go-certificates"

	"github.com/arangodb/arangosync/pkg/errors"
	"github.com/rs/zerolog"
)

type ClientCache struct {
	mutex   sync.Mutex
	clients map[string]API
}

// GetClient returns a client used to access the source with given authentication.
func (cc *ClientCache) GetClient(log zerolog.Logger, source Endpoint, auth Authentication, insecureSkipVerify bool) (API, error) {
	if len(source) == 0 {
		return nil, errors.Wrapf(PreconditionFailedError, "Cannot create master client: no source configured")
	}
	keyData := strings.Join(source, ",") + ":" + auth.String()
	key := fmt.Sprintf("%x", sha1.Sum([]byte(keyData)))

	cc.mutex.Lock()
	defer cc.mutex.Unlock()

	if cc.clients == nil {
		cc.clients = make(map[string]API)
	}

	// Get existing client (if any)
	if c, ok := cc.clients[key]; ok {
		return c, nil
	}

	// Client does not exist, create one
	log.Debug().Msg("Creating new client")
	c, err := cc.createClient(source, auth, insecureSkipVerify)
	if err != nil {
		return nil, maskAny(err)
	}

	cc.clients[key] = c
	c.SetShared()
	return c, nil
}

// createClient creates a client used to access the source with given authentication.
func (cc *ClientCache) createClient(source Endpoint, auth Authentication, insecureSkipVerify bool) (API, error) {
	if len(source) == 0 {
		return nil, errors.Wrapf(PreconditionFailedError, "Cannot create master client: no source configured")
	}
	tlsConfig, err := certificates.CreateTLSConfigFromAuthentication(AuthProxy{auth.TLSAuthentication}, insecureSkipVerify)
	if err != nil {
		return nil, maskAny(err)
	}
	ac := AuthenticationConfig{}
	if auth.JWTSecret != "" {
		ac.JWTSecret = auth.JWTSecret
	} else if auth.ClientToken != "" {
		ac.BearerToken = auth.ClientToken
	}
	c, err := NewArangoSyncClient(source, ac, tlsConfig)
	if err != nil {
		return nil, maskAny(err)
	}
	return c, nil
}

// NewAuthentication creates a new Authentication from given arguments.
func NewAuthentication(tlsAuth TLSAuthentication, jwtSecret string) Authentication {
	return Authentication{
		TLSAuthentication: tlsAuth,
		JWTSecret:         jwtSecret,
	}
}

// Authentication contains all possible authentication methods for a client.
// Order of authentication methods:
// - JWTSecret
// - ClientToken
// - ClientCertificate
type Authentication struct {
	TLSAuthentication
	JWTSecret string
}

// String returns a string used to unique identify the authentication settings.
func (a Authentication) String() string {
	return a.TLSAuthentication.String() + ":" + a.JWTSecret
}

// AuthProxy is a helper that implements github.com/arangodb-helper/go-certificates#TLSAuthentication.
type AuthProxy struct {
	TLSAuthentication
}

func (a AuthProxy) CACertificate() string     { return a.TLSAuthentication.CACertificate }
func (a AuthProxy) ClientCertificate() string { return a.TLSAuthentication.ClientCertificate }
func (a AuthProxy) ClientKey() string         { return a.TLSAuthentication.ClientKey }
