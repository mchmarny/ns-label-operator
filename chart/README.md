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

Create a `ConfigMap` in the above created namespace to hold all the deployments you want to apply whenever specific label is applied to any namespace in your cluster. You can load multiple files and each file can include multiple YAML blocks (good way to ensure order). See [role.yaml](../manifests/role.yaml) for example. The files must have `*.yaml` extension for the operator to read them.

> Note, this step is required only because Helm doesn't allow for reference of external files yes. There is a [PR](https://github.com/helm/helm/pull/8841) to enable `--include-file` and `--include-dir` which when lands, will no longer require this step. 


```shell
kubectl create cm <config-name> \
  --from-file deployments/role.yaml \
  --from-file deployments/exporter.yaml \
  -n <target-namespace>
```

## Install the Chart

First, add the Helm repo:

```shell
helm repo add ns-label-operator https://charts.chmarny.dev
helm repo update
```

Then, install the chart:

```shell
helm install <name> ns-label-operator/ns-label-operator \
  --set triggerLabel=<label-name> \
  --set manifestConfigMap=<config-name> \
  -n <target-namespace> \
``` 

> Set the `manifestConfigMap` to the name of ConfigMap created above. `triggerLabel` is the name of the label which should trigger the deployment. 

## Verify installation

Once the chart is installed, verify the `ns-label-operator` pod is running in the target namespace:

```shell
kubectl get pods -n <target-namespace>
```

## Uninstall the Chart

To uninstall/delete the `ns-label-operator` release:

```shell
helm uninstall <name> -n <target-namespace>
```

## Configuration

The `ns-label-operator` Helm chart has the follow configuration options:

| Parameter              | Description                               | Default     |
|------------------------|-------------------------------------------|-------------|
| `debug`                | Sets logging to debug mode (verbose)      | `false`     |
| `logAsJson`            | Outputs logs in JSON format               | `true`      |
| `triggerLabel`         | The name of the label to monitor          | ``          |
| `manifestConfigMap`    | Name of the ConfigMap holding deployments | ``          |

## Example 

To apply trace exporter configuration stored in ConfigMap named `trace-exporter-config` to any namespace labeled with `trace-export-enabled=true`, start by first creating a ConfigMap to hold the deployments (YAML files) that create the necessary configuration:

> This example uses `default` namespace but you can you use any other existing namespace of your choosing. 

```shell
kubectl create cm trace-exporter-config \
    --from-file ../manifests/role.yaml \
    --from-file ../manifests/exporter.yaml \
    -n default
```

Then deploy the Helm chart to start monitoring the cluster for specific label:

```shell
helm install trace-exporter-operator ns-label-operator/ns-label-operator \
  --set triggerLabel=trace-export-enabled \
  --set manifestConfigMap=trace-exporter-config \
  -n default
```

You can check on the deployment logs: 

```shell
kubectl logs -l app=trace-exporter-operator -n default
```

Now whenever you label a namespace with the `trace-export-enabled` label: 

```shell
kubectl label ns example-namespace trace-export-enabled=true
```

All the files loaded into the `trace-exporter-config` will be applied in that namespace.

## Disclaimer

This is my personal project and it does not represent my employer. While I do my best to ensure that everything works, I take no responsibility for issues caused by this code.

## License

This software is released under the [MIT](../LICENSE)

