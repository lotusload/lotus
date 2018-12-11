local grafana = import 'grafonnet/grafana.libsonnet';
local panels = import 'panels.jsonnet';
local common=import 'common.jsonnet';
local dashboard = grafana.dashboard;
local template = grafana.template;

dashboard.new(
  'HTTP',
  tags=[common.tags.http],
  time_from='now-1h',
  schemaVersion=common.dashboard.schemaVersion,
)
.addTemplate(common.templates.test)
.addPanel(panels.workerNum, { w: 12, h: 6, x: 0, y: 0 })
.addPanel(panels.virtualUserNum, { w: 12, h: 6, x: 12, y: 0 })
.addPanel(panels.httpRequestsPerSecond, { w: 12, h: 8, x: 0, y: 6 })
.addPanel(panels.percentageOf5xxRequests, { w: 12, h: 8, x: 12, y: 6 })
.addPanel(panels.httpRequestLatency, { w: 12, h: 8, x: 0, y: 14 })
.addPanel(panels.httpRequestSentBytes, { w: 12, h: 8, x: 0, y: 22 })
.addPanel(panels.httpRequestReceivedBytes, { w: 12, h: 8, x: 12, y: 22 })