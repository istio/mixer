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

package noop2

import (
	"testing"

	"istio.io/mixer/pkg/adapter"
	"istio.io/mixer/pkg/adapterManager"
)

func TestRegisteredForAllAspects(t *testing.T) {
	handlers := adapterManager.AdapterInfoMap([]adapter.GetAdapterInfoFn{GetAdapterInfo},
		adapterManager.DoesBuilderSupportsTemplate)

	name := noop2AdapterInfo.Name
	resultNoop2AdapterInfo := handlers[name]

	expectedTmpls := noop2AdapterInfo.SupportedTemplates
	for _, expectedTmpl := range expectedTmpls {
		if !contains(resultNoop2AdapterInfo.SupportedTemplates, expectedTmpl) {
			t.Errorf("%s is not registered for template %s", name, expectedTmpl)
		}
	}
}

func contains(s []adapter.SupportedTemplates, e adapter.SupportedTemplates) bool {
	for _, a := range s {
		if a == e {
			return true
		}
	}
	return false
}