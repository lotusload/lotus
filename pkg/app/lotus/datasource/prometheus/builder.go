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
	"fmt"

	promapi "github.com/prometheus/client_golang/api"
	promv1 "github.com/prometheus/client_golang/api/prometheus/v1"

	"github.com/lotusload/lotus/pkg/app/lotus/config"
	"github.com/lotusload/lotus/pkg/app/lotus/datasource"
)

type builder struct {
}

func NewBuilder() datasource.Builder {
	return &builder{}
}

func (b *builder) Build(ds *config.DataSource, opts datasource.BuildOptions) (datasource.DataSource, error) {
	configs, ok := ds.Type.(*config.DataSource_Prometheus)
	if !ok {
		return nil, fmt.Errorf("wrong datasource type for prometheus: %T", ds.Type)
	}
	cli, err := promapi.NewClient(promapi.Config{
		Address: configs.Prometheus.Address,
	})
	if err != nil {
		return nil, err
	}
	return &prometheus{
		api:    promv1.NewAPI(cli),
		logger: opts.NamedLogger("prometheus-datasource"),
	}, nil
}
