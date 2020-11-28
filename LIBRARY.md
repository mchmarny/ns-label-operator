# Use ns-label-operator as a library  

To use `ns-label-operator` as a library, first import it:

```go
import "github.com/mchmarny/ns-label-operator/pkg/watch"
```

Than create an instance of `NsWatch`. The `Config` object used to initialize `NewNsWatch` has more config options. The two main are the label name and either an array of YAML string or a path to the directory with YAML files. To use a YAML string you need to populate the `Manifests` array:

```go
nsw, err := watch.NewNsWatch(watch.Config{
    Label:     "demo-role", // label to watch for
    Manifests: []string{`apiVersion: rbac.authorization.k8s.io/v1
			kind: Role
			metadata:
			  name: secret-reader
			rules:
			- apiGroups: [""]
			  resources: ["secrets"]
			  verbs: ["get"]`,
	}, // YAML to apply in the namespace where the the specified label is applied
})
```

You can also create an instance of `NsWatch` using dir path. All YAML files (`*.yaml`) in that directory will automatically parsed in.

```go
nsw, err := watch.NewNsWatch(watch.Config{
    Label:       "demo-role", // label to watch for
    ManifestDir: "./manifests", // path to the dir with YAML files
})
```

And then run it to start watching and apply specified YAML when specific label is applied to namespace. The `Start()` method will block until either an internal error occurs or the `Stop()` method is called on the same instance of `NsWatch`.

```go
if err := nsw.Start(); err != nil {
    log.Fatalf("error running watch: %v", err)
}
```

See the [Usage section](./README.md#usage) for instructions on how to use `ns-label-operator` to apply custom deployments when any namespace in your cluster is labeled with specific label. 


## Disclaimer

This is my personal project and it does not represent my employer. While I do my best to ensure that everything works, I take no responsibility for issues caused by this code.

## License

This software is released under the [MIT](./LICENSE)
