# ns-label-trigger

WIP: kubernetes namespace watch with label trigger

## test

> Assumes namespace named `test1` and the label to trigger on when `true` is `dapr-enabled`

Create namespace:

```shell
kubectl create ns test1
```

Label namespace:

```shell
kubectl label namespaces test1 dapr-enabled=true
```

Results in log output: 

```shell
INFO[0057] triggering (dapr-enabled) on ns: test1 (labels: map[dapr-enabled:true])
```

> actual trigger implementation in progress 

Delete test namespace:


```shell
kubectl delete ns test1
```

