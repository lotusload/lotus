# Examples

### Prerequisites

Before running one of these examples, you have to deploy [`helloworld`](https://github.com/nghialv/lotus/tree/master/examples/helloworld) target service to your cluster by running the following command.
``` console
kubectl apply -f /helloworld
```
The implementation of `helloworld` server is here [helloworld.go](https://github.com/nghialv/lotus/blob/master/pkg/app/example/cmd/helloworld/helloworld.go)

When running this service will start one grpc server for handling rpcs defined in [helloworld.proto](https://github.com/nghialv/lotus/blob/master/pkg/app/example/helloworld/helloworld.proto) and one HTTP server for handling incoming HTTP requests.

### simple-http-scenario

- Scenario: [`pkg/app/example/cmd/simplehttp/scenario.go`](https://github.com/nghialv/lotus/blob/master/pkg/app/example/cmd/simplehttp/scenario.go)
- Lotus CRD: [`simple-http-scenario.yaml`](https://github.com/nghialv/lotus/blob/master/examples/simple-http-scenario.yaml)

A simple scenario that send one http request to `http://httpbin.org/`.

### simple-grpc-scenario

- Scenario: [`/pkg/app/example/cmd/simplegrpc/scenario.go`](https://github.com/nghialv/lotus/blob/master/pkg/app/example/cmd/simplegrpc/scenario.go)
- Lotus CRD: [`simple-grpc-scenario.yaml`](https://github.com/nghialv/lotus/blob/master/examples/simple-grpc-scenario.yaml)

A simple scenario that send one grpc request to `helloworld` service.

### three-steps-scenario

- Scenario: [`/pkg/app/example/cmd/threesteps/scenario.go`](https://github.com/nghialv/lotus/blob/master/pkg/app/example/cmd/threesteps/scenario.go)
- Lotus CRD: [`three-steps-scenario.yaml`](https://github.com/nghialv/lotus/blob/master/examples/three-steps-scenario.yaml)

An example containing full 3 steps of Lotus: `preparer`, `worker` and `cleaner`.

### virtual-user-scenario

- Scenario: [`/pkg/app/example/cmd/virtualuser/scenario.go`](https://github.com/nghialv/lotus/blob/master/pkg/app/example/cmd/virtualuser/scenario.go)
- Lotus CRD: [`virtualuser-scenario.yaml`](https://github.com/nghialv/lotus/blob/master/examples/virtualuser-scenario.yaml)

An example using [`virtualuser`](https://github.com/nghialv/lotus/tree/master/pkg/virtualuser) package to spawn a given number of virtual users on each worker.
