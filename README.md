# ns-label-trigger

Watches kubernetes namespaces and fires a trigger when specific label is set to true.

## config 

Create a config map with the content you want to execute when namespace is labeled:

```shell
kubectl create cm trigger-config --from-file test.yaml -n ns-watcher
```

## deployment 

Apply the deployment that will create:

* `ns-watcher` namespace 
* `ns-watcher-account` service account 
* `ns-reader-role` cluster role able to `get`, `list`, `watch` namespaces
* `ns-reader-cluster-binding` cluster role binding `ns-watcher-account` to `ns-reader-role` 
* `ns-label-trigger` deployment 

```shell
kubectl apply -f deployment.yaml
```

Check on the deployment status:

```shell
kubectl get pods -n ns-watcher
```

The result should look something like this: 

```shell
NAME                                READY   STATUS    RESTARTS   AGE
ns-label-trigger-7885dc789b-7v6rj   1/1     Running   0          10m
```

Now, follow `ns-label-trigger` logs: 

```shell
kubectl logs -f -l app=ns-label-trigger -n ns-watcher
```


## test

> Assumes namespace named `test1` and the label to trigger on when `true` is `dapr-enabled`

In a separate terminal now, create a namespace:

```shell
kubectl create ns test1
```

Label and now label that namespace:

```shell
kubectl label ns test1 dapr-enabled=true
```

The log you followed in the deployment should now include:

```json
{
    "level":"info",
    "msg":"triggering (dapr-enabled) on ns: test1 (labels: map[dapr-enabled:true])\n",
    "time":"2020-11-25T15:44:28Z"
}
```

To remove the label:

> note, removing trigger label just prevents the trigger from firming again on that namespace, it does not undo the already created resources

```shell
kubectl label ns test1 dapr-enabled-
```

To clean up, just delete the namespace:

```shell
kubectl delete ns test1
```


## cleanup 

To delete the entire deployment:

```shell
kubectl delete -f deployment.yaml
```

