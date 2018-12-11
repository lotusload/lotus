local grafana = import 'grafonnet/grafana.libsonnet';
local template = grafana.template;
local graphPanel = grafana.graphPanel;
local text = grafana.text;

{
  datasources:: {
      default:: 'thanos',
  },
  dashboard:: {
    schemaVersion:: 16,
  },
  tags:: {
    grpc:: 'grpc',
    http:: 'http',
  },
  format:: {
    short:: 'short',
    second:: 's',
    millisecond:: 'ms',
    bytes:: 'bytes',
    bytesPerSecond:: 'Bps',
    percent_0_100:: 'percent',
    percent_0_1:: 'percentunit',
  },
  templates:: {
    test:: template.new(
      name='testId',
      label='TestID',
      datasource= $.datasources.default,
      query='query_result(count by(job) (count_over_time(up[$__range])))',
      regex='/"(.*)-worker"/',
      refresh='time',
    ),
    hiddenCustom(
      name,
      value,
    ):: template.custom(
      name=name,
      query=value,
      current=value,
      hide='value',
    ),
  },
  panel:: {
    new(
      title,
      format= $.format.short,
      datasource= $.datasources.default,
    ):: graphPanel.new(
        title=title,
        datasource=datasource,
        format=format,
        fill=2,
        linewidth=2,
        legend_alignAsTable=true,
        legend_values=true,
        legend_max=true,
        legend_min=true,
        legend_avg=true,
        legend_current=true,
        legend_sort="current",
        legend_sortDesc=true,
      ),
    transparentText(
      title=''
    ):: text.new(
      title=title,
      transparent=true,
    )
  },
}
