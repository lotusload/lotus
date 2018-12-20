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

package monitor

import (
	"context"
	"fmt"
	"time"

	"github.com/spf13/cobra"
	"go.uber.org/zap"

	"github.com/lotusload/lotus/pkg/app/lotus/config"
	"github.com/lotusload/lotus/pkg/app/lotus/datasource"
	dsregistry "github.com/lotusload/lotus/pkg/app/lotus/datasource/registry"
	"github.com/lotusload/lotus/pkg/app/lotus/model"
	"github.com/lotusload/lotus/pkg/app/lotus/reporter"
	reporterregistry "github.com/lotusload/lotus/pkg/app/lotus/reporter/registry"
	"github.com/lotusload/lotus/pkg/cli"
)

type monitor struct {
	testID                   string
	runTime                  time.Duration
	checkInterval            time.Duration
	checkInitialDelay        time.Duration
	collectSummaryDataSource string
	collectAndReportTimeout  time.Duration
	configFile               string

	dataSourceMap map[string]datasource.DataSource
	checkMap      map[string][]datasource.Check
	cfg           *config.Config
	logger        *zap.Logger
}

func NewCommand() *cobra.Command {
	m := &monitor{
		runTime:                 2 * time.Minute,
		checkInterval:           30 * time.Second,
		checkInitialDelay:       10 * time.Second,
		collectAndReportTimeout: 30 * time.Minute,
	}
	cmd := &cobra.Command{
		Use:   "monitor",
		Short: "Start running Lotus monitor",
		RunE:  cli.WithContext(m.run),
	}
	cmd.Flags().StringVar(&m.testID, "test-id", m.testID, "The unique test id")
	cmd.MarkFlagRequired("test-id")
	cmd.Flags().DurationVar(&m.runTime, "run-time", m.runTime, "How long the worker should be run")
	cmd.Flags().DurationVar(&m.checkInterval, "check-interval", m.checkInterval, "How often does the monitor run the check")
	cmd.Flags().DurationVar(&m.checkInitialDelay, "check-initial-delay", m.checkInitialDelay, "How long the monitor should wait before performing the first check")
	cmd.Flags().StringVar(&m.collectSummaryDataSource, "collect-summary-datasource", m.collectSummaryDataSource, "The datasource used to collect test summary")
	cmd.MarkFlagRequired("collect-summary-datasource")
	cmd.Flags().DurationVar(&m.collectAndReportTimeout, "collect-and-report-timeout", m.collectAndReportTimeout, "How log to wait for collect and report tasks")
	cmd.Flags().StringVar(&m.configFile, "config-file", m.configFile, "Path to the configuration file")
	cmd.MarkFlagRequired("config-file")
	return cmd
}

func (m *monitor) run(ctx context.Context, logger *zap.Logger) (lastErr error) {
	startTime := time.Now()
	m.logger = logger.Named("monitor")
	ctx, cancel := context.WithTimeout(ctx, m.runTime)
	defer cancel()

	defer func() {
		if err := m.collectAndReport(startTime, time.Now(), lastErr); err != nil {
			lastErr = err
		}
	}()

	cfg, err := config.FromFile(m.configFile)
	if err != nil {
		logger.Error("failed to load configuration", zap.Error(err))
		lastErr = err
		return
	}
	m.cfg = cfg
	dataSourceMap, err := buildDataSourceMap(cfg, logger)
	if err != nil {
		logger.Error("failed to build dataSourceMap", zap.Error(err))
		lastErr = err
		return
	}
	m.dataSourceMap = dataSourceMap
	m.checkMap = buildCheckMap(cfg)

	// Waiting for initial delay
	select {
	case <-time.After(m.checkInitialDelay):
	case <-ctx.Done():
	}

	tick := time.Tick(m.checkInterval)
	for {
		select {
		case <-tick:
			lastErr = m.check(ctx)
			if lastErr != nil {
				return
			}
		case <-ctx.Done():
			m.logger.Info("breaking the check loop due to the context deadline")
			return
		}
	}
}

func (m *monitor) check(ctx context.Context) error {
	actives := make([]string, 0)
	m.logger.Info("start checking all datasources", zap.Int("num", len(m.dataSourceMap)))
	for dsn, checks := range m.checkMap {
		ds, ok := m.dataSourceMap[dsn]
		if !ok {
			err := fmt.Errorf("missing datasource: %s", dsn)
			m.logger.Error("failed to get datasource", zap.Error(err))
			return err
		}
		result, err := ds.Check(ctx, checks)
		if err != nil {
			m.logger.Error("failed to check", zap.Error(err))
			return err
		}
		actives = append(actives, result.Actives...)
	}
	if len(actives) == 0 {
		return nil
	}
	m.logger.Info("active checks", zap.Any("actives", actives))
	return checkError{
		Actives: actives,
	}
}

type checkError struct {
	Actives []string
}

func (ce checkError) Error() string {
	return fmt.Sprintf("%d checks are failed", len(ce.Actives))
}

func (m *monitor) collectAndReport(startTime, finishTime time.Time, lastErr error) error {
	ctx, cancel := context.WithTimeout(context.Background(), m.collectAndReportTimeout)
	defer cancel()
	result := &model.Result{
		TestID:            m.testID,
		Status:            model.TestSucceeded,
		StartedTimestamp:  startTime,
		FinishedTimestamp: finishTime,
	}
	if lastErr != nil {
		result.SetFailed(lastErr.Error())
	}
	if ce, ok := lastErr.(checkError); ok {
		result.FailedChecks = ce.Actives
	}

	summary, collectErr := m.collect(ctx)
	if collectErr != nil {
		m.logger.Error("failed to collect metrics summary", zap.Error(collectErr))
		if result.Status != model.TestFailed {
			result.SetFailed("failed to collect metrics summary")
		}
	} else {
		result.MetricsSummary = summary
	}
	if m.cfg != nil {
		result.SetGrafanaDashboardURLs(m.cfg.GrafanaBaseUrl)
	}
	if err := m.report(ctx, result); err != nil {
		m.logger.Error("failed to report result", zap.Error(err))
		return err
	}
	return collectErr
}

func (m *monitor) collect(ctx context.Context) (*model.MetricsSummary, error) {
	ds, ok := m.dataSourceMap[m.collectSummaryDataSource]
	if !ok {
		err := fmt.Errorf("missing datasource for collecting test summary: %s", m.collectSummaryDataSource)
		m.logger.Error("failed to get datasource", zap.Error(err))
		return nil, err
	}
	return ds.CollectSummary(ctx, time.Now())
}

func (m *monitor) report(ctx context.Context, result *model.Result) error {
	rs := make([]reporter.Reporter, 0, len(m.cfg.Receivers))
	for _, recv := range m.cfg.Receivers {
		builder, err := reporterregistry.Default().Get(recv.ReceiverType())
		if err != nil {
			return err
		}
		r, err := builder.Build(recv, reporter.BuildOptions{
			Logger: m.logger,
		})
		if err != nil {
			return err
		}
		rs = append(rs, r)
	}
	return reporter.MultiReporter(rs...).Report(ctx, result)
}

func buildDataSourceMap(cfg *config.Config, logger *zap.Logger) (map[string]datasource.DataSource, error) {
	datasources := make(map[string]datasource.DataSource, len(cfg.DataSources))
	for _, ds := range cfg.DataSources {
		builder, err := dsregistry.Default().Get(ds.DataSourceType())
		if err != nil {
			return nil, err
		}
		datasource, err := builder.Build(ds, datasource.BuildOptions{
			Logger: logger,
		})
		if err != nil {
			return nil, err
		}
		datasources[ds.Name] = datasource
	}
	return datasources, nil
}

func buildCheckMap(cfg *config.Config) map[string][]datasource.Check {
	checkMap := make(map[string][]datasource.Check)
	for _, check := range cfg.Checks {
		c := datasource.Check{
			Name: check.Name,
			Expr: check.Expr,
			For:  check.For,
		}
		if list, ok := checkMap[check.DataSource]; ok {
			checkMap[check.DataSource] = append(list, c)
			continue
		}
		checkMap[check.DataSource] = []datasource.Check{c}
	}
	return checkMap
}
