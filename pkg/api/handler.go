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

package api

// This is what implements the per-request logic for each API method.

import (
	"context"

	"github.com/golang/glog"
	"google.golang.org/genproto/googleapis/rpc/code"
	"google.golang.org/genproto/googleapis/rpc/status"

	"istio.io/mixer/pkg/aspect"
	"istio.io/mixer/pkg/attribute"

	"fmt"

	"sync/atomic"

	mixerpb "istio.io/api/mixer/v1"
	"istio.io/mixer/pkg/config"
)

// Handler holds pointers to the functions that implement
// request-level processing for incoming all public APIs.
type Handler interface {
	// Check performs the configured set of precondition checks.
	// Note that the request parameter is immutable, while the response parameter is where
	// results are specified
	Check(context.Context, attribute.Tracker, *mixerpb.CheckRequest, *mixerpb.CheckResponse)

	// Report performs the requested set of reporting operations.
	// Note that the request parameter is immutable, while the response parameter is where
	// results are specified
	Report(context.Context, attribute.Tracker, *mixerpb.ReportRequest, *mixerpb.ReportResponse)

	// Quota increments, decrements, or queries the specified quotas.
	// Note that the request parameter is immutable, while the response parameter is where
	// results are specified
	Quota(context.Context, attribute.Tracker, *mixerpb.QuotaRequest, *mixerpb.QuotaResponse)
}

// Executor executes any aspect as described by config.Combined. Only UberManager implements this interface.
type Executor interface {
	// Execute performs actions described in combined config using the attribute bag
	Execute(cfg *config.Combined, attrs attribute.Bag) (*aspect.Output, error)
}

// handlerState holds state and configuration for the handler.
type handlerState struct {
	aspectExecutor Executor
	// Configs for the aspects that'll be used to serve each API method. <*config.Runtime>
	cfg atomic.Value
	// set of aspect Kinds that should be dispatched for "check"
	checkSet config.AspectSet
	// set of aspect Kinds that should be dispatched for "report"
	reportSet config.AspectSet
	// set of aspect Kinds that should be dispatched for "quota"
	quotaSet config.AspectSet
}

// NewHandler returns a canonical Handler that implements all of the mixer's API surface
func NewHandler(aspectExecutor Executor, methodmap map[config.APIMethod]config.AspectSet) Handler {
	return &handlerState{
		aspectExecutor: aspectExecutor,
		checkSet:       methodmap[config.CheckMethod],
		reportSet:      methodmap[config.ReportMethod],
		quotaSet:       methodmap[config.QuotaMethod],
	}
}

// execute performs common function shared across the api surface.
func (h *handlerState) execute(ctx context.Context, tracker attribute.Tracker, attrs *mixerpb.Attributes, aspects config.AspectSet) *status.Status {
	ab, err := tracker.StartRequest(attrs)
	if err != nil {
		glog.Warningf("Unable to process attribute update. error: '%v'", err)
		return newStatus(code.Code_INVALID_ARGUMENT)
	}
	defer tracker.EndRequest()

	// get a new context with the attribute bag attached
	ctx = attribute.NewContext(ctx, ab)

	cfg := h.cfg.Load().(*config.Runtime)

	if cfg == nil {
		// config has NOT been loaded yet
		glog.Warningf("configuration is not available")
		return newStatusWithMessage(code.Code_UNAVAILABLE, "configuration is not available")
	}

	cfgs, err := cfg.Resolve(ab, aspects)
	if err != nil {
		return newStatusWithMessage(code.Code_INTERNAL, fmt.Sprintf("unable to resolve config %s", err.Error()))
	}

	for _, conf := range cfgs {
		// TODO: plumb ctx through uber.manager.Execute
		_ = ctx
		out, err := h.aspectExecutor.Execute(conf, ab)
		if err != nil {
			errorStr := fmt.Sprintf("Adapter %s returned err: %v", conf.Builder.Name, err)
			glog.Warning(errorStr)
			return newStatusWithMessage(code.Code_INTERNAL, errorStr)
		}
		if out.Code != code.Code_OK {
			return newStatusWithMessage(out.Code, "Rejected by builder "+conf.Builder.Name)
		}
	}
	return newStatus(code.Code_OK)
}

// Check performs 'check' function corresponding to the mixer api.
func (h *handlerState) Check(ctx context.Context, tracker attribute.Tracker, request *mixerpb.CheckRequest, response *mixerpb.CheckResponse) {
	response.RequestIndex = request.RequestIndex
	response.Result = h.execute(ctx, tracker, request.AttributeUpdate, h.checkSet)
}

// Report performs 'report' function corresponding to the mixer api.
func (h *handlerState) Report(ctx context.Context, tracker attribute.Tracker, request *mixerpb.ReportRequest, response *mixerpb.ReportResponse) {
	response.RequestIndex = request.RequestIndex
	response.Result = h.execute(ctx, tracker, request.AttributeUpdate, h.reportSet)
}

// Quota performs 'quota' function corresponding to the mixer api.
func (h *handlerState) Quota(ctx context.Context, tracker attribute.Tracker, request *mixerpb.QuotaRequest, response *mixerpb.QuotaResponse) {
	response.RequestIndex = request.RequestIndex
	status := h.execute(ctx, tracker, request.AttributeUpdate, h.quotaSet)
	response.Result = newQuotaError(code.Code(status.Code))
}

func newStatus(c code.Code) *status.Status {
	return &status.Status{Code: int32(c)}
}

func newStatusWithMessage(c code.Code, message string) *status.Status {
	return &status.Status{Code: int32(c), Message: message}
}

func newQuotaError(c code.Code) *mixerpb.QuotaResponse_Error {
	return &mixerpb.QuotaResponse_Error{Error: newStatus(c)}
}

// ConfigChange listens for config change notifications.
func (h *handlerState) ConfigChange(cfg *config.Runtime) {
	h.cfg.Store(cfg)
}
