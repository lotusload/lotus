# Configurations

This is an example of a full configuration file.

``` yaml
lotus:
  configs:
    checks:                                             // 1. The list of global checks. Those checks will be applied to all tests.
      - name: NoWorker
        expr: absent(up)
        for: 30s
      - name: HasWorkerDown
        expr: up == 0
        for: 30s
      - name: GRPCHighFailurePercentage
        expr: lotus_grpc_client_completed_rpcs_failure_percentage:method > 5
        for: 30s
      - name: HTTPHighFailurePercentage
        expr: lotus_http_client_completed_requests_5xx_percentage:host:route:method > 5
        for: 30s
      - name: VirtualUserHighFailurePercentage
        expr: lotus_virtual_user_failure_percentage > 2
        for: 10s
    receivers:                                          // 2. The list of all receivers to send the summary result.
      - name: gcs
        gcs:
          bucket: lotus-result-bucket
          credentials:
            secret: gcs-credentials
            file: gcs-credentials.json
      - name: lotus-slack-channel
        slack:
          hookUrl: https://hooks.slack.com/services/YOUR-HOOK
      - name: logger
        logger:
    timeSeriesStorage:                                  // 3. A long-term storage for storing time series data.
      gcs:
        bucket: lotus-timeseries-bucket
        credentials:
          secret: gcs-credentials
          file: gcs-credentials.json
    grafanaBaseUrl: http://your-grafana-domain:3000
```

### 1. Global checks setup

You can define the checks on your Lotus CRD for each test, but you can also define some global checks on the configuration file.
I recommend to add these 2 checks to your global checks: the first one is for checking if no worker has started, the second one is for checking if has any worker down.

```
lotus:
  configs:
    checks:
      - name: NoWorker
        expr: absent(up)
        for: 30s
      - name: HasWorkerDown
        expr: up == 0
        for: 30s
```

### 2. Receivers setup

Currently we are supporting 3 types of receiver: GCS, Slack, Logger.

#### GCS

To configure Google Cloud Storage as a receiver, you need to set receiver with GCS bucket name and k8s secret that contains the Google Application credentials.

```
receivers:
  - name: gcs
    gcs:
      bucket: lotus-result-bucket       // The bucket name created on Google Cloud Storage
      credentials:
        secret: gcs-credentials         // The name of k8s secret that contains Google Application credentials 
        file: gcs-credentials.json      // The credentials file name inside the secret
```


### 3. Long term storage setup

To able to access the time series data after your test is deleted you have to configure to store those time series data to a long-term storage like GCS, S3, Azure...
You can do that by adding configuration for `lotus.configs.timeSEriesStorage` field.
