# ns-label-operator

Watches kubernetes namespaces and fires a trigger to apply pre-configured yaml to that namespace when specific label on that namespace is set to true. Helpful to configure new namespace with specific roles, trace forwarders, or other common settings.

## config 

Create the `ns-watcher` namespace and `trigger-config` config map with the content you want to execute when namespace is labeled. That file can include multiple YAML blocks. See [test.yaml](manifests/test.yaml) for example.

> Note, you can load as many files as you like but the files must have `*.yaml` extension for the operator to read them

```shell
kubectl create ns ns-watcher
kubectl create cm trigger-config --from-file manifests/test.yaml -n ns-watcher
```

## deployment 

Apply the `ns-label-operator` deployment that will create:

* `ns-watcher` namespace 
* `ns-watcher-account` service account 
* `ns-reader-role` cluster role able to `get`, `list`, `watch` namespaces
* `ns-reader-cluster-binding` cluster role binding `ns-watcher-account` to `ns-reader-role` 
* `ns-label-operator` deployment 

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
ns-label-operator-7885dc789b-7v6rj  1/1     Running   0          10m
```

Now, follow `ns-label-operator` logs: 

```shell
kubectl logs -f -l app=ns-label-operator -n ns-watcher
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
    "msg":"triggering (dapr-enabled) on ns: test1 (labels: map[dapr-enabled:true])",
    "time":"2020-11-25T15:44:28Z"
}
```

You can also check on the namespace impact, in this case role and role binding:

```shell
kubectl get Roles,RoleBindings -n test1
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
kubectl delete cm trigger-config -n ns-watcher
```

