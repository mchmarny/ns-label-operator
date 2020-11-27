# ns-label-operator

![Release](https://github.com/mchmarny/ns-label-operator/workflows/Release/badge.svg) ![Head](https://github.com/mchmarny/ns-label-operator/workflows/Test/badge.svg) [![Go Report Card](https://goreportcard.com/badge/github.com/mchmarny/ns-label-operator)](https://goreportcard.com/report/github.com/mchmarny/ns-label-operator) ![GitHub go.mod Go version](https://img.shields.io/github/go-mod/go-version/mchmarny/ns-label-operator) [![codecov](https://codecov.io/gh/mchmarny/ns-label-operator/branch/main/graph/badge.svg)](https://codecov.io/gh/mchmarny/ns-label-operator)

Watches kubernetes namespaces for specific label and applies pre-configured YAML files to those namespaces that have that label set to true. Helpful in configuring common roles, trace forwarders, or other common settings on a new namespace (e.g. Dapr.io role, role binding, and trace exporter).

## config 

Create the `ns-watcher` namespace:

```shell
kubectl create ns ns-watcher
```

Next, create `trigger-label` config map with the name of the namespace label for which you want watch.

```shell
kubectl create cm trigger-label \
    --from-literal label=dapr-enabled \
    -n ns-watcher
```

Next, create `trigger-config` config map with the content you want to execute when the specific label is applied to a namespace.

> You can load multiple files and each file can include multiple YAML blocks. See [role.yaml](manifests/role.yaml) for example. The files must have `*.yaml` extension for the operator to read them.

```shell
kubectl create cm trigger-config \
    --from-file manifests/role.yaml \
    --from-file manifests/exporter.yaml \
    -n ns-watcher
```

## deployment 

Apply the deployment:

```shell
kubectl apply -f deployments
```

Check on the deployment status:

```shell
kubectl get pods -n ns-watcher
```

The result should look something like this: 

```shell
NAME                                 READY   STATUS    RESTARTS   AGE
ns-label-operator-67d47c58b6-46vx6   1/1     Running   0          12s
```

Also, check the `ns-label-operator` logs: 

```shell
kubectl logs -f -l app=ns-label-operator -n ns-watcher
```

On successfully deployment, it should include something like this: 

```json
{
    "level":"info",
    "msg":"starting ns-label-operator for label: dapr-enabled",
    "time":"2020-11-27T06:59:39-08:00"
}
```

Now, try testing it.

## test

> Assumes namespace named `demo1` and the label to trigger on when `true` is `dapr-enabled`

In a separate terminal now, create a namespace:

```shell
kubectl create ns demo1
```

Label and now label that namespace:

```shell
kubectl label ns demo1 dapr-enabled=true
```

The log you followed in the deployment should now also include the 3 entries, one for each YAML manifest loaded from the 2 files in [manifests](./manifests) directory:

```json
{
    "level":"info",
    "msg":"name:zipkin, ns:demo8 kind:Component.dapr.io, version:v1alpha1",
    "time":"2020-11-27T07:01:10-08:00"
}
{
    "level":"info",
    "msg":"name:secret-reader, ns:demo8 kind:Role.rbac.authorization.k8s.io, version:v1",
    "time":"2020-11-27T07:01:10-08:00"
}
{
    "level":"info",
    "msg":"name:dapr-secret-reader, ns:demo8 kind:RoleBinding.rbac.authorization.k8s.io, version:v1",
    "time":"2020-11-27T07:01:10-08:00"
}
{
    "level":"info",
    "msg":"trigger:dapr-enabled applied on namespace:demo8",
    "time":"2020-11-27T07:01:10-08:00"
}
```

You can also check on the changes made in the namespace:

```shell
kubectl get all,Roles,RoleBindings -n demo1
```

> Note, you can remove trigger label but that's just prevents the trigger from firming again on that namespace, it does not undo the already created resources

```shell
kubectl label ns test1 dapr-enabled-
```

## library 

To use `ns-label-operator` as a library, first import it:

```go
import "github.com/mchmarny/ns-label-operator/pkg/watch"
```

Than create an instance of `NsWatch`:

```go
nsw, err := watch.NewNsWatch(watch.Config{
    Label:       label, // label to watch for
    ConfigFile:  configPath, // kube config file path or empty for in cluster config
    ManifestDir: dirPath, // dir path, alternatively set the Manifests with YAML strings 
    Logger:      logger, // (optional), logrus logger instance, will be created if nil
})
handleErr(err)
```

And then run it to start watching and apply specified YAML when specific label is applied to namespace. The `Run()` method will block until either an internal error occurs or the `Stop()` method is called on the same `NsWatch`.

```go
if err := nsw.Run(); err != nil {
    log.Fatalf("error running watch: %v", err)
}
```

## cleanup 

To delete the entire deployment:

```shell
kubectl delete -f deployments
kubectl delete ns ns-watcher
```

## Disclaimer

This is my personal project and it does not represent my employer. While I do my best to ensure that everything works, I take no responsibility for issues caused by this code.

## License

This software is released under the [MIT](./LICENSE)
