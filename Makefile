.PHONY: build
build:
	bazel build -k -- //cmd/... //pkg/...

.PHONY: test
test:
	bazel test -- //cmd/... //pkg/...

.PHONY: push
push:
	bazel run --platforms=@io_bazel_rules_go//go/toolchain:linux_amd64 //cmd:push_all_images

.PHONY: dep
dep:
	dep ensure
	bazel run //:gazelle -- update-repos -from_file=Gopkg.lock

.PHONY: gazelle
gazelle:
	bazel run //:gazelle

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
