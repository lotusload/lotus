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
	"time"

	"go.opencensus.io/tag"
	"golang.org/x/net/context"
	"google.golang.org/grpc/grpclog"
	"google.golang.org/grpc/stats"
)

type ClientHandler struct {
}

func (c *ClientHandler) HandleConn(ctx context.Context, cs stats.ConnStats) {
	// no-op
}

func (c *ClientHandler) TagConn(ctx context.Context, cti *stats.ConnTagInfo) context.Context {
	// no-op
	return ctx
}

func (c *ClientHandler) HandleRPC(ctx context.Context, rs stats.RPCStats) {
	statsHandleRPC(ctx, rs)
}

func (c *ClientHandler) TagRPC(ctx context.Context, rti *stats.RPCTagInfo) context.Context {
	ctx = c.statsTagRPC(ctx, rti)
	return ctx
}

func (h *ClientHandler) statsTagRPC(ctx context.Context, info *stats.RPCTagInfo) context.Context {
	startTime := time.Now()
	if info == nil {
		if grpclog.V(2) {
			grpclog.Infoln("Failed to retrieve *rpcData from context.")
		}
		return ctx
	}
	d := &rpcData{
		startTime: startTime,
		method:    info.FullMethodName,
	}
	ts := tag.FromContext(ctx)
	if ts != nil {
		encoded := tag.Encode(ts)
		ctx = stats.SetTags(ctx, encoded)
	}
	return context.WithValue(ctx, rpcDataKey, d)
}
