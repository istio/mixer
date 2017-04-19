// Copyright 2017 the Istio Authors.
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

package noop

import (
	"testing"

	"istio.io/mixer/pkg/adapter"
	"istio.io/mixer/pkg/adapterManager"
	"istio.io/mixer/pkg/config"
)

func TestRegisteredForAllAspects(t *testing.T) {
	builders := adapterManager.BuilderMap([]adapter.RegisterFn{Register})

	name := builder{}.Name()
	noop := builders[name]

	var i uint
	for i = 0; i < uint(config.NumKinds); i++ {
		k := config.Kind(i)
		if !noop.Kinds.IsSet(k) {
			t.Errorf("%s is not registered for kind %s", name, k)
		}
	}
}
