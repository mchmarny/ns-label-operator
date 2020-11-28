# Introduction

This chart deploys `ns-label-operator` into Kubernetes cluster using the Helm package manager.

## Prerequisites

* Kubernetes cluster with RBAC (Role-Based Access Control) enabled is required
* Helm 3.4.0 or newer

## Configuration 

Create namespace:

```shell
kubectl create ns <target-namespace>
```

To deploy `ns-label-operator` you will have to first create a ConfigMap in the target namespace as Helm doesn't allow for reference of external files. There is a [PR](https://github.com/helm/helm/pull/8841) to enable `--include-file` and `--include-dir`. 

> You can load multiple files and each file can include multiple YAML blocks. See [role.yaml](../manifests/role.yaml) for example. The files must have `*.yaml` extension for the operator to read them.

```shell
kubectl create cm <config-name> \
  --from-file ../manifests/role.yaml \
  --from-file ../manifests/exporter.yaml \
  -n <target-namespace>
```

> The `-n` flag has to be same to the one used in `helm install` command.


## Install the Chart

Ensure Helm is initialized in your Kubernetes cluster.

For more details on initializing Helm, [read the Helm docs](https://helm.sh/docs/)

1. Add the helm repo

```shell
helm repo add ns-label-operator https://github.com/mchmarny/ns-label-operator/charts
helm repo update
```

2. Install the chart

```shell
helm install <name> ns-label-operator/ns-label-operator \
  -n <target-namespace> \
  --set triggerLabel=<label-name> \
  --set manifestConfigMap=<config-name>
``` 

## Verify installation

Once the chart is installed, verify the `ns-label-operator` pod is running in the target namespace:
```
kubectl get pods -n <target-namespace>
```

## Uninstall the Chart

To uninstall/delete the `ns-label-operator` release:

```
helm uninstall <name> -n <target-namespace>
```

## Configuration

The Helm chart has the follow configuration options that can be supplied:

| Parameter                                 | Description         | Default       |
|-------------------------------------------|----------------------|-------------------------|
| `debug`                         | Set logging to debug        | `false`      |
| `logAsJson`                     | Output logs in JSON format  | `true`       |
| `triggerLabel`                        | The name of the label to monitor     | ``        |
| `manifestConfigMap`                  | Name of the ConfigMap holding files that will be executed when the specified label is applied to namespace  | ``                |

## Example 

To apply files in `trace-exporter-config` ConfigMap whenever namespace has the label `trace-exporter-enabled=true` applied:

```shell
# namespace
kubectl create ns trace-exporter

# configmap
kubectl create cm trace-exporter-config \
    --from-file ../manifests/role.yaml \
    --from-file ../manifests/exporter.yaml \
    -n trace-exporter

# operator
helm install trace-exporter-operator chart/ \
  -n trace-exporter \
  --set triggerLabel=trace-exporter-enabled \
  --set manifestConfigMap=trace-exporter-config
```

## Disclaimer

This is my personal project and it does not represent my employer. While I do my best to ensure that everything works, I take no responsibility for issues caused by this code.

## License

This software is released under the [MIT](../LICENSE)

