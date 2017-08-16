package token

// Copyright 2017 Istio Authors.
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

import (
	"testing"

	"istio.io/mixer/pkg/adapter"
	"istio.io/mixer/pkg/adapterManager"
	"istio.io/mixer/pkg/config"
)

func TestRegisteredForAllAspects(t *testing.T) {
	builders := adapterManager.BuilderMap([]adapter.RegisterFn{Register})

	k := config.AttributesKind
	found := false
	for _, sample := range builders {
		if sample.Kinds.IsSet(k) {
			found = true
		}
		if !found {
			t.Errorf("sample is not registered for kind %s", k)
		}
	}
}
