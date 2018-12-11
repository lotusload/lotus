package model

const (
	textTemplate = `
TestID:        {{ .TestID }}
TestStatus:    {{ .Status }}
{{- if eq .Status "Failed" }}
    Reason: {{ .FailureReason }}
{{- if gt (len .FailedChecks) 0 }}
    FailedChecks:  {{ .FailedChecks }}
{{- end }}
{{- end }}
Start:         {{ formatTime .StartedTimestamp }}
End:           {{ formatTime .FinishedTimestamp }}

MetricsSummary:

{{- if .MetricsSummary }}

1. Virtual User
  - Started:             {{ formatValue .MetricsSummary.VirtualUserStartedTotal }}
  - Failed:              {{ formatValue .MetricsSummary.VirtualUserFailedTotal }}

2. GRPC
  - RPCTotal:            {{ formatValue .MetricsSummary.GRPCRPCTotal }}
  - FailurePercentage:   {{ formatValue .MetricsSummary.GRPCFailurePercentage }}

GroupByMethod:
{{ formatGRPCByMethod .MetricsSummary.GRPCByMethod .MetricsSummary.GRPCAll }}
Grafana: {{ .GrafanaGRPCDashboardsURL }}

3. HTTP
  - RequestTotal:        {{ formatValue .MetricsSummary.HTTPRequestTotal }}
  - FailurePercentage:   {{ formatValue .MetricsSummary.HTTPFailurePercentage }}

GroupByPath:
{{ formatHTTPByPath .MetricsSummary.HTTPByPath .MetricsSummary.HTTPAll }}
Grafana: {{ .GrafanaHTTPDashboardsURL }}
{{- else }}

  No data
{{- end }}
`
)

var (
	markdownTemplate = textTemplate
)
