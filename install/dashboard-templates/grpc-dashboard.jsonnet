local grafana = import 'grafonnet/grafana.libsonnet';
local panels = import 'panels.jsonnet';
local common=import 'common.jsonnet';
local dashboard = grafana.dashboard;
local template = grafana.template;

dashboard.new(
  'GRPC',
  tags=[common.tags.grpc],
  time_from='now-1h',
  schemaVersion=common.dashboard.schemaVersion,
)
.addTemplate(common.templates.test)
.addPanel(panels.workerNum, { w: 12, h: 6, x: 0, y: 0 })
.addPanel(panels.virtualUserNum, { w: 12, h: 6, x: 12, y: 0 })
.addPanel(panels.rpcsPerSecond, { w: 12, h: 8, x: 0, y: 6 })
.addPanel(panels.rpcLatency, { w: 12, h: 8, x: 12, y: 6 })
.addPanel(panels.rpcsPerSecondByStatus, { w: 12, h: 8, x: 0, y: 14 })
.addPanel(panels.percentageFailedRPCs, { w: 12, h: 8, x: 12, y: 14 })
.addPanel(panels.rpcSentBytes, { w: 12, h: 8, x: 0, y: 22 })
.addPanel(panels.rpcReceivedBytes, { w: 12, h: 8, x: 12, y: 22 })