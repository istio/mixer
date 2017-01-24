// Copyright 2017 Google Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// Package stdioLogger provides an implementation of the mixer logger aspect
// that writes logs (serialized as JSON) to a standard stream (stdout | stderr).
package stdioLogger

import (
	"encoding/json"
	"io"
	"os"

	"istio.io/mixer/adapter/stdioLogger/config"
	"istio.io/mixer/pkg/adapter"

	me "github.com/hashicorp/go-multierror"
)

type (
	builderState struct{}
	aspectImpl   struct {
		logStream io.Writer
	}
)

// MustRegister records the builders exposed by this adapter.
func MustRegister(r adapter.Registrar) { r.RegisterLogger(builderState{}) }

func (builderState) Name() string { return "istio/stdioLogger" }
func (builderState) Description() string {
	return "Writes structured log entries to a standard I/O stream"
}
func (builderState) DefaultConfig() adapter.AspectConfig                              { return &config.Params{} }
func (builderState) Close() error                                                     { return nil }
func (builderState) ValidateConfig(c adapter.AspectConfig) (ce *adapter.ConfigErrors) { return nil }
func (builderState) NewLogger(env adapter.Env, cfg adapter.AspectConfig) (adapter.LoggerAspect, error) {
	c := cfg.(*config.Params)

	w := os.Stderr
	if c.LogStream == config.Params_STDOUT {
		w = os.Stdout
	}

	return &aspectImpl{w}, nil
}

func (a *aspectImpl) Close() error { return nil }

func (a *aspectImpl) Log(entries []adapter.LogEntry) error {
	var errors *me.Error
	for _, entry := range entries {
		if err := writeJSON(a.logStream, entry); err != nil {
			errors = me.Append(errors, err)
		}
	}

	return errors.ErrorOrNil()
}

func writeJSON(w io.Writer, le adapter.LogEntry) error {
	return json.NewEncoder(w).Encode(le)
}
