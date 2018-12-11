#!/usr/bin/env bash

set -o errexit
set -o nounset
set -o pipefail

OPTION=${1:-""}
SUFFIX=""
if [ ! -z "$OPTION" ]; then
  SUFFIX="-$OPTION"
fi

ROOT=$(dirname ${BASH_SOURCE})/..
MANIFESTS_DIR="${ROOT}/install/manifests${SUFFIX}"
VALUE_FILE="${ROOT}/install/manifest-generate-values${SUFFIX}.yaml"
HELM_CHART_DIR="${ROOT}/install/helm"

echo "Generating manifests to tmp..."
helm template --name lotus -f $VALUE_FILE $HELM_CHART_DIR --output-dir /tmp

echo "Deleting all old manifests..."
mkdir -p ${MANIFESTS_DIR}
rm -rf ${MANIFESTS_DIR}/*

echo "Copying generated manifests to manifests folder..."
cp /tmp/lotus/templates/* ${MANIFESTS_DIR}
for f in $(find /tmp/lotus/charts/grafana/templates -type f); do
  cp $f ${MANIFESTS_DIR}/grafana-${f##*/};
done

echo "Deleting tmp data..."
rm -rf /tmp/lotus

echo "Done"

