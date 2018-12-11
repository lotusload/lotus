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
	"go.opencensus.io/stats"
	"go.opencensus.io/stats/view"
	"go.opencensus.io/tag"
)

var (
	ClientSentMessagesPerRPC = stats.Int64(
		"grpc/client/sent_messages_per_rpc",
		"Number of messages sent in the RPC (always 1 for non-streaming RPCs).",
		stats.UnitDimensionless,
	)

	ClientSentBytesPerRPC = stats.Int64(
		"grpc/client/sent_bytes_per_rpc",
		"Total bytes sent across all request messages per RPC.",
		stats.UnitBytes,
	)

	ClientReceivedMessagesPerRPC = stats.Int64(
		"grpc/client/received_messages_per_rpc",
		"Number of response messages received per RPC (always 1 for non-streaming RPCs).",
		stats.UnitDimensionless,
	)

	ClientReceivedBytesPerRPC = stats.Int64(
		"grpc/client/received_bytes_per_rpc",
		"Total bytes received across all response messages per RPC.",
		stats.UnitBytes,
	)

	ClientRoundtripLatency = stats.Float64(
		"grpc/client/roundtrip_latency",
		"Time between first byte of request sent to last byte of response received, or terminal error.",
		stats.UnitMilliseconds,
	)
)

var (
	KeyClientMethod, _ = tag.NewKey("grpc_client_method")
	KeyClientStatus, _ = tag.NewKey("grpc_client_status")
)

var (
	ClientSentBytesPerRPCView = &view.View{
		Measure:     ClientSentBytesPerRPC,
		Name:        "grpc/client/sent_bytes_per_rpc",
		Description: "Distribution of bytes sent per RPC, by method.",
		TagKeys:     []tag.Key{KeyClientMethod},
		Aggregation: DefaultBytesDistribution,
	}

	ClientReceivedBytesPerRPCView = &view.View{
		Measure:     ClientReceivedBytesPerRPC,
		Name:        "grpc/client/received_bytes_per_rpc",
		Description: "Distribution of bytes received per RPC, by method.",
		TagKeys:     []tag.Key{KeyClientMethod},
		Aggregation: DefaultBytesDistribution,
	}

	ClientRoundtripLatencyView = &view.View{
		Measure:     ClientRoundtripLatency,
		Name:        "grpc/client/roundtrip_latency",
		Description: "Distribution of round-trip latency, by method.",
		TagKeys:     []tag.Key{KeyClientMethod},
		Aggregation: DefaultMillisecondsDistribution,
	}

	ClientCompletedRPCsView = &view.View{
		Measure:     ClientRoundtripLatency,
		Name:        "grpc/client/completed_rpcs",
		Description: "Count of RPCs by method and status.",
		TagKeys:     []tag.Key{KeyClientMethod, KeyClientStatus},
		Aggregation: view.Count(),
	}

	ClientSentMessagesPerRPCView = &view.View{
		Measure:     ClientSentMessagesPerRPC,
		Name:        "grpc/client/sent_messages_per_rpc",
		Description: "Distribution of sent messages count per RPC, by method.",
		TagKeys:     []tag.Key{KeyClientMethod},
		Aggregation: DefaultMessageCountDistribution,
	}

	ClientReceivedMessagesPerRPCView = &view.View{
		Measure:     ClientReceivedMessagesPerRPC,
		Name:        "grpc/client/received_messages_per_rpc",
		Description: "Distribution of received messages count per RPC, by method.",
		TagKeys:     []tag.Key{KeyClientMethod},
		Aggregation: DefaultMessageCountDistribution,
	}
)

var DefaultClientViews = []*view.View{
	ClientSentBytesPerRPCView,
	ClientReceivedBytesPerRPCView,
	ClientRoundtripLatencyView,
	ClientCompletedRPCsView,
	ClientSentMessagesPerRPCView,
	ClientReceivedMessagesPerRPCView,
}

var (
	DefaultBytesDistribution        = view.Distribution(0, 1024, 2048, 4096, 16384, 65536, 262144, 1048576, 4194304, 16777216, 67108864, 268435456, 1073741824, 4294967296)
	DefaultMillisecondsDistribution = view.Distribution(0, 0.01, 0.05, 0.1, 0.3, 0.6, 0.8, 1, 2, 3, 4, 5, 6, 8, 10, 13, 16, 20, 25, 30, 40, 50, 65, 80, 100, 130, 160, 200, 250, 300, 400, 500, 650, 800, 1000, 2000, 5000, 10000, 20000, 50000, 100000)
	DefaultMessageCountDistribution = view.Distribution(0, 1, 2, 4, 8, 16, 32, 64, 128, 256, 512, 1024, 2048, 4096, 8192, 16384, 32768, 65536)
)
