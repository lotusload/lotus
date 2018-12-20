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
