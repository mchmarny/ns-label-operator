# Install ns-label-operator Manually

## Config

Create the `ns-watcher` namespace:

```shell
kubectl create ns ns-watcher
```

Next, create `trigger-label` config map with the name of the namespace label for which you want watch.

```shell
kubectl create cm trigger-label \
    --from-literal label=dapr-demo \
    -n ns-watcher
```

Next, create `trigger-config` config map with the content you want to execute when the specific label is applied to a namespace.

> You can load multiple files and each file can include multiple YAML blocks. See [role.yaml](manifests/role.yaml) for example. The files must have `*.yaml` extension for the operator to read them.

```shell
kubectl create secret generic demo-ns-config \
    --from-file manifests/role.yaml \
    --from-file manifests/exporter.yaml \
    -n ns-watcher
```

## Deployment 

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


## Cleanup 

To delete the entire deployment:

```shell
kubectl delete -f deployments
kubectl delete ns ns-watcher
```

## Disclaimer

This is my personal project and it does not represent my employer. While I do my best to ensure that everything works, I take no responsibility for issues caused by this code.

## License

This software is released under the [MIT](./LICENSE)
