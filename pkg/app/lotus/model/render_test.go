package model

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRender(t *testing.T) {
	metricsSummary := &MetricsSummary{
		GRPCRPCTotal:          25000000,
		GRPCFailurePercentage: 2.507,
		GRPCAll: map[string]float64{
			GRPCRPCsKey:              25000000,
			GRPCFailurePercentageKey: 1.207,
			GRPCLatencyAvgKey:        135,
			GRPCSentBytesAvgKey:      12,
			GRPCReceivedBytesAvgKey:  245,
		},
		GRPCByMethod: map[string]ValueByLabel{
			"helloworld.Hello": map[string]float64{
				GRPCRPCsKey:              12500000,
				GRPCFailurePercentageKey: 1.015,
				GRPCLatencyAvgKey:        105,
				GRPCSentBytesAvgKey:      15,
				GRPCReceivedBytesAvgKey:  8,
			},
			"helloworld.Profile": map[string]float64{
				GRPCRPCsKey:              12500000,
				GRPCFailurePercentageKey: 1.415,
				GRPCLatencyAvgKey:        152,
				GRPCSentBytesAvgKey:      8,
				GRPCReceivedBytesAvgKey:  256,
			},
		},
		HTTPRequestTotal:      10,
		HTTPFailurePercentage: 1.05890,
		HTTPAll: map[string]float64{
			HTTPRequestsKey:          10,
			HTTPFailurePercentageKey: 1.05890,
		},
		VirtualUserStartedTotal: 1000000,
		VirtualUserFailedTotal:  0,
	}
	testcases := []struct {
		Result *Result
	}{
		{
			Result: &Result{
				TestID:            "test-scenario-12345",
				Status:            TestSucceeded,
				MetricsSummary:    metricsSummary,
				StartedTimestamp:  time.Now().Add(-10 * time.Minute),
				FinishedTimestamp: time.Now(),
			},
		},
	}
	for _, tc := range testcases {
		tc.Result.SetGrafanaDashboardURLs("http://localhost:3000")
		out, err := tc.Result.Render(RenderFormatText)
		require.NoError(t, err)
		fmt.Println(string(out))
	}
}

func TestFormatValue(t *testing.T) {
	testcases := []struct {
		Value    float64
		Expected string
	}{
		{
			Value:    -1,
			Expected: noDataMark,
		},
		{
			Value:    0.0,
			Expected: "0",
		},
		{
			Value:    2.15,
			Expected: "2.15",
		},
		{
			Value:    12345678,
			Expected: "12.35M",
		},
		{
			Value:    0.123456,
			Expected: "123.5m",
		},
	}
	for _, tc := range testcases {
		out := formatValue(tc.Value)
		assert.Equal(t, tc.Expected, out)
	}
}
