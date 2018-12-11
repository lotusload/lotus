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

package httpmetrics

import (
	"go.opencensus.io/stats"
	"go.opencensus.io/stats/view"
	"go.opencensus.io/tag"
)

var (
	ClientSentBytes = stats.Int64(
		"http/client/sent_bytes",
		"Total bytes sent in request body (not including headers)",
		stats.UnitBytes,
	)
	ClientReceivedBytes = stats.Int64(
		"http/client/received_bytes",
		"Total bytes received in response bodies (not including headers but including error responses with bodies)",
		stats.UnitBytes,
	)
	ClientRoundtripLatency = stats.Float64(
		"http/client/roundtrip_latency",
		"Time between first byte of request headers sent to last byte of response received, or terminal error",
		stats.UnitMilliseconds,
	)
)

// KeyClientRoute is a low cardinality string representing the logical
// handler of the request. This is usually the pattern of the request.
var (
	KeyClientHost, _   = tag.NewKey("http_client_host")
	KeyClientRoute, _  = tag.NewKey("http_client_route")
	KeyClientMethod, _ = tag.NewKey("http_client_method")
	KeyClientStatus, _ = tag.NewKey("http_client_status")

	HTTPClientTagKeys = []tag.Key{
		KeyClientHost,
		KeyClientMethod,
		KeyClientRoute,
		KeyClientStatus,
	}
)

var (
	ClientSentBytesDistribution = &view.View{
		Name:        "http/client/sent_bytes",
		Measure:     ClientSentBytes,
		Aggregation: DefaultSizeDistribution,
		Description: "Total bytes sent in request body (not including headers)",
		TagKeys:     HTTPClientTagKeys,
	}

	ClientReceivedBytesDistribution = &view.View{
		Name:        "http/client/received_bytes",
		Measure:     ClientReceivedBytes,
		Aggregation: DefaultSizeDistribution,
		Description: "Total bytes received in response bodies (not including headers but including error responses with bodies)",
		TagKeys:     HTTPClientTagKeys,
	}

	ClientRoundtripLatencyDistribution = &view.View{
		Name:        "http/client/roundtrip_latency",
		Measure:     ClientRoundtripLatency,
		Aggregation: DefaultLatencyDistribution,
		Description: "End-to-end latency",
		TagKeys:     HTTPClientTagKeys,
	}

	ClientCompletedCount = &view.View{
		Name:        "http/client/completed_count",
		Measure:     ClientRoundtripLatency,
		Aggregation: view.Count(),
		Description: "Count of completed requests",
		TagKeys:     HTTPClientTagKeys,
	}
)

var DefaultClientViews = []*view.View{
	ClientCompletedCount,
	ClientSentBytesDistribution,
	ClientReceivedBytesDistribution,
	ClientRoundtripLatencyDistribution,
}

var (
	DefaultSizeDistribution    = view.Distribution(0, 1024, 2048, 4096, 16384, 65536, 262144, 1048576, 4194304, 16777216, 67108864, 268435456, 1073741824, 4294967296)
	DefaultLatencyDistribution = view.Distribution(0, 1, 2, 3, 4, 5, 6, 8, 10, 13, 16, 20, 25, 30, 40, 50, 65, 80, 100, 130, 160, 200, 250, 300, 400, 500, 650, 800, 1000, 2000, 5000, 10000, 20000, 50000, 100000)
)
