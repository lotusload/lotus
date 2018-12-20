// Copyright (c) 2018 Lotus Load
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in all
// copies or substantial portions of the Software.

// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.

package resource

import (
	"bytes"
	"text/template"

	"github.com/lotusload/lotus/pkg/app/lotus/apis/lotus/v1beta1"
)

func renderTemplate(params interface{}, tpl string) ([]byte, error) {
	template, err := template.New("template").Parse(tpl)
	if err != nil {
		return nil, err
	}
	var buffer bytes.Buffer
	err = template.Execute(&buffer, params)
	if err != nil {
		return nil, err
	}
	return buffer.Bytes(), nil
}

type prometheusConfigParams struct {
	Name        string
	Namespace   string
	ServiceName string
	RuleFiles   []string
}

const prometheusConfigTemplate = `
global:
  scrape_interval: 5s
  scrape_timeout: 5s
  evaluation_interval: 5s
  external_labels:
    monitor: prometheus
    replica: {{ .Name }}
scrape_configs:
- job_name: lotus-runer
  metrics_path: /metrics
  scheme: http
  kubernetes_sd_configs:
  - api_server: null
    role: endpoints
    namespaces:
      names:
      - {{ .Namespace  }}
  relabel_configs:
  - source_labels: [__meta_kubernetes_service_name]
    separator: ;
    regex: {{ .ServiceName }}
    replacement: $1
    action: keep
  - source_labels: [__meta_kubernetes_namespace]
    separator: ;
    regex: (.*)
    target_label: namespace
    replacement: $1
    action: replace
  - source_labels: [__meta_kubernetes_pod_name]
    separator: ;
    regex: (.*)
    target_label: pod
    replacement: $1
    action: replace
  - source_labels: [__meta_kubernetes_service_name]
    separator: ;
    regex: (.*)
    target_label: service
    replacement: $1
    action: replace
  - source_labels: [__meta_kubernetes_service_name]
    separator: ;
    regex: (.*)
    target_label: job
    replacement: ${1}
    action: replace
{{- if gt (len .RuleFiles) 0 }}
rule_files:
{{- range .RuleFiles }}
  - {{ . }}
{{- end }}
{{- end }}
`

type prometheusRuleParams struct {
	Alerts []v1beta1.LotusCheck
}

const prometheusRuleTemplate = `
groups:
- name: lotus
  rules:
{{- range .Alerts }}
  - alert: {{ .Name }}
    expr: {{ .Expr }}
    for: {{ .For }}
{{- end }}
  - record: lotus_virtual_user_failure_percentage
    expr: 100 * sum by (job) (lotus_virtual_user_count{virtual_user_status="failed"}) / sum by (job) (lotus_virtual_user_count{virtual_user_status="started"})
  - record: lotus_grpc_client_completed_rpcs_per_second:method
    expr: sum by (job, grpc_client_method) (rate(lotus_grpc_client_completed_rpcs[1m]))
  - record: lotus_grpc_client_completed_rpcs_per_second:status
    expr: sum by (job, grpc_client_status) (rate(lotus_grpc_client_completed_rpcs[1m]))
  - record: lotus_grpc_client_completed_rpcs_failure_percentage:method
    expr: 100 * sum by (job, grpc_client_method) (rate(lotus_grpc_client_completed_rpcs{grpc_client_status!~"OK|NOT_FOUND|ALREADY_EXISTS"}[1m])) / sum by (job, grpc_client_method) (rate(lotus_grpc_client_completed_rpcs[1m]))
  - record: lotus_grpc_client_roundtrip_latency:method
    expr: sum by (job, grpc_client_method) (rate(lotus_grpc_client_roundtrip_latency_sum[1m])) / sum by (job, grpc_client_method) (rate(lotus_grpc_client_roundtrip_latency_count[1m]))
  - record: lotus_grpc_client_sent_bytes_per_rpc:method
    expr: sum by (job, grpc_client_method) (rate(lotus_grpc_client_sent_bytes_per_rpc_sum[1m])) / sum by (job, grpc_client_method) (rate(lotus_grpc_client_sent_bytes_per_rpc_count[1m]))
  - record: lotus_grpc_client_received_bytes_per_rpc:method
    expr: sum by (job, grpc_client_method) (rate(lotus_grpc_client_received_bytes_per_rpc_sum[1m])) / sum by (job, grpc_client_method) (rate(lotus_grpc_client_received_bytes_per_rpc_count[1m]))
  - record: lotus_http_client_completed_requests_per_second:host:route:method
    expr: sum by (job, http_client_host, http_client_route, http_client_method) (rate(lotus_http_client_completed_count[1m]))
  - record: lotus_http_client_completed_requests_5xx_percentage:host:route:method
    expr: 100 * sum by (job, http_client_host, http_client_route, http_client_method) (rate(lotus_http_client_completed_count{http_client_status=~"5.."}[1m])) / sum by (job, http_client_host, http_client_route, http_client_method) (rate(lotus_http_client_completed_count[1m]))
  - record: lotus_http_client_roundtrip_latency:host:route:method
    expr: sum by (job, http_client_host, http_client_route, http_client_method) (rate(lotus_http_client_roundtrip_latency_sum[1m])) / sum by (job, http_client_host, http_client_route, http_client_method) (rate(lotus_http_client_roundtrip_latency_count[1m]))
  - record: lotus_http_client_sent_bytes:host:route:method
    expr: sum by (job, http_client_host, http_client_route, http_client_method) (rate(lotus_http_client_sent_bytes_sum[1m])) / sum by (job, http_client_host, http_client_route, http_client_method) (rate(lotus_http_client_sent_bytes_count[1m]))
  - record: lotus_http_client_received_bytes:host:route:method
    expr: sum by (job, http_client_host, http_client_route, http_client_method) (rate(lotus_http_client_received_bytes_sum[1m])) / sum by (job, http_client_host, http_client_route, http_client_method) (rate(lotus_http_client_received_bytes_count[1m]))
`
