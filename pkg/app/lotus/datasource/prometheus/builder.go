package prometheus

import (
	"fmt"

	promapi "github.com/prometheus/client_golang/api"
	promv1 "github.com/prometheus/client_golang/api/prometheus/v1"

	"github.com/nghialv/lotus/pkg/app/lotus/config"
	"github.com/nghialv/lotus/pkg/app/lotus/datasource"
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
