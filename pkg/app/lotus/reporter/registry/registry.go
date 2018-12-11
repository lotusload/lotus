package registry

import (
	"fmt"

	"github.com/nghialv/lotus/pkg/app/lotus/config"
	"github.com/nghialv/lotus/pkg/app/lotus/reporter"
	"github.com/nghialv/lotus/pkg/app/lotus/reporter/gcs"
	"github.com/nghialv/lotus/pkg/app/lotus/reporter/logger"
	"github.com/nghialv/lotus/pkg/app/lotus/reporter/slack"
)

var defaultRegistry = New()

func init() {
	defaultRegistry.Register(config.Receiver_LOGGER, logger.NewBuilder())
	defaultRegistry.Register(config.Receiver_GCS, gcs.NewBuilder())
	defaultRegistry.Register(config.Receiver_SLACK, slack.NewBuilder())
}

func Default() *registry {
	return defaultRegistry
}

type registry struct {
	builders map[config.Receiver_Type]reporter.Builder
}

func New() *registry {
	return &registry{
		builders: make(map[config.Receiver_Type]reporter.Builder),
	}
}

func (r *registry) Register(rt config.Receiver_Type, b reporter.Builder) {
	if r.builders[rt] != nil {
		panic(fmt.Sprintf("duplicate builder registered: %v", rt))
	}
	r.builders[rt] = b
}

func (r *registry) Get(rt config.Receiver_Type) (reporter.Builder, error) {
	if b, ok := r.builders[rt]; ok {
		return b, nil
	}
	return nil, fmt.Errorf("unknown builder: %v", rt)
}
