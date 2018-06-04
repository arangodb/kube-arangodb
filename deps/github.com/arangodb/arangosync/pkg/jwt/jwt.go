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

package jwt

import (
	"fmt"
	"net/http"
	"strings"

	jg "github.com/dgrijalva/jwt-go"
)

const (
	issArangod    = "arangodb"
	issArangoSync = "arangosync"
)

// AddArangodJwtHeader calculates a JWT authorization header, for authorization
// of a request to an arangod server, based on the given secret
// and adds it to the given request.
// If the secret is empty, nothing is done.
func AddArangodJwtHeader(req *http.Request, jwtSecret string) error {
	if jwtSecret == "" {
		return nil
	}
	value, err := CreateArangodJwtAuthorizationHeader(jwtSecret)
	if err != nil {
		return maskAny(err)
	}

	req.Header.Set("Authorization", value)
	return nil
}

// CreateArangodJwtAuthorizationHeader calculates a JWT authorization header, for authorization
// of a request to an arangod server, based on the given secret.
// If the secret is empty, nothing is done.
func CreateArangodJwtAuthorizationHeader(jwtSecret string) (string, error) {
	if jwtSecret == "" {
		return "", nil
	}
	// Create a new token object, specifying signing method and the claims
	// you would like it to contain.
	token := jg.NewWithClaims(jg.SigningMethodHS256, jg.MapClaims{
		"iss":       issArangod,
		"server_id": "foo",
	})

	// Sign and get the complete encoded token as a string using the secret
	signedToken, err := token.SignedString([]byte(jwtSecret))
	if err != nil {
		return "", maskAny(err)
	}

	return "bearer " + signedToken, nil
}

// AddArangoSyncJwtHeader calculates a JWT authorization header, for authorization
// of a request to an arangosync server, based on the given secret
// and adds it to the given request.
// If the secret is empty, nothing is done.
func AddArangoSyncJwtHeader(req *http.Request, jwtSecret string) error {
	if jwtSecret == "" {
		return nil
	}
	// Create a new token object, specifying signing method and the claims
	// you would like it to contain.
	token := jg.NewWithClaims(jg.SigningMethodHS256, jg.MapClaims{
		"iss":       issArangoSync,
		"server_id": "foo",
	})

	// Sign and get the complete encoded token as a string using the secret
	signedToken, err := token.SignedString([]byte(jwtSecret))
	if err != nil {
		return maskAny(err)
	}

	req.Header.Set("Authorization", "bearer "+signedToken)
	return nil
}

// VerifyArangoSyncJwtHeader verifies the bearer token in the given request with
// the given secret.
// If returns nil when verification succeed, an error if verification fails.
// If the secret is empty, nothing is done.
func VerifyArangoSyncJwtHeader(req *http.Request, jwtSecret string) error {
	if jwtSecret == "" {
		return nil
	}
	// Extract Authorization header
	authHdr := strings.TrimSpace(req.Header.Get("Authorization"))
	if authHdr == "" {
		return maskAny(fmt.Errorf("No Authorization found"))
	}
	prefix := "bearer "
	if !strings.HasPrefix(strings.ToLower(authHdr), prefix) {
		// Missing bearer prefix
		return maskAny(fmt.Errorf("No bearer prefix"))
	}
	tokenStr := strings.TrimSpace(authHdr[len(prefix):])
	// Parse token
	claims := jg.MapClaims{
		"iss":       issArangoSync,
		"server_id": "foo",
	}
	_, err := jg.ParseWithClaims(tokenStr, claims, func(*jg.Token) (interface{}, error) { return []byte(jwtSecret), nil })
	if err != nil {
		return maskAny(err)
	}

	return nil
}
