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

package registry

import (
	"fmt"

	"github.com/lotusload/lotus/pkg/app/lotus/config"
	"github.com/lotusload/lotus/pkg/app/lotus/datasource"
	"github.com/lotusload/lotus/pkg/app/lotus/datasource/prometheus"
)

var defaultRegistry = New()

func init() {
	defaultRegistry.Register(config.DataSource_PROMETHEUS, prometheus.NewBuilder())
}

func Default() *registry {
	return defaultRegistry
}

type registry struct {
	builders map[config.DataSource_Type]datasource.Builder
}

func New() *registry {
	return &registry{
		builders: make(map[config.DataSource_Type]datasource.Builder),
	}
}

func (r *registry) Register(dst config.DataSource_Type, b datasource.Builder) {
	if r.builders[dst] != nil {
		panic(fmt.Sprintf("duplicate builder registered: %v", dst))
	}
	r.builders[dst] = b
}

func (r *registry) Get(dst config.DataSource_Type) (datasource.Builder, error) {
	if b, ok := r.builders[dst]; ok {
		return b, nil
	}
	return nil, fmt.Errorf("unknown builder: %v", dst)
}
