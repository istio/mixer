// Copyright 2016 Istio Authors
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

package e2e

import (
	"fmt"
	"google.golang.org/grpc"
	"istio.io/api/mixer/v1"
	adp "istio.io/mixer/adapter"
	"istio.io/mixer/pkg/adapter"
	"istio.io/mixer/pkg/adapterManager"
	"istio.io/mixer/pkg/api"
	"istio.io/mixer/pkg/aspect"
	"istio.io/mixer/pkg/config"
	"istio.io/mixer/pkg/config/store"
	"istio.io/mixer/pkg/expr"
	"istio.io/mixer/pkg/il/evaluator"
	"istio.io/mixer/pkg/pool"
	mixerRuntime "istio.io/mixer/pkg/runtime"
	"istio.io/mixer/pkg/template"
	"net"
	"testing"
	"time"
)

const (
	configIdentityAttribute = "target.service"
	identityDomainAttribute = "svc.cluster.local"
)

type testState struct {
	client     istio_mixer_v1.MixerClient
	gs         *grpc.Server
	gp         *pool.GoroutinePool
	server     istio_mixer_v1.MixerServer
	connection *grpc.ClientConn
}

// fail fatal if dispatcher cannot be constructed
func initMixer(t *testing.T, configStore2URL string, adptInfos []adapter.InfoFn, tmplInfos map[string]template.Info) *testState {
	// TODO replace
	useIL := false
	apiPoolSize := 1024
	adapterPoolSize := 1024
	//loopDelay := time.Second * 5
	configFetchIntervalSec := 5
	singleThreadedGoRoutinePool := false
	configDefaultNamespace := "istio-config-default"
	gp := getGoRoutinePool(apiPoolSize, singleThreadedGoRoutinePool)
	adapterGP := getAdapterGoRoutinePool(adapterPoolSize, singleThreadedGoRoutinePool)
	adapterMap := adp.InventoryMap(adptInfos)
	eval, err := expr.NewCEXLEvaluator(expr.DefaultCacheSize)
	if err != nil {
		t.Errorf("Failed to create expression evaluator: %v", err)
	}
	var ilEval *evaluator.IL
	if useIL {
		ilEval, err = evaluator.NewILEvaluator(expr.DefaultCacheSize)
		if err != nil {
			t.Fatalf("Failed to create IL expression evaluator with cache size %d: %v", 1024, err)
		}
		eval = ilEval
	}
	var dispatcher mixerRuntime.Dispatcher

	store2, err := store.NewRegistry2(config.Store2Inventory()...).NewStore2(configStore2URL)
	if err != nil {
		t.Fatalf("Failed to connect to the configuration2 server. %v", err)
	}
	dispatcher, err = mixerRuntime.New(eval, gp, adapterGP,
		configIdentityAttribute, configDefaultNamespace,
		store2, adapterMap, tmplInfos,
	)
	if err != nil {
		t.Fatalf("Failed to create runtime dispatcher. %v", err)
	}
	adapterMgr := adapterManager.NewManager(adp.Inventory(), aspect.Inventory(), eval, gp, adapterGP)

	repo := template.NewRepository(tmplInfos)

	var st store.KeyValueStore
	registry := store.NewRegistry(config.StoreInventory()...)
	if st, err = registry.NewStore(configStore2URL); err != nil {
		t.Fatalf("Failed to get config store 1: %v", err)
	}

	configManager := config.NewManager(eval, adapterMgr.AspectValidatorFinder, adapterMgr.BuilderValidatorFinder, adptInfos,
		adapterMgr.SupportedKinds,
		repo, st, time.Second*time.Duration(configFetchIntervalSec),
		configIdentityAttribute,
		identityDomainAttribute)
	configManager.Register(adapterMgr)
	if useIL {
		configManager.Register(ilEval)
	}
	configManager.Start()

	ts := testState{}
	ts.server = api.NewGRPCServer(adapterMgr, dispatcher, gp)

	// get the network stuff setup
	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", 0))
	if err != nil {
		t.Fatalf("%v", err)
	}

	var grpcOptions []grpc.ServerOption
	grpcOptions = append(grpcOptions, grpc.MaxConcurrentStreams(32))
	grpcOptions = append(grpcOptions, grpc.MaxMsgSize(1024*1024))

	// get everything wired up
	ts.gs = grpc.NewServer(grpcOptions...)

	ts.gp = pool.NewGoroutinePool(128, false)
	ts.gp.AddWorkers(32)

	istio_mixer_v1.RegisterMixerServer(ts.gs, ts.server)

	go func() {
		_ = ts.gs.Serve(listener)
	}()

	dial := listener.Addr().String()

	if err = ts.createAPIClient(dial); err != nil {
		ts.deleteGRPCServer()
		t.Fatalf("%s", err)
	}

	///////
	return &ts
}

func (ts *testState) deleteGRPCServer() {
	ts.gs.GracefulStop()
	ts.gp.Close()
}

func (ts *testState) createAPIClient(dial string) error {
	var opts []grpc.DialOption
	opts = append(opts, grpc.WithInsecure())

	var err error
	if ts.connection, err = grpc.Dial(dial, opts...); err != nil {
		return err
	}

	ts.client = istio_mixer_v1.NewMixerClient(ts.connection)
	return nil
}

func (ts *testState) cleanupTestState() {
	ts.deleteAPIClient()
	ts.deleteGRPCServer()
}

func (ts *testState) deleteAPIClient() {
	_ = ts.connection.Close()
	ts.client = nil
	ts.connection = nil
}

func getAdapterGoRoutinePool(adapterPoolSize int, singleThreadedGoRoutinePool bool) *pool.GoroutinePool {
	adapterGP := pool.NewGoroutinePool(adapterPoolSize, singleThreadedGoRoutinePool)
	adapterGP.AddWorkers(adapterPoolSize)
	return adapterGP
}
func getGoRoutinePool(apiPoolSize int, singleThreadedGoRoutinePool bool) *pool.GoroutinePool {
	gp := pool.NewGoroutinePool(apiPoolSize, singleThreadedGoRoutinePool)
	gp.AddWorkers(apiPoolSize)
	gp.AddWorkers(apiPoolSize)
	return gp
}
