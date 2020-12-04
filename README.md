# ns-label-operator

![Release](https://github.com/mchmarny/ns-label-operator/workflows/Release/badge.svg) ![Head](https://github.com/mchmarny/ns-label-operator/workflows/Test/badge.svg) [![Go Report Card](https://goreportcard.com/badge/github.com/mchmarny/ns-label-operator)](https://goreportcard.com/report/github.com/mchmarny/ns-label-operator) ![GitHub go.mod Go version](https://img.shields.io/github/go-mod/go-version/mchmarny/ns-label-operator) [![codecov](https://codecov.io/gh/mchmarny/ns-label-operator/branch/main/graph/badge.svg?token=COOCQF289Q)](https://codecov.io/gh/mchmarny/ns-label-operator)

Watches Kubernetes namespaces and applies pre-configured YAML when specific label is applied to a namespace. Helpful in configuring common roles, trace forwarders, or any other common settings on a new namespace (e.g. Dapr.io role, role binding, and trace exporter).

## Installation Options

* [Install using Helm Chart](./chart)
* [Install manually](./MANUAL.md)
* [use as a library in your project](./LIBRARY.md)

## Usage

To illustrate the usage, let's assume you want to apply custom trace export configuration to any namespace in your cluster that's labeled with the `dapr-demo=true` label. 

To start, create a ConfigMap to hold all the deployments (YAML files) that you need to create the necessary configuration:

> This example uses `default` namespace but you can deploy the `ns-label-operator` to any existing namespace in your cluster.

```shell
kubectl create secret generic demo-ns-config \
    --from-file manifests/role.yaml \
    --from-file manifests/exporter.yaml \
    -n default
```

Then deploy the `ns-label-operator` into your cluster to start monitoring for specific label:

> This example uses Helm chart, See [Installation Options](#installation-options) for other ways to use `ns-label-operator`

```shell
helm install dapr-demo-operator ns-label-operator/ns-label-operator \
  --set triggerLabel=dapr-demo \
  --set manifestConfigMap=demo-ns-config \
  -n default
```

> Make sure that the ConfigMap and `ns-label-operator` are deployed into the same namespace.

Now whenever someone labels namespace in your cluster with with the `dapr-demo` label: 

```shell
kubectl label ns example-namespace dapr-demo=true
```

All the files loaded into the `demo-ns-config` ConfigMap will be applied in that namespace.


> Note, you can remove trigger label to prevent the trigger from firming again on that namespace but that will not undo the already created resources.

```shell
kubectl label ns test1 dapr-enabled-
```

## Cleanup 

```shell
kubectl delete secret demo-ns-config
helm uninstall dapr-demo-operator
```

## Disclaimer

This is my personal project and it does not represent my employer. While I do my best to ensure that everything works, I take no responsibility for issues caused by this code.

## License

This software is released under the [MIT](./LICENSE)
