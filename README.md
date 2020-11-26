# ns-label-operator

Watches kubernetes namespaces for specific label and applies pre-configured YAML files to those namespaces that have that label set to true. Helpful in configuring common roles, trace forwarders, or other common settings on a new namespace (e.g. Dapr.io role, role binding, and trace exporter).

> This is fully functional now although still little too manual. Helm chart in progress. 

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
kubectl apply -f deployments/deployment.yaml
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
{"level":"info","msg":"starting ns-label-operator","time":"2020-11-26T19:35:14Z"}
{"level":"info","msg":"using cluster config","time":"2020-11-26T19:35:14Z"}
{"level":"info","msg":"loading manifests from: /config","time":"2020-11-26T19:35:14Z"}
{"level":"info","msg":"parsing /config/exporter.yaml file","time":"2020-11-26T19:35:14Z"}
{"level":"info","msg":"parsing /config/role.yaml file","time":"2020-11-26T19:35:14Z"}
{"level":"info","msg":"found 3 YAML manifest(s) from 2 file(s)","time":"2020-11-26T19:35:14Z"}
```

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

The log you followed in the deployment should now also include:

```json
{"level":"info","msg":"name:zipkin, ns:demo2 kind:Component.dapr.io, version:v1alpha1","time":"2020-11-26T19:37:45Z"}
{"level":"info","msg":"name:secret-reader, ns:demo2 kind:Role.rbac.authorization.k8s.io, version:v1","time":"2020-11-26T19:37:46Z"}
{"level":"info","msg":"name:dapr-secret-reader, ns:demo2 kind:RoleBinding.rbac.authorization.k8s.io, version:v1","time":"2020-11-26T19:37:46Z"}
{"level":"info","msg":"trigger:dapr-enabled applied on namespace:demo2","time":"2020-11-26T19:37:46Z"}
```

You can also check on the changes made in the namespace:

```shell
kubectl get all,Roles,RoleBindings -n demo1
```

> Note, you can remove trigger label but that's just prevents the trigger from firming again on that namespace, it does not undo the already created resources

```shell
kubectl label ns test1 dapr-enabled-
```

## cleanup 

To delete the entire deployment:

```shell
kubectl delete -f deployments/deployment.yaml
kubectl delete ns ns-watcher
```

## Disclaimer

This is my personal project and it does not represent my employer. While I do my best to ensure that everything works, I take no responsibility for issues caused by this code.

## License

This software is released under the [MIT](./LICENSE)
