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

// Package aspect contains the various aspect managers which are responsible for
// mapping incoming requests into the interface expected by individual types of
// aspects.
package aspect

import (
	"io"
	"time"

	rpc "github.com/googleapis/googleapis/google/rpc"

	"istio.io/mixer/pkg/adapter"
	"istio.io/mixer/pkg/attribute"
	"istio.io/mixer/pkg/config"
	"istio.io/mixer/pkg/config/descriptor"
	cpb "istio.io/mixer/pkg/config/proto"
	"istio.io/mixer/pkg/expr"
)

type (
	// Manager is responsible for a specific aspect and presents a uniform interface
	// to the rest of the system.
	Manager interface {
		config.AspectValidator

		// Kind return the kind of aspect handled by this manager
		Kind() Kind
	}

	// CheckManager take care of aspects used to implement the Check API method
	CheckManager interface {
		Manager

		// NewCheckExecutor creates a new aspect executor given configuration.
		NewCheckExecutor(cfg *cpb.Combined, builder adapter.Builder, env adapter.Env, df descriptor.Finder) (CheckExecutor, error)
	}

	// ReportManager take care of aspects used to implement the Report API method
	ReportManager interface {
		Manager

		// NewReportExecutor creates a new aspect executor given configuration.
		NewReportExecutor(cfg *cpb.Combined, builder adapter.Builder, env adapter.Env, df descriptor.Finder) (ReportExecutor, error)
	}

	// QuotaManager take care of aspects used to implement the Quota API method
	QuotaManager interface {
		Manager

		// NewQuotaExecutor creates a new aspect executor given configuration.
		NewQuotaExecutor(cfg *cpb.Combined, builder adapter.Builder, env adapter.Env, df descriptor.Finder) (QuotaExecutor, error)
	}

	// Executor encapsulates a single aspect and allows it to be invoked.
	Executor interface {
		io.Closer
	}

	// CheckExecutor encapsulates a single CheckManager aspect and allows it to be invoked.
	CheckExecutor interface {
		Executor

		// Execute dispatches to the aspect manager.
		Execute(attrs attribute.Bag, mapper expr.Evaluator) rpc.Status
	}

	// ReportExecutor encapsulates a single ReportManager aspect and allows it to be invoked.
	ReportExecutor interface {
		Executor

		// Execute dispatches to the aspect manager.
		Execute(attrs attribute.Bag, mapper expr.Evaluator) rpc.Status
	}

	// QuotaExecutor encapsulates a single QuotaManager aspect and allows it to be invoked.
	QuotaExecutor interface {
		Executor

		// Execute dispatches to the aspect manager.
		Execute(attrs attribute.Bag, mapper expr.Evaluator, qma *QuotaMethodArgs) (rpc.Status, *QuotaMethodResp)
	}

	// QuotaMethodArgs is supplied by invocations of the Quota method.
	QuotaMethodArgs struct {
		// Used for deduplicating quota allocation/free calls in the case of
		// failed RPCs and retries. This should be a UUID per call, where the same
		// UUID is used for retries of the same quota allocation call.
		DeduplicationID string

		// The quota to allocate from.
		Quota string

		// The amount of quota to allocate.
		Amount int64

		// If true, allows a response to return less quota than requested. When
		// false, the exact requested amount is returned or 0 if not enough quota
		// was available.
		BestEffort bool
	}

	// QuotaMethodResp is returned by invocations of the Quota method.
	QuotaMethodResp struct {
		// The amount of time until which the returned quota expires, this is 0 for non-expiring quotas.
		Expiration time.Duration

		// The total amount of quota returned, may be less than requested.
		Amount int64
	}
)
