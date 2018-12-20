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

package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFromFile(t *testing.T) {
	cfg, err := FromFile("testdata/valid.yaml")
	require.NoError(t, err)
	require.NotNil(t, cfg)

	require.NotNil(t, cfg.TimeSeriesStorage)
	gcs, ok := cfg.TimeSeriesStorage.Type.(*TimeSeriesStorage_Gcs)
	require.True(t, ok)
	assert.Equal(t, "gcs-bucket", gcs.Gcs.Bucket)
	assert.NotNil(t, gcs.Gcs.Credentials)
	assert.Equal(t, 1, len(cfg.DataSources))
	assert.Equal(t, 1, len(cfg.Checks))
	assert.Equal(t, 3, len(cfg.Receivers))
}

func TestMarshaling(t *testing.T) {
	configs := []*Config{
		&Config{},
		&Config{
			DataSources: []*DataSource{
				&DataSource{
					Name: "prometheus",
					Type: &DataSource_Prometheus{
						Prometheus: &PrometheusConfigs{
							Address: "https://127.0.0.1:9090",
						},
					},
				},
			},
			Checks: []*Check{
				&Check{
					Name: "HighErrorRate",
					Expr: "error_rate > 0.5",
					For:  "1m",
				},
				&Check{
					Name:       "HighLatency",
					Expr:       "latency > 125",
					For:        "1m",
					DataSource: "prometheus",
				},
			},
			Receivers: []*Receiver{
				&Receiver{
					Name: "gcs",
					Type: &Receiver_Gcs{
						Gcs: &GCSReceiverConfigs{
							Bucket: "bucket-2",
							Credentials: &SecretFileSelector{
								Secret: "foo",
								File:   "credentials-2",
							},
						},
					},
				},
				&Receiver{
					Name: "slack",
					Type: &Receiver_Slack{
						Slack: &SlackReceiverConfigs{
							HookUrl: "http://api-2.slack.com",
						},
					},
				},
			},
			TimeSeriesStorage: &TimeSeriesStorage{
				Type: &TimeSeriesStorage_Gcs{
					Gcs: &GCSTimeSeriesStorageConfigs{
						Bucket: "gcs-bucket",
						Credentials: &SecretFileSelector{
							Secret: "secret-name",
							File:   "filename",
						},
					},
				},
			},
		},
	}
	for _, cfg := range configs {
		require.NoError(t, cfg.Validate())

		data, err := cfg.MarshalToYaml()
		require.NoError(t, err)

		unmarshaledCfg, err := UnmarshalFromYaml(data)
		require.NoError(t, err)
		assert.Equal(t, cfg, unmarshaledCfg)
	}
}
