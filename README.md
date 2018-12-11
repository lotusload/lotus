# Lotus [![PRs Welcome](https://img.shields.io/badge/PRs-welcome-brightgreen.svg?style=flat)](http://makeapullrequest.com) [![MIT Licensed](https://img.shields.io/badge/license-MIT-blue.svg)](https://github.com/nghialv/lotus/blob/master/LICENSE)

Lotus is a Kubernetes controller for running load testing. Lotus schedules & monitors the load test workers, collects & stores the metrics and notifies the test result.

Once installed, Lotus provides the following features:
- GRPC and HTTP support
- Ability to write the scenario by any language you want
- Automation-friendly
  - `Checks` (like asserts, fails in normal test) for easy and flexible CI configuration
  - Test is configured by using declarative Kubernetes CRD for version control friendliness
- Flexible metrics storage and visualization
  - Ability to view visualized time series by Grafana
  - Ability to persist time series data to a long-term storage like GCS or S3
  - Ability to notify and store the result summary to multiple receivers: GCS, Slack, Logger...

> _I am thinking about adding a feature that helps us determine the maximum number of  users (requests) the target services can handle. This can be done by automatically running the load tests with the number of virtual users increasing gradually until one of the checks fails. Or a feature that helps us determine the needed resources of the target services so that they can handle the given number of users. [more](https://github.com/nghialv/lotus/issues/1)_

### Installation
Firstly, you need to install Lotus controller on your Kubernetes cluster to start using.
Lotus requires a Kubernetes cluster of version `>=1.9.0`.

The Lotus controller can be installed either by using the helm [`chart`](https://github.com/nghialv/lotus/tree/master/install/helm) or by using Kubernetes [`manifests`](https://github.com/nghialv/lotus/tree/master/install/manifests) directly.
(Using the helm chart is recommended.)

``` console
helm install --name lotus ./install/helm
```

See [`install`](https://github.com/nghialv/lotus/tree/master/install) for more details.

### Running Lotus
We have 2 steps to start running a load test:
- Writing a load test scenario
- Writing a Lotus CRD configuration

#### 1. Writing a load test scenario

Theoretically, you can write your scenarios by using any language you like. The only thing you need to have is a metrics exporter for Prometheus.

In the case of Golang, I have already prepared some util packages (e.g. [`metrics`](https://github.com/nghialv/lotus/tree/master/pkg/metrics), [`virtualuser`](https://github.com/nghialv/lotus/tree/master/pkg/virtualuser)) that help you write your scenarios faster and easier.

- Expose a metrics server in your scenario's `main.go`
``` go
import "github.com/nghialv/lotus/pkg/metrics"

m, err := metrics.NewServer(8081)
if err != nil {
    return err
}
defer m.Stop()
go m.Run()
```
- In case you want to send gRPC's rpcs to your load server, let's set `grpcmetrics.ClientHandler` as the `StatsHandler` of your gRPC connection.
``` go
grpc.Dial(
    grpc.WithStatsHandler(&grpcmetrics.ClientHandler{}),
)
```

- In case you want to send HTTP requests to your load server, let's use the `Transport` from `httpmetrics` package.
``` go
http.Client{
    Transport: &httpmetrics.Transport{},
}
```
- That is all. Now let's build your scenario image and publish to your container registry.

#### 2. Writing a Lotus CRD configuration

``` yaml
apiVersion: lotus.nghialv.com/v1beta1
kind: Lotus
metadata:
  name: simple-scenario-12345                                  // The unique testID
spec:
  worker:
    runTime: 10m                                               // How long the load test will be run
    replicas: 15                                               // How many workers should be created
    metricsPort: 8081                                          // What port number should be used to collect metrics
    containers:
      - name: worker
        image: your-registry/your-worker-image                 // The scenario image you published above
        ports:
          - name: metrics
            containerPort: 8081
  checks:                                                      // You can add some checks to be checked while running
    - name: GRPCHighErrorRate
      expr: lotus_grpc_client_failure_percentage > 10
      for: 30s
```

Then apply this file to your Kubernetes cluster. Lotus will handle this test for you.

See [`crd-configurations.md`](https://github.com/nghialv/lotus/blob/master/docs/lotus-crd-configurations.md) for all configurable fields.

See [`examples`](https://github.com/nghialv/lotus/tree/master/examples) for more examples.

### Outputs

- Test summary

Lotus collects the metrics data and evaluates the `checks` to build a summary result for each test.
Lotus can be configured to upload this summary file to external services (e.g: GCS, Slack...) or to log into `stdout`.
3 formats of the summary file are supported: `Text`, `Markdown`, `JSON`.

``` yaml
TestID:        test-scenario-12345
TestStatus:    Succeeded
Start:         09:02:59 2018-12-03
End:           09:12:59 2018-12-03

MetricsSummary:

1. Virtual User
  - Started:             1M
  - Failed:              0

2. GRPC
  - RPCTotal:            25M
  - FailurePercentage:   2.507

GroupByMethod:
                        RPCs      Failure%  Latency   SentBytes  RecvBytes

  - helloworld.Hello    12.5M     1.015     105       15        8
  - helloworld.Profile  12.5M     1.415     152       8         256
  - all                 25M       1.207     135       12        245

Grafana: http://localhost:3000/dashboard/db/grpc?from=1543827779598&to=1543828379598
```


- Grafana dashboards

To be able to fully explore and understand your test, Lotus is providing some Grafana dashboards to view the visualizations of the metrics.
You can also set up Lotus to persist the time series data to a long-term storage (GCS or S3) for accessing after the test is deleted.

- Test Status

After applying the Lotus CRD to your Kubernetes cluster you can also use the following command to check the status of your test.

``` console
kubectl describe Lotus your-lotus-name
```

Your test can be one of these status: `Pending`, `Preparing`, `Running`, `Cleaning`, `FailureCleaning`, `Failed`, `Succeeded`

### Examples

Please checkout [`/examples`](https://github.com/nghialv/lotus/tree/master/examples) directory that contains some prepared examples.

### FQA

Refer to [FQA.md](https://github.com/nghialv/lotus/blob/master/docs/fqa.md)

### Development

Refer to [development.md](https://github.com/nghialv/lotus/blob/master/docs/development.md)

### LICENSE
Lotus is released under the MIT license. See [LICENSE](https://github.com/nghialv/lotus/blob/master/LICENSE) file for the details.
