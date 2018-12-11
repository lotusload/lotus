// Copyright 2018, OpenCensus Authors
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

package grpcmetrics

import (
	"context"
	"strconv"
	"strings"
	"sync/atomic"
	"time"

	ocstats "go.opencensus.io/stats"
	"go.opencensus.io/tag"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/grpclog"
	"google.golang.org/grpc/stats"
	"google.golang.org/grpc/status"
)

type rpcData struct {
	sentCount, sentBytes, recvCount, recvBytes int64
	startTime                                  time.Time
	method                                     string
}

type grpcInstrumentationKey string

var (
	rpcDataKey = grpcInstrumentationKey("lotus-rpcData")
)

func methodName(fullname string) string {
	return strings.TrimLeft(fullname, "/")
}

func statsHandleRPC(ctx context.Context, s stats.RPCStats) {
	switch st := s.(type) {
	case *stats.Begin, *stats.OutHeader, *stats.InHeader, *stats.InTrailer, *stats.OutTrailer:
		// do nothing for client
	case *stats.OutPayload:
		handleRPCOutPayload(ctx, st)
	case *stats.InPayload:
		handleRPCInPayload(ctx, st)
	case *stats.End:
		handleRPCEnd(ctx, st)
	default:
		grpclog.Infof("unexpected stats: %T", st)
	}
}

func handleRPCOutPayload(ctx context.Context, s *stats.OutPayload) {
	d, ok := ctx.Value(rpcDataKey).(*rpcData)
	if !ok {
		if grpclog.V(2) {
			grpclog.Infoln("Failed to retrieve *rpcData from context.")
		}
		return
	}
	atomic.AddInt64(&d.sentBytes, int64(s.Length))
	atomic.AddInt64(&d.sentCount, 1)
}

func handleRPCInPayload(ctx context.Context, s *stats.InPayload) {
	d, ok := ctx.Value(rpcDataKey).(*rpcData)
	if !ok {
		if grpclog.V(2) {
			grpclog.Infoln("Failed to retrieve *rpcData from context.")
		}
		return
	}
	atomic.AddInt64(&d.recvBytes, int64(s.Length))
	atomic.AddInt64(&d.recvCount, 1)
}

func handleRPCEnd(ctx context.Context, s *stats.End) {
	d, ok := ctx.Value(rpcDataKey).(*rpcData)
	if !ok {
		if grpclog.V(2) {
			grpclog.Infoln("Failed to retrieve *rpcData from context.")
		}
		return
	}
	elapsedTime := time.Since(d.startTime)

	var st string
	if s.Error != nil {
		s, ok := status.FromError(s.Error)
		if ok {
			st = statusCodeToString(s)
		}
	} else {
		st = "OK"
	}

	latencyMillis := float64(elapsedTime) / float64(time.Millisecond)
	if !s.Client {
		return
	}
	ocstats.RecordWithTags(ctx,
		[]tag.Mutator{
			tag.Upsert(KeyClientMethod, methodName(d.method)),
			tag.Upsert(KeyClientStatus, st),
		},
		ClientSentBytesPerRPC.M(atomic.LoadInt64(&d.sentBytes)),
		ClientSentMessagesPerRPC.M(atomic.LoadInt64(&d.sentCount)),
		ClientReceivedMessagesPerRPC.M(atomic.LoadInt64(&d.recvCount)),
		ClientReceivedBytesPerRPC.M(atomic.LoadInt64(&d.recvBytes)),
		ClientRoundtripLatency.M(latencyMillis))
}

func statusCodeToString(s *status.Status) string {
	// see https://github.com/grpc/grpc/blob/master/doc/statuscodes.md
	switch c := s.Code(); c {
	case codes.OK:
		return "OK"
	case codes.Canceled:
		return "CANCELLED"
	case codes.Unknown:
		return "UNKNOWN"
	case codes.InvalidArgument:
		return "INVALID_ARGUMENT"
	case codes.DeadlineExceeded:
		return "DEADLINE_EXCEEDED"
	case codes.NotFound:
		return "NOT_FOUND"
	case codes.AlreadyExists:
		return "ALREADY_EXISTS"
	case codes.PermissionDenied:
		return "PERMISSION_DENIED"
	case codes.ResourceExhausted:
		return "RESOURCE_EXHAUSTED"
	case codes.FailedPrecondition:
		return "FAILED_PRECONDITION"
	case codes.Aborted:
		return "ABORTED"
	case codes.OutOfRange:
		return "OUT_OF_RANGE"
	case codes.Unimplemented:
		return "UNIMPLEMENTED"
	case codes.Internal:
		return "INTERNAL"
	case codes.Unavailable:
		return "UNAVAILABLE"
	case codes.DataLoss:
		return "DATA_LOSS"
	case codes.Unauthenticated:
		return "UNAUTHENTICATED"
	default:
		return "CODE_" + strconv.FormatInt(int64(c), 10)
	}
}
