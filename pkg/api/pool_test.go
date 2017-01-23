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

package api

import (
	"context"
	"testing"
	"time"

	"google.golang.org/genproto/googleapis/rpc/code"
	"google.golang.org/genproto/googleapis/rpc/status"
	"istio.io/mixer/pkg/adapter"
	"istio.io/mixer/pkg/aspect"
	"istio.io/mixer/pkg/attribute"
	"istio.io/mixer/pkg/expr"

	"github.com/golang/protobuf/ptypes/struct"
	istioconfig "istio.io/api/mixer/v1/config"
	"istio.io/mixer/pkg/adapterManager"
)

type testAspectFn func() (*aspect.Output, error)

type testManager struct {
	name     string
	instance testAspect
}

func newTestManager(name string, fn testAspectFn) testManager {
	return testManager{name, testAspect{fn}}
}
func (testManager) Close() error                                                { return nil }
func (testManager) DefaultConfig() adapter.AspectConfig                         { return nil }
func (testManager) ValidateConfig(c adapter.AspectConfig) *adapter.ConfigErrors { return nil }
func (testManager) Kind() string                                                { return "denyChecker" }
func (m testManager) Name() string                                              { return m.name }
func (testManager) Description() string                                         { return "deny checker aspect manager for testing" }

func (m testManager) NewAspect(cfg *aspect.CombinedConfig, adapter adapter.Builder, env adapter.Env) (aspect.Wrapper, error) {
	return m.instance, nil
}
func (m testManager) NewDenyChecker(env adapter.Env, c adapter.AspectConfig) (adapter.DenyCheckerAspect, error) {
	return m.instance, nil
}

type testAspect struct {
	fn testAspectFn
}

func (testAspect) Close() error { return nil }
func (t testAspect) Execute(attrs attribute.Bag, mapper expr.Evaluator) (*aspect.Output, error) {
	return t.fn()
}
func (testAspect) Deny() status.Status { return status.Status{Code: int32(code.Code_INTERNAL)} }

var (
	name = "denyChecker"

	cfg = &aspect.CombinedConfig{
		Aspect: &istioconfig.Aspect{
			Kind:    name,
			Adapter: "",
			Inputs:  make(map[string]string),
			Params:  new(structpb.Struct),
		},
		Builder: &istioconfig.Adapter{
			Name:   "",
			Kind:   "",
			Impl:   name,
			Params: new(structpb.Struct),
		},
	}
)

func TestPoolSize(t *testing.T) {
	blockChan := make(chan struct{})
	testMngr := newTestManager(name, func() (*aspect.Output, error) {
		<-blockChan
		return &aspect.Output{Code: code.Code_OK}, nil
	})
	b := StaticBinding{
		RegisterFn: func(r adapter.Registrar) error { return r.RegisterDenyChecker(testMngr) },
		Manager:    testMngr,
		Config:     cfg,
		Methods:    []Method{Check},
	}

	// Easier than creating a new manager directly and having to register everything. We need all the config either way.
	mgr, _ := NewMethodHandlers(1, b).(*methodHandlers)
	underTest := newPool(1)

	// Note: we allocate a result buffer of size two by passing in numAdapters = 2
	res, enqueue := underTest.requestGroup(mgr.mngr, nil, nil, 2)

	// Enqueue work which will not complete until blockChan is closed; since the pool size == 1 this blocks the queue
	enqueue(context.Background(), cfg)

	second := make(chan struct{})
	go func() {
		// this second enqueue will block until blockChan is closed
		enqueue(context.Background(), cfg)
		close(second)
	}()

	if len(res) != 0 {
		t.Errorf("len(res) = %d, wanted 0", len(res))
	}

	close(blockChan) // unblock the queue
	<-second         // block the test go routine till the second enqueue completes

	// It takes a little time for the two goroutines to write their results; we loop to make the test more reliable.
	for count := 0; len(res) != 2 && count < 5; count++ {
		time.Sleep(10 * time.Millisecond)
	}
	if len(res) != 2 {
		t.Errorf("got %d finished tasks, wanted 2", len(res))
	}

	for i := 0; i < 2; i++ {
		r := <-res
		if r.out.Code != code.Code_OK {
			t.Errorf("r.out.Code = %s, wanted %s", code.Code_name[int32(r.out.Code)], code.Code_name[int32(code.Code_OK)])
		}
	}
}

func TestShutdown(t *testing.T) {
	fail := make(chan struct{})
	succeed := make(chan struct{})
	p := newPool(1)

	go func() {
		time.Sleep(1 * time.Second)
		close(fail)
	}()

	go func() {
		p.shutdown()
		close(succeed)
	}()

	select {
	case <-fail:
		t.Error("pool.shutdown() didn't complete in the expected time")
	case <-succeed:
	}
}

func TestEnqueuePanics(t *testing.T) {
	p := newPool(1)
	mgr := adapterManager.NewManager(nil)

	numCalls := 1
	_, enqueue := p.requestGroup(mgr, nil, nil, numCalls)
	for i := 0; i < numCalls; i++ {
		enqueue(nil, cfg)
	}

	defer func() {
		if r := recover(); r == nil {
			t.Error("enqueue(nil, cfg) got nil, want panic and non-nil recover.")
		}
	}()
	enqueue(nil, cfg)
	t.Fail()
}
