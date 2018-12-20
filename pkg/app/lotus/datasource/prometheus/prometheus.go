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

package prometheus

import (
	"context"
	"fmt"
	"strings"
	"time"

	promv1 "github.com/prometheus/client_golang/api/prometheus/v1"
	prommodel "github.com/prometheus/common/model"
	"go.uber.org/zap"

	"github.com/lotusload/lotus/pkg/app/lotus/datasource"
	"github.com/lotusload/lotus/pkg/app/lotus/model"
	"github.com/lotusload/lotus/pkg/metrics/grpcmetrics"
	"github.com/lotusload/lotus/pkg/metrics/httpmetrics"
)

const (
	AlertNameLabel   = prommodel.AlertNameLabel
	AlertStateLabel  = "alertstate"
	AlertStateFiring = "firing"
)

type prometheus struct {
	api    promv1.API
	logger *zap.Logger
}

func (p *prometheus) Query(ctx context.Context, query string, ts time.Time) ([]*datasource.Sample, error) {
	v, err := p.api.Query(ctx, query, ts)
	if err != nil {
		return nil, err
	}
	vector, ok := v.(prommodel.Vector)
	if !ok {
		return nil, fmt.Errorf("unsupported value type: %s, %v", v.Type(), v)
	}
	return vectorToSamples(vector), nil
}

func vectorToSamples(vector prommodel.Vector) []*datasource.Sample {
	samples := make([]*datasource.Sample, 0, len(vector))
	for _, s := range vector {
		sample := &datasource.Sample{
			Labels: make(map[string]string, len(s.Metric)),
		}
		for k, v := range s.Metric {
			sample.Labels[string(k)] = string(v)
		}
		sample.Value = float64(s.Value)
		sample.Timestamp = s.Timestamp.Time()
		samples = append(samples, sample)
	}
	return samples
}

func (p *prometheus) Check(ctx context.Context, checks []datasource.Check) (*datasource.CheckResult, error) {
	var ts time.Time
	v, err := p.api.Query(ctx, "ALERTS", ts)
	if err != nil {
		p.logger.Error("failed to run query to get alerts", zap.Error(err))
		return nil, err
	}
	vector, ok := v.(prommodel.Vector)
	if !ok {
		p.logger.Error("unsupported value type", zap.Any("value", v))
		return nil, fmt.Errorf("unsupported value type: %s", v.Type())
	}
	p.logger.Debug("extracting actives", zap.Any("vector", vector), zap.Any("checks", checks))
	return &datasource.CheckResult{
		Actives: extractActives(vector, checks),
	}, nil
}

func extractActives(vector prommodel.Vector, checks []datasource.Check) []string {
	targets := make(map[string]struct{}, len(checks))
	for _, check := range checks {
		targets[check.Name] = struct{}{}
	}
	actives := make(map[string]struct{})
	for _, sample := range vector {
		if sample.Value == 0 {
			continue
		}
		name, ok := sample.Metric[AlertNameLabel]
		if !ok {
			continue
		}
		if _, ok := targets[string(name)]; !ok {
			continue
		}
		state, ok := sample.Metric[AlertStateLabel]
		if !ok {
			continue
		}
		if state != AlertStateFiring {
			continue
		}
		actives[string(name)] = struct{}{}
	}
	list := make([]string, 0, len(actives))
	for k, _ := range actives {
		list = append(list, k)
	}
	return list
}

func (p *prometheus) CollectSummary(ctx context.Context, ts time.Time) (*model.MetricsSummary, error) {
	grpcByMethod, err := p.collectGRPCByMethod(ctx, ts)
	if err != nil {
		return nil, err
	}
	httpByPath, err := p.collectHTTPByPath(ctx, ts)
	if err != nil {
		return nil, err
	}
	grpcAll, err := p.multiQuery(ctx, map[string]string{
		model.GRPCRPCsKey:              grpcRPCTotalQuery,
		model.GRPCFailurePercentageKey: grpcFailurePercentageQuery,
		model.GRPCLatencyAvgKey:        grpcLatencyAvgQuery,
		model.GRPCSentBytesAvgKey:      grpcSentBytesAvgQuery,
		model.GRPCReceivedBytesAvgKey:  grpcReceivedBytesAvgQuery,
	}, ts)
	if err != nil {
		return nil, err
	}
	httpAll, err := p.multiQuery(ctx, map[string]string{
		model.HTTPRequestsKey:          httpRequestTotalQuery,
		model.HTTPFailurePercentageKey: httpFailurePercentageQuery,
		model.HTTPLatencyAvgKey:        httpLatencyAvgQuery,
		model.HTTPSentBytesAvgKey:      httpSentBytesAvgQuery,
		model.HTTPReceivedBytesAvgKey:  httpReceivedBytesAvgQuery,
	}, ts)
	if err != nil {
		return nil, err
	}
	summary := &model.MetricsSummary{
		GRPCRPCTotal:          grpcAll[model.GRPCRPCsKey],
		GRPCFailurePercentage: grpcAll[model.GRPCFailurePercentageKey],
		GRPCAll:               grpcAll,
		GRPCByMethod:          grpcByMethod,
		HTTPRequestTotal:      httpAll[model.HTTPRequestsKey],
		HTTPFailurePercentage: httpAll[model.HTTPFailurePercentageKey],
		HTTPAll:               httpAll,
		HTTPByPath:            httpByPath,
	}
	queries := []struct {
		Query  string
		Target *float64
	}{
		{
			Query:  vuStartedTotalQuery,
			Target: &summary.VirtualUserStartedTotal,
		},
		{
			Query:  vuFailedTotalQuery,
			Target: &summary.VirtualUserFailedTotal,
		},
	}
	for i := range queries {
		value, err := p.queryOne(ctx, queries[i].Query, ts)
		if err != nil {
			return nil, err
		}
		*queries[i].Target = value
	}
	return summary, nil
}

func (p *prometheus) collectGRPCByMethod(ctx context.Context, ts time.Time) (map[string]model.ValueByLabel, error) {
	result := make(map[string]model.ValueByLabel)
	queries := map[string]string{
		model.GRPCRPCsKey:              grpcRPCsByMethodQuery,
		model.GRPCFailurePercentageKey: grpcFailurePercentageByMethodQuery,
		model.GRPCLatencyAvgKey:        grpcLatencyAvgByMethodQuery,
		model.GRPCSentBytesAvgKey:      grpcSentBytesAvgByMethodQuery,
		model.GRPCReceivedBytesAvgKey:  grpcReceivedBytesAvgByMethodQuery,
	}
	for name, query := range queries {
		values, err := p.queryByLabel(
			ctx,
			func(labels map[string]string) (string, bool) {
				value, ok := labels[grpcmetrics.KeyClientMethod.Name()]
				return value, ok
			},
			query,
			ts,
		)
		if err != nil {
			return nil, err
		}
		for lv, v := range values {
			if _, ok := result[lv]; !ok {
				result[lv] = make(map[string]float64)
			}
			result[lv][name] = v
		}
	}
	return result, nil
}

func (p *prometheus) collectHTTPByPath(ctx context.Context, ts time.Time) (map[string]model.ValueByLabel, error) {
	result := make(map[string]model.ValueByLabel)
	queries := map[string]string{
		model.HTTPRequestsKey:          httpRequestsByPathQuery,
		model.HTTPFailurePercentageKey: httpFailurePercentageByPathQuery,
		model.HTTPLatencyAvgKey:        httpLatencyAvgByPathQuery,
		model.HTTPSentBytesAvgKey:      httpSentBytesAvgByPathQuery,
		model.HTTPReceivedBytesAvgKey:  httpReceivedBytesAvgByPathQuery,
	}
	for name, query := range queries {
		values, err := p.queryByLabel(
			ctx,
			func(labels map[string]string) (string, bool) {
				host, ok := labels[httpmetrics.KeyClientHost.Name()]
				if !ok {
					return "", false
				}
				route, ok := labels[httpmetrics.KeyClientRoute.Name()]
				if !ok {
					return "", false
				}
				method, ok := labels[httpmetrics.KeyClientMethod.Name()]
				if !ok {
					return "", false
				}
				return fmt.Sprintf("%s/%s/%s", method, host, strings.TrimLeft(route, "/")), true
			},
			query,
			ts,
		)
		if err != nil {
			return nil, err
		}
		for lv, v := range values {
			if _, ok := result[lv]; !ok {
				result[lv] = make(map[string]float64)
			}
			result[lv][name] = v
		}
	}
	return result, nil
}

func (p *prometheus) queryOne(ctx context.Context, query string, ts time.Time) (float64, error) {
	samples, err := p.Query(ctx, query, ts)
	if err != nil {
		return 0, err
	}
	if len(samples) > 1 {
		return 0, fmt.Errorf("response must contain only one sample")
	}
	if len(samples) == 0 {
		return model.NoDataValue, nil
	}
	return samples[0].Value, nil
}

type labelsToKey func(labels map[string]string) (string, bool)

func (p *prometheus) queryByLabel(ctx context.Context, toKey labelsToKey, query string, ts time.Time) (map[string]float64, error) {
	values := make(map[string]float64)
	samples, err := p.Query(ctx, query, ts)
	if err != nil {
		return nil, err
	}
	for _, sample := range samples {
		key, ok := toKey(sample.Labels)
		if !ok {
			continue
		}
		values[key] = sample.Value
	}
	return values, nil
}

func (p *prometheus) multiQuery(ctx context.Context, queries map[string]string, ts time.Time) (map[string]float64, error) {
	values := make(map[string]float64, len(queries))
	for key, query := range queries {
		value, err := p.queryOne(ctx, query, ts)
		if err != nil {
			return nil, err
		}
		values[key] = value
	}
	return values, nil
}
