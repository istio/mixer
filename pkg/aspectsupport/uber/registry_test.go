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

package uber

import (
	"fmt"
	"testing"

	"github.com/golang/protobuf/proto"
	"istio.io/mixer/pkg/aspect"
	"istio.io/mixer/pkg/aspect/denyChecker"
	"istio.io/mixer/pkg/aspect/listChecker"
	al "istio.io/mixer/pkg/aspect/logger"
	"istio.io/mixer/pkg/aspect/quota"
)

type testAdapter struct {
	name string
}

func (t testAdapter) Name() string                                               { return t.name }
func (testAdapter) Close() error                                                 { return nil }
func (testAdapter) Description() string                                          { return "mock adapter for testing" }
func (testAdapter) DefaultConfig() proto.Message                                 { return nil }
func (testAdapter) ValidateConfig(implConfig proto.Message) *aspect.ConfigErrors { return nil }

type denyAdapter struct{ testAdapter }

func (denyAdapter) NewAspect(env aspect.Env, cfg proto.Message) (denyChecker.Aspect, error) {
	return nil, fmt.Errorf("not implemented")
}

func TestRegistry_RegisterDeny(t *testing.T) {
	reg := NewRegistry()
	adapter := denyAdapter{testAdapter{name: "foo"}}

	if err := reg.RegisterDeny(adapter); err != nil {
		t.Errorf("Failed to register deny adapter with err: %v", err)
	}

	impl, ok := reg.ByImpl(adapter.Name())
	if !ok {
		t.Errorf("No adapter by impl with name %s, expected adapter: %v", adapter.Name(), adapter)
	}

	if deny, ok := impl.(denyAdapter); !ok || deny != adapter {
		t.Errorf("reg.ByImpl(%s) expected adapter '%v', actual '%v'", adapter.Name(), adapter, impl)
	}
}

type listAdapter struct{ testAdapter }

func (listAdapter) NewAspect(env aspect.Env, cfg proto.Message) (listChecker.Aspect, error) {
	return nil, fmt.Errorf("not implemented")
}

func TestRegistry_RegisterCheckList(t *testing.T) {
	reg := NewRegistry()
	adapter := listAdapter{testAdapter{name: "foo"}}

	if err := reg.RegisterCheckList(adapter); err != nil {
		t.Errorf("Failed to register deny adapter with err: %v", err)
	}

	impl, ok := reg.ByImpl(adapter.Name())
	if !ok {
		t.Errorf("No adapter by impl with name %s, expected adapter: %v", adapter.Name(), adapter)
	}

	if deny, ok := impl.(listAdapter); !ok || deny != adapter {
		t.Errorf("reg.ByImpl(%s) expected adapter '%v', actual '%v'", adapter.Name(), adapter, impl)
	}
}

type loggerAdapter struct{ testAdapter }

func (loggerAdapter) NewAspect(env aspect.Env, cfg proto.Message) (al.Aspect, error) {
	return nil, fmt.Errorf("not implemented")
}

func TestRegistry_RegisterLogger(t *testing.T) {
	reg := NewRegistry()
	adapter := loggerAdapter{testAdapter{name: "foo"}}

	if err := reg.RegisterLogger(adapter); err != nil {
		t.Errorf("Failed to register deny adapter with err: %v", err)
	}

	impl, ok := reg.ByImpl(adapter.Name())
	if !ok {
		t.Errorf("No adapter by impl with name %s, expected adapter: %v", adapter.Name(), adapter)
	}

	if deny, ok := impl.(loggerAdapter); !ok || deny != adapter {
		t.Errorf("reg.ByImpl(%s) expected adapter '%v', actual '%v'", adapter.Name(), adapter, impl)
	}
}

type quotaAdapter struct{ testAdapter }

func (quotaAdapter) NewAspect(env aspect.Env, cfg proto.Message) (quota.Aspect, error) {
	return nil, fmt.Errorf("not implemented")
}

func TestRegistry_RegisterQuota(t *testing.T) {
	reg := NewRegistry()
	adapter := quotaAdapter{testAdapter{name: "foo"}}

	if err := reg.RegisterQuota(adapter); err != nil {
		t.Errorf("Failed to register deny adapter with err: %v", err)
	}

	impl, ok := reg.ByImpl(adapter.Name())
	if !ok {
		t.Errorf("No adapter by impl with name %s, expected adapter: %v", adapter.Name(), adapter)
	}

	if deny, ok := impl.(quotaAdapter); !ok || deny != adapter {
		t.Errorf("reg.ByImpl(%s) expected adapter '%v', actual '%v'", adapter.Name(), adapter, impl)
	}
}

func TestRegistry_VerifyCollision(t *testing.T) {
	reg := NewRegistry()
	name := "some name that they both have"

	a1 := denyAdapter{testAdapter{name}}
	if err := reg.RegisterDeny(a1); err != nil {
		t.Errorf("Failed to insert first adapter with err: %s", err)
	}
	if a, ok := reg.ByImpl(name); !ok || a != a1 {
		t.Errorf("Failed to get first adapter by impl name; expected: '%v', actual: '%v'", a1, a)
	}

	a2 := listAdapter{testAdapter{name}}
	if err := reg.RegisterCheckList(a2); err != nil {
		t.Errorf("Failed to insert second adapter with err: %s", err)
	}
	if a, ok := reg.ByImpl(name); !ok || a != a2 {
		t.Errorf("Expected registering adapter with identical name to overwrite existing one; expected: '%v', actual: '%v'", a2, a)
	}
}
