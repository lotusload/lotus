# Lotus CRD Configurations

The following is an example of the full configurations.

``` yaml
apiVersion: lotus.lotusload.com/v1beta1
kind: Lotus
metadata:
  name: scenario-12345
spec:
  ttlSecondsAfterFinished: 300
  checkIntervalSeconds: 10
  checkInitialDelaySeconds: 15
  worker:
    runTime: 30m
    replicas: 20
    metricsPort: 8081
    containers:
      - name: worker
        image: lotusload/lotus-example:v0.1.5
        args:
          - three-steps-scenario
          - --step=worker
          - --helloworld-grpc-address=helloworld:8080
          - --helloworld-http-address=http://helloworld:9090
        ports:
          - name: metrics
            containerPort: 8081
        volumeMounts:
          - name: data
            mountPath: /etc/data
    volumes:
      - name: data
        configMap:
          name: worker-data
  preparer:
    containers:
      - name: preparer
        image: lotusload/lotus-example:v0.1.5
        args:
          - three-steps-scenario
          - --step=preparer
  cleaner:
    containers:
      - name: cleaner
        image: lotusload/lotus-example:v0.1.5
        args:
          - three-steps-scenario
          - --step=cleaner
  checks:
    - name: GRPCHighLatency
      expr: lotus_grpc_client_roundtrip_latency:method > 250
      for: 30s
    - name: VirtualUserHighFailurePercentage
      expr: lotus_virtual_user_failure_percentage > 10
      for: 10s
```