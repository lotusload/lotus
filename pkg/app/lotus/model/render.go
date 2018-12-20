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
	"bytes"
	"fmt"
	"math"
	"text/template"
	"time"
)

type RenderFormat string

const (
	RenderFormatMarkdown RenderFormat = "markdown"
	RenderFormatText                  = "text"
	RenderFormatJson                  = "json"

	noDataMark = "--"
	timeFormat = "15:04:05 2006-01-02"
)

type valueByLabelList struct {
	name   string
	values ValueByLabel
}

func renderTemplate(result *Result, tpl string) ([]byte, error) {
	funcMap := template.FuncMap{
		"formatValue":        formatValue,
		"formatTime":         formatTime,
		"formatGRPCByMethod": formatGRPCByMethod,
		"formatHTTPByPath":   formatHTTPByPath,
	}
	template, err := template.New("result").Funcs(funcMap).Parse(tpl)
	if err != nil {
		return nil, err
	}
	var buffer bytes.Buffer
	err = template.Execute(&buffer, result)
	if err != nil {
		return nil, err
	}
	return buffer.Bytes(), nil
}

func formatGRPCByMethod(data map[string]ValueByLabel, all ValueByLabel) string {
	keys := []struct {
		Desc string
		Key  string
	}{
		{Desc: "RPCs", Key: GRPCRPCsKey},
		{Desc: `Failure%`, Key: GRPCFailurePercentageKey},
		{Desc: "Latency(ms)", Key: GRPCLatencyAvgKey},
		{Desc: "SentBytes", Key: GRPCSentBytesAvgKey},
		{Desc: "RecvBytes", Key: GRPCReceivedBytesAvgKey},
	}
	methodMaxLength := 5
	for method, _ := range data {
		if len(method) > methodMaxLength {
			methodMaxLength = len(method)
		}
	}
	var b bytes.Buffer
	format := func(key string, values ValueByLabel) string {
		value, ok := values[key]
		if !ok {
			value = NoDataValue
		}
		return formatValue(value)
	}
	titleFormat := fmt.Sprintf("    %%-%ds  %%-8s  %%-8s  %%-12s  %%-8s  %%-8s\n\n", methodMaxLength)
	detailFormat := fmt.Sprintf("  - %%-%ds  %%-8s  %%-8s  %%-12s  %%-8s  %%-8s\n", methodMaxLength)

	b.WriteString(fmt.Sprintf(titleFormat, "", keys[0].Desc, keys[1].Desc, keys[2].Desc, keys[3].Desc, keys[4].Desc))
	rows := make([]valueByLabelList, 0, len(data)+1)
	for method, values := range data {
		rows = append(rows, valueByLabelList{
			name:   method,
			values: values,
		})
	}
	rows = append(rows, valueByLabelList{
		name:   "all",
		values: all,
	})
	for _, row := range rows {
		b.WriteString(fmt.Sprintf(detailFormat,
			row.name,
			format(keys[0].Key, row.values),
			format(keys[1].Key, row.values),
			format(keys[2].Key, row.values),
			format(keys[3].Key, row.values),
			format(keys[4].Key, row.values),
		))
	}
	return b.String()
}

func formatHTTPByPath(data map[string]ValueByLabel, all ValueByLabel) string {
	keys := []struct {
		Desc string
		Key  string
	}{
		{Desc: "Requests", Key: HTTPRequestsKey},
		{Desc: `Failure%`, Key: HTTPFailurePercentageKey},
		{Desc: "Latency(ms)", Key: HTTPLatencyAvgKey},
		{Desc: "SentBytes", Key: HTTPSentBytesAvgKey},
		{Desc: "RecvBytes", Key: HTTPReceivedBytesAvgKey},
	}
	pathMaxLength := 5
	for path, _ := range data {
		if len(path) > pathMaxLength {
			pathMaxLength = len(path)
		}
	}
	var b bytes.Buffer
	format := func(key string, values ValueByLabel) string {
		value, ok := values[key]
		if !ok {
			value = NoDataValue
		}
		return formatValue(value)
	}
	titleFormat := fmt.Sprintf("    %%-%ds  %%-8s  %%-8s  %%-12s  %%-8s  %%-8s\n\n", pathMaxLength)
	detailFormat := fmt.Sprintf("  - %%-%ds  %%-8s  %%-8s  %%-12s  %%-8s  %%-8s\n", pathMaxLength)

	b.WriteString(fmt.Sprintf(titleFormat, "", keys[0].Desc, keys[1].Desc, keys[2].Desc, keys[3].Desc, keys[4].Desc))
	rows := make([]valueByLabelList, 0, len(data)+1)
	for path, values := range data {
		rows = append(rows, valueByLabelList{
			name:   path,
			values: values,
		})
	}
	rows = append(rows, valueByLabelList{
		name:   "all",
		values: all,
	})
	for _, row := range rows {
		b.WriteString(fmt.Sprintf(detailFormat,
			row.name,
			format(keys[0].Key, row.values),
			format(keys[1].Key, row.values),
			format(keys[2].Key, row.values),
			format(keys[3].Key, row.values),
			format(keys[4].Key, row.values),
		))
	}
	return b.String()
}

// https://en.wikipedia.org/wiki/Metric_prefix
func formatValue(v float64) string {
	if v == NoDataValue {
		return noDataMark
	}
	if v == 0 || math.IsNaN(v) || math.IsInf(v, 0) {
		return fmt.Sprintf("%.4g", v)
	}
	if math.Abs(v) >= 1 {
		prefix := ""
		for _, p := range []string{"k", "M", "G", "T", "P", "E", "Z", "Y"} {
			if math.Abs(v) < 1000 {
				break
			}
			prefix = p
			v /= 1000
		}
		return fmt.Sprintf("%.4g%s", v, prefix)
	}
	prefix := ""
	for _, p := range []string{"m", "u", "n", "p", "f", "a", "z", "y"} {
		if math.Abs(v) >= 1 {
			break
		}
		prefix = p
		v *= 1000
	}
	return fmt.Sprintf("%.4g%s", v, prefix)
}

func formatTime(t time.Time) string {
	return t.Format(timeFormat)
}
