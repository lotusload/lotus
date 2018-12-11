package registry

import (
	"fmt"

	"github.com/nghialv/lotus/pkg/app/lotus/config"
	"github.com/nghialv/lotus/pkg/app/lotus/datasource"
	"github.com/nghialv/lotus/pkg/app/lotus/datasource/prometheus"
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
