package main

import (
	"context"
	"encoding/json"
	"io/ioutil"
	"path"
	"path/filepath"
	"strings"

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/serializer/yaml"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/discovery"
	"k8s.io/client-go/discovery/cached/memory"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/restmapper"
)

type triggerCmd struct {
	cfg         *rest.Config
	cs          *kubernetes.Clientset
	logger      *logrus.Logger
	fileManager string
	manifests   []string
}

func (r *triggerCmd) init(dir string) error {
	r.logger.Debugf("config dir: %s", dir)

	files, err := getFiles(dir, "*.yaml")
	if err != nil {
		return errors.Wrapf(err, "error reading files from: %s", dir)
	}

	r.manifests = make([]string, 0)

	for _, f := range files {
		r.logger.Debugf("parsing %s file", f)
		b, err := ioutil.ReadFile(f)
		if err != nil {
			return errors.Wrapf(err, "error reading manifest: %s", f)
		}
		for _, y := range strings.Split(string(b), "---") {
			r.manifests = append(r.manifests, y)
		}
	}

	r.logger.Debugf("found %d YAML manifest(s) from %d file(s)", len(r.manifests), len(files))

	return nil
}

func (r *triggerCmd) run(ns *corev1.Namespace) error {
	if ns == nil {
		return errors.New("nil namespace")
	}

	r.logger.Debugf("running %d files on %s namespace", len(r.manifests), ns.Name)

	ctx := context.Background()
	for i, y := range r.manifests {
		if err := r.serverApply(ctx, ns, y); err != nil {
			r.logger.Errorf("error applying yaml (%s): %v", y, err)
			continue
		}
		r.logger.Debugf("file %d applied successfully", i)
	}
	return nil
}

func (r *triggerCmd) serverApply(ctx context.Context, ns *corev1.Namespace, deploymentYAML string) error {
	if ns == nil {
		return errors.New("nil namespace")
	}
	if deploymentYAML == "" {
		return errors.New("empty deployment YAML vairable")
	}

	decUnstructured := yaml.NewDecodingSerializer(unstructured.UnstructuredJSONScheme)
	dc, err := discovery.NewDiscoveryClientForConfig(r.cfg)
	if err != nil {
		return errors.Wrapf(err, "error creating discovery client using: %v", r.cfg)
	}

	mapper := restmapper.NewDeferredDiscoveryRESTMapper(memory.NewMemCacheClient(dc))
	dyn, err := dynamic.NewForConfig(r.cfg)
	if err != nil {
		return errors.Wrapf(err, "error creating dynamic client using: %v", r.cfg)
	}

	obj := &unstructured.Unstructured{}
	_, gvk, err := decUnstructured.Decode([]byte(deploymentYAML), nil, obj)
	if err != nil {
		return errors.Wrapf(err, "error decoding YAML: %v", deploymentYAML)
	}
	obj.SetNamespace(ns.GetName())
	r.logger.Infof("name:%s, ns:%s kind:%s, version:%s",
		obj.GetName(), obj.GetNamespace(), gvk.GroupKind(), gvk.Version)

	mapping, err := mapper.RESTMapping(gvk.GroupKind(), gvk.Version)
	if err != nil {
		return errors.Wrapf(err, "error creating REST mapping: %v", gvk.GroupKind())
	}

	r.logger.Infof("resource: %v, scope: %v", mapping.Resource, mapping.Scope)
	dr := dyn.Resource(mapping.Resource).Namespace(ns.Name)

	data, err := json.Marshal(obj)
	if err != nil {
		return errors.Wrapf(err, "error marshaling object: %v", obj)
	}

	r.logger.Infof("patching %s in %s... ", obj.GetName(), obj.GetNamespace())
	_, err = dr.Patch(ctx, obj.GetName(), types.ApplyPatchType, data, metav1.PatchOptions{
		FieldManager: r.fileManager,
	})

	if err != nil {
		return errors.Wrapf(err, "error applying %s to %s", string(data), obj.GetName())
	}

	r.logger.Infof("object %s applied in %s... ", obj.GetName(), obj.GetNamespace())
	return nil
}

// k8s configmap mounts include version subdirectories
// so no walking down, just list the top dir files
func getFiles(dir, pattern string) ([]string, error) {
	var matches []string
	files, err := ioutil.ReadDir(dir)
	if err != nil {
		return nil, errors.Wrapf(err, "error reading dir: %s", dir)
	}
	for _, f := range files {
		if matched, err := filepath.Match(pattern, f.Name()); err != nil {
			return nil, errors.Wrapf(err, "error matching file: %s", f)
		} else if matched {
			matches = append(matches, path.Join(dir, f.Name()))
		}
	}
	return matches, nil
}
