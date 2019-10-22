# Development

### Prerequisites

- [bazelisk](https://github.com/bazelbuild/bazelisk)
- [jsonnet](https://jsonnet.org/) (Only if you want to make a change for the grafana dashboards.)

### Getting started

- Building
``` console
make build
```

- Testing
``` console
make test
```

- Adding a new go dependency
``` console
### 1. Update Gopkg.toml file

### 2. Fetch the dependency and update Gopkg.lock by running
make dep

### 3. Update bazel's BUILD files by running
make gazelle
```

- Making a change on [`Lotus model`](https://github.com/lotusload/lotus/blob/master/pkg/app/lotus/apis/lotus/v1beta1/types.go)

We are using [`code-generator`](https://github.com/kubernetes/code-generator) to generate a typed client, informers, listers and deep-copy functions for `Lotus model`.
Then after making a change on the `Lotus model` you have to run the following command to update the generated codes.
``` console
make codegen
```
The following files and directories will be updated.

```
pkg/lotus/apis/lotus/v1beta1/zz_generated.deepcopy.go
pkg/lotus/client/
```

- Making a change on grafana dashboards

We are using `jsonnet` to do dashboard templating. The templates is located at [/install/dashboard-templates](https://github.com/lotusload/lotus/tree/master/install/dashboard-templates)

``` console
### Regenerate grafana dashoards
make generate-dashboards

### Regenerate kubernetes manifests with the new updates
make generate-manifests
```
