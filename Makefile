.PHONY: build
build:
	bazelisk build -- //cmd/... //pkg/...

.PHONY: test
test:
	bazelisk test -- //cmd/... //pkg/...

.PHONY: push
push:
	./hack/push.sh

.PHONY: dep
dep:
	GO111MODULE=on go mod tidy
	GO111MODULE=on go mod vendor
	bazelisk run //:gazelle -- update-repos -from_file=go.mod -to_macro=repositories.bzl%go_repositories

.PHONY: gazelle
gazelle:
	bazelisk run //:gazelle

.PHONY: codegen
codegen:
	./hack/update-codegen.sh

.PHONY: libsonnet
libsonnet:
	jb update --jsonnetpkg-home=libsonnet

.PHONY: install
install:
	helm install --name lotus -f ./install/values.yaml ./install/helm

.PHONY: upgrade
upgrade:
	helm upgrade lotus -f ./install/values.yaml ./install/helm

.PHONY: generate-manifests
generate-manifests:
	./hack/generate-manifests.sh
	./hack/generate-manifests.sh norbac

.PHONY: generate-dashboards
generate-dashboards:
	./hack/generate-dashboards.sh
