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
