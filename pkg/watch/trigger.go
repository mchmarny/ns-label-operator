package watch

import (
	"context"
	"encoding/json"

	"github.com/pkg/errors"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/serializer/yaml"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/discovery"
	"k8s.io/client-go/discovery/cached/memory"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/restmapper"
)

func (w *NsWatch) apply(ns *corev1.Namespace) error {
	if ns == nil {
		return errors.New("nil namespace")
	}

	w.logger.Debugf("running %d files on %s namespace", len(w.manifests), ns.Name)
	ctx := context.Background()
	for i, y := range w.manifests {
		if err := w.applyManifest(ctx, ns, y); err != nil {
			w.logger.Errorf("error applying yaml (%s): %v", y, err)
			continue
		}
		w.logger.Debugf("file %d applied successfully", i)
	}
	return nil
}

func (w *NsWatch) applyManifest(ctx context.Context, ns *corev1.Namespace, manifestYAML string) error {
	if ns == nil {
		return errors.New("nil namespace")
	}
	if manifestYAML == "" {
		return errors.New("empty deployment YAML variable")
	}

	decUnstructured := yaml.NewDecodingSerializer(unstructured.UnstructuredJSONScheme)
	dc, err := discovery.NewDiscoveryClientForConfig(w.config)
	if err != nil {
		return errors.Wrapf(err, "error creating discovery client using: %v", w.config)
	}

	mapper := restmapper.NewDeferredDiscoveryRESTMapper(memory.NewMemCacheClient(dc))
	dyn, err := dynamic.NewForConfig(w.config)
	if err != nil {
		return errors.Wrapf(err, "error creating dynamic client using: %v", w.config)
	}

	obj := &unstructured.Unstructured{}
	_, gvk, err := decUnstructured.Decode([]byte(manifestYAML), nil, obj)
	if err != nil {
		return errors.Wrapf(err, "error decoding YAML: %v", manifestYAML)
	}
	obj.SetNamespace(ns.GetName())
	w.logger.Infof("name:%s, ns:%s kind:%s, version:%s",
		obj.GetName(), obj.GetNamespace(), gvk.GroupKind(), gvk.Version)

	mapping, err := mapper.RESTMapping(gvk.GroupKind(), gvk.Version)
	if err != nil {
		return errors.Wrapf(err, "error creating REST mapping: %v", gvk.GroupKind())
	}

	w.logger.Debugf("resource: %v, scope: %v", mapping.Resource, mapping.Scope)
	dr := dyn.Resource(mapping.Resource).Namespace(ns.Name)

	data, err := json.Marshal(obj)
	if err != nil {
		return errors.Wrapf(err, "error marshaling object: %v", obj)
	}

	w.logger.Debugf("patching %s in %s... ", obj.GetName(), obj.GetNamespace())
	_, err = dr.Patch(ctx, obj.GetName(), types.ApplyPatchType, data, metav1.PatchOptions{
		FieldManager: w.fileManager,
	})

	if err != nil {
		return errors.Wrapf(err, "error applying %s to %s", string(data), obj.GetName())
	}

	w.logger.Debugf("object %s applied in %s... ", obj.GetName(), obj.GetNamespace())
	return nil
}
