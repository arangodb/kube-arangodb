//
// DISCLAIMER
//
// Copyright 2016-2022 ArangoDB GmbH, Cologne, Germany
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

package service

import (
	"context"
	"encoding/json"
	"io"
	"net/http"

	"github.com/julienschmidt/httprouter"

	"github.com/arangodb/kube-arangodb/pkg/logging"
	"github.com/arangodb/kube-arangodb/pkg/storage/provisioner"
	"github.com/arangodb/kube-arangodb/pkg/util/errors"
)

const (
	contentTypeJSON = "application/json"
)

// runServer runs a HTTP server serving the given API
func runServer(ctx context.Context, log logging.Logger, addr string, api provisioner.API) error {
	mux := httprouter.New()
	mux.GET("/nodeinfo", getNodeInfoHandler(api))
	mux.POST("/info", getInfoHandler(api))
	mux.POST("/prepare", getPrepareHandler(api))
	mux.POST("/remove", getRemoveHandler(api))

	httpServer := &http.Server{
		Addr:    addr,
		Handler: mux,
	}

	serverErrors := make(chan error)
	go func() {
		defer close(serverErrors)
		log.Info("Listening on %s", addr)
		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			serverErrors <- errors.WithStack(err)
		}
	}()

	select {
	case err := <-serverErrors:
		return errors.WithStack(err)
	case <-ctx.Done():
		// Close server
		log.Debug("Closing server...")
		httpServer.Close()
		return nil
	}
}

func getNodeInfoHandler(api provisioner.API) func(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	return func(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
		ctx := r.Context()
		result, err := api.GetNodeInfo(ctx)
		if err != nil {
			handleError(w, err)
		} else {
			sendJSON(w, result)
		}
	}
}

func getInfoHandler(api provisioner.API) func(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	return func(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
		ctx := r.Context()
		var input provisioner.Request
		if err := parseBody(r, &input); err != nil {
			handleError(w, err)
		} else {
			result, err := api.GetInfo(ctx, input.LocalPath)
			if err != nil {
				handleError(w, err)
			} else {
				sendJSON(w, result)
			}
		}
	}
}

func getPrepareHandler(api provisioner.API) func(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	return func(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
		ctx := r.Context()
		var input provisioner.Request
		if err := parseBody(r, &input); err != nil {
			handleError(w, err)
		} else {
			if err := api.Prepare(ctx, input.LocalPath); err != nil {
				handleError(w, err)
			} else {
				sendJSON(w, struct{}{})
			}
		}
	}
}

func getRemoveHandler(api provisioner.API) func(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	return func(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
		ctx := r.Context()
		var input provisioner.Request
		if err := parseBody(r, &input); err != nil {
			handleError(w, err)
		} else {
			if err := api.Remove(ctx, input.LocalPath); err != nil {
				handleError(w, err)
			} else {
				sendJSON(w, struct{}{})
			}
		}
	}
}

// sendJSON encodes given body as JSON and sends it to the given writer with given HTTP status.
func sendJSON(w http.ResponseWriter, body interface{}) error {
	w.Header().Set("Content-Type", contentTypeJSON)
	w.WriteHeader(http.StatusOK)
	if body == nil {
		w.Write([]byte("{}"))
	} else {
		encoder := json.NewEncoder(w)
		if err := encoder.Encode(body); err != nil {
			return errors.WithStack(err)
		}
	}
	return nil
}

func parseBody(r *http.Request, data interface{}) error {
	defer r.Body.Close()
	body, err := io.ReadAll(r.Body)
	if err != nil {
		return errors.WithStack(err)
	}
	if err := json.Unmarshal(body, data); err != nil {
		return errors.Wrapf(provisioner.BadRequestError, "Cannot parse request body: %v", err.Error())
	}
	return nil
}

func handleError(w http.ResponseWriter, err error) {
	if provisioner.IsBadRequest(err) {
		writeError(w, http.StatusBadRequest, err.Error())
	} else {
		writeError(w, http.StatusInternalServerError, err.Error())
	}
}

func writeError(w http.ResponseWriter, status int, message string) {
	if message == "" {
		message = "Unknown error"
	}
	resp := provisioner.ErrorResponse{Error: message}
	b, _ := json.Marshal(resp)
	w.Header().Set("Content-Type", contentTypeJSON)
	w.WriteHeader(status)
	w.Write(b)
}
