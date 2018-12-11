package model

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"
)

type TestStatus string

const (
	TestSucceeded TestStatus = "Succeeded"
	TestFailed               = "Failed"
	TestCancelled            = "Cancelled"
)

type Result struct {
	TestID                   string
	Status                   TestStatus
	MetricsSummary           *MetricsSummary
	FailureReason            string
	FailedChecks             []string
	StartedTimestamp         time.Time
	FinishedTimestamp        time.Time
	GrafanaGRPCDashboardsURL string
	GrafanaHTTPDashboardsURL string
}

func (r *Result) SetFailed(reason string) {
	r.Status = TestFailed
	r.FailureReason = reason
}

func (r *Result) SetGrafanaDashboardURLs(base string) {
	base = strings.TrimRight(base, "/")
	var from int64 = r.StartedTimestamp.Add(-time.Minute).UnixNano() / 1e6
	var to int64 = r.FinishedTimestamp.Add(time.Minute).UnixNano() / 1e6
	r.GrafanaGRPCDashboardsURL = fmt.Sprintf("%s/dashboard/db/grpc?from=%d&to=%d&var-testId=%s", base, from, to, r.TestID)
	r.GrafanaHTTPDashboardsURL = fmt.Sprintf("%s/dashboard/db/http?from=%d&to=%d&var-testId=%s", base, from, to, r.TestID)
}

func (r *Result) Render(format RenderFormat) ([]byte, error) {
	switch format {
	case RenderFormatMarkdown:
		return renderTemplate(r, markdownTemplate)
	case RenderFormatText:
		return renderTemplate(r, textTemplate)
	case RenderFormatJson:
		return json.Marshal(r)
	}
	return nil, fmt.Errorf("unsupported render format: %s", format)
}
