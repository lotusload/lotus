#!/usr/bin/env bash

set -o errexit
set -o nounset
set -o pipefail

ROOT=$(dirname ${BASH_SOURCE})/..
LIBSONNET_DIR="$ROOT/libsonnet"
TEMPLATES_DIR="$ROOT/install/dashboard-templates"
DASHBOARDS_DIR="$ROOT/install/helm/dashboards"

mkdir -p $DASHBOARDS_DIR
rm -rf $DASHBOARDS_DIR/*

for f in $(find $TEMPLATES_DIR -name "*-dashboard.jsonnet"); do
  fn=${f##*/}
  echo "Rendering $fn..."
  jsonnet -J $LIBSONNET_DIR $f > $DASHBOARDS_DIR/${fn%.*}.json
done

