package prometheus

const (
	// VirtualUser Queries
	vuStartedTotalQuery = `sum(max_over_time(lotus_virtual_user_count{virtual_user_status="started"}[1h]))`

	vuFailedTotalQuery = `sum(max_over_time(lotus_virtual_user_count{virtual_user_status="failed"}[1h]))`

	// GRPC Queries
	grpcRPCTotalQuery = `sum(max_over_time(lotus_grpc_client_completed_rpcs[1h]))`

	grpcFailurePercentageQuery = `100 * sum(max_over_time(lotus_grpc_client_completed_rpcs{grpc_client_status!~"OK|NOT_FOUND"}[1h])) /
		sum(max_over_time(lotus_grpc_client_completed_rpcs[1h]))`

	grpcLatencyAvgQuery = `sum(max_over_time(lotus_grpc_client_roundtrip_latency_sum[1h])) /
		sum(max_over_time(lotus_grpc_client_roundtrip_latency_count[1h]))`

	grpcSentBytesAvgQuery = `sum(max_over_time(lotus_grpc_client_sent_bytes_per_rpc_sum[1h])) /
		sum(max_over_time(lotus_grpc_client_sent_bytes_per_rpc_count[1h]))`

	grpcReceivedBytesAvgQuery = `sum(max_over_time(lotus_grpc_client_received_bytes_per_rpc_sum[1h])) /
		sum(max_over_time(lotus_grpc_client_received_bytes_per_rpc_count[1h]))`

	// GRPCByMethod Queries
	grpcMethodLabel = "grpc_client_method"

	grpcRPCsByMethodQuery = `sum by(grpc_client_method) (max_over_time(lotus_grpc_client_completed_rpcs[1h]))`

	grpcFailurePercentageByMethodQuery = `100 * sum by(grpc_client_method) (max_over_time(lotus_grpc_client_completed_rpcs{grpc_client_status!~"OK|NOT_FOUND"}[1h])) /
		sum by(grpc_client_method) (max_over_time(lotus_grpc_client_completed_rpcs[1h]))`

	grpcLatencyAvgByMethodQuery = `sum by(grpc_client_method) (max_over_time(lotus_grpc_client_roundtrip_latency_sum[1h])) /
		sum by(grpc_client_method) (max_over_time(lotus_grpc_client_roundtrip_latency_count[1h]))`

	grpcSentBytesAvgByMethodQuery = `sum by(grpc_client_method) (max_over_time(lotus_grpc_client_sent_bytes_per_rpc_sum[1h])) /
		sum by(grpc_client_method) (max_over_time(lotus_grpc_client_sent_bytes_per_rpc_count[1h]))`

	grpcReceivedBytesAvgByMethodQuery = `sum by(grpc_client_method) (max_over_time(lotus_grpc_client_received_bytes_per_rpc_sum[1h])) /
		sum by(grpc_client_method) (max_over_time(lotus_grpc_client_received_bytes_per_rpc_count[1h]))`

	// HTTP Queries
	httpRequestTotalQuery = `sum(max_over_time(lotus_http_client_completed_count[1h]))`

	httpFailurePercentageQuery = `100 * sum(max_over_time(lotus_http_client_completed_count{http_client_status=~"5.."}[1h])) /
		sum(max_over_time(lotus_http_client_completed_count[1h]))`

	httpLatencyAvgQuery = `sum(max_over_time(lotus_http_client_roundtrip_latency_sum[1h])) /
		sum(max_over_time(lotus_http_client_roundtrip_latency_count[1h]))`

	httpSentBytesAvgQuery = `sum(max_over_time(lotus_http_client_sent_bytes_sum[1h])) /
		sum(max_over_time(lotus_http_client_sent_bytes_count[1h]))`

	httpReceivedBytesAvgQuery = `sum(max_over_time(lotus_http_client_received_bytes_sum[1h])) /
		sum(max_over_time(lotus_http_client_received_bytes_count[1h]))`

	// HTTPByPath Queries
	httpRequestsByPathQuery = `sum by(http_client_host,http_client_route,http_client_method) (max_over_time(lotus_http_client_completed_count[1h]))`

	httpFailurePercentageByPathQuery = `100 * sum by(http_client_host,http_client_route,http_client_method) (max_over_time(lotus_http_client_completed_count{http_client_status=~"5.."}[1h])) /
		sum by(http_client_host,http_client_route,http_client_method) (max_over_time(lotus_http_client_completed_count[1h]))`

	httpLatencyAvgByPathQuery = `sum by(http_client_host,http_client_route,http_client_method) (max_over_time(lotus_http_client_roundtrip_latency_sum[1h])) /
		sum by(http_client_host,http_client_route,http_client_method) (max_over_time(lotus_http_client_roundtrip_latency_count[1h]))`

	httpSentBytesAvgByPathQuery = `sum by(http_client_host,http_client_route,http_client_method) (max_over_time(lotus_http_client_sent_bytes_sum[1h])) /
		sum by(http_client_host,http_client_route,http_client_method) (max_over_time(lotus_http_client_sent_bytes_count[1h]))`

	httpReceivedBytesAvgByPathQuery = `sum by(http_client_host,http_client_route,http_client_method) (max_over_time(lotus_http_client_received_bytes_sum[1h])) /
		sum by(http_client_host,http_client_route,http_client_method) (max_over_time(lotus_http_client_received_bytes_count[1h]))`
)
