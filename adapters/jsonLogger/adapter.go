// Copyright 2016 Google Inc.
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

package jsonLogger

import (
	"os"

	"istio.io/mixer"
	"istio.io/mixer/adapters"
)

type (
	adapter struct{}
)

func newLogger(c adapters.AdapterConfig) (adapters.Logger, error) {
	return &adapter{}, nil
}

func (a *adapter) Log(l []mixer.LogEntry) error {
	var logsErr error
	for _, le := range l {
		if err := mixer.WriteJSON(os.Stdout, le); err != nil {
			logsErr = err
			continue
		}
	}
	return logsErr
}

func (a *adapter) Close() error { return nil }
