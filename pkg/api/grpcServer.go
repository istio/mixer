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

package api

// gRPC server. The GRPCServer type handles incoming streaming gRPC traffic and invokes method-specific
// handlers to implement the method-specific logic.
//
// When you create a GRPCServer instance, you specify a number of transport-level options, along with the
// set of method handlers responsible for the logic of individual API methods

// TODO: Once the gRPC code is updated to use context objects from "context" as
// opposed to from "golang.org/x/net/context", this code should be updated to
// pass the context from the gRPC streams to downstream calls as opposed to merely
// using context.Background.

import (
	"context"
	"io"
	"sync"
	"time"

	"github.com/golang/glog"
	"github.com/golang/protobuf/proto"
	rpc "github.com/googleapis/googleapis/google/rpc"
	"github.com/opentracing/opentracing-go/log"
	"google.golang.org/grpc"

	mixerpb "istio.io/api/mixer/v1"
	"istio.io/mixer/pkg/adapterManager"
	"istio.io/mixer/pkg/aspect"
	"istio.io/mixer/pkg/attribute"
	"istio.io/mixer/pkg/pool"
	"istio.io/mixer/pkg/status"
	"istio.io/mixer/pkg/tracing"
)

// grpcServer holds the state for the gRPC API server.
type grpcServer struct {
	aspectDispatcher adapterManager.AspectDispatcher
	attrMgr          *attribute.Manager
	tracer           tracing.Tracer
	gp               *pool.GoroutinePool

	// replaceable sendMsg so we can inject errors in tests
	sendMsg func(grpc.Stream, proto.Message) error
}

// NewGRPCServer creates a gRPC serving stack.
func NewGRPCServer(aspectDispatcher adapterManager.AspectDispatcher, tracer tracing.Tracer, gp *pool.GoroutinePool) mixerpb.MixerServer {
	return &grpcServer{
		aspectDispatcher: aspectDispatcher,
		attrMgr:          attribute.NewManager(),
		tracer:           tracer,
		gp:               gp,
		sendMsg: func(stream grpc.Stream, m proto.Message) error {
			return stream.SendMsg(m)
		},
	}
}

// dispatcher does all the nitty-gritty details of handling the mixer's low-level API
// protocol and dispatching to the right API handler.
func (s *grpcServer) dispatcher(stream grpc.Stream, methodName string,
	getState func() (request proto.Message, response proto.Message, requestAttrs *mixerpb.Attributes, responseAttrs *mixerpb.Attributes, result *rpc.Status),
	worker func(ctx context.Context, requestBag *attribute.MutableBag, responseBag *attribute.MutableBag,
		request proto.Message, response proto.Message)) error {

	// tracks attribute state for this stream
	reqTracker := s.attrMgr.NewTracker()
	respTracker := s.attrMgr.NewTracker()
	defer reqTracker.Done()
	defer respTracker.Done()

	// used to serialize sending on the grpc stream, since the grpc stream is not multithread-safe
	sendLock := &sync.Mutex{}

	root, ctx := s.tracer.StartRootSpan(stream.Context(), methodName)
	defer root.Finish()

	// ensure pending stuff is done before leaving
	wg := sync.WaitGroup{}
	defer wg.Wait()

	for {
		request, response, requestAttrs, responseAttrs, result := getState()

		// get a single message
		err := stream.RecvMsg(request)
		if err == io.EOF {
			return nil
		} else if err != nil {
			glog.Errorf("Stream error %s", err)
			return err
		}

		requestBag, err := reqTracker.ApplyProto(requestAttrs)
		if err != nil {
			msg := "Request could not be processed due to invalid 'attribute_update'."
			glog.Error(msg, "\n", err)
			details := status.NewBadRequest("attribute_update", err)
			*result = status.InvalidWithDetails(msg, details)

			sendLock.Lock()
			err = s.sendMsg(stream, response)
			sendLock.Unlock()

			if err != nil {
				glog.Errorf("Unable to send gRPC response message: %v", err)
			}

			continue
		}

		// throw the message into the work queue
		wg.Add(1)
		s.gp.ScheduleWork(func() {
			span, ctx2 := s.tracer.StartSpanFromContext(ctx, "RequestProcessing")
			span.LogFields(log.Object("gRPC request", request))

			responseBag := attribute.GetMutableBag(nil)

			// do the actual work for the message
			worker(ctx2, requestBag, responseBag, request, response)

			sendLock.Lock()
			respTracker.ApplyBag(responseBag, 0, responseAttrs)
			err := s.sendMsg(stream, response)
			sendLock.Unlock()

			if err != nil {
				glog.Errorf("Unable to send gRPC response message: %v", err)
			}

			requestBag.Done()
			responseBag.Done()

			span.LogFields(log.Object("gRPC response", response))
			span.Finish()

			wg.Done()
		})
	}
}

// Check is the entry point for the external Check method
func (s *grpcServer) Check(stream mixerpb.Mixer_CheckServer) error {
	return s.dispatcher(stream, "/istio.mixer.v1.Mixer/Check",
		func() (proto.Message, proto.Message, *mixerpb.Attributes, *mixerpb.Attributes, *rpc.Status) {
			request := &mixerpb.CheckRequest{}
			response := &mixerpb.CheckResponse{}
			response.AttributeUpdate = &mixerpb.Attributes{}
			return request, response, &request.AttributeUpdate, response.AttributeUpdate, &response.Result
		},
		s.handleCheck)
}

// Report is the entry point for the external Report method
func (s *grpcServer) Report(stream mixerpb.Mixer_ReportServer) error {
	return s.dispatcher(stream, "/istio.mixer.v1.Mixer/Report",
		func() (proto.Message, proto.Message, *mixerpb.Attributes, *mixerpb.Attributes, *rpc.Status) {
			request := &mixerpb.ReportRequest{}
			response := &mixerpb.ReportResponse{}
			response.AttributeUpdate = &mixerpb.Attributes{}
			return request, response, &request.AttributeUpdate, response.AttributeUpdate, &response.Result
		},
		s.handleReport)
}

// Quota is the entry point for the external Quota method
func (s *grpcServer) Quota(stream mixerpb.Mixer_QuotaServer) error {
	return s.dispatcher(stream, "/istio.mixer.v1.Mixer/Quota",
		func() (proto.Message, proto.Message, *mixerpb.Attributes, *mixerpb.Attributes, *rpc.Status) {
			request := &mixerpb.QuotaRequest{}
			response := &mixerpb.QuotaResponse{}
			response.AttributeUpdate = &mixerpb.Attributes{}
			return request, response, &request.AttributeUpdate, response.AttributeUpdate, &response.Result
		},
		s.handleQuota)
}

func (s *grpcServer) handleCheck(ctx context.Context, requestBag *attribute.MutableBag, responseBag *attribute.MutableBag,
	request proto.Message, response proto.Message) {
	req := request.(*mixerpb.CheckRequest)
	resp := response.(*mixerpb.CheckResponse)

	if glog.V(2) {
		glog.Infof("Check [%x]", req.RequestIndex)
	}

	resp.RequestIndex = req.RequestIndex
	resp.Result = s.aspectDispatcher.Check(ctx, requestBag, responseBag)
	// TODO: this value needs to initially come from config, and be modulated by the kind of attribute
	//       that was used in the check and the in-used aspects (for example, maybe an auth check has a
	//       30s TTL but a whitelist check has got a 120s TTL)
	resp.Expiration = time.Duration(5) * time.Second

	if glog.V(2) {
		glog.Infof("Check [%x] <-- %s", req.RequestIndex, response)
	}
}

func (s *grpcServer) handleReport(ctx context.Context, requestBag *attribute.MutableBag, responseBag *attribute.MutableBag,
	request proto.Message, response proto.Message) {
	req := request.(*mixerpb.ReportRequest)
	resp := response.(*mixerpb.ReportResponse)

	if glog.V(2) {
		glog.Infof("Report [%x]", req.RequestIndex)
	}

	resp.RequestIndex = req.RequestIndex
	resp.Result = s.aspectDispatcher.Report(ctx, requestBag, responseBag)

	if glog.V(2) {
		glog.Infof("Report [%x] <-- %s", req.RequestIndex, response)
	}
}

func (s *grpcServer) handleQuota(ctx context.Context, requestBag *attribute.MutableBag, responseBag *attribute.MutableBag,
	request proto.Message, response proto.Message) {
	req := request.(*mixerpb.QuotaRequest)
	resp := response.(*mixerpb.QuotaResponse)

	if glog.V(2) {
		glog.Infof("Quota [%x]", req.RequestIndex)
	}

	qma := &aspect.QuotaMethodArgs{
		Quota:           req.Quota,
		Amount:          req.Amount,
		DeduplicationID: req.DeduplicationId,
		BestEffort:      req.BestEffort,
	}
	var qmr *aspect.QuotaMethodResp

	resp.RequestIndex = req.RequestIndex
	qmr, resp.Result = s.aspectDispatcher.Quota(ctx, requestBag, responseBag, qma)

	if qmr != nil {
		resp.Amount = qmr.Amount
		resp.Expiration = qmr.Expiration
	}

	if glog.V(2) {
		glog.Infof("Quota [%x] <-- %s", req.RequestIndex, response)
	}
}
