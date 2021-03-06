module github.com/lotusload/lotus

go 1.12

require (
	cloud.google.com/go v0.38.0
	contrib.go.opencensus.io/exporter/prometheus v0.1.0
	github.com/ghodss/yaml v0.0.0-20150909031657-73d445a93680
	github.com/golang/protobuf v1.3.2
	github.com/prometheus/client_golang v1.2.1
	github.com/prometheus/common v0.7.0
	github.com/spf13/cobra v0.0.5
	github.com/stretchr/testify v1.3.0
	go.opencensus.io v0.21.0
	go.uber.org/atomic v1.4.0 // indirect
	go.uber.org/multierr v1.2.0 // indirect
	go.uber.org/zap v1.11.0
	golang.org/x/net v0.0.0-20190812203447-cdfb69ac37fc
	golang.org/x/sync v0.0.0-20190423024810-112230192c58
	google.golang.org/api v0.4.0
	google.golang.org/grpc v1.19.0
	k8s.io/api v0.0.0-20191016225839-816a9b7df678
	k8s.io/apimachinery v0.0.0-20191016225534-b1267f8c42b4
	k8s.io/client-go v0.0.0-20191016230210-14c42cd304d9
	k8s.io/code-generator v0.0.0-00010101000000-000000000000
)

replace (
	k8s.io/api => k8s.io/api v0.0.0-20191016225839-816a9b7df678
	k8s.io/apimachinery => k8s.io/apimachinery v0.0.0-20191016225534-b1267f8c42b4
	k8s.io/client-go => k8s.io/client-go v0.0.0-20191016230210-14c42cd304d9
	k8s.io/code-generator => k8s.io/code-generator v0.0.0-20191003035328-700b1226c0bd
)
