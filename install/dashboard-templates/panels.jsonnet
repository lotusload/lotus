local grafana = import 'grafonnet/grafana.libsonnet';
local common=import 'common.jsonnet';
local prometheus = grafana.prometheus;
local graphPanel = grafana.graphPanel;

{
  ### Worker & VirtualUser pannels

  workerNum:: common.panel.new(
    title='Number of workers',
  )
  .addTarget(prometheus.target(
    'sum (up{job=~"$testId-worker"})',
    legendFormat=' ')
  ),

  virtualUserNum:: common.panel.new(
    title='Number of virtual users',
  )
  .addTarget(prometheus.target(
    'sum by (virtual_user_status) (lotus_virtual_user_count{job=~"$testId-worker"})',
    legendFormat='virtual_user_status')
  ),

  ### GRPC pannels

  rpcsPerSecond:: common.panel.new(
    title='RPCs / second',
  )
  .addTarget(prometheus.target(
    'lotus_grpc_client_completed_rpcs_per_second:method{job=~"$testId-worker"}',
    legendFormat='{{ grpc_client_method }}')
  ),

  rpcsPerSecondByStatus:: common.panel.new(
    title='RPCs / seconds grouping by status',
  )
  .addTarget(prometheus.target(
    'lotus_grpc_client_completed_rpcs_per_second:status{job=~"$testId-worker"}',
    legendFormat='{{ grpc_client_status }}')
  ),

  percentageFailedRPCs:: common.panel.new(
    title='Percentage of failed RPCs',
    format=common.format.percent_0_100,
  )
  .addTarget(prometheus.target(
    'lotus_grpc_client_completed_rpcs_failure_percentage:method{job=~"$testId-worker"}',
    legendFormat='{{ grpc_client_method }}')
  ),

  rpcLatency:: common.panel.new(
    title='Latency',
    format=common.format.millisecond,
  )
  .addTarget(prometheus.target(
    'lotus_grpc_client_roundtrip_latency:method{job=~"$testId-worker"}',
    legendFormat='{{ grpc_client_method }}')
  ),

  rpcSentBytes:: common.panel.new(
    title='Sent Bytes',
    format=common.format.bytes,
  )
  .addTarget(prometheus.target(
    'lotus_grpc_client_sent_bytes_per_rpc:method{job=~"$testId-worker"}',
    legendFormat='{{ grpc_client_method }}')
  ),

  rpcReceivedBytes:: common.panel.new(
    title='Received Bytes',
    format=common.format.bytes,
  )
  .addTarget(prometheus.target(
    'lotus_grpc_client_received_bytes_per_rpc:method{job=~"$testId-worker"}',
    legendFormat='{{ grpc_client_method }}')
  ),

  ### HTTP Pannels

  httpRequestsPerSecond:: common.panel.new(
    title='Requests / second',
  )
  .addTarget(prometheus.target(
    'lotus_http_client_completed_requests_per_second:host:route:method{job=~"$testId-worker"}',
    legendFormat='{{ http_client_method }}/{{ http_client_host }}{{ http_client_route }}')
  ),

  percentageOf5xxRequests:: common.panel.new(
    title='Percentage of 5xx Requests',
    format=common.format.percent_0_100,
  )
  .addTarget(prometheus.target(
    'lotus_http_client_completed_requests_5xx_percentage:host:route:method{job=~"$testId-worker"}',
    legendFormat='{{ http_client_method }}/{{ http_client_host }}{{ http_client_route }}')
  ),

  httpRequestLatency:: common.panel.new(
    title='Latency',
    format=common.format.millisecond,
  )
  .addTarget(prometheus.target(
    'lotus_http_client_roundtrip_latency:host:route:method{job=~"$testId-worker"}',
    legendFormat='{{ http_client_method }}/{{ http_client_host }}{{ http_client_route }}')
  ),

  httpRequestSentBytes:: common.panel.new(
    title='Sent Bytes',
    format=common.format.bytes,
  )
  .addTarget(prometheus.target(
    'lotus_http_client_sent_bytes:host:route:method{job=~"$testId-worker"}',
    legendFormat='{{ http_client_method }}/{{ http_client_host }}{{ http_client_route }}')
  ),

  httpRequestReceivedBytes:: common.panel.new(
    title='Received Bytes',
    format=common.format.bytes,
  )
  .addTarget(prometheus.target(
    'lotus_http_client_received_bytes:host:route:method{job=~"$testId-worker"}',
    legendFormat='{{ http_client_method }}/{{ http_client_host }}{{ http_client_route }}')
  ),
}