# Installation

We are supporting 2 ways to install Lotus:
- using helm chart
- using kubernetes manifests directly

**Using the helm chart is recommended.**

### Using Helm chart

The Lotus chart is put in `./helm` directory.

```
helm install --name lotus ./helm

### If you want to override the values

helm install --name lotus -f ./path/to/your/values.yaml ./helm
```

Please check out [`values.yaml`](https://github.com/lotusload/lotus/blob/master/install/helm/values.yaml) for configurable fields.
Note: Please change [`grafana.adminPassword`](https://github.com/lotusload/lotus/tree/master/install/helm/values.yaml#L27) value. The current password is `admin`.

### Using kubernetes manifests

All kubernetes manifests for Lotus are put in `./manifests` (RBAC is enabled) and `./manifests-norbac` (RBAC is disabled) directories. We generated those manifests from the helm chart above.

```
kubectl apply -f ./manifests

### Or for disabling RBAC

kubectl apply -f ./manifests-norbac
```

Note: Please change [`grafana adminPassword`](https://github.com/lotusload/lotus/tree/master/install/manifests/grafana-secret.yaml#L15) value to a `base64` encoded value. The current password is `admin`.